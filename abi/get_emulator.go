package abi

import (
	"context"
	"encoding/base64"
	"fmt"
	"math/big"
	"reflect"

	"github.com/tonkeeper/tongo/ton"
	"github.com/tonkeeper/tongo/txemulator"

	"github.com/pkg/errors"

	"github.com/tonkeeper/tongo"
	"github.com/tonkeeper/tongo/boc"
	"github.com/tonkeeper/tongo/tlb"
	"github.com/tonkeeper/tongo/tvm"

	"github.com/xssnick/tonutils-go/address"
	tutlb "github.com/xssnick/tonutils-go/tlb"
	"github.com/xssnick/tonutils-go/ton/nft"
	"github.com/xssnick/tonutils-go/tvm/cell"

	"github.com/tonindexer/anton/addr"
)

type VmValue struct {
	VmValueDesc
	Payload any `json:"payload"`
}

type VmStack []VmValue

type GetMethodExecution struct {
	Name string `json:"name,omitempty"`

	Address *addr.Address `json:"address,omitempty"`

	Arguments []VmValueDesc `json:"arguments,omitempty"`
	Receives  []any         `json:"receives,omitempty"`

	ReturnValues []VmValueDesc `json:"return_values,omitempty"`
	Returns      []any         `json:"returns,omitempty"`

	Error string `json:"error,omitempty"`
}

var ErrWrongValueFormat = errors.New("wrong value for this format")

type Emulator struct {
	Emulator  *tvm.Emulator
	AccountID tongo.AccountID
}

func newEmulator(addr *address.Address, e *tvm.Emulator) (*Emulator, error) {
	accId, err := ton.AccountIDFromBase64Url(addr.String())
	if err != nil {
		return nil, errors.Wrap(err, "parse address")
	}

	return &Emulator{Emulator: e, AccountID: accId}, nil
}

func NewEmulator(addr *address.Address, code, data, cfg *cell.Cell) (*Emulator, error) {
	e, err := tvm.NewEmulatorFromBOCsBase64(
		base64.StdEncoding.EncodeToString(code.ToBOC()),
		base64.StdEncoding.EncodeToString(data.ToBOC()),
		base64.StdEncoding.EncodeToString(cfg.ToBOC()),
	)
	if err != nil {
		return nil, err
	}
	return newEmulator(addr, e)
}

func NewEmulatorBase64(addr *address.Address, code, data, cfg, libraries string) (*Emulator, error) {
	var (
		e   *tvm.Emulator
		err error
	)

	if libraries != "" {
		e, err = tvm.NewEmulatorFromBOCsBase64(
			code,
			data,
			cfg,
			tvm.WithLazyC7Optimization(),
			tvm.WithLibrariesBase64(libraries),
			tvm.WithVerbosityLevel(txemulator.PrintsAllStackValuesForCommand),
		)
	} else {
		e, err = tvm.NewEmulatorFromBOCsBase64(code, data, cfg, tvm.WithLazyC7Optimization(), tvm.WithVerbosityLevel(txemulator.PrintsAllStackValuesForCommand))
	}

	if err != nil {
		return nil, err
	}

	return newEmulator(addr, e)
}

