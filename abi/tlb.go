package abi

import (
	"fmt"
	"math/big"
	"reflect"
	"strconv"
	"strings"

	"github.com/iancoleman/strcase"
	"github.com/pkg/errors"

	"github.com/xssnick/tonutils-go/address"
	"github.com/xssnick/tonutils-go/tlb"
	"github.com/xssnick/tonutils-go/tvm/cell"
)

type TLBFieldDesc struct {
	Name     string        `json:"name"`
	Type     string        `json:"tlb_type"`
	Format   TLBType       `json:"format,omitempty"`
	Optional bool          `json:"optional,omitempty"`
	Fields   TLBFieldsDesc `json:"struct_fields,omitempty"` // Format = "struct"
}

type TLBFieldsDesc []TLBFieldDesc

type OperationDesc struct {
	Name string        `json:"op_name"`
	Code string        `json:"op_code"`
	Type string        `json:"type,omitempty"`
	Body TLBFieldsDesc `json:"body"`
}

func tlbMakeDesc(t reflect.Type, skipMagic ...bool) (ret TLBFieldsDesc, err error) {
	for i := 0; i < t.NumField(); i++ {
		f := t.Field(i)

		schema := TLBFieldDesc{
			Name: strcase.ToSnake(f.Name),
			Type: f.Tag.Get("tlb"),
		}

		if len(skipMagic) > 0 && skipMagic[0] && i == 0 && f.Type == reflect.TypeOf(tlb.Magic{}) {
			continue // skip tlb constructor tag as it has to be inside OperationDesc
		}

		ft, ok := typeNameRMap[f.Type]
		switch {
		case ok:
			schema.Format = ft

		case f.Type.Kind() == reflect.Pointer && f.Type.Elem().Kind() == reflect.Struct:
			schema.Format = TLBStructCell
			schema.Fields, err = tlbMakeDesc(f.Type.Elem())
			if err != nil {
				return nil, fmt.Errorf("%s: %w", f.Name, err)
			}

		case f.Type.Kind() == reflect.Struct:
			schema.Format = TLBStructCell
			schema.Fields, err = tlbMakeDesc(f.Type)
			if err != nil {
				return nil, fmt.Errorf("%s: %w", f.Name, err)
			}

		case strings.HasPrefix(schema.Type, "[") && strings.HasSuffix(schema.Type, "]"):
			// no format for union

		default:
			return nil, fmt.Errorf("%s: unknown structField type %s", f.Name, f.Type)
		}

		ret = append(ret, schema)
	}

	return ret, nil
}

func tlbParseSettingsInt(settings []string) (reflect.Type, error) {
	if settings[0] != "##" {
		return nil, fmt.Errorf("wrong int settings: %v", settings)
	}

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
	default:
		return nil, fmt.Errorf("too much bits for ## tag: %d", num)
	}
}

func tlbParseSettingsDict(settings []string) (reflect.Type, error) {
	if settings[0] != "dict" {
		return nil, fmt.Errorf("wrong dict settings: %v", settings)
	}

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

// tlbParseSettings automatically determines go type to map field into (refactor of tlb.LoadFromCell)
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

	if strings.HasPrefix(settings[0], "[") && strings.HasSuffix(settings[0], "]") {
		for _, dn := range strings.Split(tag[1:len(tag)-1], ",") {
			// iterate through union definitions
			// check that all definitions are known
			_, ok := registeredDefinitions[TLBType(dn)]
			if !ok {
				return nil, fmt.Errorf("cannot find definition for '%s' type inside union", dn)
			}
		}
		var v any
		return reflect.ValueOf(&v).Type().Elem(), nil // return type with interface kind
	}

	switch settings[0] {
	case "maybe":
		return reflect.TypeOf((*cell.Cell)(nil)), nil

	case "either":
		if len(settings) < 3 {
			return nil, errors.New("either tag should have 2 args")
		}
		return reflect.TypeOf((*cell.Cell)(nil)), nil

	case "##": // bits
		return tlbParseSettingsInt(settings)

	case "addr":
		return reflect.TypeOf((*address.Address)(nil)), nil

	case "bool":
		return reflect.TypeOf(false), nil

	case "bits":
		return reflect.TypeOf([]byte(nil)), nil

	case "^", ".":
		return reflect.TypeOf((*cell.Cell)(nil)), nil

	case "dict":
		return tlbParseSettingsDict(settings)

	default:
		return nil, fmt.Errorf("cannot deserialize field as tag '%s'", tag)
	}
}

func tlbMapFormat(format TLBType, tag reflect.StructTag) (reflect.Type, error) {
	t, ok := typeNameMap[format]
	if ok {
		return t, nil
	}

	switch format {
	case "":
		// parse tlb tag and get default type
		t, err := tlbParseSettings(tag.Get("tlb"))
		if t == nil || err != nil {
			return nil, fmt.Errorf("parse tlb settings with tag %s: %w", tag.Get("tlb"), err)
		}
		return t, nil

	default:
		d, ok := registeredDefinitions[format]
		if !ok {
			return nil, fmt.Errorf("cannot find definition for '%s' format", format)
		}
		t, err := tlbParseDesc(nil, d)
		if err != nil {
			return nil, errors.Wrap(err, "creating new type from definition")
		}
		return reflect.PointerTo(t), nil
	}
}

