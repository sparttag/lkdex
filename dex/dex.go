package dex

import (
	"encoding/json"
	"errors"
	"math/big"
	"sync"
	"time"

	"github.com/lianxiangcloud/linkchain/libs/common"
	"github.com/lianxiangcloud/linkchain/libs/hexutil"
	"github.com/lianxiangcloud/linkchain/libs/log"
	"github.com/lianxiangcloud/linkchain/rpc/rtypes"
	"github.com/lianxiangcloud/lkdex/config"
	"github.com/lianxiangcloud/lkdex/types"
)

const FreshInterval = 5 * time.Second

var (
	defaultInitBlockHeight = uint64(0)
)

func ArgsNull() []byte {
	return []byte("{}")
}

func Args1(a1 interface{}) ([]byte, error) {
	v := struct {
		A0 interface{} `json:"0"`
	}{
		a1,
	}
	return json.Marshal(v)
}

func Args2(a1 interface{}, a2 interface{}) ([]byte, error) {
	v := struct {
		A0 interface{} `json:"0"`
		A1 interface{} `json:"1"`
	}{
		a1,
		a2,
	}
	return json.Marshal(v)
}

func Args3(a1 interface{}, a2 interface{}, a3 interface{}) ([]byte, error) {
	v := struct {
		A0 interface{} `json:"0"`
		A1 interface{} `json:"1"`
		A2 interface{} `json:"2"`
	}{
		a1,
		a2,
		a3,
	}
	return json.Marshal(v)
}

func Ret(ret []byte) (string, error) {
	v := struct {
		Ret string `json:"ret"`
	}{}
	err := json.Unmarshal(ret, &v)
	return v.Ret, err
}

func CheckOrder(order *types.Order) error {
	zero := big.NewInt(0)
	if order == nil {
		return errors.New("order is nil")
	}

	if (*big.Int)(order.AmountGet).Cmp(zero) <= 0 {
		return errors.New("order format error: amountGet less or equal 0")
	}

	if (*big.Int)(order.AmountGive).Cmp(zero) <= 0 {
		return errors.New("order format error: amountGive less or equal 0")
	}

	if order.TokenGet == order.TokenGive {
		return errors.New("order format error: TokenGet == TokenGive")
	}
	return nil
}

// Dex dex
type Dex struct {
	Logger log.Logger
	lock   sync.Mutex
	dexDB  *SQLDBBackend
	config *config.Config
	dexSub *DexSubscription
	//currAccount *common.Address
	isFreshNonce map[common.Address]bool
	freshLock    sync.Mutex
}

func NewDex(config *config.Config, logger log.Logger, db *SQLDBBackend) (*Dex, error) {

	dexSub, err := NewDexSubscription(config.Daemon.PeerWS, config.ContractAddr, logger, db)
	if err != nil {
		return nil, err
	}
	dex := &Dex{
		config:       config,
		dexDB:        db,
		Logger:       logger,
		dexSub:       dexSub,
		isFreshNonce: make(map[common.Address]bool),
	}
	dex.Logger.Info("Dex client create")
	db.AutoMigrate(&OrderModel{}, &TradeModel{}, &AccountModel{}, &BlockSyncModel{})
	db.CreateSync()
	db.SetLogger(logger)

	height, err := GenesisBlockNumber()
	if err != nil {
		dex.Logger.Error("GenesisBlockNumber fail", "err", err)
		return nil, err
	}
	defaultInitBlockHeight = uint64(*height)

	dex.Logger.Info("NewDex", "defaultInitBlockHeight", defaultInitBlockHeight)
	if err != nil {
		return nil, err
	}
	dex.dexSub.OnStart()
	return dex, nil
}

func (dex *Dex) SignDexOrder(order *types.Order) ([]byte, error) {

	err := CheckOrder(order)
	if err != nil {
		dex.Logger.Debug("walletSignHashErr", "order", order, "err", err.Error())
		return nil, err
	}

	hash := order.OrderToHash()
	sign, err := WalletSignHash(order.Maker, hash)
	if err != nil {
		dex.Logger.Debug("walletSignHashErr", "order", order, "err", err.Error())
		return nil, err
	}

	return sign, nil
}

func (dex *Dex) DexDeposit(a common.Address, token common.Address, amount *hexutil.Big) (common.Hash, error) {
	callArgs := ArgsNull()
	callData := []byte("deposit|" + string(callArgs))
	dex.Logger.Debug("Deposit", "call", string(callData))
	addr := common.HexToAddress(dex.config.ContractAddr)
	send := rtypes.SendTxArgs{
		TokenAddress: token,
		From:         a,
		To:           &addr,
		Data:         (*hexutil.Bytes)(&callData),
		Value:        amount,
	}

	dex.Logger.Debug("SendTx", "TX", send)

	hash, err := dex.PostChainTx(&send)
	if err != nil {
		return common.EmptyHash, err
	}
	return hash, nil
}

