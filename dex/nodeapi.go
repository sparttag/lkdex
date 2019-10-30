package dex

import (
	"encoding/json"
	"fmt"
	"math/big"

	"github.com/lianxiangcloud/lkdex/daemon"
	"github.com/lianxiangcloud/linkchain/libs/common"
	"github.com/lianxiangcloud/linkchain/libs/hexutil"
	"github.com/lianxiangcloud/linkchain/rpc/rtypes"
	lktypes "github.com/lianxiangcloud/linkchain/types"
	wtypes "github.com/lianxiangcloud/linkchain/wallet/types"
)

// GenesisBlockNumber return genesisBlock init height
func GenesisBlockNumber() (*hexutil.Uint64, error) {
	p := make([]interface{}, 0)
	body, err := daemon.CallJSONRPC("eth_genesisBlockNumber", p)
	if err != nil || body == nil || len(body) == 0 {
		return nil, wtypes.ErrNoConnectionToDaemon
	}

	var jsonRes wtypes.RPCResponse
	if err = json.Unmarshal(body, &jsonRes); err != nil {
		//return nil, fmt.Errorf("GenesisBlockNumber json.Unmarshal(body, &jsonRes) fail, err:%v, body:%s", err, string(body))
		return nil, wtypes.ErrDaemonResponseBody
	}
	if jsonRes.Error.Code != 0 {
		//return nil, fmt.Errorf("json RPC error:%v,body:[%s]", jsonRes.Error, string(body))
		return nil, wtypes.ErrDaemonResponseCode
	}

	var blockNumber hexutil.Uint64
	if err = json.Unmarshal(jsonRes.Result, &blockNumber); err != nil {
		//return nil, fmt.Errorf("json.Unmarshal jsonRes.Result fail, err:%v, body:%s", err, string(body))
		return nil, wtypes.ErrDaemonResponseData
	}

	return &blockNumber, nil
}

// GenesisBlockNumber return genesisBlock init height
func EthCall(args *rtypes.SendTxArgs) (hexutil.Bytes, error) {
	p := make([]interface{}, 2)
	p[0] = MarshalTx(args)
	p[1] = "latest"
	body, err := daemon.CallJSONRPC("eth_call", p)
	if err != nil || body == nil || len(body) == 0 {
		return nil, wtypes.ErrNoConnectionToDaemon
	}

	var jsonRes wtypes.RPCResponse
	if err = json.Unmarshal(body, &jsonRes); err != nil {
		//return nil, fmt.Errorf("GenesisBlockNumber json.Unmarshal(body, &jsonRes) fail, err:%v, body:%s", err, string(body))
		return nil, wtypes.ErrDaemonResponseBody
	}
	if jsonRes.Error.Code != 0 {
		//return nil, fmt.Errorf("json RPC error:%v,body:[%s]", jsonRes.Error, string(body))
		return nil, wtypes.ErrDaemonResponseCode
	}

	var ret hexutil.Bytes
	if err = json.Unmarshal(jsonRes.Result, &ret); err != nil {
		//return nil, fmt.Errorf("json.Unmarshal jsonRes.Result fail, err:%v, body:%s", err, string(body))
		return nil, wtypes.ErrDaemonResponseData
	}

	return ret, nil
}

func WalletSignHash(addr common.Address, hash common.Hash) (hexutil.Bytes, error) {
	p := make([]interface{}, 2)
	p[0] = addr
	p[1] = hash
	body, err := daemon.WalletCallJSONRPC("ltk_signHash", p)
	if err != nil || body == nil || len(body) == 0 {
		return nil, wtypes.ErrNoConnectionToDaemon
	}

	var jsonRes wtypes.RPCResponse
	if err = json.Unmarshal(body, &jsonRes); err != nil {
		//return nil, fmt.Errorf("GenesisBlockNumber json.Unmarshal(body, &jsonRes) fail, err:%v, body:%s", err, string(body))
		return nil, wtypes.ErrDaemonResponseBody
	}
	if jsonRes.Error.Code != 0 {
		//return nil, fmt.Errorf("json RPC error:%v,body:[%s]", jsonRes.Error, string(body))
		return nil, wtypes.ErrDaemonResponseCode
	}

	var signData hexutil.Bytes
	if err = json.Unmarshal(jsonRes.Result, &signData); err != nil {
		//return nil, fmt.Errorf("json.Unmarshal jsonRes.Result fail, err:%v, body:%s", err, string(body))
		return nil, wtypes.ErrDaemonResponseData
	}
	return signData, nil
}

