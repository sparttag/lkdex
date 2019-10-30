package rpc

import (
	"math/big"

	"github.com/lianxiangcloud/lkdex/dex"
	"github.com/lianxiangcloud/lkdex/types"
	"github.com/lianxiangcloud/linkchain/libs/common"
	"github.com/lianxiangcloud/linkchain/libs/hexutil"
)

//MaxAmountDecPrecision = 8
//MaxPriceDecPrecision  = 8
const (
	MaxOrderCount         = 20
	MaxAmountDecPrecision = 8 //trade amount
	MaxPriceDecPrecision  = 8 //trade
)

type TxPair [2]common.Address

// PublicTransactionPoolAPI exposes methods for the RPC interface
type PublicOrderPoolAPI struct {
	b     Backend
	dex   *dex.Dex
	dexDB *dex.SQLDBBackend
}

// NewPublicTransactionPoolAPI creates a new RPC service with methods specific for the transaction pool.
func NewPublicOrderPoolAPI(b Backend) *PublicOrderPoolAPI {
	return &PublicOrderPoolAPI{b, b.GetDex(), b.GetDexDB()}
}

func (s *PublicOrderPoolAPI) GetOrderByHash(hash common.Hash) (*types.SignOrder, error) {
	return s.dexDB.ReadOrder(hash)
}

func (s *PublicOrderPoolAPI) GetTxPair() ([]TxPair, error) {
	return nil, nil
}

/*
Sell   order
						price   amount
						13000	1000
						12000	1
						11000	100
		-------------------------------
						10000	60
						 900	 4
						 8000	50
Buying order
*/
func (s *PublicOrderPoolAPI) GetOrderByTxPair(getToken common.Address, giveToken common.Address, count uint64) ([]*types.SignOrder, error) {
	return s.dexDB.QueryOrderByTxPair(getToken, giveToken, 0, count)
}

func (s *PublicOrderPoolAPI) GetPriceByTxPair(getToken common.Address, giveToken common.Address) (*big.Int, error) {
	return nil, nil
}

func (s *PublicOrderPoolAPI) GetDealOrderByTaker(a common.Address, token common.Address) ([]types.Order, error) {
	return nil, nil
}

func (s *PublicOrderPoolAPI) GetDepositAmount(a common.Address, token common.Address) (*hexutil.Big, error) {
	ret, err := s.dex.DexGetDepositAmount(&a, &token)
	if err != nil {
		return nil, err
	}
	return (*hexutil.Big)(ret), nil
}
