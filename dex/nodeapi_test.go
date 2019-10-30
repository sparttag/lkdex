package dex

import (
	"fmt"
	"math/big"
	"testing"

	"github.com/lianxiangcloud/lkdex/config"
	"github.com/lianxiangcloud/lkdex/daemon"
	"github.com/lianxiangcloud/linkchain/libs/common"
	"github.com/lianxiangcloud/linkchain/libs/hexutil"
	"github.com/lianxiangcloud/linkchain/rpc/rtypes"
)

var adminAddr = common.HexToAddress("0xa73810e519e1075010678d706533486d8ecc8000")

//TODO: Mock Wallet Test
func TestBasic(t *testing.T) {
	daemon.InitDaemonClient(config.DefaultDaemonConfig())
	daemon.InitWalletDaemonClient(config.DefaultWalletDaemonConfig())
	n, err := GenesisBlockNumber()
	if err != nil {
		t.Error("GenesisBlockNumber")
		fmt.Println(err)
	}
	fmt.Println(n)

	sign, err := WalletSignHash(adminAddr, common.HexToHash("0x6c554f11cc33de44e2687e6539c27d9fad08db76803f92008d7cfcea55ad597a"))
	if err != nil {
		t.Error("WalletSignHash")
		fmt.Println(err)
	} else {
		fmt.Println(sign)
	}

	nonce, err := WalletGetTransactionCount(adminAddr)
	if err != nil {
		t.Error("WalletSignHash")
		fmt.Println(err)
	} else {
		fmt.Println(nonce)
	}

	tx := rtypes.SendTxArgs{
		From:     adminAddr,
		To:       &adminAddr,
		Gas:      (*hexutil.Uint64)(new(uint64)),
		GasPrice: (*hexutil.Big)(big.NewInt(0x174876e800)),
		Value:    (*hexutil.Big)(big.NewInt(0x1)),
		Nonce:    (*hexutil.Uint64)(new(uint64)),
	}
	*tx.Gas = 0x1
	*tx.Nonce = hexutil.Uint64(nonce)

	signTxResult, err := WalletSignTx(&tx)
	if err != nil {
		t.Error("WalletSignTx")
		fmt.Println(err)
	} else {
		fmt.Println(signTxResult)
	}
}
