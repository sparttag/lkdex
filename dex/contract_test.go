package dex

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math/big"
	"testing"

	"github.com/lianxiangcloud/linkchain/libs/common"
	"github.com/lianxiangcloud/linkchain/libs/crypto"
	"github.com/lianxiangcloud/linkchain/libs/crypto/secp256k1"
	"github.com/lianxiangcloud/linkchain/libs/db"
	"github.com/lianxiangcloud/linkchain/libs/hexutil"
	"github.com/lianxiangcloud/linkchain/libs/log"
	"github.com/lianxiangcloud/linkchain/state"
	"github.com/lianxiangcloud/linkchain/types"
	"github.com/lianxiangcloud/linkchain/vm/wasm"
	dextype "github.com/lianxiangcloud/lkdex/types"
	"github.com/xunleichain/tc-wasm/vm"
)

var DexContract *wasm.Contract
var DexContractAddr = common.HexToAddress("0x0000000000000000000006616c696461746f7273")

var LKToken = common.HexToAddress("0x0000000000000000000000000000000000000000")
var Token1 = common.HexToAddress("0x0000000000000000000000000000000000000011")
var Token2 = common.HexToAddress("0x0000000000000000000000000000000000000022")
var Token3 = common.HexToAddress("0x0000000000000000000000000000000000000033")

var user1 = common.HexToAddress("0x1000000000000000000000000000000000000000")
var user2 = common.HexToAddress("0x2000000000000000000000000000000000000000")

var TestUser = common.HexToAddress("0xa73810e519e1075010678d706533486d8ecc8000")
var TestToken = common.HexToAddress("0x95ccc08ab44ac6d071a0c5911df64ad2394a4c54")
var caller = common.HexToAddress("0x54fb1c7d0f011dd63b08f85ed7b518ab82028100")

type MockChainContext struct {
	// GetHeader returns the hash corresponding to their hash.
}

func (m *MockChainContext) GetHeader(uint64) *types.Header {
	h := types.Header{}
	return &h
}

type MockAccountRef common.Address

// Address casts AccountRef to a Address
func (ar MockAccountRef) Address() common.Address { return (common.Address)(ar) }

func InitContract(addr common.Address, code []byte) *wasm.Contract {
	caller := MockAccountRef(common.HexToAddress("0x54fb1c7d0f011dd63b08f85ed7b518ab82028100"))
	to := MockAccountRef(addr)
	value := big.NewInt(0)
	gas := uint64(10000000000)
	contract := wasm.NewContract(caller, to, value, gas)
	contract.SetCallCode(&common.EmptyAddress, crypto.Keccak256Hash(code), code)
	contract.CreateCall = true
	return contract
}

func CallContract(st *state.StateDB, caller common.Address, contract *wasm.Contract, input string, value *types.TokenValue) error {
	fmt.Println("-----------------input------------------")
	fmt.Println(input)

	if value == nil {
		value = &types.TokenValue{common.Address{}, nil}
	}
	gas := uint64(10000000000)
	testHeader := types.Header{}
	ctx := wasm.NewWASMContext(&testHeader, &MockChainContext{}, &common.EmptyAddress, 1000)
	ctx.Origin = common.EmptyAddress
	ctx.GasPrice = big.NewInt(1999)
	ctx.Token = value.TokenAddr

	encodeinput := hex.EncodeToString([]byte(input))
	strInput, _ := hex.DecodeString(encodeinput)
	contract.Input = strInput

	innerContract := vm.NewContract(caller.Bytes(), contract.Address().Bytes(), value.Value, gas)
	innerContract.SetCallCode(contract.CodeAddr.Bytes(), contract.CodeHash.Bytes(), contract.Code)
	innerContract.Input = contract.Input
	innerContract.CreateCall = contract.CreateCall
	eng := vm.NewEngine(innerContract, contract.Gas, st, log.Test())
	eng.Ctx = wasm.NewWASM(ctx, st, nil)
	eng.SetTrace(false)
	app, err := eng.NewApp(contract.Address().String(), contract.Code, false)
	if err != nil {
		return fmt.Errorf("NewApp failed")
	}

	fnIndex := app.GetExportFunction(vm.APPEntry)
	if fnIndex < 0 {
		fmt.Printf("eng.GetExportFunction Not Exist: func=%s\n", "thunderchain_main")
		return fmt.Errorf("Function Not Exist")
	}
	app.EntryFunc = vm.APPEntry
	ret, err := eng.Run(app, contract.Input)
	if err != nil {
		fmt.Printf("eng.Run done: gas_used=%d, gas_left=%d\n", eng.GasUsed(), eng.Gas())
		fmt.Printf("eng.Run fail: index=%d, err=%s, input=%s\n", fnIndex, err, input)
		return err
	}
	vmem := app.VM.VMemory()
	pBytes, err := vmem.GetString(ret)
	if err != nil {
		fmt.Printf("vmem.GetString fail: err=%v", err)
		return err
	}
	fmt.Printf("eng.Run  done: gas_used=%d, gas_left=%d, return with(%d) %s\n", eng.GasUsed(), eng.Gas(), len(pBytes), string(pBytes))
	return nil
}