func MarshalTx(args *rtypes.SendTxArgs) map[string]interface{} {
	req := make(map[string]interface{})
	req["from"] = args.From
	req["tokenAddress"] = args.TokenAddress
	if args.To != nil {
		req["to"] = *args.To
	}

	if args.Gas != nil && uint64(*args.Gas) > 0 {
		req["gas"] = args.Gas
	} else {
		req["gas"] = "0x0"
	}
	if args.GasPrice != nil && args.GasPrice.ToInt().Cmp(big.NewInt(0)) > 0 {
		req["gasPrice"] = args.GasPrice
	} else {
		req["gasPrice"] = (*hexutil.Big)(big.NewInt(1e11))
	}
	if args.Value != nil {
		req["value"] = args.Value
	}
	if args.Data != nil && len(*args.Data) > 0 {
		req["data"] = args.Data
	}
	if args.Nonce != nil {
		req["nonce"] = args.Nonce
	}
	return req
}

func WalletSignTx(args *rtypes.SendTxArgs) (*rtypes.SignTransactionResult, error) {
	body, err := daemon.WalletCallJSONRPC("ltk_signTransaction", []interface{}{MarshalTx(args)})
	if err != nil || body == nil || len(body) == 0 {
		return nil, wtypes.ErrNoConnectionToDaemon
	}

	var jsonRes wtypes.RPCResponse
	if err = json.Unmarshal(body, &jsonRes); err != nil {
		//return nil, fmt.Errorf("GenesisBlockNumber json.Unmarshal(body, &jsonRes) fail, err:%v, body:%s", err, string(body))
		return nil, wtypes.ErrDaemonResponseBody
	}
	if jsonRes.Error.Code != 0 {
		//return nil, fmt.Errorf("json RPC error:%v,body:[%s]", jsonRes.Error, string(body))
		return nil, wtypes.ErrDaemonResponseCode
	}

	result := rtypes.SignTransactionResult{Raw: nil, Tx: &lktypes.Transaction{}}
	if err = json.Unmarshal(jsonRes.Result, &result); err != nil {
		//return nil, fmt.Errorf("json.Unmarshal jsonRes.Result fail, err:%v, body:%s", err, string(body))
		return nil, wtypes.ErrDaemonResponseData
	}
	return &result, nil
}

func WalletSendRawTransaction(b hexutil.Bytes) (common.Hash, error) {
	p := make([]interface{}, 1)
	p[0] = b
	body, err := daemon.WalletCallJSONRPC("ltk_sendRawTransaction", p)
	if err != nil || body == nil || len(body) == 0 {
		return common.EmptyHash, wtypes.ErrNoConnectionToDaemon
	}

	var jsonRes wtypes.RPCResponse
	if err = json.Unmarshal(body, &jsonRes); err != nil {
		//return nil, fmt.Errorf("GenesisBlockNumber json.Unmarshal(body, &jsonRes) fail, err:%v, body:%s", err, string(body))
		return common.EmptyHash, wtypes.ErrDaemonResponseBody
	}
	if jsonRes.Error.Code != 0 {
		//return nil, fmt.Errorf("json RPC error:%v,body:[%s]", jsonRes.Error, string(body))
		return common.EmptyHash, wtypes.ErrDaemonResponseCode
	}

	var hash common.Hash
	if err = json.Unmarshal(jsonRes.Result, &hash); err != nil {
		//return nil, fmt.Errorf("json.Unmarshal jsonRes.Result fail, err:%v, body:%s", err, string(body))
		return common.EmptyHash, wtypes.ErrDaemonResponseData
	}
	return hash, nil
}

