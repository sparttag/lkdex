//go:generate mockgen -destination mock_wallet.go -package rpc -self_package github.com/lianxiangcloud/linkchain/wallet/rpc github.com/lianxiangcloud/linkchain/wallet/rpc Wallet

package rpc

import (
	"github.com/lianxiangcloud/lkdex/config"
	"github.com/lianxiangcloud/lkdex/dex"
	"github.com/lianxiangcloud/linkchain/libs/log"
)

// Context RPC context
type Context struct {
	cfg    *config.Config
	dex    *dex.Dex
	dexDB  *dex.SQLDBBackend
	logger log.Logger

	//	accManager *accounts.Manager
}

func NewContext() *Context {
	return &Context{}
}

func (c *Context) SetLogger(logger log.Logger) {
	c.logger = logger.With("module", "rpc.service")
}
func (c *Context) SetDex(dex *dex.Dex) {
	c.dex = dex
}

func (c *Context) SetDexDB(db *dex.SQLDBBackend) {
	c.dexDB = db
}