func vmMakeValueInt(v *VmValue) (ret tlb.VmStackValue, _ error) {
	var bi *big.Int
	var ok bool

	switch v.Format {
	case "", TLBBigInt:
		bi, ok = v.Payload.(*big.Int)
	case "uint8":
		ui, uok := v.Payload.(uint8)
		bi, ok = big.NewInt(int64(ui)), uok
	case "uint16":
		ui, uok := v.Payload.(uint16)
		bi, ok = big.NewInt(int64(ui)), uok
	case "uint32":
		ui, uok := v.Payload.(uint32)
		bi, ok = big.NewInt(int64(ui)), uok
	case "uint64":
		ui, uok := v.Payload.(uint64)
		bi, ok = big.NewInt(int64(ui)), uok
	case "int8":
		ui, uok := v.Payload.(int8)
		bi, ok = big.NewInt(int64(ui)), uok
	case "int16":
		ui, uok := v.Payload.(int16)
		bi, ok = big.NewInt(int64(ui)), uok
	case "int32":
		ui, uok := v.Payload.(int32)
		bi, ok = big.NewInt(int64(ui)), uok
	case "int64":
		ui, uok := v.Payload.(int64)
		bi, ok = big.NewInt(ui), uok
	case "bytes":
		ui, uok := v.Payload.([]byte)
		bi, ok = new(big.Int).SetBytes(ui), uok
	}
	if !ok {
		return ret, errors.Wrapf(ErrWrongValueFormat, "'%s' type with '%s' format", v.StackType, v.Format)
	}

	ret.SumType = "VmStkInt"
	ret.VmStkInt = *(*tlb.Int257)(bi)

	return ret, nil
}

func vmMakeValueCell(v *VmValue) (tlb.VmStackValue, error) {
	var c *cell.Cell
	var ok bool

	switch v.Format {
	case "", TLBCell:
		c, ok = v.Payload.(*cell.Cell)
	case TLBAddr:
		a, aok := v.Payload.(*address.Address)
		if aok {
			b := cell.BeginCell()
			if err := b.StoreAddr(a); err != nil {
				return tlb.VmStackValue{}, errors.Wrap(err, "store address")
			}
			c, ok = b.EndCell(), aok
		}
	case TLBString:
		s, sok := v.Payload.(string)
		if sok {
			b := cell.BeginCell()
			if err := b.StoreStringSnake(s); err != nil {
				return tlb.VmStackValue{}, errors.Wrap(err, "store string snake")
			}
			c, ok = b.EndCell(), sok
		}
	case TLBStructCell:
		var err error
		c, err = tutlb.ToCell(v.Payload)
		if err != nil {
			return tlb.VmStackValue{}, err
		}
	}
	if !ok {
		return tlb.VmStackValue{}, errors.Wrapf(ErrWrongValueFormat, "'%s' type with '%s' format", v.StackType, v.Format)
	}

	if c == nil {
		return tlb.VmStackValue{SumType: "VmStkNull"}, nil
	}

	tgc, err := boc.DeserializeSinglRootBase64(base64.StdEncoding.EncodeToString(c.ToBOC()))
	if err != nil {
		return tlb.VmStackValue{}, errors.Wrapf(err, "tongo deserialize boc cell")
	}

	ret, err := tlb.TlbStructToVmCell(tgc)
	return ret, err
}

func vmMakeValueSlice(v *VmValue) (tlb.VmStackValue, error) {
	var s *cell.Slice
	var ok bool

	switch v.Format {
	case "", TLBType(VmSlice):
		s, ok = v.Payload.(*cell.Slice)
	case TLBAddr:
		a, aok := v.Payload.(*address.Address)
		if aok {
			b := cell.BeginCell()
			if err := b.StoreAddr(a); err != nil {
				return tlb.VmStackValue{}, errors.Wrap(err, "store address")
			}
			s, ok = b.EndCell().BeginParse(), aok
		}
	case TLBString:
		a, aok := v.Payload.(string)
		if aok {
			b := cell.BeginCell()
			if err := b.StoreStringSnake(a); err != nil {
				return tlb.VmStackValue{}, errors.Wrap(err, "store string")
			}
			s, ok = b.EndCell().BeginParse(), aok
		}
	case TLBStructCell:
		c, err := tutlb.ToCell(v.Payload)
		if err != nil {
			return tlb.VmStackValue{}, err
		}
		s = c.BeginParse()
	}
	if !ok {
		return tlb.VmStackValue{}, errors.Wrapf(ErrWrongValueFormat, "'%s' type with '%s' format", v.StackType, v.Format)
	}

	c, err := s.ToCell()
	if err != nil {
		return tlb.VmStackValue{}, errors.Wrap(err, "convert slice to cell")
	}

	tgc, err := boc.DeserializeSinglRootBase64(base64.StdEncoding.EncodeToString(c.ToBOC()))
	if err != nil {
		return tlb.VmStackValue{}, errors.Wrapf(err, "tongo deserialize boc cell")
	}

	ret, err := tlb.TlbStructToVmCellSlice(tgc)
	return ret, err
}