func tlbParseDesc(fields []reflect.StructField, schema TLBFieldsDesc, skipOptional ...bool) (reflect.Type, error) {
	var err error

	for i := range schema {
		f := &schema[i]

		if len(skipOptional) > 0 && skipOptional[0] && f.Optional {
			continue
		}
		if len(f.Fields) > 0 && f.Format == "" {
			f.Format = TLBStructCell
		}

		var sf = reflect.StructField{
			Name: strcase.ToCamel(f.Name),
			Tag:  reflect.StructTag(fmt.Sprintf("tlb:%q json:%q", f.Type, strcase.ToSnake(f.Name))),
		}

		if f.Format == TLBStructCell {
			sf.Type, err = tlbParseDesc(nil, f.Fields, skipOptional...)
			if err != nil {
				return nil, fmt.Errorf("%s field with struct: %w", sf.Name, err)
			}
			sf.Type = reflect.PointerTo(sf.Type)
		} else {
			sf.Type, err = tlbMapFormat(f.Format, sf.Tag)
			if err != nil {
				return nil, errors.Wrapf(err, "%s field", f.Name)
			}
		}

		fields = append(fields, sf)
	}

	return reflect.StructOf(fields), nil
}

func NewTLBDesc(x any) (TLBFieldsDesc, error) {
	rv := reflect.ValueOf(x)
	if rv.Kind() != reflect.Pointer {
		return nil, fmt.Errorf("x should be a pointer")
	}
	return tlbMakeDesc(rv.Type().Elem())
}

func (desc TLBFieldsDesc) New(skipOptional ...bool) (any, error) {
	t, err := tlbParseDesc(nil, desc, skipOptional...)
	if err != nil {
		return nil, err
	}
	return reflect.New(t).Interface(), nil
}

func (desc TLBFieldsDesc) FromCell(c *cell.Cell) (any, error) {
	parsed, err := desc.New()
	if err != nil {
		return nil, errors.Wrapf(err, "creating struct")
	}
	if err := tlb.LoadFromCell(parsed, c.BeginParse()); err == nil {
		return parsed, nil
	}
	if !strings.Contains(err.Error(), "not enough data in reader") && !strings.Contains(err.Error(), "no more refs exists") {
		return nil, errors.Wrap(err, "load from cell")
	}

	// skipping optional fields
	parsed, err = desc.New(true)
	if err != nil {
		return nil, errors.Wrapf(err, "creating struct (skip optional)")
	}
	if err := tlb.LoadFromCell(parsed, c.BeginParse()); err != nil {
		return nil, errors.Wrap(err, "load from cell (skip optional)")
	}

	return parsed, nil
}

func operationID(t reflect.Type) (uint32, error) {
	op := t.Field(0)
	if op.Type != reflect.TypeOf(tlb.Magic{}) {
		return 0, fmt.Errorf("no magic type in first struct field")
	}

	opValueStr, ok := op.Tag.Lookup("tlb")
	if !ok || len(opValueStr) != 9 || opValueStr[0] != '#' {
		return 0, fmt.Errorf("wrong tlb tag format (%s)", opValueStr)
	}

	opValue, err := strconv.ParseUint(opValueStr[1:], 16, 32)
	if err != nil {
		return 0, errors.Wrap(err, "parse hex uint32")
	}

	return uint32(opValue), nil
}

func NewOperationDesc(x any) (*OperationDesc, error) {
	var ret OperationDesc

	rv := reflect.ValueOf(x)
	if rv.Kind() != reflect.Pointer {
		return nil, fmt.Errorf("x should be a pointer")
	}

	t := rv.Type().Elem()

	ret.Name = strcase.ToSnake(t.Name())

	opCode, err := operationID(t)
	if err != nil {
		return nil, errors.Wrap(err, "lookup operation id")
	}
	ret.Code = fmt.Sprintf("0x%x", opCode)

	ret.Body, err = tlbMakeDesc(t, true)
	if err != nil {
		return nil, errors.Wrap(err, "make tlb schema")
	}

	return &ret, nil
}

func (desc *OperationDesc) New(skipOptional ...bool) (any, error) {
	var fields = []reflect.StructField{
		{
			Name: "Op",
			Tag:  reflect.StructTag(fmt.Sprintf("tlb:\"#%08s\" json:\"-\"", strings.ReplaceAll(desc.Code, "0x", ""))),
			Type: reflect.TypeOf(tlb.Magic{}),
		},
	}
	t, err := tlbParseDesc(fields, desc.Body, skipOptional...)
	if err != nil {
		return nil, err
	}
	return reflect.New(t).Interface(), nil
}

func (desc *OperationDesc) FromCell(c *cell.Cell) (any, error) {
	parsed, err := desc.New()
	if err != nil {
		return nil, errors.Wrapf(err, "creating struct")
	}
	if err = tlb.LoadFromCell(parsed, c.BeginParse()); err == nil {
		return parsed, nil
	}
	if !strings.Contains(err.Error(), "not enough data in reader") && !strings.Contains(err.Error(), "no more refs exists") {
		return nil, errors.Wrap(err, "load from cell")
	}

	// skipping optional fields
	parsed, err = desc.New(true)
	if err != nil {
		return nil, errors.Wrapf(err, "creating struct (skip optional)")
	}
	if err = tlb.LoadFromCell(parsed, c.BeginParse()); err != nil {
		return nil, errors.Wrap(err, "load from cell (skip optional)")
	}

	return parsed, nil
}
