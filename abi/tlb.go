package abi

import (
	"fmt"
	"math/big"
	"reflect"
	"strconv"
	"strings"

	"github.com/pkg/errors"
	"github.com/xssnick/tonutils-go/address"

	"github.com/iancoleman/strcase"

	"github.com/xssnick/tonutils-go/tlb"
	"github.com/xssnick/tonutils-go/tvm/cell"
)

type TLBFieldDesc struct {
	Name   string        `json:"name"`
	Type   string        `json:"tlb_type"`
	MapTo  string        `json:"map_to"`
	Fields TLBFieldsDesc `json:"struct_fields,omitempty"` // MapTo = "struct"
}

type TLBFieldsDesc []*TLBFieldDesc

type OperationDesc struct {
	Name string        `json:"op_name"`
	Code string        `json:"op_code"`
	Body TLBFieldsDesc `json:"body"`
}

func operationID(t reflect.Type) (uint32, error) {
	op := t.Field(0)
	if op.Type != reflect.TypeOf(tlb.Magic{}) {
		return 0, fmt.Errorf("no magic type in first struct field")
	}

	opValueStr, ok := op.Tag.Lookup("tlb")
	if !ok || len(opValueStr) != 9 || opValueStr[0] != '#' {
		return 0, fmt.Errorf("wrong tlb tag format")
	}

	opValue, err := strconv.ParseUint(opValueStr[1:], 16, 32)
	if err != nil {
		return 0, errors.Wrap(err, "parse hex uint32")
	}

	return uint32(opValue), nil
}

func tlbMakeDesc(t reflect.Type) (ret TLBFieldsDesc, err error) {
	for i := 0; i < t.NumField(); i++ {
		f := t.Field(i)

		schema := &TLBFieldDesc{
			Name: strcase.ToSnake(f.Name),
			Type: f.Tag.Get("tlb"),
		}

		if schema.Name == "_" && f.Type == reflect.TypeOf(tlb.Magic{}) {
			continue // skip tlb constructor tag as it has to be inside OperationDesc
		}

		ft, ok := typeNameRMap[f.Type]
		switch {
		case ok:
			schema.MapTo = ft

		case f.Type.Kind() == reflect.Pointer && f.Type.Elem().Kind() == reflect.Struct:
			schema.MapTo = structTypeName
			schema.Fields, err = tlbMakeDesc(f.Type.Elem())
			if err != nil {
				return nil, fmt.Errorf("%s: %w", f.Name, err)
			}

		case f.Type.Kind() == reflect.Struct:
			schema.MapTo = structTypeName
			schema.Fields, err = tlbMakeDesc(f.Type)
			if err != nil {
				return nil, fmt.Errorf("%s: %w", f.Name, err)
			}

		default:
			return nil, fmt.Errorf("%s: unknown structField type %s", f.Name, f.Type)
		}

		ret = append(ret, schema)
	}

	return ret, nil
}

func NewTLBDesc(x any) (TLBFieldsDesc, error) {
	rv := reflect.ValueOf(x)
	if rv.Kind() != reflect.Pointer {
		return nil, fmt.Errorf("x should be a pointer")
	}
	return tlbMakeDesc(rv.Type().Elem())
}

func opMakeDesc(t reflect.Type) (*OperationDesc, error) {
	var ret OperationDesc

	ret.Name = t.Name()

	opCode, err := operationID(t)
	if err != nil {
		return nil, errors.Wrap(err, "lookup operation id")
	}
	ret.Code = fmt.Sprintf("0x%x", opCode)

	ret.Body, err = tlbMakeDesc(t)
	if err != nil {
		return nil, errors.Wrap(err, "make tlb schema")
	}

	return &ret, nil
}

func NewOperationDesc(x any) (*OperationDesc, error) {
	rv := reflect.ValueOf(x)
	if rv.Kind() != reflect.Pointer {
		return nil, fmt.Errorf("x should be a pointer")
	}
	return opMakeDesc(rv.Type().Elem())
}

