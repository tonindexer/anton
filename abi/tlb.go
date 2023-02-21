package abi

import (
	"encoding/json"
	"fmt"
	"math/big"
	"reflect"
	"strconv"

	"github.com/pkg/errors"
	"github.com/xssnick/tonutils-go/address"
	"github.com/xssnick/tonutils-go/tlb"
	"github.com/xssnick/tonutils-go/tvm/cell"
)

type structField struct {
	Name         string         `json:"name,omitempty"`
	Type         string         `json:"type"`
	Tag          string         `json:"tag,omitempty"`
	StructFields []*structField `json:"struct_fields,omitempty"` // Type = "struct"
}

type TelemintText struct {
	Len  uint8  // ## 8
	Text []byte // bits (len * 8)
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
	x.Text = t

	return nil
}

var (
	structTypeName = "struct"
	typeNameRMap   = map[reflect.Type]string{
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
		"magic":        reflect.TypeOf(tlb.Magic{}),
		"coins":        reflect.TypeOf(tlb.Coins{}),
		"address":      reflect.TypeOf((*address.Address)(nil)),
		"telemintText": reflect.TypeOf((*TelemintText)(nil)),
	}
)

func init() {
	for t, n := range typeNameMap {
		typeNameRMap[n] = t
	}
}

func structToFields(t reflect.Type) (ret []*structField, err error) {
	for i := 0; i < t.NumField(); i++ {
		f := t.Field(i)
		ft, ok := typeNameRMap[f.Type]

		schema := &structField{
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
			return nil, fmt.Errorf("%s: unknown structField type %s", f.Name, f.Type)
		}

		if schema.Name == "_" && schema.Type == "magic" {
			schema.Name = "Op"
		}
		ret = append(ret, schema)
	}

	return ret, nil
}

func fieldsToStruct(schema []*structField) (reflect.Type, error) {
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
	var schema []*structField

	if err := json.Unmarshal(raw, &schema); err != nil {
		return nil, err
	}

	t, err := fieldsToStruct(schema)
	if err != nil {
		return nil, err
	}

	return reflect.New(t).Interface(), nil
}

func OperationID(x any) (uint32, error) {
	if reflect.TypeOf(x).Kind() != reflect.Pointer {
		return 0, fmt.Errorf("x should be a pointer")
	}

	s := reflect.TypeOf(x).Elem()
	if s.NumField() < 1 {
		return 0, fmt.Errorf("no struct fields")
	}

	op := s.Field(0)
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

func ParseOperationID(body []byte) (opId uint32, comment string, err error) {
	payload, err := cell.FromBOC(body)
	if err != nil {
		return 0, "", errors.Wrap(err, "msg body from boc")
	}
	slice := payload.BeginParse()

	op, err := slice.LoadUInt(32)
	if err != nil {
		return 0, "", errors.Wrap(err, "load uint")
	}

	if opId = uint32(op); opId != 0 {
		return opId, "", nil
	}

	// simple transfer with comment
	if comment, err = slice.LoadStringSnake(); err != nil {
		return 0, "", errors.Wrap(err, "load transfer comment")
	}

	return opId, comment, nil
}
