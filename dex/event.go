package dex

import (
	"context"
	"encoding/json"
	"fmt"
	"math/big"

	"github.com/lianxiangcloud/linkchain/libs/common"
	"github.com/lianxiangcloud/linkchain/libs/hexutil"
	"github.com/lianxiangcloud/linkchain/libs/log"
	"github.com/lianxiangcloud/linkchain/libs/rpc"
	"github.com/lianxiangcloud/linkchain/rpc/filters"
	lktypes "github.com/lianxiangcloud/linkchain/types"
	"github.com/lianxiangcloud/lkdex/types"
)

type DexSubscription struct {
	nodeUrl      string
	client       *rpc.Client
	contractAddr common.Address
	logger       log.Logger
	db           *SQLDBBackend
	begin        uint64
	//	restart      chan bool
	// handerLog func(lktypes.Log)
}

func NewDexSubscription(peer string, contractAddr string, logger log.Logger, db *SQLDBBackend) (*DexSubscription, error) {
	url := fmt.Sprintf("%s", peer)
	client, err := rpc.Dial(url)
	if err != nil {
		return nil, fmt.Errorf("failed to dial peer(%s): %v", url, err)
	}
	return &DexSubscription{
		nodeUrl:      url,
		client:       client,
		contractAddr: common.HexToAddress(contractAddr),
		logger:       logger,
		db:           db,
		//restart:      make(chan bool, 1),
	}, nil
}

func (c *DexSubscription) SubLoop(chanLog chan lktypes.Log, cli *rpc.ClientSubscription) {
	for {
		select {
		case err := <-cli.Err():
			c.logger.Error("connection peer(%s) error: %v", "URL", c.nodeUrl, "err", err)
			c.client.Close()
		//	c.restart <- true
		case vLog := <-chanLog:
			c.logger.Debug("Subscription", "block", vLog.BlockNumber) // pointer to event log
			c.FilterrLog(&vLog)
		}
	}
}

type OrderRet struct {
	TokenGet   common.Address `json:"tokenGet"`
	AmountGet  string         `json:"amountGet"`
	TokenGive  common.Address `json:"tokenGive"`
	AmountGive string         `json:"amountGive"`
	Expires    uint64         `json:"expires"`
	Nonce      uint64         `json:"nonce"`
	Maker      common.Address `json:"maker"`
}

func (o *OrderRet) ToOrder() (*types.Order, error) {
	order := types.Order{
		TokenGet:  o.TokenGet,
		TokenGive: o.TokenGive,
		Maker:     o.Maker,
		Nonce:     hexutil.Uint64(o.Nonce),
		Expires:   hexutil.Uint64(o.Expires),
	}

	amountGet, ok := new(big.Int).SetString(o.AmountGet, 0)
	if !ok {
		return nil, fmt.Errorf("Unmarshal AmountGet Error")
	}
	order.AmountGet = (*hexutil.Big)(amountGet)
	amountGive, ok := new(big.Int).SetString(o.AmountGive, 0)
	if !ok {
		return nil, fmt.Errorf("Unmarshal AmountGive Error")
	}
	order.AmountGive = (*hexutil.Big)(amountGive)
	return &order, nil
}

type TradeRet struct {
	FilledAmount string         `json:"filled"`
	DealAmount   string         `json:"deal"`
	Taker        common.Address `json:"taker"`
	Hash         common.Hash    `json:"hash"`
}

type SignOrderRet struct {
	OrderRet `json:"order"`
	R        string `json:"R"`
	S        string `json:"S"`
	V        string `json:"V"`
}

func (o *SignOrderRet) ToSignOrder() (*types.SignOrder, error) {
	or, err := o.OrderRet.ToOrder()
	if err != nil {
		return nil, err
	}
	signOrder := types.SignOrder{
		Order: *or,
	}
	r, ok := new(big.Int).SetString(o.R, 0)
	if !ok {
		return nil, fmt.Errorf("Unmarshal R Error")
	}

	signOrder.R = (*hexutil.Big)(r)
	s, ok := new(big.Int).SetString(o.S, 0)
	if !ok {
		return nil, fmt.Errorf("Unmarshal S Error")
	}
	signOrder.S = (*hexutil.Big)(s)
	v, ok := new(big.Int).SetString(o.V, 0)
	if !ok {
		return nil, fmt.Errorf("Unmarshal V Error")
	}
	signOrder.V = (*hexutil.Big)(v)

	return &signOrder, nil
}

