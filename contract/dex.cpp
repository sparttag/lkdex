#include "tctpl.hpp"
struct Order {
	tc::Address tokenGet; 
	tc::BInt amountGet;
	tc::Address tokenGive;
	tc::BInt amountGive; 
	uint64 expires; 
	uint64 nonce; 
	tc::Address maker;
};
TC_STRUCT(Order,
		TC_FIELD_NAME(tokenGet, "tokenGet"), 
		TC_FIELD_NAME(amountGet, "amountGet"),
		TC_FIELD_NAME(tokenGive, "tokenGive"),
		TC_FIELD_NAME(amountGive, "amountGive"),
		TC_FIELD_NAME(expires, "expires"),
		TC_FIELD_NAME(nonce, "nonce"),
		TC_FIELD_NAME(maker, "maker"))

struct TradeRet{
	tc::BInt filledAmount;
	tc::BInt dealAmount;
	tc::Address taker;
	tc::Hash hash;
};
TC_STRUCT(TradeRet,
		TC_FIELD_NAME(filledAmount, "filled"),
		TC_FIELD_NAME(dealAmount, "deal"),
		TC_FIELD_NAME(taker, "taker"),
		TC_FIELD_NAME(hash, "hash"))

struct SignOrder {
	Order order;
	tc::BInt v;
	tc::BInt r;
	tc::BInt s;
};
TC_STRUCT(SignOrder,
		TC_FIELD_NAME(order, "order"),
		TC_FIELD_NAME(v, "v"),
		TC_FIELD_NAME(r, "r"),
		TC_FIELD_NAME(s, "s"))

//internal struct 
struct OrderState{
	bool isCancel{false};
	tc::BInt filledAmount{0};
};
TC_STRUCT(OrderState,
		TC_FIELD_NAME(isCancel, "isCancel"),
		TC_FIELD_NAME(filledAmount, "filledAmount"))

class Dex : public TCBaseContract {
private:
	StorMap<Key<tc::Address, tc::Address>, tc::BInt> depositAmount{"dps"};//deposit amount
	StorMap<Key<tc::Hash>, OrderState> orderState{"st"};//order state

public:
	void postOrder(const SignOrder& signOrder);

	void trade(const SignOrder& order, const tc::BInt& amount);

	void cancelOrder(const SignOrder& order);

	void withdraw(const tc::Address& token, const tc::BInt& amount);
	void deposit();

	//local interface
	tc::BInt availableVolume(const Order& order);
	tc::BInt usedVolumeByHash(const tc::Hash& hash);
	tc::BInt getDepositAmount(const tc::Address& user, const tc::Address& token);

	std::string testTakerTrade(const Order& order, const tc::Address& taker, const tc::BInt& amount);
	std::string testTrade(const Order& order, const tc::BInt& amount);

private:
	void setDepositAmount(const tc::Address& user, const tc::Address& token, const tc::BInt& amount);
	void exchange(const tc::Address& maker, const tc::Address& taker, 
			const tc::Address& tokenGet, const tc::Address& tokenGive, 
			const tc::BInt& amountGet, const tc::BInt& amountGive, const tc::BInt& amount);
};

TC_ABI(Dex, (postOrder)(trade)(cancelOrder)(withdraw)\
		(deposit)(availableVolume)(usedVolumeByHash)(getDepositAmount)\
		(testTakerTrade)(testTrade))

tc::BInt Dex::getDepositAmount(const tc::Address& user, const tc::Address& token){
	TC_Payable(false);
	return depositAmount.get(user, token);
}

void Dex::setDepositAmount(const tc::Address& user, const tc::Address& token, const tc::BInt& amount){
	return depositAmount.set(amount, user, token);
}

void Dex::exchange(const tc::Address& maker, const tc::Address& taker, 
		const tc::Address& tokenGet, const tc::Address& tokenGive, 
		const tc::BInt& amountGet, const tc::BInt& amountGive, const tc::BInt& amount){

	setDepositAmount(taker, tokenGive,  getDepositAmount(taker, tokenGive) + amountGive*amount/amountGet);
	setDepositAmount(taker, tokenGet,  getDepositAmount(taker, tokenGet) - amount);
	setDepositAmount(maker, tokenGet, getDepositAmount(maker, tokenGet) + amount);
	setDepositAmount(maker, tokenGive,  getDepositAmount(maker, tokenGive) - amountGive*amount/amountGet);
}

tc::Hash getOrderHash(const Order& order) {
	char* orderData = tc::json::Marshal(order);
	return tc::Hash{TC_Keccak256(orderData)};
}

void checkOrder(const Order& order){
	TC_RequireWithMsg(order.tokenGive != order.tokenGet, "order format error: tokenGive == tokenGet");
	TC_RequireWithMsg(order.amountGet > 0, "order format error: amountGet <= 0");
	TC_RequireWithMsg(order.amountGive > 0, "order format error: amountGive <= 0");
	TC_RequireWithMsg(order.expires > TC_Now(), "order error: order expired");
}

tc::BInt min(const tc::BInt& a, const tc::BInt& b){
	if(a < b){
		return a;
	} else {
		return b;
	}
}

tc::BInt min(const tc::BInt& a, const tc::BInt& b, const tc::BInt& c){
	tc::BInt min = (a < b) ? a : b;
	min = (min < c) ? min : c;
	return min;
}

void Dex::postOrder(const SignOrder& signOrder){
	TC_Payable(false);

	const Order& order = signOrder.order;
	checkOrder(order);

	tc::Hash hash = getOrderHash(signOrder.order);
	OrderState state = orderState.get(hash);

	TC_RequireWithMsg(state.filledAmount < order.amountGet, "order is already finished");
	tc::Address addr = tc::Sign::recover(hash, signOrder.v, signOrder.r, signOrder.s);
	TC_RequireWithMsg(addr == order.maker, "Order sign error");
	TC_RequireWithMsg(!state.isCancel, "Order is canceled");

	TC_Log1(tc::json::Marshal(signOrder), "Order");
}