func vmMakeValue(v *VmValue) (ret tlb.VmStackValue, _ error) {
	switch v.StackType {
	case VmInt:
		return vmMakeValueInt(v)

	case VmCell:
		return vmMakeValueCell(v)

	case VmSlice:
		return vmMakeValueSlice(v)

	default:
		return ret, fmt.Errorf("unsupported '%s' type", v.StackType)
	}
}

func vmParseValueInt(v *tlb.VmStackValue, d *VmValueDesc) (any, error) {
	var bi *big.Int

	switch v.SumType {
	case "VmStkInt":
		bi = (*big.Int)(&v.VmStkInt)
	case "VmStkTinyInt":
		bi = big.NewInt(v.VmStkTinyInt)
	default:
		return nil, fmt.Errorf("wrong descriptor '%s' type as method returned '%s'", d.StackType, v.SumType)
	}

	switch d.Format {
	case "", TLBBigInt:
		return bi, nil
	case "uint8":
		return uint8(bi.Uint64()), nil
	case "uint16":
		return uint16(bi.Uint64()), nil
	case "uint32":
		return uint32(bi.Uint64()), nil
	case "uint64":
		return bi.Uint64(), nil
	case "int8":
		return int8(bi.Int64()), nil
	case "int16":
		return int16(bi.Int64()), nil
	case "int32":
		return int32(bi.Int64()), nil
	case "int64":
		return bi.Int64(), nil
	case TLBBool:
		return bi.Cmp(big.NewInt(0)) != 0, nil
	case TLBBytes:
		return bi.Bytes(), nil
	default:
		return nil, fmt.Errorf("unsupported '%s' format for '%s' type", d.Format, d.StackType)
	}
}

func vmParseCell(c *cell.Cell, desc *VmValueDesc) (any, error) {
	switch desc.Format {
	case TLBCell:
		return c, nil

	case TLBSlice:
		return c.BeginParse(), nil

	case TLBString:
		s, err := c.BeginParse().LoadStringSnake()
		if err != nil {
			return nil, errors.Wrap(err, "load string snake")
		}
		return s, nil

	case TLBAddr:
		a, err := c.BeginParse().LoadAddr()
		if err != nil {
			return nil, errors.Wrap(err, "load address")
		}
		return a, nil

	case TLBContentCell:
		content, err := nft.ContentFromCell(c)
		if err != nil {
			return nil, errors.Wrap(err, "load content from cell")
		}
		return content, nil

	case TLBStructCell:
		parsed, err := desc.Fields.FromCell(c)
		if err != nil {
			return nil, errors.Wrapf(err, "load struct from cell on %s value description schema", desc.Name)
		}
		return parsed, nil

	default:
		d, ok := registeredDefinitions[desc.Format]
		if !ok {
			t, ok := typeNameMap[desc.Format]
			if !ok {
				return nil, fmt.Errorf("cannot find definition or type for '%s' format", desc.Format)
			}
			if t.Kind() == reflect.Pointer {
				t = t.Elem()
			}
			tv := reflect.New(t).Interface()
			if err := tutlb.LoadFromCell(tv, c.BeginParse()); err != nil {
				return nil, fmt.Errorf("load type '%s' from cell: %w", desc.Format, err)
			}
			return tv, nil
		}
		parsed, err := d.FromCell(c)
		if err != nil {
			return nil, errors.Wrapf(err, "'%s' definition from cell", desc.Format)
		}
		return parsed, nil
	}
}