func (c *DexSubscription) FilterrLog(vlog *lktypes.Log) error {
	c.logger.Debug("EthSubscribe logs")
	if len(vlog.Topics) > 0 {
		ret := string(vlog.Data)
		switch vlog.Topics[0] {
		case common.BytesToHash([]byte("Order")):
			//Save Order
			c.logger.Debug("event", "Order", ret)

			var s SignOrderRet
			err := json.Unmarshal([]byte(ret), &s)
			if err != nil {
				c.logger.Error("Event Order unmarshal err", "ret", ret, "err", err)
				return err
			}
			retS, err := s.ToSignOrder()
			if err != nil {
				c.logger.Error("Event Order unmarshal err", "ret", ret, "err", err)
				return err
			}
			err = c.db.CreateOrder(retS, 1)
			if err != nil {
				c.logger.Error("Order Create err", "err", err)
				return err
			}
		case common.BytesToHash([]byte("Trade")):
			//Save Trade
			c.logger.Debug("event", "Trade", ret)
			var r TradeRet
			err := json.Unmarshal([]byte(ret), &r)
			if err != nil {
				c.logger.Error("Event Order unmarshal err", "ret", ret, "err", err)
				return err
			}

			takerAddr := r.Taker
			orderHash := r.Hash

			c.logger.Debug("Trade Amount", "taker", takerAddr.String(), "dealAmount", r.DealAmount, "fillAmount", r.FilledAmount, "hash", orderHash.Hex())
			err = c.db.UpdateFillAmount(orderHash, r.FilledAmount)
			if err != nil {
				c.logger.Error("Trade Fill Amount err", "err", err)
				return err
			}

			FilledAmount, ok := new(big.Int).SetString(r.FilledAmount, 0)
			if !ok {
				c.logger.Error("Trade Fill Amount err", "err", err)
				return err
			}
			DealAmount, ok := new(big.Int).SetString(r.DealAmount, 0)
			if !ok {
				c.logger.Error("Trade Deal Amount err", "err", err)
				return err
			}
			err = c.db.CreateTrade(orderHash, FilledAmount, DealAmount, vlog.BlockNumber, vlog.TxHash, takerAddr)
			if err != nil {
				c.logger.Error("Trade Create err", "err", err)
				return err
			}

		case common.BytesToHash([]byte("Cancel")):
			//Save CancelOrder
			c.logger.Debug("event", "Cancel", ret)
			err := c.db.UpdateOrderState(common.HexToHash(ret), Finish)
			if err != nil {
				c.logger.Error("Order update err", "err", err)
				return err
			}

		case common.BytesToHash([]byte("Withdraw")):
			//Save Withdraw
			c.logger.Debug("event", "Withdraw", ret)

		case common.BytesToHash([]byte("Deposit")):
			//Save Deposit
			c.logger.Debug("event", "Deposit", ret)
		}
	}
	return nil
}

func (c *DexSubscription) OnStart() error {
	query := filters.FilterCriteria{
		FromBlock: (*hexutil.Big)(new(big.Int).SetUint64(c.begin)),
		Addresses: []common.Address{c.contractAddr},
	}
	arg, err := toFilterArg(&query)
	if err != nil {
		return err
	}

	//TODO: init result is too big
	var result = make([]*lktypes.Log, 0)
	if err := c.client.Call(&result, "lk_getLogs", query); err != nil {
		log.Error("getLogs", "err", err.Error())
	} else {
		log.Debug("getLogs", "lenNum", len(result))
	}

	for _, log := range result {
		c.FilterrLog(log)
	}

	chanLog := make(chan lktypes.Log)
	sub, err := c.client.Subscribe(context.Background(), "lk", chanLog, "logsSubscribe", arg)
	if err != nil {
		return err
	}
	go c.SubLoop(chanLog, sub)
	return nil
}

func toFilterArg(q *filters.FilterCriteria) (interface{}, error) {
	arg := map[string]interface{}{
		"addrs":  q.Addresses,
		"topics": q.Topics,
	}
	if q.FromBlock == nil {
		arg["fromBlock"] = "0x0"
	} else {
		arg["fromBlock"] = q.FromBlock
	}
	return arg, nil
}

func toBlockNumArg(number *big.Int) string {
	if number == nil {
		return "latest"
	}
	return hexutil.EncodeBig(number)
}
