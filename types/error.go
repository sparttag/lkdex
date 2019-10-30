package types

import (
	"errors"
)

var (
	ErrNoConnectionToDaemon       = errors.New("no_connection_to_daemon")
	ErrNoConnectionToWalletDaemon = errors.New("no_connection_to_wallet_daemon")
	// ErrDaemonBusy           = errors.New("daemon_busy")
	// ErrGetHashes            = errors.New("get_hashes_error")
	// ErrGetBlocks            = errors.New("get_blocks_error")
	ErrWalletNotOpen     = errors.New("wallet not open")
	ErrTxNotFound        = errors.New("tx not found")
	ErrNoTransInTx       = errors.New("no trans in tx")
	ErrArgsInvalid       = errors.New("args invalid")
	ErrBlockNotFound     = errors.New("block not found")
	ErrBlockParentHash   = errors.New("err block parent hash")
	ErrSaveAccountSubCnt = errors.New("save AccountSubCnt fail")
	ErrDBOrderError      = errors.New("Read Order db error")
)
