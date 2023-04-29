package abi

// #cgo darwin LDFLAGS: -L ./lib/darwin/ -Wl,-rpath,./lib/darwin/ -l emulator
// #cgo linux LDFLAGS: -L ./lib/linux/ -Wl,-rpath,./lib/linux/ -l emulator
// #include "./lib/emulator-extern.h"
// #include <stdlib.h>
// #include <stdbool.h>
import "C"
import (
	"context"
	"encoding/base64"
	"fmt"
	"math/big"
	"strings"

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

func parseVmValue(v VmValue) (ret tlb.VmStackValue, _ error) {
	switch v.FuncType {
	case "int":
		var bi *big.Int
		var ok bool

		switch strings.ToLower(v.Format) {
		case "", "bigint":
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
		}
		if !ok {
			return ret, errors.Wrapf(ErrWrongValueFormat, "'%s' type with '%s' format", v.FuncType, v.Format)
		}

		ret.SumType = "VmStkInt"
		ret.VmStkInt = *(*tlb.Int257)(bi)
		return ret, nil

	case "cell":
		var c *cell.Cell
		var ok bool

		switch strings.ToLower(v.Format) {
		case "", "cell":
			c, ok = v.Payload.(*cell.Cell)
		case "string":
			s, sok := v.Payload.(string)
			if sok {
				b := cell.BeginCell()
				if err := b.StoreStringSnake(s); err != nil {
					return ret, errors.Wrap(err, "store string snake")
				}
				c, ok = b.EndCell(), sok
			}
		}
		if !ok {
			return ret, errors.Wrapf(ErrWrongValueFormat, "'%s' type with '%s' format", v.FuncType, v.Format)
		}

		tgc, err := boc.DeserializeSinglRootBase64(base64.StdEncoding.EncodeToString(c.ToBOC()))
		if err != nil {
			return ret, errors.Wrapf(err, "tongo deserialize boc cell")
		}

		ret, err = tlb.TlbStructToVmCell(tgc)
		return ret, nil

	case "slice":
		var s *cell.Slice
		var ok bool

		switch strings.ToLower(v.Format) {
		case "", "slice":
			s, ok = v.Payload.(*cell.Slice)
		case "addr":
			a, aok := v.Payload.(*address.Address)
			if aok {
				b := cell.BeginCell()
				if err := b.StoreAddr(a); err != nil {
					return ret, errors.Wrap(err, "store address")
				}
				s, ok = b.EndCell().BeginParse(), aok
			}
		}
		if !ok {
			return ret, errors.Wrapf(ErrWrongValueFormat, "'%s' type with '%s' format", v.FuncType, v.Format)
		}

		c, err := s.ToCell()
		if err != nil {
			return ret, errors.Wrap(err, "convert slice to cell")
		}

		tgc, err := boc.DeserializeSinglRootBase64(base64.StdEncoding.EncodeToString(c.ToBOC()))
		if err != nil {
			return ret, errors.Wrapf(err, "tongo deserialize boc cell")
		}

		ret, err = tlb.TlbStructToVmCellSlice(tgc)
		return ret, err

	default:
		return ret, fmt.Errorf("unsupported '%s' type", v.FuncType)
	}

}

func mapToVmValue(v tlb.VmStackValue, d VmValueDesc) (any, error) {
	switch d.FuncType {
	case "int":
		var bi *big.Int

		switch v.SumType {
		case "VmStkInt":
			bi = (*big.Int)(&v.VmStkInt)
		case "VmStkTinyInt":
			bi = big.NewInt(v.VmStkTinyInt)
		default:
			return nil, fmt.Errorf("wrong descriptor '%s' type as method returned '%s'", d.FuncType, v.SumType)
		}

		switch strings.ToLower(d.Format) {
		case "", "bigint":
			return bi, nil
		case "uint8":
			return uint8(bi.Uint64()), nil
		case "uint16":
			return uint16(bi.Uint64()), nil
		case "uint32":
			return uint32(bi.Uint64()), nil
		case "uint64":
			return bi.Uint64(), nil
		case "bool":
			return bi.Cmp(big.NewInt(0)) != 0, nil
		default:
			return nil, fmt.Errorf("unsupported '%s' format for '%s' type", d.Format, d.FuncType)
		}

	case "cell":
		switch v.SumType {
		case "VmStkNull":
			return (*cell.Cell)(nil), nil
		case "VmStkCell":
		default:
			return nil, fmt.Errorf("wrong descriptor '%s' type as method returned '%s'", d.FuncType, v.SumType)
		}

		tgcBoc, err := v.VmStkCell.Value.ToBocCustom(false, false, false, 0)
		if err != nil {
			return nil, errors.Wrap(err, "convert stack cell to boc")
		}
		c, err := cell.FromBOC(tgcBoc)
		if err != nil {
			return nil, errors.Wrap(err, "convert boc to cell")
		}

		switch strings.ToLower(d.Format) {
		case "", "cell":
			return c, nil
		case "string":
			s, err := c.BeginParse().LoadStringSnake()
			if err != nil {
				return nil, errors.Wrap(err, "load string snake")
			}
			return s, nil
		case "content":
			content, err := nft.ContentFromCell(c)
			if err != nil {
				return nil, errors.Wrap(err, "load content from cell")
			}
			return content, nil
		default:
			return nil, fmt.Errorf("unsupported '%s' format for '%s' type", d.Format, d.FuncType)
		}

	case "slice":
		switch v.SumType {
		case "VmStkNull":
			return (*cell.Slice)(nil), nil
		case "VmStkSlice":
		default:
			return nil, fmt.Errorf("wrong descriptor '%s' type as method returned '%s'", d.FuncType, v.SumType)
		}

		tgcBoc, err := v.VmStkSlice.Cell().ToBocCustom(false, false, false, 0)
		if err != nil {
			return nil, errors.Wrap(err, "convert stack cell to boc")
		}
		c, err := cell.FromBOC(tgcBoc)
		if err != nil {
			return nil, errors.Wrap(err, "convert boc to cell")
		}

		switch strings.ToLower(d.Format) {
		case "", "slice":
			return c.BeginParse(), nil
		case "addr":
			a, err := c.BeginParse().LoadAddr()
			if err != nil {
				return nil, errors.Wrap(err, "load address")
			}
			return a, nil
		default:
			return nil, fmt.Errorf("unsupported '%s' format for '%s' type", d.Format, d.FuncType)
		}

	default:
		return nil, fmt.Errorf("unsupported '%s' type", d.FuncType)
	}
}

func (e *Emulator) RunGetMethod(ctx context.Context, method string, args VmStack, retDesc []VmValueDesc) (ret []any, err error) {
	var params tlb.VmStack

	for _, a := range args {
		v, err := parseVmValue(a)
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
		r, err := mapToVmValue(stk[i], retDesc[i])
		if err != nil {
			return nil, err
		}
		ret = append(ret, r)
	}

	return ret, nil
}
