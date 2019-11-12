package dex

import (
	"database/sql"
	"math/big"

	"github.com/jinzhu/gorm"
	"github.com/lianxiangcloud/linkchain/libs/common"
	"github.com/lianxiangcloud/linkchain/libs/hexutil"
	"github.com/lianxiangcloud/linkchain/libs/log"
	"github.com/lianxiangcloud/lkdex/types"
	_ "github.com/mattn/go-sqlite3"
)

type SQLDBBackend struct {
	gorm.DB
	logger log.Logger
}

//sync contract event
type BlockSyncModel struct {
	gorm.Model
	BeginBlock sql.NullInt64 `gorm:"not null"`
	EndBlock   sql.NullInt64 `gorm:"not null"`
}

type AccountModel struct {
	UserID string `gorm:"primary_key;type:char(42)"`
	Token  string `gorm:"type:char(42)"`
	Amount string
}

const (
	Sending = iota
	Trading
	Finish
)

type OrderModel struct {
	HashID       string          `gorm:"primary_key;type:char(66)"`
	TokenGet     string          `gorm:"type:char(42);not null"`
	AmountGet    string          `gorm:"not null"`
	TokenGive    string          `gorm:"type:char(42);not null"`
	AmountGive   string          `gorm:"not null"`
	Nonce        sql.NullInt64   `gorm:"not null"`
	Expires      sql.NullInt64   `gorm:"not null"`
	Maker        string          `gorm:"type:char(42);not null"`
	R            string          `gorm:"type:char(34);not null"`
	S            string          `gorm:"type:char(34);not null"`
	V            string          `gorm:"type:char(4);not null"`
	State        sql.NullInt64   `gorm:"not null"` //0:Sending(not save in block)  1:Trading  2:Finish(Cancel)
	Price        sql.NullFloat64 `gorm:"type:numeric(225,20);not null"`
	FilledAmount string          `gorm:""`
}

type TradeModel struct {
	gorm.Model
	HashID     string        `gorm:"type:char(66);FOREIGNKEY"` //Order Hash
	DealAmount string        `gorm:"not null"`                 //Deal amount
	BlockNum   sql.NullInt64 `gorm:"not null"`                 //Deal BlockNum
	TxHash     string        `gorm:"type:char(66);not null"`   //Deal Tx hash
	Taker      string        `gorm:"type:char(42);not null"`
}

func (o *OrderModel) ToSignOrder() (*types.SignOrder, error) {
	amountGet, ok := new(big.Int).SetString(o.AmountGet, 0)
	if !ok {
		return nil, types.ErrDBOrderError
	}
	amountGive, ok := new(big.Int).SetString(o.AmountGive, 0)
	if !ok {
		return nil, types.ErrDBOrderError
	}
	r, ok := new(big.Int).SetString(o.R, 0)
	if !ok {
		return nil, types.ErrDBOrderError
	}
	s, ok := new(big.Int).SetString(o.S, 0)
	if !ok {
		return nil, types.ErrDBOrderError
	}
	v, ok := new(big.Int).SetString(o.V, 0)
	if !ok {
		return nil, types.ErrDBOrderError
	}

	return &types.SignOrder{
		Order: types.Order{
			TokenGet:   common.HexToAddress(o.TokenGet),
			AmountGet:  (*hexutil.Big)(amountGet),
			TokenGive:  common.HexToAddress(o.TokenGive),
			AmountGive: (*hexutil.Big)(amountGive),
			Expires:    (hexutil.Uint64)(o.Expires.Int64),
			Nonce:      (hexutil.Uint64)(o.Nonce.Int64),
			Maker:      common.HexToAddress(o.Maker),
		},
		V: (*hexutil.Big)(v),
		S: (*hexutil.Big)(s),
		R: (*hexutil.Big)(r),
	}, nil
}

//Order: CURD
func (db *SQLDBBackend) CreateOrder(order *types.SignOrder, state uint64) error {
	hash := order.OrderToHash()
	db.logger.Debug("Create Hash", "hash", hash.Hex())

	if db.NewRecord(&OrderModel{HashID: hash.Hex()}) {
		db.logger.Debug("Record is create", "hash", hash.Hex())
		return nil
	}
	price := new(big.Float).Quo(
		new(big.Float).SetInt((*big.Int)(order.AmountGet)),
		new(big.Float).SetInt((*big.Int)(order.AmountGive)))

	pricef, acy := price.Float64()
	if acy != 0 {
		db.logger.Debug("Price not Exact", "price", pricef, "Accuracy", acy, "Get", order.AmountGet.String(), "Give", order.AmountGive.String())
	}

	saveOrder := &OrderModel{
		HashID:     hash.Hex(),
		TokenGet:   order.TokenGet.Hex(),
		AmountGet:  order.AmountGet.String(),
		TokenGive:  order.TokenGive.Hex(),
		AmountGive: order.AmountGive.String(),
		Nonce:      sql.NullInt64{(int64)(order.Nonce), true},
		Expires:    sql.NullInt64{(int64)(order.Expires), true},
		Maker:      order.Maker.Hex(),
		R:          order.R.String(),
		S:          order.S.String(),
		V:          order.V.String(),
		State:      sql.NullInt64{(int64)(state), true},
		Price:      sql.NullFloat64{pricef, true},
	}
	if err := db.Create(saveOrder).Error; err != nil {
		return err
	}
	db.logger.Debug("Save Order", "order", hash.Hex())
	return nil
}

