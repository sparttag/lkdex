package rpc

import (
	"math/big"

	"github.com/lianxiangcloud/linkchain/libs/common"
	"github.com/lianxiangcloud/linkchain/libs/hexutil"
	"github.com/lianxiangcloud/lkdex/dex"
	"github.com/lianxiangcloud/lkdex/types"
)

// PrivateWalletAPI provides an API to send Dex contract transaction.
// It offers methods to call Dex contract method. All methods need wallet daemon connected.
type PrivateWalletAPI struct {
	b     Backend
	dex   *dex.Dex
	dexDB *dex.SQLDBBackend
}

// NewPrivateAccountAPI create a new PrivateAccountAPI.
func NewPrivateWalletAPI(b Backend) *PrivateWalletAPI {
	return &PrivateWalletAPI{
		b:     b,
		dex:   b.GetDex(),
		dexDB: b.GetDexDB(),
	}
}

func (s *PrivateWalletAPI) PostOrder(order *types.Order) ([]common.Hash, error) {
	//TODO:off-chain check order
	sig, err := s.dex.SignDexOrder(order)
	if err != nil {
		return nil, err
	}
	signOrder := types.SignOrder{
		Order: *order,
		R:     (*hexutil.Big)(new(big.Int).SetBytes(sig[:32])),
		S:     (*hexutil.Big)(new(big.Int).SetBytes(sig[32:64])),
		V:     (*hexutil.Big)(new(big.Int).SetBytes([]byte{sig[64] + 27})),
	}

	hash, err := s.dex.DexPostOrder(&signOrder)
	if err != nil {
		return nil, err
	}
	orderHash := order.OrderToHash()
	return []common.Hash{hash, orderHash}, nil
}

func (s *PrivateWalletAPI) PostSignOrder(order *types.SignOrder) ([]common.Hash, error) {

	hash, err := s.dex.DexPostOrder(order)
	if err != nil {
		return nil, err
	}
	orderHash := order.OrderToHash()
	return []common.Hash{hash, orderHash}, nil
}

type SignOrderRet struct {
	SO   *types.SignOrder `json:"Order"`
	Hash common.Hash      `json:"hash"`
}

func (s *PrivateWalletAPI) SignOrder(order *types.Order) (*SignOrderRet, error) {

	sig, err := s.dex.SignDexOrder(order)
	if err != nil {
		return nil, err
	}
	signOrder := types.SignOrder{
		Order: *order,
		R:     (*hexutil.Big)(new(big.Int).SetBytes(sig[:32])),
		S:     (*hexutil.Big)(new(big.Int).SetBytes(sig[32:64])),
		V:     (*hexutil.Big)(new(big.Int).SetBytes([]byte{sig[64] + 27})),
	}

	return &SignOrderRet{&signOrder, signOrder.OrderToHash()}, nil
}

func (s *PrivateWalletAPI) Trade(a common.Address, order *types.SignOrder, amount *hexutil.Big) ([]common.Hash, error) {
	hash, err := s.dex.DexTrade(a, order, amount)
	if err != nil {
		return nil, err
	}
	orderHash := order.OrderToHash()
	return []common.Hash{hash, orderHash}, nil
}

func (s *PrivateWalletAPI) TakerOrderByHash(a common.Address, hash common.Hash, amount *hexutil.Big) ([]common.Hash, error) {
	order, err := s.dexDB.ReadOrder(hash)
	if err != nil {
		return nil, err
	}
	if order == nil {
		return nil, err
	}
	return s.Trade(a, order, amount)
}

func (s *PrivateWalletAPI) WithdrawToken(a common.Address, token common.Address, amount *hexutil.Big) (common.Hash, error) {
	return s.dex.DexWithDraw(a, token, amount)
}

func (s *PrivateWalletAPI) DepositToken(a common.Address, token common.Address, amount *hexutil.Big) (common.Hash, error) {
	return s.dex.DexDeposit(a, token, amount)
}
