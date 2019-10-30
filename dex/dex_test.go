package dex

import (
	"encoding/json"
	"fmt"
	"testing"
)

type A struct {
	As int `json:"a"`
}

func TestOrderJson(t *testing.T) {
	order := A{1}
	s, _ := Args1(&order)
	bd, _ := json.Marshal(order)

	fmt.Println(string(s))
	fmt.Println(string(bd))
}
