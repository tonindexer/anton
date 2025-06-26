package msg_hash

import (
	"fmt"
	"math/big"
	"reflect"
	"strconv"
	"strings"

	"github.com/pkg/errors"
	"github.com/xssnick/tonutils-go/address"
	"github.com/xssnick/tonutils-go/tlb"
	"github.com/xssnick/tonutils-go/tvm/cell"
)

func GetMessageHash(msg tlb.AnyMessage) ([]byte, error) {
	msgCell, err := toCell(msg)
	if err != nil {
		return nil, errors.Wrap(err, "cannot convert message to cell")
	}
	return msgCell.Hash(), nil
}

// copied from tlb.ToCell of tonutils-go@v1.13.0
// the only patch is to disable "store cell as ref directly" in storeField

func toCell(v any) (*cell.Cell, error) {
	rv := reflect.ValueOf(v)
	if rv.Kind() == reflect.Pointer {
		if rv.IsNil() {
			return nil, fmt.Errorf("v should not be nil")
		}
		rv = rv.Elem()
	}

	if ld, ok := v.(tlb.Marshaller); ok {
		c, err := ld.ToCell()
		if err != nil {
			return nil, fmt.Errorf("failed to store to cell for %s, using manual storer, err: %w", reflect.TypeOf(v).PkgPath(), err)
		}
		return c, nil
	}

	root := cell.BeginCell()

next:
	for i := 0; i < rv.NumField(); i++ {
		structField := rv.Type().Field(i)
		parseType := structField.Type
		fieldVal := rv.Field(i)
		tag := strings.TrimSpace(structField.Tag.Get("tlb"))
		if tag == "-" {
			continue
		}
		settings := strings.Split(tag, " ")

		if len(settings) == 0 {
			continue
		}

		if settings[0][0] == '?' {
			// conditional tlb parse depending on some field value of this struct
			cond := rv.FieldByName(settings[0][1:])
			if !cond.Bool() {
				continue
			}
			settings = settings[1:]
		}

		if settings[0] == "maybe" {
			if structField.Type.Kind() != reflect.Pointer && structField.Type.Kind() != reflect.Interface && structField.Type.Kind() != reflect.Slice {
				return nil, fmt.Errorf("maybe flag can only be applied to interface or pointer, field %s", structField.Name)
			}

			if fieldVal.IsNil() {
				if err := root.StoreBoolBit(false); err != nil {
					return nil, fmt.Errorf("cannot store maybe bit: %w", err)
				}
				continue
			}

			if err := root.StoreBoolBit(true); err != nil {
				return nil, fmt.Errorf("cannot store maybe bit: %w", err)
			}
			settings = settings[1:]
		}

		if structField.Type.Kind() == reflect.Pointer && structField.Type.Elem().Kind() != reflect.Struct {
			// to same process both pointers and types
			parseType = parseType.Elem()
			fieldVal = fieldVal.Elem()
		}

		if settings[0] == "either" {
			settings = settings[1:]

			if len(settings) < 2 {
				panic("either tag should have 2 args")
			}

			leaveBits, leaveRefs := 0, 0
			if settings[0] == "leave" {
				settings = settings[1:]

				if len(settings) < 3 {
					panic("either tag should have 2 args and leave tag should have 1 arg")
				}

				spl := strings.Split(settings[0], ",")
				settings = settings[1:]

				val, err := strconv.ParseUint(spl[0], 10, 10)
				if err != nil {
					panic("invalid argument for either leave bits")
				}
				// set how many free bits we need to have after either written
				leaveBits = int(val)

				if len(spl) > 1 {
					val, err = strconv.ParseUint(spl[1], 10, 10)
					if err != nil {
						panic("invalid argument for either leave refs")
					}
					// set how many free efs we need to have after either written
					leaveRefs = int(val)
				}
			}

			// we try first option, if it is overflows then we try second
			for x := 0; x < 2; x++ {
				builder := cell.BeginCell()
				if err := storeField([]string{settings[x]}, builder, structField, fieldVal, parseType); err != nil {
					return nil, fmt.Errorf("failed to serialize field %s to cell as either %d: %w", structField.Name, x, err)
				}

				// check if we have enough free bits
				if x == 0 && (int(root.BitsLeft())-int(builder.BitsUsed()+1) < leaveBits || int(root.RefsLeft())-int(builder.RefsUsed()) < leaveRefs) {
					// if not, then we try second option
					continue
				}

				if err := root.StoreUInt(uint64(x), 1); err != nil {
					return nil, fmt.Errorf("cannot store either bit: %w", err)
				}
				if err := root.StoreBuilder(builder); err != nil {
					return nil, fmt.Errorf("failed to concat builder of field %s to cell as either %d: %w", structField.Name, x, err)
				}

				continue next
			}

			return nil, fmt.Errorf("failed to serialize either field %s to cell: no valid options", structField.Name)
		}

		if err := storeField(settings, root, structField, fieldVal, parseType); err != nil {
			return nil, fmt.Errorf("failed to serialize field %s to cell: %w", structField.Name, err)
		}
	}

	return root.EndCell(), nil
}