// tlbParseSettings automatically determines go type to map field into (copy from tlb.LoadFromCell)
// ## N - means integer with N bits, if size <= 64 it loads to uint of any size, if > 64 it loads to *big.Int
// ^ - loads ref and calls recursively, if field type is *cell.Cell, it loads without parsing
// . - calls recursively to continue load from current loader (inner struct)
// [^]dict N [-> array [^]] - loads dictionary with key size N, transformation '->' can be applied to convert dict to array, example: 'dict 256 -> array ^' will give you array of deserialized refs (^) of values
// bits N - loads bit slice N len to []byte
// bool - loads 1 bit boolean
// addr - loads ton address
// maybe - reads 1 bit, and loads rest if its 1, can be used in combination with others only
// either X Y - reads 1 bit, if its 0 - loads X, if 1 - loads Y
// Some tags can be combined, for example "dict 256", "maybe ^"
func tlbParseSettings(tag string) (reflect.Type, error) {
	tag = strings.TrimSpace(tag)
	if tag == "-" {
		return nil, nil
	}
	settings := strings.Split(tag, " ")

	if len(settings) == 0 {
		return nil, nil
	}

	if settings[0] == "maybe" {
		return reflect.TypeOf((*cell.Cell)(nil)), nil
	}

	if settings[0] == "either" {
		if len(settings) < 3 {
			return nil, errors.New("either tag should have 2 args")
		}
		return reflect.TypeOf((*cell.Cell)(nil)), nil
	}

	// bits
	if settings[0] == "##" {
		num, err := strconv.ParseUint(settings[1], 10, 64)
		if err != nil {
			return nil, errors.New("corrupted num bits in ## tag")
		}
		switch {
		case num <= 8:
			return reflect.TypeOf(uint8(0)), nil
		case num <= 16:
			return reflect.TypeOf(uint16(0)), nil
		case num <= 32:
			return reflect.TypeOf(uint32(0)), nil
		case num <= 64:
			return reflect.TypeOf(uint64(0)), nil
		case num <= 256:
			return reflect.TypeOf((*big.Int)(nil)), nil
		}
	}

	if settings[0] == "addr" {
		return reflect.TypeOf((*address.Address)(nil)), nil
	}
	if settings[0] == "bool" {
		return reflect.TypeOf(false), nil
	}
	if settings[0] == "bits" {
		return reflect.TypeOf([]byte(nil)), nil
	}

	if settings[0] == "^" || settings[0] == "." {
		return reflect.TypeOf((*cell.Cell)(nil)), nil
	}

	if settings[0] == "dict" {
		_, err := strconv.ParseUint(settings[1], 10, 64)
		if err != nil {
			return nil, fmt.Errorf("cannot deserialize field as dict, bad size '%s'", settings[1])
		}
		if len(settings) >= 4 {
			// transformation
			return nil, errors.New("dict transformation is not supported")
		}
		return reflect.TypeOf((*cell.Dictionary)(nil)), nil
	}

	return nil, fmt.Errorf("cannot deserialize field as tag '%s'", tag)
}

func tlbParseDesc(fields []reflect.StructField, schema TLBFieldsDesc) (reflect.Type, error) {
	var (
		err error
		ok  bool
	)

	for _, field := range schema {
		var f = reflect.StructField{
			Name: strcase.ToCamel(field.Name),
			Tag:  reflect.StructTag(fmt.Sprintf("tlb:\"%s\" json:\"%s\"", field.Type, strcase.ToSnake(field.Name))),
		}

		// get type from map_to field
		f.Type, ok = typeNameMap[field.MapTo]
		if !ok {
			// parse tlb tag and get default type
			f.Type, err = tlbParseSettings(f.Tag.Get("tlb"))
			if f.Type == nil || err != nil {
				return nil, fmt.Errorf("%s (tag = %s) parse tlb settings: %w", f.Name, f.Tag.Get("tlb"), err)
			}
		}

		// make new struct
		if len(field.Fields) > 0 {
			f.Type, err = tlbParseDesc(nil, field.Fields)
			if err != nil {
				return nil, fmt.Errorf("%s: %w", f.Name, err)
			}
			f.Type = reflect.PointerTo(f.Type)
		}

		fields = append(fields, f)
	}

	return reflect.StructOf(fields), nil
}

func (d TLBFieldsDesc) New() (any, error) {
	t, err := tlbParseDesc(nil, d)
	if err != nil {
		return nil, err
	}
	return reflect.New(t).Interface(), nil
}

func (d *OperationDesc) New() (any, error) {
	var fields = []reflect.StructField{
		{
			Name: "Op",
			Tag:  reflect.StructTag(fmt.Sprintf("tlb:\"#%x\"", strings.ReplaceAll(d.Code, "0x", ""))),
			Type: reflect.TypeOf(tlb.Magic{}),
		},
	}
	t, err := tlbParseDesc(fields, d.Body)
	if err != nil {
		return nil, err
	}
	return reflect.New(t).Interface(), nil
}