/*
 * Trade
 * (1) Verify that the order signature is the signature of the Maker 
 * (2) Check the Maker-Taker's mortgage and calculate the amount of the amount that can be filled based on the mortgage amount.  
 * (3) The transaction is recorded on the corresponding account 
 * (4) Record the order turnover 
 * (5) Send a trade event 
 *
 * An error occurred during the transaction, reported a revert error 
 */
//amount is the quantity maker needs token
void Dex::trade(const SignOrder& signOrder, const tc::BInt& amount){
	TC_Payable(false);

	const Order& order = signOrder.order;
	checkOrder(order);

	tc::Hash hash = getOrderHash(signOrder.order);
	OrderState state = orderState.get(hash);
	TC_RequireWithMsg(state.filledAmount < order.amountGet, "order is already finished");
	tc::Address addr = tc::Sign::recover(hash, signOrder.v, signOrder.r, signOrder.s);
	TC_RequireWithMsg(addr == order.maker, "Order sign error");
	TC_RequireWithMsg(!state.isCancel, "Order is canceled");

	tc::BInt takerBalance = min(amount, getDepositAmount(tc::App::getInstance()->sender(), order.tokenGet));

	tc::BInt deal = min(takerBalance, order.amountGet - state.filledAmount, 
			getDepositAmount(order.maker, order.tokenGive)*order.amountGet/order.amountGive);

	TC_RequireWithMsg(deal > 0, "deal amount is less than or equal to zero");

	const tc::Address& taker = tc::App::getInstance()->sender();
	exchange(order.maker, taker, order.tokenGet, order.tokenGive, order.amountGet, order.amountGive, deal);

	state.filledAmount = deal + state.filledAmount;
	orderState.set(state, hash);
	TradeRet ret{state.filledAmount,deal,taker,hash};
	TC_Log1(tc::json::Marshal(ret), "Trade");
}

void Dex::cancelOrder(const SignOrder& signOrder){
	TC_Payable(false);

	const Order& order = signOrder.order;
	checkOrder(order);

	tc::Hash hash = getOrderHash(signOrder.order);
	tc::Address addr = tc::Sign::recover(hash, signOrder.v, signOrder.r, signOrder.s);

	OrderState state = orderState.get(hash);
	state.isCancel = true;
	orderState.set(state, hash);
	TC_Log1(hash.toString(), "Cancel");
}

//check token
void Dex::withdraw(const tc::Address& token, const tc::BInt& amount){
	TC_Payable(false);
	tc::BInt balance = getDepositAmount(tc::App::getInstance()->sender(), token);
	TC_RequireWithMsg(balance >= amount && amount > 0, "Insufficient balance");
	depositAmount.set(balance - amount, tc::App::getInstance()->sender(), token);
	TC_TransferToken(tc::App::getInstance()->sender().toString(), token.toString(), amount.toString());
	TC_Log1(amount.toString(), "Withdraw");
}

void Dex::deposit(){
	tc::BInt balance = depositAmount.get(tc::App::getInstance()->sender(), tc::App::getInstance()->tokenAddress());
	if (tc::App::getInstance()->tokenAddress() == tc::Address{}){
		depositAmount.set(balance + tc::App::getInstance()->value(), tc::App::getInstance()->sender(), tc::App::getInstance()->tokenAddress());
		TC_Log1(tc::App::getInstance()->value().toString(), "Deposit");
	} else{
		depositAmount.set(balance + tc::App::getInstance()->tokenValue(), tc::App::getInstance()->sender(), tc::App::getInstance()->tokenAddress());
		TC_Log1(tc::App::getInstance()->tokenValue().toString(), "Deposit");
	}
}

tc::BInt Dex::availableVolume(const Order& order){
	TC_Payable(false);

	if(order.amountGive <= 0 || order.amountGet <=0 || order.tokenGet == order.tokenGive){
		TC_RequireWithMsg(false, "order format error");
	}

	tc::Hash hash = getOrderHash(order);
	OrderState state = orderState.get(hash);
	if(state.isCancel || order.expires <= TC_Now()){
		return tc::BInt{"0"};
	}
	return order.amountGet - state.filledAmount;
}

tc::BInt Dex::usedVolumeByHash(const tc::Hash& hash){
	TC_Payable(false);
	OrderState state = orderState.get(hash);
	return state.filledAmount;
}

//local interface
std::string Dex::testTakerTrade(const Order& order, const tc::Address& taker, const tc::BInt& amount){
	TC_Payable(false);

	if(order.amountGive <= 0 || order.amountGet <=0 || order.tokenGet == order.tokenGive){
		return "fail: order format error";
	}

	if (order.expires <= TC_Now()){
		return "fail: order is expired";
	}

	tc::Hash hash = getOrderHash(order);
	OrderState state = orderState.get(hash);
	if(state.isCancel){
		return "fail: order is canceled";
	}

	tc::BInt takerBalance = min(amount, getDepositAmount(taker, order.tokenGet));
	tc::BInt deal = min(takerBalance*order.amountGive/order.amountGet, 
			availableVolume(order), 
			getDepositAmount(order.maker, order.tokenGive));

	if(deal <= 0){
		return "fail: deal amount is zero";
	}
	return "success: this transaction can be executed";
}

std::string Dex::testTrade(const Order& order, const tc::BInt& amount){
	TC_Payable(false);
	return testTakerTrade(order, tc::App::getInstance()->sender(), amount);
}
