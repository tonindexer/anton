package abi

import (
	"encoding/json"
	"fmt"
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
	TLBSlice       TLBType = "slice"
	TLBContentCell TLBType = "content"
	TLBStructCell  TLBType = "struct"
	TLBTag         TLBType = "tag"
)

func init() {
	tlb.Register(DedustAssetNative{})
	tlb.Register(DedustAssetJetton{})
	tlb.Register(DedustAssetExtraCurrency{})
}

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

type DedustAssetNative struct {
	_ tlb.Magic `tlb:"$0000"`
}

type DedustAssetJetton struct {
	_         tlb.Magic `tlb:"$0001"`
	Workchain int8      `tlb:"## 8"`
	Address   []byte    `tlb:"bits 256"`
}

type DedustAssetExtraCurrency struct {
	_          tlb.Magic `tlb:"$0010"`
	CurrencyID int32     `tlb:"## 32"`
}

type DedustAsset struct {
	Asset any `tlb:"[DedustAssetNative,DedustAssetJetton,DedustAssetExtraCurrency]"`
}

func (x *DedustAsset) LoadFromCell(loader *cell.Slice) error {
	pfx, err := loader.LoadUInt(4)
	if err != nil {
		return err
	}

	switch pfx {
	case 0b0000:
		x.Asset = new(DedustAssetNative)
		return nil
	case 0b0001:
		x.Asset = new(DedustAssetJetton)
		err = tlb.LoadFromCell(x.Asset, loader, true)
		if err != nil {
			return fmt.Errorf("failed to parse DedustAssetJetton: %w", err)
		}
		return nil
	case 0b0010:
		x.Asset = new(DedustAssetExtraCurrency)
		err = tlb.LoadFromCell(x.Asset, loader, true)
		if err != nil {
			return fmt.Errorf("failed to parse DedustAssetExtraCurrency: %w", err)
		}
		return nil
	}

	return fmt.Errorf("unknown dedust asset type: %x", pfx)
}

func (x *DedustAsset) MarshalJSON() ([]byte, error) {
	if x == nil || x.Asset == nil {
		return json.Marshal(nil)
	}

	var ret struct {
		Type       string `json:"type"`
		Workchain  *int8  `json:"workchain,omitempty"`
		Address    []byte `json:"address,omitempty"`
		CurrencyID int32  `json:"currency_id,omitempty"`
	}
	switch v := x.Asset.(type) {
	case *DedustAssetNative:
		ret.Type = "native"
	case *DedustAssetJetton:
		ret.Type = "jetton"
		ret.Workchain = &v.Workchain
		ret.Address = v.Address
	case *DedustAssetExtraCurrency:
		ret.Type = "extra_currency"
		ret.CurrencyID = v.CurrencyID
	}

	return json.Marshal(ret)
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
		"dedustAsset":  reflect.TypeOf((*DedustAsset)(nil)),
	}

	registeredDefinitions = map[TLBType]TLBFieldsDesc{}
)

func init() {
	for n, t := range typeNameMap {
		typeNameRMap[t] = n
	}
}
