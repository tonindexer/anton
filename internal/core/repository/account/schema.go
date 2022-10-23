package account

import (
	"encoding/json"
	"fmt"
	"math/big"
	"reflect"

	"github.com/xssnick/tonutils-go/address"
	"github.com/xssnick/tonutils-go/tlb"
	"github.com/xssnick/tonutils-go/tvm/cell"
)

type opFieldSchema struct {
	Name string
	Type string
	Tag  string
}

var (
	typeNameMap = map[reflect.Type]string{
		reflect.TypeOf(uint32(0)):               "uint32",
		reflect.TypeOf(uint64(0)):               "uint64",
		reflect.TypeOf(big.NewInt(0)):           "bigInt",
		reflect.TypeOf((*cell.Cell)(nil)):       "tlbCell",
		reflect.TypeOf(tlb.Magic{}):             "tlbMagic",
		reflect.TypeOf(tlb.Coins{}):             "tlbCoins",
		reflect.TypeOf((*address.Address)(nil)): "address",
	}
	typeNameRMap = map[string]reflect.Type{}
)

func init() {
	for t, n := range typeNameMap {
		typeNameRMap[n] = t
	}
}

func marshalStructSchema(fields []reflect.StructField) (string, error) {
	var ret []*opFieldSchema
	var ok bool

	for it := range fields {
		f := &fields[it]
		tmp := opFieldSchema{Name: f.Name, Tag: string(f.Tag)}
		tmp.Type, ok = typeNameMap[f.Type]
		if !ok {
			return "", fmt.Errorf("cannot marshal type %s", f.Type)
		}
		ret = append(ret, &tmp)
	}

	raw, err := json.Marshal(ret)
	if err != nil {
		return "", err
	}

	return string(raw), nil
}

func unmarshalStructSchema(raw string) ([]reflect.StructField, error) {
	var fields []reflect.StructField
	var tmp []*opFieldSchema
	var ok bool

	if err := json.Unmarshal([]byte(raw), &tmp); err != nil {
		return nil, err
	}

	for _, t := range tmp {
		f := reflect.StructField{Name: t.Name, Tag: reflect.StructTag(t.Tag)}
		f.Type, ok = typeNameRMap[t.Type]
		if !ok {
			return nil, fmt.Errorf("cannot unmarshal type %s", f.Type)
		}
		fields = append(fields, f)
	}

	return fields, nil
}
