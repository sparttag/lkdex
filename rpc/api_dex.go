package rpc

import (
	"fmt"

	"github.com/lianxiangcloud/linkchain/libs/common"
	"github.com/lianxiangcloud/linkchain/libs/hexutil"
	"github.com/lianxiangcloud/lkdex/dex"
	"github.com/lianxiangcloud/lkdex/types"
)

//MaxAmountDecPrecision = 8
//MaxPriceDecPrecision  = 8
const (
	MaxOrderCount         = 20
	MaxAmountDecPrecision = 8 //trade amount
	MaxPriceDecPrecision  = 8 //trade
)

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
func (s *PublicOrderPoolAPI) GetOrderHash(order *types.Order) (common.Hash, error) {
	if order == nil {
		return common.EmptyHash, fmt.Errorf("order is nil")
	}
	return order.OrderToHash(), nil
}

func (s *PublicOrderPoolAPI) GetSignOrderHash(order *types.SignOrder) (common.Hash, error) {
	if order == nil {
		return common.EmptyHash, fmt.Errorf("order is nil")
	}
	return order.OrderToHash(), nil
}

func (s *PublicOrderPoolAPI) GetOrderByHash(hash common.Hash) (*types.SignOrder, error) {
	return s.dexDB.ReadOrder(hash)
}

func (s *PublicOrderPoolAPI) GetOrderByTxPair(getToken common.Address, giveToken common.Address, count uint64) ([]*types.SignOrder, error) {
	return s.dexDB.QueryOrderByTxPair(getToken, giveToken, 0, count)
}

func (s *PublicOrderPoolAPI) GetDepositAmount(a common.Address, token common.Address) (*hexutil.Big, error) {
	ret, err := s.dex.DexGetDepositAmount(&a, &token)
	if err != nil {
		return nil, err
	}
	return (*hexutil.Big)(ret), nil
}
