package node

import (
	"path/filepath"

	"github.com/jinzhu/gorm"
	cmn "github.com/lianxiangcloud/linkchain/libs/common"
	"github.com/lianxiangcloud/linkchain/libs/log"
	cfg "github.com/lianxiangcloud/lkdex/config"
	"github.com/lianxiangcloud/lkdex/daemon"
	"github.com/lianxiangcloud/lkdex/dex"
	"github.com/lianxiangcloud/lkdex/rpc"
	_ "github.com/mattn/go-sqlite3"
)

// DBContext specifies config information for loading a new DB.
type DBContext struct {
	ID     string
	Config *cfg.Config
}

// DBProvider takes a DBContext and returns an instantiated DB.
type SQLDBProvider func(*DBContext) (*dex.SQLDBBackend, error)

// DefaultDBProvider returns a database using the DBBackend and DBDir
// specified in the ctx.Config.
func DefaultSQLDBProvider(ctx *DBContext) (*dex.SQLDBBackend, error) {
	db, err := NewSQLDB(ctx.ID, ctx.Config.RootDir, ctx.Config.ContractAddr)
	if err != nil {
		return nil, err
	}
	return db, nil
}

func NewSQLDB(ID string, dbdir string, contractAddr string) (*dex.SQLDBBackend, error) {
	//	fmt.Println(filepath.Join(dbdir, ID+contractAddr+".db"))
	db, err := gorm.Open("sqlite3", filepath.Join(dbdir, ID+contractAddr+".db"))
	if err != nil {
		return nil, err
	}
	return &dex.SQLDBBackend{
		DB: *db,
	}, nil

}

// NodeProvider takes a config and a logger and returns a ready to go Node.
type NodeProvider func(*cfg.Config, log.Logger) (*Node, error)

// DefaultNewNode returns a blockchain node with default settings for the
// PrivValidator, and DBProvider.
// It implements NodeProvider.
func DefaultNewNode(config *cfg.Config, logger log.Logger) (*Node, error) {
	return NewNode(config, logger, DefaultSQLDBProvider)
}

func NewNode(config *cfg.Config, logger log.Logger, dbProvider SQLDBProvider) (*Node, error) {
	logger.Info("DefaultNewNode", "conf", *config)

	// init db
	dexDB, err := dbProvider(&DBContext{"dex", config})
	if err != nil {
		return nil, err
	}
	err = dexDB.DB.DB().Ping()
	if err != nil {
		return nil, err
	}

	// init daemon
	daemon.InitDaemonClient(config.Daemon)
	daemon.InitWalletDaemonClient(config.WalletDaemon)

	// init eventListen
	localDex, err := dex.NewDex(config, logger.With("module", "dex"), dexDB)
	if err != nil {
		return nil, err
	}
	logger.Info("DefaultDexCreate")

	// init rpc
	//Make rpc context and service
	rpcContext := rpc.NewContext()
	rpcContext.SetLogger(logger)
	rpcContext.SetDex(localDex)
	rpcContext.SetDexDB(dexDB)
	//rpcContext.SetAccountManager(accountManager)

	rpcSrv, err := rpc.NewService(config, rpcContext)
	if err != nil {
		return nil, err
	}

	node := &Node{
		config: config,
		rpcSrv: rpcSrv,
		dex:    localDex,
	}

	node.BaseService = *cmn.NewBaseService(logger, "Node", node)

	return node, nil
}

type Node struct {
	cmn.BaseService

	// config
	config *cfg.Config

	//accounts manager
	//accountManager *accounts.Manager

	// rpc
	rpcSrv *rpc.Service
	dex    *dex.Dex
}

// OnStart starts the Node. It implements cmn.Service.
func (n *Node) OnStart() error {
	n.Logger.Info("starting Node")
	n.rpcSrv.Start()
	//	n.localWallet.Start()
	return nil
}

// OnStop stops the Node. It implements cmn.Service.
func (n *Node) OnStop() {
	n.BaseService.OnStop()
	n.Logger.Info("Stopping Node")
	//n.localWallet.Stop()
	n.rpcSrv.Stop()
}

// RunForever waits for an interrupt signal and stops the node.
func (n *Node) RunForever() {
	// Sleep forever and then...
	cmn.TrapSignal(func() {
		n.Stop()
	})
}