func (dex *Dex) DexWithDraw(a common.Address, token common.Address, amount *hexutil.Big) (common.Hash, error) {
	zero := big.NewInt(0)
	if (*big.Int)(amount).Cmp(zero) <= 0 {
		return common.EmptyHash, errors.New("arg format error: amount less or equal 0")
	}

	callArgs, err := Args2(token, amount)
	if err != nil {
		return common.EmptyHash, err
	}

	callData := []byte("withdraw|" + string(callArgs))
	dex.Logger.Debug("withdraw", "call", string(callData))
	addr := common.HexToAddress(dex.config.ContractAddr)
	send := rtypes.SendTxArgs{
		From: a,
		To:   &addr,
		Data: (*hexutil.Bytes)(&callData),
	}

	dex.Logger.Debug("SendTx", "TX", send)

	hash, err := dex.PostChainTx(&send)
	if err != nil {
		return common.EmptyHash, err
	}
	return hash, nil
}

func (dex *Dex) DexPostOrder(order *types.SignOrder) (common.Hash, error) {
	err := CheckOrder(&order.Order)
	if err != nil {
		return common.EmptyHash, err

	}

	callArgs, err := Args1(order)
	if err != nil {
		return common.EmptyHash, err
	}

	callData := []byte("postOrder|" + string(callArgs))
	dex.Logger.Debug("PostOrder", "call", string(callData))

	return dex.DexPostRequest(order.Maker, callData)
}

func (dex *Dex) DexTrade(a common.Address, order *types.SignOrder, amount *hexutil.Big) (common.Hash, error) {
	zero := big.NewInt(0)
	if (*big.Int)(amount).Cmp(zero) <= 0 {
		return common.EmptyHash, errors.New("arg format error: amount less or equal 0")
	}

	err := CheckOrder(&order.Order)
	if err != nil {
		return common.EmptyHash, err

	}
	callArgs, err := Args2(order, amount)
	if err != nil {
		return common.EmptyHash, err
	}
	callData := []byte("trade|" + string(callArgs))
	dex.Logger.Debug("trade", "call", string(callData))

	return dex.DexPostRequest(a, callData)
}

func (dex *Dex) DexCancelOrder(order *types.SignOrder) (common.Hash, error) {
	err := CheckOrder(&order.Order)
	if err != nil {
		return common.EmptyHash, err

	}
	callArgs, err := Args1(order)
	if err != nil {
		return common.EmptyHash, err
	}
	callData := []byte("cancelOrder|" + string(callArgs))
	dex.Logger.Debug("cancelOrder", "call", string(callData))

	return dex.DexPostRequest(order.Maker, callData)
}

func (dex *Dex) DexAvailableVolume(order *types.Order) (*big.Int, error) {
	err := CheckOrder(order)
	if err != nil {
		return nil, err
	}

	callArgs, err := Args1(order)
	if err != nil {
		return nil, err
	}

	callData := []byte("availableVolume|" + string(callArgs))
	dex.Logger.Debug("availableVolume", "call", string(callData))

	result, err := dex.DexCallRequest(order.Maker, callData)
	vol, err := hexutil.DecodeBig(string(result))
	if err != nil {
		return nil, err
	}
	return vol, nil

}

func (dex *Dex) DexUsedVolumeByHash(hash *common.Hash) (*big.Int, error) {

	callArgs, err := Args1(hash)
	if err != nil {
		return nil, err
	}
	callData := []byte("usedVolumeByHash|" + string(callArgs))
	dex.Logger.Debug("usedVolumeByHash", "call", string(callData))

	result, err := dex.DexCallRequest(common.EmptyAddress, callData)
	if err != nil {
		return nil, err
	}

	vol, err := hexutil.DecodeBig(string(result))
	if err != nil {
		return nil, err
	}
	return vol, nil
}

func (dex *Dex) DexGetDepositAmount(user *common.Address, token *common.Address) (*big.Int, error) {
	callArgs, err := Args2(user, token)
	if err != nil {
		return nil, err
	}
	callData := []byte("getDepositAmount|" + string(callArgs))
	dex.Logger.Debug("getDepositAmount", "call", string(callData))

	result, err := dex.DexCallRequest(*user, callData)
	if err != nil {
		return nil, err
	}
	ret, err := Ret(result)
	if err != nil {
		return nil, err
	}
	amount, ok := new(big.Int).SetString(ret, 0)
	if !ok {
		return nil, err
	}
	return amount, nil
}