func (db *SQLDBBackend) ReadOrder(hash common.Hash) (*types.SignOrder, error) {
	var order OrderModel
	if err := db.Where(&OrderModel{HashID: hash.Hex()}).Find(&order).Error; err != nil {
		return nil, err
	}
	return order.ToSignOrder()
}

func (db *SQLDBBackend) UpdateOrderState(hash common.Hash, state uint64) error {
	stateSql := sql.NullInt64{int64(state), true}
	var order OrderModel
	if err := db.Model(&order).Where(&OrderModel{HashID: hash.Hex()}).Update("state", stateSql).Error; err != nil {
		return err
	}
	return nil
}

func (db *SQLDBBackend) UpdateFillAmount(hash common.Hash, amount string) error {
	var order OrderModel
	if err := db.Model(&order).Where(&OrderModel{HashID: hash.Hex()}).Update("filledAmount", amount).Error; err != nil {
		return err
	}
	return nil
}
func (db *SQLDBBackend) DeleteOrder(hash common.Hash) error {
	if err := db.Delete(&OrderModel{HashID: hash.Hex()}).Error; err != nil {
		return err
	}
	return nil
}

//Trade: CURD
func (db *SQLDBBackend) CreateTrade(orderHash common.Hash, DealAmount *big.Int, BlockNum uint64, txHash common.Hash, taker common.Address) error {
	last := TradeModel{}
	found := db.Where(&TradeModel{HashID: orderHash.Hex()}).Last(last).RecordNotFound()
	if found {
		if last.BlockNum.Int64 > int64(BlockNum) {
			return nil
		}
		if last.BlockNum.Int64 == int64(BlockNum) {
			amount, _ := new(big.Int).SetString(last.DealAmount, 0)
			if DealAmount.Cmp(amount) <= 0 {
				return nil
			}
		}
	}
	trade := TradeModel{
		HashID:     orderHash.Hex(),
		DealAmount: DealAmount.String(),
		BlockNum:   sql.NullInt64{(int64)(BlockNum), true},
		TxHash:     txHash.Hex(),
		Taker:      taker.Hex(),
	}
	db.Save(&trade)
	return nil
}

//TODO:Query transactions on special demand
//QueryOrderByTxPair: order by price
func (db *SQLDBBackend) QueryOrderByTxPair(tokenGet common.Address, tokenGive common.Address, index uint64, count uint64) ([]*types.SignOrder, error) {

	var orders []OrderModel
	var rets []*types.SignOrder

	if err := db.Model(&OrderModel{}).Limit(count).Offset(index).Where(&OrderModel{
		TokenGive: tokenGive.Hex(),
		TokenGet:  tokenGet.Hex(),
	}).Order("price").Find(&orders).Error; err != nil {
		if gorm.IsRecordNotFoundError(err) {
			return nil, nil
		}
		return nil, err
	}
	for _, a := range orders {
		ret, err := a.ToSignOrder()
		if err != nil {
			return nil, err
		}
		rets = append(rets, ret)
	}
	return rets, nil
}

//Account: CURD
func (db *SQLDBBackend) CreateAccountBalance(account common.Address) error {
	if !db.NewRecord(&AccountModel{UserID: account.Hex()}) {
		return nil
	}
	return db.Create(&AccountModel{UserID: account.Hex()}).Error
}
func (db *SQLDBBackend) ReadAccountBalance(account common.Address) (*AccountModel, error) {
	acc := &AccountModel{}
	err := db.Model(&acc).Where(&AccountModel{UserID: account.Hex()}).First(&acc).Error

	if err != nil {
		acc = nil
		if !gorm.IsRecordNotFoundError(err) {
			return nil, err
		}
	}
	return acc, nil
}

func (db *SQLDBBackend) UpdateAccountBalance(account common.Address, token common.Address, amount *big.Int) error {
	acc := &AccountModel{UserID: account.Hex()}
	err := db.First(acc).Error
	if err != nil {
		return err
	}
	acc.Amount = amount.String()
	acc.Token = token.String()
	err = db.Save(acc).Error
	if err != nil {
		return err
	}
	return nil
}

func (db *SQLDBBackend) DeleteAccountBalance(account common.Address) error {
	acc := &AccountModel{UserID: account.Hex()}
	return db.Delete(acc).Error
}

//Block: CURD
func (db *SQLDBBackend) CreateSync() error {
	b := &BlockSyncModel{
		BeginBlock: sql.NullInt64{(int64)(defaultInitBlockHeight), true},
		EndBlock:   sql.NullInt64{(int64)(defaultInitBlockHeight), true},
	}
	err := db.First(&b).Error
	if gorm.IsRecordNotFoundError(err) {
		return db.Create(&b).Error
	}
	return nil
}

func (db *SQLDBBackend) ReadSync() (uint64, uint64, error) {
	b := &BlockSyncModel{}
	err := db.First(&b).Error
	return (uint64)(b.BeginBlock.Int64), (uint64)(b.EndBlock.Int64), err
}

func (db *SQLDBBackend) UpdateSync(begin, end uint64) error {
	b := &BlockSyncModel{}
	err := db.First(&b).Error
	if err != nil {
		return err
	}
	b.BeginBlock.Int64 = int64(begin)
	b.EndBlock.Int64 = int64(end)
	return db.Save(&b).Error
}

func (db *SQLDBBackend) SetLogger(logger log.Logger) {
	db.logger = logger
}