var cellType = reflect.TypeOf(&cell.Cell{})

func storeField(settings []string, root *cell.Builder, structField reflect.StructField, fieldVal reflect.Value, parseType reflect.Type) error {
	builder := root

	asRef := false
	if settings[0] == "^" {
		// if cellType == parseType { // we disable this
		// 	// store cell as ref directly
		// 	if err := root.StoreRef(fieldVal.Interface().(*cell.Cell)); err != nil {
		// 		return fmt.Errorf("failed to store cell to ref for %s, err: %w", structField.Name, err)
		// 	}
		// 	return nil
		// }

		asRef = true
		settings = settings[1:]
		builder = cell.BeginCell()
	}

	if structField.Type.Kind() == reflect.Interface {
		allowed := strings.Join(settings, "")
		if !strings.HasPrefix(allowed, "[") || !strings.HasSuffix(allowed, "]") {
			panic("corrupted allowed list tag of field " + structField.Name + ", should be [a,b,c], got " + allowed)
		}

		// cut brackets
		allowed = allowed[1 : len(allowed)-1]
		types := strings.Split(allowed, ",")

		t := fieldVal.Elem().Type()
		if t.Kind() == reflect.Pointer {
			t = t.Elem()
		}

		found := false
		for _, typ := range types {
			if t.Name() == typ {
				found = true
				break
			}
		}

		if !found {
			return fmt.Errorf("unexpected data to serialize, not registered magic in tag for %s", t.String())
		}
		settings = settings[:0]
	}

	if len(settings) == 0 || settings[0] == "." {
		c, err := structStore(fieldVal, structField.Type.Name())
		if err != nil {
			return err
		}

		err = builder.StoreBuilder(c.ToBuilder())
		if err != nil {
			return fmt.Errorf("failed to store cell to builder for %s, err: %w", structField.Name, err)
		}
	} else if settings[0] == "##" {
		num, err := strconv.ParseUint(settings[1], 10, 64)
		if err != nil {
			// we panic, because its developer's issue, need to fix tag
			panic("corrupted num bits in ## tag")
		}

		switch {
		case num <= 64:
			switch parseType.Kind() {
			case reflect.Int64, reflect.Int32, reflect.Int16, reflect.Int8, reflect.Int:
				err = builder.StoreInt(fieldVal.Int(), uint(num))
				if err != nil {
					return fmt.Errorf("failed to store int %d, err: %w", num, err)
				}
			case reflect.Uint64, reflect.Uint32, reflect.Uint16, reflect.Uint8, reflect.Uint:
				err = builder.StoreUInt(fieldVal.Uint(), uint(num))
				if err != nil {
					return fmt.Errorf("failed to store int %d, err: %w", num, err)
				}
			default:
				if parseType == reflect.TypeOf(&big.Int{}) {
					err = builder.StoreBigInt(fieldVal.Interface().(*big.Int), uint(num))
					if err != nil {
						return fmt.Errorf("failed to store bigint %d, err: %w", num, err)
					}
				} else {
					panic("unexpected field type for tag ## - " + parseType.String())
				}
			}
		case num <= 256:
			err := builder.StoreBigInt(fieldVal.Interface().(*big.Int), uint(num))
			if err != nil {
				return fmt.Errorf("failed to store bigint %d, err: %w", num, err)
			}
		}
	} else if settings[0] == "addr" {
		err := builder.StoreAddr(fieldVal.Interface().(*address.Address))
		if err != nil {
			return fmt.Errorf("failed to store address, err: %w", err)
		}
	} else if settings[0] == "bool" {
		err := builder.StoreBoolBit(fieldVal.Bool())
		if err != nil {
			return fmt.Errorf("failed to store bool, err: %w", err)
		}
	} else if settings[0] == "bits" {
		num, err := strconv.Atoi(settings[1])
		if err != nil {
			// we panic, because its developer's issue, need to fix tag
			panic("corrupted num bits in bits tag")
		}

		err = builder.StoreSlice(fieldVal.Bytes(), uint(num))
		if err != nil {
			return fmt.Errorf("failed to store bits %d, err: %w", num, err)
		}
	} else if parseType == reflect.TypeOf(tlb.Magic{}) {
		var sz, base int
		if strings.HasPrefix(settings[0], "#") {
			base = 16
			sz = (len(settings[0]) - 1) * 4
		} else if strings.HasPrefix(settings[0], "$") {
			base = 2
			sz = len(settings[0]) - 1
		} else {
			panic("unknown magic value type in tag")
		}

		if sz > 64 {
			panic("too big magic value type in tag")
		}

		magic, err := strconv.ParseInt(settings[0][1:], base, 64)
		if err != nil {
			panic("corrupted magic value in tag")
		}

		err = builder.StoreUInt(uint64(magic), uint(sz))
		if err != nil {
			return fmt.Errorf("failed to store magic: %w", err)
		}
	} else if settings[0] == "dict" {
		var dict *cell.Dictionary

		settings = settings[1:]

		isInline := len(settings) > 0 && settings[0] == "inline"
		if isInline {
			settings = settings[1:]
		}

		if len(settings) < 3 || settings[1] != "->" {
			dict = fieldVal.Interface().(*cell.Dictionary)
		} else {
			if fieldVal.Kind() != reflect.Map {
				return fmt.Errorf("want to create dictionary from map, but instead got %s type", fieldVal.Type())
			}
			if fieldVal.Type().Key() != reflect.TypeOf("") {
				return fmt.Errorf("map key should be string, but instead got %s type", fieldVal.Type().Key())
			}

			sz, err := strconv.ParseUint(settings[0], 10, 64)
			if err != nil {
				panic(fmt.Sprintf("cannot deserialize field '%s' as dict, bad size '%s'", structField.Name, settings[0]))
			}

			dict = cell.NewDict(uint(sz))

			for _, mapK := range fieldVal.MapKeys() {
				mapKI, ok := big.NewInt(0).SetString(mapK.Interface().(string), 10)
				if !ok {
					return fmt.Errorf("cannot parse '%s' map key to big int of '%s' field", mapK.Interface().(string), structField.Name)
				}

				mapKB := cell.BeginCell()
				if err := mapKB.StoreBigInt(mapKI, uint(sz)); err != nil {
					return fmt.Errorf("store big int of size %d to %s field", sz, structField.Name)
				}

				mapV := fieldVal.MapIndex(mapK)

				cellVT := reflect.StructOf([]reflect.StructField{{
					Name: "Value",
					Type: mapV.Type(),
					Tag:  reflect.StructTag(fmt.Sprintf("tlb:%q", strings.Join(settings[2:], " "))),
				}})
				cellV := reflect.New(cellVT).Elem()
				cellV.Field(0).Set(mapV)

				mapVC, err := toCell(cellV.Interface())
				if err != nil {
					return fmt.Errorf("creating cell for dict value of '%s' field: %w", structField.Name, err)
				}

				if err := dict.Set(mapKB.EndCell(), mapVC); err != nil {
					return fmt.Errorf("set dict key/value on '%s' field: %w", structField.Name, err)
				}
			}
		}

		if isInline {
			dCell, err := dict.ToCell()
			if err != nil {
				return fmt.Errorf("failed to serialize inline dict to cell for %s, err: %w", structField.Name, err)
			}

			if dCell == nil {
				return fmt.Errorf("inline dict in field %s cannot be empty", structField.Name)
			}

			if err = builder.StoreBuilder(dCell.ToBuilder()); err != nil {
				return fmt.Errorf("failed to store inline dict for %s, err: %w", structField.Name, err)
			}
		} else {
			if err := builder.StoreDict(dict); err != nil {
				return fmt.Errorf("failed to store dict for %s, err: %w", structField.Name, err)
			}
		}
	} else if settings[0] == "var" {
		if settings[1] == "uint" {
			sz, err := strconv.Atoi(settings[2])
			if err != nil {
				panic(err.Error())
			}

			err = builder.StoreBigVarUInt(fieldVal.Interface().(*big.Int), uint(sz))
			if err != nil {
				return fmt.Errorf("failed to store var uint: %w", err)
			}
		} else {
			panic("var of type " + settings[1] + " is not supported")
		}
	} else {
		panic(fmt.Sprintf("cannot serialize field '%s' as tag '%s', use manual serialization", structField.Name, structField.Tag.Get("tlb")))
	}

	if asRef {
		err := root.StoreRef(builder.EndCell())
		if err != nil {
			return fmt.Errorf("failed to store cell to ref for %s, err: %w", structField.Name, err)
		}
	}

	return nil
}

func structStore(field reflect.Value, name string) (*cell.Cell, error) {
	if field.Type() == cellType {
		if field.IsNil() {
			return cell.BeginCell().EndCell(), nil
		}
		return field.Interface().(*cell.Cell), nil
	}

	inf := field.Interface()

	c, err := toCell(inf)
	if err != nil {
		return nil, fmt.Errorf("failed to store to cell for %s of type %s, err: %w", name, field.Type().String(), err)
	}
	return c, nil
}