func (dex *Dex) DexTestTakerTrade(order *types.Order, taker common.Address, amount *big.Int) (string, error) {
	zero := big.NewInt(0)
	if (*big.Int)(amount).Cmp(zero) <= 0 {
		return "", errors.New("arg format error: amount less or equal 0")
	}

	err := CheckOrder(order)
	if err != nil {
		return "", err
	}

	callArgs, err := Args3(order, taker, amount)
	if err != nil {
		return "", err
	}

	callData := []byte("testTakerTrade|" + string(callArgs))
	dex.Logger.Debug("TestTakerTrade", "call", string(callData))

	result, err := dex.DexCallRequest(taker, callData)
	if err != nil {
		return "", err
	}
	ret, err := Ret(result)
	if err != nil {
		return "", err
	}

	return ret, nil
}

func (dex *Dex) DexCallRequest(from common.Address, txData []byte) (hexutil.Bytes, error) {
	addr := common.HexToAddress(dex.config.ContractAddr)
	send := rtypes.SendTxArgs{
		From: from,
		To:   &addr,
		Data: (*hexutil.Bytes)(&txData),
	}
	dex.Logger.Debug("CallRequest", "TX", send)
	result, err := EthCall(&send)
	if err != nil {
		return nil, err
	}
	return result, nil
}

//user call contract
func (dex *Dex) DexPostRequest(from common.Address, txData []byte) (common.Hash, error) {
	addr := common.HexToAddress(dex.config.ContractAddr)
	send := rtypes.SendTxArgs{
		From: from,
		To:   &addr,
		Data: (*hexutil.Bytes)(&txData),
	}

	dex.Logger.Debug("SendTx", "TX", send)

	hash, err := dex.PostChainTx(&send)
	if err != nil {
		return common.EmptyHash, err
	}
	return hash, nil
}

func (dex *Dex) startFreshTimer(from common.Address) {
	timer := time.NewTimer(FreshInterval)
	go func(t *time.Timer) {
		for {
			<-t.C
			for k, v := range dex.isFreshNonce {
				if v {
					dex.freshLock.Lock()
					dex.isFreshNonce[k] = false
					dex.freshLock.Unlock()
				}
			}
			t.Reset(FreshInterval)
		}
	}(timer)
}

func (dex *Dex) setFresh(from common.Address) {
	dex.freshLock.Lock()
	dex.isFreshNonce[from] = true
	dex.freshLock.Unlock()
}

func (dex *Dex) resetFresh(from common.Address) {
	dex.freshLock.Lock()
	dex.isFreshNonce[from] = false
	dex.freshLock.Unlock()
}

func (dex *Dex) PostChainTx(tx *rtypes.SendTxArgs) (common.Hash, error) {
	tx.GasPrice = (*hexutil.Big)(big.NewInt(1e11))

	// EstimateGas return gas
	gas, err := WalletEstimateGas(tx)
	if err != nil {
		dex.Logger.Debug("WalletEstimateGas err")
		return common.EmptyHash, err
	}
	tx.Gas = &gas

	if _, ok := dex.isFreshNonce[tx.From]; ok {
		dex.setFresh(tx.From)
		dex.startFreshTimer(tx.From)
	}

	var nonce uint64
	if dex.isFreshNonce[tx.From] {
		nonce, err = WalletGetTransactionCount(tx.From)
		if err != nil {
			dex.Logger.Debug("WalletGetNonce err")
			return common.EmptyHash, err
		}
		dex.dexDB.UpdateAccountNonce(tx.From, nonce)
	} else {
		nonce, err = dex.dexDB.ReadAccountNonce(tx.From)
		if err != nil {
			dex.Logger.Debug("DexDBGetNonce err")
			return common.EmptyHash, err
		}
	}

	dex.resetFresh(tx.From)

	s := hexutil.Uint64(nonce)
	tx.Nonce = &s

	result, err := WalletSignTx(tx)
	if err != nil {
		dex.Logger.Debug("WalletSignTx err")
		return common.EmptyHash, err
	}

	var txType string
	if tx.TokenAddress == common.EmptyAddress {
		txType = "tx"
	} else {
		txType = "txt"
	}

	err = dex.dexDB.UpdateAccountNonce(tx.From, nonce+1)
	if err != nil {
		dex.Logger.Debug("UpdateAccountNonce err")
		return common.EmptyHash, err
	}

	hash, err := SendRawTx(result.Raw, txType)
	if err != nil {
		dex.Logger.Debug("SendRawTransacion err")
		return common.EmptyHash, err
	}
	return hash, nil
}
