//go:generate mockgen -destination mock_backend.go -package rpc -self_package github.com/lianxiangcloud/linkchain/wallet/rpc github.com/lianxiangcloud/linkchain/wallet/rpc Backend

package rpc

import (
	"github.com/lianxiangcloud/lkdex/dex"
	"github.com/lianxiangcloud/linkchain/libs/rpc"
)

// Backend interface provides the common API services (that are provided by
// both full and light clients) with access to necessary functions.
type Backend interface {
	GetDex() *dex.Dex
	GetDexDB() *dex.SQLDBBackend
}

func GetAPIs(apiBackend Backend) []rpc.API {
	return []rpc.API{
		{
			Namespace: "dex",
			Version:   "1.0",
			Service:   NewPublicOrderPoolAPI(apiBackend),
			Public:    true,
		},
		{
			Namespace: "wlt",
			Version:   "1.0",
			Service:   NewPrivateWalletAPI(apiBackend),
			Public:    false,
		},
	}
}

type ApiBackend struct {
	s     *Service
	dex   *dex.Dex
	dexDB *dex.SQLDBBackend
}

func (b *ApiBackend) context() *Context {
	return b.s.context()
}

// func (b *ApiBackend) bloomService() *BloomService {
// 	return b.s.bloom
// }

func NewApiBackend(s *Service, dex *dex.Dex, db *dex.SQLDBBackend) *ApiBackend {
	return &ApiBackend{
		s:     s,
		dex:   dex,
		dexDB: db,
	}
}
func (b *ApiBackend) GetDex() *dex.Dex {
	return b.dex
}

func (b *ApiBackend) GetDexDB() *dex.SQLDBBackend {
	return b.dexDB
}