func vmParseValueCell(v *tlb.VmStackValue, desc *VmValueDesc) (any, error) {
	switch v.SumType {
	case "VmStkNull":
		switch desc.Format {
		case "", TLBCell, TLBStructCell:
			return (*cell.Cell)(nil), nil
		case TLBString:
			return "", nil
		case TLBContentCell:
			return nft.ContentAny(nil), nil
		default:
			return nil, fmt.Errorf("unsupported '%s' format for '%s' type", desc.Format, desc.StackType)
		}

	case "VmStkCell":
		// go further

	default:
		return nil, fmt.Errorf("wrong descriptor '%s' type as method returned '%s'", desc.StackType, v.SumType)
	}

	tgcBoc, err := v.VmStkCell.Value.ToBocCustom(false, false, false, 0)
	if err != nil {
		return nil, errors.Wrap(err, "convert stack cell to boc")
	}
	c, err := cell.FromBOC(tgcBoc)
	if err != nil {
		return nil, errors.Wrap(err, "convert boc to cell")
	}

	if desc.Format == "" && len(desc.Fields) > 0 {
		desc.Format = TLBStructCell
	} else if desc.Format == "" {
		desc.Format = TLBCell
	}

	return vmParseCell(c, desc)
}

func vmParseValueSlice(v *tlb.VmStackValue, desc *VmValueDesc) (any, error) {
	switch v.SumType {
	case "VmStkNull":
		switch desc.Format {
		case "":
			return (*cell.Slice)(nil), nil
		case TLBAddr:
			return address.NewAddressNone(), nil
		case TLBString:
			return "", nil
		default:
			return nil, fmt.Errorf("unsupported '%s' format for '%s' type", desc.Format, desc.StackType)
		}

	case "VmStkSlice":
		// go further

	default:
		return nil, fmt.Errorf("wrong descriptor '%s' type as method returned '%s'", desc.StackType, v.SumType)
	}

	tgcBoc, err := v.VmStkSlice.Cell().ToBocCustom(false, false, false, 0)
	if err != nil {
		return nil, errors.Wrap(err, "convert stack cell to boc")
	}
	c, err := cell.FromBOC(tgcBoc)
	if err != nil {
		return nil, errors.Wrap(err, "convert boc to cell")
	}

	if desc.Format == "" && len(desc.Fields) > 0 {
		desc.Format = TLBStructCell
	} else if desc.Format == "" {
		desc.Format = TLBSlice
	}

	return vmParseCell(c, desc)
}

func vmParseValue(v *tlb.VmStackValue, d *VmValueDesc) (any, error) {
	switch d.StackType {
	case "int":
		return vmParseValueInt(v, d)

	case "cell":
		return vmParseValueCell(v, d)

	case "slice":
		return vmParseValueSlice(v, d)

	default:
		return nil, fmt.Errorf("unsupported '%s' type", d.StackType)
	}
}

func (e *Emulator) RunGetMethod(ctx context.Context, method string, args VmStack, retDesc []VmValueDesc) (ret VmStack, err error) {
	var params tlb.VmStack

	for it := range args {
		v, err := vmMakeValue(&args[it])
		if err != nil {
			return nil, err
		}
		params.Put(v)
	}

	exit, stk, err := e.Emulator.RunSmcMethod(ctx, e.AccountID, method, params)
	if err != nil {
		return nil, errors.Wrap(err, "run smc method")
	}
	if exit != 0 && exit != 1 { // 1 - alternative success code
		return nil, fmt.Errorf("tvm execution failed with code %d", exit)
	}
	if len(stk) < len(retDesc) {
		return nil, fmt.Errorf("tvm execution returned stack with length %d, but expected length %d", len(stk), len(retDesc))
	}

	for i := range retDesc {
		r, err := vmParseValue(&stk[i], &retDesc[i])
		if err != nil {
			return nil, err
		}
		ret = append(ret, VmValue{VmValueDesc: retDesc[i], Payload: r})
	}

	return ret, nil
}
