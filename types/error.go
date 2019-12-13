package types

import (
	"errors"
)

var (
	ErrNoConnectionToDaemon       = errors.New("no_connection_to_daemon")
	ErrNoConnectionToWalletDaemon = errors.New("no_connection_to_wallet_daemon")
	ErrDBOrderError      = errors.New("Read Order db error")
)