func SendRawTx(b hexutil.Bytes, txType string) (common.Hash, error) {
	p := make([]interface{}, 2)
	p[0] = b
	p[1] = txType
	body, err := daemon.CallJSONRPC("eth_sendRawTx", p)
	if err != nil || body == nil || len(body) == 0 {
		return common.EmptyHash, wtypes.ErrNoConnectionToDaemon
	}

	var jsonRes wtypes.RPCResponse
	if err = json.Unmarshal(body, &jsonRes); err != nil {
		//return nil, fmt.Errorf("GenesisBlockNumber json.Unmarshal(body, &jsonRes) fail, err:%v, body:%s", err, string(body))
		return common.EmptyHash, wtypes.ErrDaemonResponseBody
	}
	if jsonRes.Error.Code != 0 {
		//return nil, fmt.Errorf("json RPC error:%v,body:[%s]", jsonRes.Error, string(body))
		return common.EmptyHash, wtypes.ErrDaemonResponseCode
	}

	var hash common.Hash
	if err = json.Unmarshal(jsonRes.Result, &hash); err != nil {
		//return nil, fmt.Errorf("json.Unmarshal jsonRes.Result fail, err:%v, body:%s", err, string(body))
		return common.EmptyHash, wtypes.ErrDaemonResponseData
	}
	return hash, nil
}

func WalletGetTransactionCount(addr common.Address) (uint64, error) {
	p := make([]interface{}, 2)
	p[0] = addr
	p[1] = `latest`
	body, err := daemon.WalletCallJSONRPC("ltk_getTransactionCount", p)
	if err != nil || body == nil || len(body) == 0 {
		return 0, wtypes.ErrNoConnectionToDaemon
	}

	var jsonRes wtypes.RPCResponse
	if err = json.Unmarshal(body, &jsonRes); err != nil {
		//return nil, fmt.Errorf("GenesisBlockNumber json.Unmarshal(body, &jsonRes) fail, err:%v, body:%s", err, string(body))
		return 0, wtypes.ErrDaemonResponseBody
	}
	if jsonRes.Error.Code != 0 {
		//return nil, fmt.Errorf("json RPC error:%v,body:[%s]", jsonRes.Error, string(body))
		return 0, wtypes.ErrDaemonResponseCode
	}

	var nonce hexutil.Uint64
	if err = json.Unmarshal(jsonRes.Result, &nonce); err != nil {
		//return nil, fmt.Errorf("json.Unmarshal jsonRes.Result fail, err:%v, body:%s", err, string(body))
		fmt.Println(err)
		return 0, wtypes.ErrDaemonResponseData
	}
	uNonce := uint64(nonce)
	return uNonce, nil
}

func WalletEstimateGas(args *rtypes.SendTxArgs) (hexutil.Uint64, error) {
	body, err := daemon.WalletCallJSONRPC("ltk_estimateGas", []interface{}{MarshalTx(args)})
	if err != nil || body == nil || len(body) == 0 {
		return 0, wtypes.ErrNoConnectionToDaemon
	}

	var jsonRes wtypes.RPCResponse
	if err = json.Unmarshal(body, &jsonRes); err != nil {
		return 0, fmt.Errorf("estimateGas json.Unmarshal(body, &jsonRes) fail, err:%v, body:%s", err, string(body))
		//		return 0, wtypes.ErrDaemonResponseBody
	}

	if jsonRes.Error.Code != 0 {
		return 0, fmt.Errorf("json RPC error:%v,body:[%s]", jsonRes.Error, string(body))
		//	return 0, wtypes.ErrDaemonResponseCode
	}

	var gas hexutil.Uint64
	if err = json.Unmarshal(jsonRes.Result, &gas); err != nil {
		return 0, fmt.Errorf("json.Unmarshal jsonRes.Result fail, err:%v, body:%s", err, string(body))
		//	return gas, wtypes.ErrDaemonResponseData
	}
	return gas, nil
}
