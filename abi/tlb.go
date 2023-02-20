package abi

import (
	"encoding/json"
	"fmt"
	"math/big"
	"reflect"

	"github.com/xssnick/tonutils-go/address"
	"github.com/xssnick/tonutils-go/tlb"
	"github.com/xssnick/tonutils-go/tvm/cell"
)

type fieldSchema struct {
	Name         string         `json:"name,omitempty"`
	Type         string         `json:"type"`
	Tag          string         `json:"tag,omitempty"`
	StructFields []*fieldSchema `json:"struct_fields,omitempty"` // Type = "struct"
}

var (
	structTypeName = "struct"
	typeNameRMap   = map[reflect.Type]string{
		reflect.TypeOf([]uint8{}): "bytes",
	}
	typeNameMap = map[string]reflect.Type{
		"bool":    reflect.TypeOf(false),
		"uint16":  reflect.TypeOf(uint16(0)),
		"uint32":  reflect.TypeOf(uint32(0)),
		"uint64":  reflect.TypeOf(uint64(0)),
		"bytes":   reflect.TypeOf([]byte{}),
		"bigInt":  reflect.TypeOf(big.NewInt(0)),
		"cell":    reflect.TypeOf((*cell.Cell)(nil)),
		"magic":   reflect.TypeOf(tlb.Magic{}),
		"coins":   reflect.TypeOf(tlb.Coins{}),
		"address": reflect.TypeOf((*address.Address)(nil)),
	}
)

func init() {
	for t, n := range typeNameMap {
		typeNameRMap[n] = t
	}
}

func structToFields(t reflect.Type) (ret []*fieldSchema, err error) {
	for i := 0; i < t.NumField(); i++ {
		f := t.Field(i)
		ft, ok := typeNameRMap[f.Type]

		schema := &fieldSchema{
			Name: f.Name,
			Tag:  string(f.Tag),
		}

		switch {
		case ok:
			schema.Type = ft

		case f.Type.Kind() == reflect.Pointer && f.Type.Elem().Kind() == reflect.Struct:
			schema.Type = structTypeName
			schema.StructFields, err = structToFields(f.Type.Elem())
			if err != nil {
				return nil, fmt.Errorf("%s: %w", f.Name, err)
			}

		case f.Type.Kind() == reflect.Struct:
			schema.Type = structTypeName
			schema.StructFields, err = structToFields(f.Type)
			if err != nil {
				return nil, fmt.Errorf("%s: %w", f.Name, err)
			}

		default:
			return nil, fmt.Errorf("%s: unknown field type %s", f.Name, f.Type)
		}

		if schema.Name == "_" && schema.Type == "magic" {
			schema.Name = "Op"
		}
		ret = append(ret, schema)
	}

	return ret, nil
}

func fieldsToStruct(schema []*fieldSchema) (reflect.Type, error) {
	var fields []reflect.StructField
	var err error

	for _, field := range schema {
		f := reflect.StructField{
			Name: field.Name,
			Tag:  reflect.StructTag(field.Tag),
		}
		ft, ok := typeNameMap[field.Type]

		switch {
		case ok:
			f.Type = ft

		case field.Type == structTypeName:
			f.Type, err = fieldsToStruct(field.StructFields)
			if err != nil {
				return nil, fmt.Errorf("%s: %w", f.Name, err)
			}

		default:
			return nil, fmt.Errorf("cannot unmarshal type %s", f.Type)
		}

		fields = append(fields, f)
	}

	return reflect.StructOf(fields), nil
}

func MarshalSchema(x any) ([]byte, error) {
	rv := reflect.ValueOf(x)

	if rv.Kind() != reflect.Pointer {
		return nil, fmt.Errorf("x should be a pointer")
	}

	fields, err := structToFields(rv.Type().Elem())
	if err != nil {
		return nil, err
	}

	return json.Marshal(fields)
}

func UnmarshalSchema(raw []byte) (any, error) {
	var schema []*fieldSchema

	if err := json.Unmarshal(raw, &schema); err != nil {
		return nil, err
	}

	t, err := fieldsToStruct(schema)
	if err != nil {
		return nil, err
	}

	return reflect.New(t).Interface(), nil
}