func InitState() *state.StateDB {
	sdb := db.NewMemDB()
	st, _ := state.New(common.EmptyHash, state.NewDatabase(sdb))
	st.IntermediateRoot(false)

	st.AddBalance(user1, new(big.Int).SetUint64(10000000000))
	st.AddBalance(user2, new(big.Int).SetUint64(10000000000))

	st.AddTokenBalance(user1, Token1, new(big.Int).SetUint64(10000000000))
	st.AddTokenBalance(user1, Token2, new(big.Int).SetUint64(10000000000))
	st.AddTokenBalance(user1, Token3, new(big.Int).SetUint64(10000000000))

	st.AddTokenBalance(user2, Token1, new(big.Int).SetUint64(10000000000))
	st.AddTokenBalance(user2, Token2, new(big.Int).SetUint64(10000000000))
	st.AddTokenBalance(user2, Token3, new(big.Int).SetUint64(10000000000))

	st.AddBalance(TestUser, new(big.Int).SetUint64(10000000000))
	st.AddTokenBalance(TestUser, TestToken, new(big.Int).SetUint64(10000000000))

	return st
}

func TestMain(m *testing.M) {
	code, err := ioutil.ReadFile("../contract/output.wasm")
	if err != nil {
		fmt.Println("ReadFile Fail")
		return
	}
	DexContract = InitContract(DexContractAddr, code)
	m.Run()
}

func TestDepositToken(t *testing.T) {
	st := InitState()
	deposit := "deposit|{}"
	err := CallContract(st, caller, DexContract, deposit, &types.TokenValue{Token1, big.NewInt(10000000)})
	if err != nil {
		t.Fatal("deposit Error")
	}
}

func TestDepositLKToken(t *testing.T) {
	st := InitState()
	deposit := "deposit|{}"
	err := CallContract(st, caller, DexContract, deposit, &types.TokenValue{LKToken, big.NewInt(10000000)})
	if err != nil {
		t.Fatal("deposit Error")
	}
}

func TestWithDrawLKToken(t *testing.T) {
	st := InitState()
	err := CallContract(st, caller, DexContract, "deposit|{}", &types.TokenValue{LKToken, big.NewInt(10000000)})
	if err != nil {
		t.Fatal("deposit Error")
	}
	wasm.Transfer(st, caller, DexContractAddr, LKToken, big.NewInt(10000000))

	fmt.Println("TokenBalance", st.GetTokenBalance(DexContractAddr, LKToken))
	withdraw := fmt.Sprintf(`withdraw|{"0":"%s","1":"%s"}`, LKToken.Hex(), big.NewInt(10000000).String())
	err = CallContract(st, caller, DexContract, withdraw, nil)
	if err != nil {
		t.Fatal("withdraw Error")
	}
}

func TestWithDrawToken(t *testing.T) {
	st := InitState()
	err := CallContract(st, caller, DexContract, "deposit|{}", &types.TokenValue{Token1, big.NewInt(10000000)})
	if err != nil {
		t.Fatal("deposit Error")
	}
	wasm.Transfer(st, caller, DexContractAddr, Token1, big.NewInt(10000000))
	withdraw := fmt.Sprintf(`withdraw|{"0":"%s","1":"%s"}`, Token1.Hex(), big.NewInt(10000000).String())
	err = CallContract(st, caller, DexContract, withdraw, nil)
	if err != nil {
		t.Fatal("withdraw Error")
	}
}

/*
 */

