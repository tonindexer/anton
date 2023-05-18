package abi

import (
	"context"
	"encoding/base64"
	"fmt"
	"math/big"

	"github.com/pkg/errors"

	"github.com/tonkeeper/tongo"
	"github.com/tonkeeper/tongo/boc"
	"github.com/tonkeeper/tongo/tlb"
	"github.com/tonkeeper/tongo/tvm"

	"github.com/xssnick/tonutils-go/address"
	"github.com/xssnick/tonutils-go/ton/nft"
	"github.com/xssnick/tonutils-go/tvm/cell"
)

type VmValue struct {
	VmValueDesc
	Payload any
}

type VmStack []VmValue

var ErrWrongValueFormat = errors.New("wrong value for this format")

type Emulator struct {
	Emulator  *tvm.Emulator
	AccountID tongo.AccountID
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
	err = e.SetVerbosityLevel(0)
	if err != nil {
		return nil, errors.Wrap(err, "set verbosity level")
	}
	accId, err := tongo.AccountIDFromBase64Url(addr.String())
	if err != nil {
		return nil, errors.Wrap(err, "parse address")
	}
	return &Emulator{Emulator: e, AccountID: accId}, nil
}

func vmMakeValueInt(v *VmValue) (ret tlb.VmStackValue, _ error) {
	var bi *big.Int
	var ok bool

	switch v.Format {
	case "", VmBigInt:
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
	case "", VmCell:
		c, ok = v.Payload.(*cell.Cell)
	case VmStringCell:
		s, sok := v.Payload.(string)
		if sok {
			b := cell.BeginCell()
			if err := b.StoreStringSnake(s); err != nil {
				return tlb.VmStackValue{}, errors.Wrap(err, "store string snake")
			}
			c, ok = b.EndCell(), sok
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
	case "", VmSlice:
		s, ok = v.Payload.(*cell.Slice)
	case VmAddrSlice:
		a, aok := v.Payload.(*address.Address)
		if aok {
			b := cell.BeginCell()
			if err := b.StoreAddr(a); err != nil {
				return tlb.VmStackValue{}, errors.Wrap(err, "store address")
			}
			s, ok = b.EndCell().BeginParse(), aok
		}
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
	case "", VmBigInt:
		return bi, nil
	case "uint8":
		return uint8(bi.Uint64()), nil
	case "uint16":
		return uint16(bi.Uint64()), nil
	case "uint32":
		return uint32(bi.Uint64()), nil
	case "uint64":
		return bi.Uint64(), nil
	case VmBool:
		return bi.Cmp(big.NewInt(0)) != 0, nil
	case VmBytes:
		return bi.Bytes(), nil
	default:
		return nil, fmt.Errorf("unsupported '%s' format for '%s' type", d.Format, d.StackType)
	}
}

func vmParseValueCell(v *tlb.VmStackValue, d *VmValueDesc) (any, error) {
	switch v.SumType {
	case "VmStkNull":
		switch d.Format {
		case "", VmCell:
			return (*cell.Cell)(nil), nil
		case VmStringCell:
			return "", nil
		case VmContentCell:
			return nft.ContentAny(nil), nil
		default:
			return nil, fmt.Errorf("unsupported '%s' format for '%s' type", d.Format, d.StackType)
		}

	case "VmStkCell":
		// go further

	default:
		return nil, fmt.Errorf("wrong descriptor '%s' type as method returned '%s'", d.StackType, v.SumType)
	}

	tgcBoc, err := v.VmStkCell.Value.ToBocCustom(false, false, false, 0)
	if err != nil {
		return nil, errors.Wrap(err, "convert stack cell to boc")
	}
	c, err := cell.FromBOC(tgcBoc)
	if err != nil {
		return nil, errors.Wrap(err, "convert boc to cell")
	}

	switch d.Format {
	case "", VmCell:
		return c, nil
	case VmStringCell:
		s, err := c.BeginParse().LoadStringSnake()
		if err != nil {
			return nil, errors.Wrap(err, "load string snake")
		}
		return s, nil
	case VmContentCell:
		content, err := nft.ContentFromCell(c)
		if err != nil {
			return nil, errors.Wrap(err, "load content from cell")
		}
		return content, nil
	default:
		return nil, fmt.Errorf("unsupported '%s' format for '%s' type", d.Format, d.StackType)
	}
}

func vmParseValueSlice(v *tlb.VmStackValue, d *VmValueDesc) (any, error) {
	switch v.SumType {
	case "VmStkNull":
		switch d.Format {
		case "", VmSlice:
			return (*cell.Slice)(nil), nil
		case VmAddrSlice:
			return address.NewAddressNone(), nil
		default:
			return nil, fmt.Errorf("unsupported '%s' format for '%s' type", d.Format, d.StackType)
		}

	case "VmStkSlice":
		// go further

	default:
		return nil, fmt.Errorf("wrong descriptor '%s' type as method returned '%s'", d.StackType, v.SumType)
	}

	tgcBoc, err := v.VmStkSlice.Cell().ToBocCustom(false, false, false, 0)
	if err != nil {
		return nil, errors.Wrap(err, "convert stack cell to boc")
	}
	c, err := cell.FromBOC(tgcBoc)
	if err != nil {
		return nil, errors.Wrap(err, "convert boc to cell")
	}

	switch d.Format {
	case "", VmSlice:
		return c.BeginParse(), nil
	case VmAddrSlice:
		a, err := c.BeginParse().LoadAddr()
		if err != nil {
			return nil, errors.Wrap(err, "load address")
		}
		return a, nil
	default:
		return nil, fmt.Errorf("unsupported '%s' format for '%s' type", d.Format, d.StackType)
	}
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
	if len(stk) != len(retDesc) {
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
