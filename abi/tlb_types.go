package abi

import (
	"math/big"
	"reflect"

	"github.com/pkg/errors"
	"github.com/xssnick/tonutils-go/address"
	"github.com/xssnick/tonutils-go/tlb"
	"github.com/xssnick/tonutils-go/tvm/cell"
)

type TelemintText struct {
	Len  uint8  // ## 8
	Text string // bits (len * 8)
}

func (x *TelemintText) LoadFromCell(loader *cell.Slice) error {
	l, err := loader.LoadUInt(8)
	if err != nil {
		return errors.Wrap(err, "load len uint8")
	}

	t, err := loader.LoadSlice(8 * uint(l))
	if err != nil {
		return errors.Wrap(err, "load text slice")
	}

	x.Len = uint8(l)
	x.Text = string(t)

	return nil
}

type StringSnake string

func (x *StringSnake) LoadFromCell(loader *cell.Slice) error {
	s, err := loader.LoadStringSnake()
	if err != nil {
		return err
	}
	*x = StringSnake(s)
	return nil
}

var (
	structTypeName = "struct"

	typeNameRMap = map[reflect.Type]string{
		reflect.TypeOf([]uint8{}): "bytes",
	}
	typeNameMap = map[string]reflect.Type{
		"bool":         reflect.TypeOf(false),
		"uint8":        reflect.TypeOf(uint8(0)),
		"uint16":       reflect.TypeOf(uint16(0)),
		"uint32":       reflect.TypeOf(uint32(0)),
		"uint64":       reflect.TypeOf(uint64(0)),
		"bytes":        reflect.TypeOf([]byte{}),
		"bigInt":       reflect.TypeOf(big.NewInt(0)),
		"cell":         reflect.TypeOf((*cell.Cell)(nil)),
		"dict":         reflect.TypeOf((*cell.Dictionary)(nil)),
		"magic":        reflect.TypeOf(tlb.Magic{}),
		"coins":        reflect.TypeOf(tlb.Coins{}),
		"addr":         reflect.TypeOf((*address.Address)(nil)),
		"telemintText": reflect.TypeOf((*TelemintText)(nil)),
	}
)

func init() {
	for n, t := range typeNameMap {
		typeNameRMap[t] = n
	}
}
