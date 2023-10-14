package abi

import (
	"math/big"
	"reflect"

	"github.com/pkg/errors"

	"github.com/xssnick/tonutils-go/address"
	"github.com/xssnick/tonutils-go/tlb"
	"github.com/xssnick/tonutils-go/tvm/cell"
)

type TLBType string

const (
	TLBAddr        TLBType = "addr"
	TLBBool        TLBType = "bool"
	TLBBigInt      TLBType = "bigInt"
	TLBString      TLBType = "string"
	TLBBytes       TLBType = "bytes"
	TLBCell        TLBType = "cell"
	TLBContentCell TLBType = "content"
	TLBStructCell  TLBType = "struct"
	TLBTag         TLBType = "tag"
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
	typeNameRMap = map[reflect.Type]TLBType{
		reflect.TypeOf([]uint8{}): TLBBytes,
	}
	typeNameMap = map[TLBType]reflect.Type{
		TLBBool:        reflect.TypeOf(false),
		"int8":         reflect.TypeOf(int8(0)),
		"int16":        reflect.TypeOf(int16(0)),
		"int32":        reflect.TypeOf(int32(0)),
		"int64":        reflect.TypeOf(int64(0)),
		"uint8":        reflect.TypeOf(uint8(0)),
		"uint16":       reflect.TypeOf(uint16(0)),
		"uint32":       reflect.TypeOf(uint32(0)),
		"uint64":       reflect.TypeOf(uint64(0)),
		TLBBytes:       reflect.TypeOf([]byte{}),
		TLBBigInt:      reflect.TypeOf(big.NewInt(0)),
		TLBCell:        reflect.TypeOf((*cell.Cell)(nil)),
		"dict":         reflect.TypeOf((*cell.Dictionary)(nil)),
		TLBTag:         reflect.TypeOf(tlb.Magic{}),
		"coins":        reflect.TypeOf(tlb.Coins{}),
		TLBAddr:        reflect.TypeOf((*address.Address)(nil)),
		TLBString:      reflect.TypeOf((*StringSnake)(nil)),
		"telemintText": reflect.TypeOf((*TelemintText)(nil)),
	}

	registeredDefinitions = map[TLBType]TLBFieldsDesc{}
)

func init() {
	for n, t := range typeNameMap {
		typeNameRMap[t] = n
	}
}
