package types

import (
	"fmt"

	"github.com/lianxiangcloud/linkchain/libs/common"
	"github.com/lianxiangcloud/linkchain/libs/crypto"
	"github.com/lianxiangcloud/linkchain/libs/hexutil"
)

type OrderType int

//{"txType":"", "tokenGet":"", "amountGet":"", "tokenGive":"", "amountGive":"", "expires":"", "nonce":"", "blockHash":""}
type Order struct {
	TokenGet   common.Address `json:"tokenGet"`
	AmountGet  *hexutil.Big   `json:"amountGet"`
	TokenGive  common.Address `json:"tokenGive"`
	AmountGive *hexutil.Big   `json:"amountGive"`
	Expires    hexutil.Uint64 `json:"expires"`
	Nonce      hexutil.Uint64 `json:"nonce"`
	Maker      common.Address `json:"maker"`
}

//Adaptation contract to generate hash
func (r *Order) OrderToHash() common.Hash {
	b := fmt.Sprintf(`{"amountGet":"%s","amountGive":"%s","expires":"%d","nonce":%d,"tokenGet":"%s","tokenGive":"%s","maker":"%s"}`,
		r.AmountGet.ToInt().String(),
		r.AmountGive.ToInt().String(),
		uint64(r.Expires),
		uint64(r.Nonce),
		r.TokenGet.Hex(),
		r.TokenGive.Hex(),
		r.Maker.Hex(),
	)
	return common.BytesToHash(crypto.Keccak256([]byte(b)))
}

type SignOrder struct {
	Order `json:"order"`
	V     *hexutil.Big `json:"v"`
	S     *hexutil.Big `json:"s"`
	R     *hexutil.Big `json:"r"`
}
