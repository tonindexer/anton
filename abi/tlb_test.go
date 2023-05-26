package abi_test

import (
	"encoding/json"
	"math/big"
	"testing"

	"github.com/xssnick/tonutils-go/address"
	"github.com/xssnick/tonutils-go/tlb"
	"github.com/xssnick/tonutils-go/tvm/cell"

	"github.com/stretchr/testify/require"

	"github.com/tonindexer/anton/abi"
)

type Payload struct {
	SmallInt  uint32   `tlb:"## 32"`
	BigInt    *big.Int `tlb:"## 128"`
	RefStruct struct {
		Addr *address.Address `tlb:"addr"`
	} `tlb:"^"`
	EmbedStruct struct {
		Bits []byte `tlb:"bits 32"`
	} `tlb:"^"`
	MaybeCell  *cell.Cell `tlb:"maybe ^"`
	EitherCell *cell.Cell `tlb:"either ^ ."`
}

type Operation struct {
	_       tlb.Magic `tlb:"#00000001"`
	Payload Payload   `tlb:"^"`
}

var testPayloadShortSchema = `[{"name":"small_int","tlb_type":"## 32"},{"name":"big_int","tlb_type":"## 128"},{"name":"ref_struct","tlb_type":"^","struct_fields":[{"name":"addr","tlb_type":"addr"}]},{"name":"embed_struct","tlb_type":"^","struct_fields":[{"name":"bits","tlb_type":"bits 32"}]},{"name":"maybe_cell","tlb_type":"maybe ^"},{"name":"either_cell","tlb_type":"either ^ ."}]`
var testPayloadFullSchema = `[{"name":"small_int","tlb_type":"## 32","format":"uint32"},{"name":"big_int","tlb_type":"## 128","format":"bigInt"},{"name":"ref_struct","tlb_type":"^","format":"struct","struct_fields":[{"name":"addr","tlb_type":"addr","format":"addr"}]},{"name":"embed_struct","tlb_type":"^","format":"struct","struct_fields":[{"name":"bits","tlb_type":"bits 32","format":"bytes"}]},{"name":"maybe_cell","tlb_type":"maybe ^","format":"cell"},{"name":"either_cell","tlb_type":"either ^ .","format":"cell"}]`

func TestNewTLBDesc(t *testing.T) {
	var d1, d2 abi.TLBFieldsDesc

	// test json unmarshal
	err := json.Unmarshal([]byte(testPayloadShortSchema), &d1)
	require.Nil(t, err)

	x, err := d1.New()
	require.Nil(t, err)

	// test structure description
	d2, err = abi.NewTLBDesc(x)
	require.Nil(t, err)

	_, err = d2.New()
	require.Nil(t, err)

	j, err := json.Marshal(d2)
	require.Nil(t, err)
	require.Equal(t, testPayloadFullSchema, string(j))
}

func TestNewOperationDesc(t *testing.T) {
	d1, err := abi.NewOperationDesc(&Operation{})
	require.Nil(t, err)

	x, err := d1.New()
	require.Nil(t, err)

	d2, err := abi.NewOperationDesc(x)
	require.Nil(t, err)

	_, err = d2.New()
	require.Nil(t, err)

	d2.Name = d1.Name
	j1, err := json.Marshal(d1)
	require.Nil(t, err)
	j2, err := json.Marshal(d2)
	require.Nil(t, err)

	require.Equal(t, string(j1), string(j2))
}

func TestTLBFieldsDesc_LoadFromCell(t *testing.T) {
	var (
		p Payload
		d abi.TLBFieldsDesc
	)

	p.SmallInt = 42
	p.BigInt, _ = new(big.Int).SetString("8000000000000000000000000", 10)
	p.RefStruct.Addr = address.MustParseAddr("EQDj5AA8mQvM5wJEQsFFFof79y3ZsuX6wowktWQFhz_Anton")
	p.EmbedStruct.Bits = []byte("asdf")
	p.MaybeCell = nil
	p.EitherCell = cell.BeginCell().MustStoreStringSnake("either").EndCell()

	err := json.Unmarshal([]byte(testPayloadShortSchema), &d)
	require.Nil(t, err)

	x, err := d.New()
	require.Nil(t, err)

	c, err := tlb.ToCell(&p)
	require.Nil(t, err)

	err = tlb.LoadFromCell(x, c.BeginParse())
	require.Nil(t, err)

	j, err := json.Marshal(&x)
	require.Nil(t, err)

	exp := `{"small_int":42,"big_int":8000000000000000000000000,"ref_struct":{"addr":"EQDj5AA8mQvM5wJEQsFFFof79y3ZsuX6wowktWQFhz_Anton"},"embed_struct":{"bits":"YXNkZg=="},"maybe_cell":null,"either_cell":"te6cckEBAQEACAAADGVpdGhlcskJ1lc="}`
	require.Equal(t, exp, string(j))
}