func TestTradeNormal(t *testing.T) {
	st := InitState()

	deposit := "deposit|{}"
	err := CallContract(st, caller, DexContract, deposit, &types.TokenValue{TestToken, big.NewInt(10000000)})
	if err != nil {
		t.Fatal("deposit Error")
	}

	err = CallContract(st, TestUser, DexContract, deposit, &types.TokenValue{TestToken, big.NewInt(10000000)})
	if err != nil {
		t.Fatal("deposit Error")
	}

	//zero := (*hexutil.Big)(big.NewInt(0))
	tokenGet := common.HexToAddress("0x95ccc08ab44ac6d071a0c5911df64ad2394a4c54")
	tokenGive := common.HexToAddress("0x95ccc08ab44ac6d071a0c5911df64ad2394a4c54")
	one := (*hexutil.Big)(big.NewInt(1))

	v := new(big.Int)
	s := new(big.Int)
	r := new(big.Int)

	v.SetString("0x1b", 0)
	s.SetString("0x2a6f594a320bd15c355bcae57fd9892cad04eeba2b51d09646daab5596d14657", 0)
	r.SetString("0xe3f70a209f0b38c06d26333915acc705561e1fc0a7b088e22470ff134fe61a31", 0)

	signOrder := dextype.SignOrder{
		dextype.Order{
			tokenGet,
			one,
			tokenGive,
			one,
			1,
			1,
			TestUser,
		}, (*hexutil.Big)(v), (*hexutil.Big)(s), (*hexutil.Big)(r),
	}

	fmt.Println("order Hash", signOrder.Order.OrderToHash().Hex())

	str, _ := json.Marshal(signOrder)
	trade := fmt.Sprintf(`trade|{"0":%s,"1":"%s"}`, str, big.NewInt(10000).String())
	err = CallContract(st, caller, DexContract, trade, nil)
	if err != nil {
		t.Fatal("withdraw Error")
	}

	//t.Fatal("not implemented")
}

func TestTradeNoMoney(t *testing.T) {
	// t.Fatal("not implemented")
}

func TestTradeSignError(t *testing.T) {
	// t.Fatal("not implemented")
}
func isProtectedV(V *big.Int) bool {
	if V != nil && V.BitLen() <= 8 {
		v := V.Uint64()
		return v != 27 && v != 28
	}
	// anything not 27 or 28 are considered unprotected
	return true
}

func TestSign(t *testing.T) {
	sign := "e3f70a209f0b38c06d26333915acc705561e1fc0a7b088e22470ff134fe61a312a6f594a320bd15c355bcae57fd9892cad04eeba2b51d09646daab5596d1465700"
	hash := "b7804371ebae4e9dc99e306e835f96f7056f87c0178730bbdb86e7b0a1ea6717"
	// Hex2Bytes returns the bytes represented by the hexadecimal string str.
	s, _ := hex.DecodeString(sign)
	h, _ := hex.DecodeString(hash)
	pub, err := secp256k1.RecoverPubkey(h, s)
	if err != nil {
		fmt.Println(err.Error())
	}
	ret1 := fmt.Sprintf("0x%x", crypto.Keccak256(pub[1:])[12:])
	fmt.Println(ret1)

	R := new(big.Int).SetBytes(s[:32])
	S := new(big.Int).SetBytes(s[32:64])
	V := new(big.Int).SetBytes([]byte{s[64] + 27})
	fmt.Println("R", hexutil.EncodeBig(R))
	fmt.Println("S", hexutil.EncodeBig(S))
	fmt.Println("V", hexutil.EncodeBig(V))

	signT := make([]byte, 65)
	copy(signT[:32], R.Bytes())
	copy(signT[32:64], S.Bytes())

	var realV byte
	if isProtectedV(V) {
		signParam := types.DeriveSignParam(V).Uint64()
		realV = byte(V.Uint64() - 35 - 2*signParam)
	} else {
		realV = byte(V.Uint64() - 27)
	}

	// tighter sig s values input homestead only apply to tx sigs
	if !crypto.ValidateSignatureValues(realV, R, S, false) {
		t.Fail()
	}
	signT[64] = realV
	// v needs to be at the end for libsecp256k1
	pubKey, err := secp256k1.RecoverPubkey(h, signT)
	// make sure the public key is a valid one
	if err != nil {
		t.Fail()
	}
	ret2 := fmt.Sprintf("0x%x", crypto.Keccak256(pubKey[1:])[12:])
	fmt.Println(ret2)
	if ret1 != ret2 {
		t.Fail()
	}
}
