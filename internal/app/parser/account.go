package parser

import (
	"bytes"
	"context"
	"fmt"
	"reflect"
	"runtime"

	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
	"github.com/xssnick/tonutils-go/address"
	"github.com/xssnick/tonutils-go/tlb"
	"github.com/xssnick/tonutils-go/ton/wallet"
	"github.com/xssnick/tonutils-go/tvm/cell"

	"github.com/iam047801/tonidx/abi"
	"github.com/iam047801/tonidx/internal/core"
)

func matchByAddress(acc *tlb.Account, addr string) bool {
	if addr == "" {
		return false
	}
	return acc.State != nil && addr == acc.State.Address.String()
}

func matchByCode(acc *tlb.Account, code []byte) bool {
	if len(code) == 0 {
		return false
	}

	codeCell, err := cell.FromBOC(code)
	if err != nil {
		log.Error().Err(err).Msg("parse contract interface code")
		return false
	}

	return acc.Code != nil && bytes.Equal(acc.Code.Hash(), codeCell.Hash())
}

func (s *Service) DetermineInterfaces(ctx context.Context, acc *tlb.Account) ([]abi.ContractName, error) {
	var ret []abi.ContractName

	version := wallet.GetWalletVersion(acc)
	if version != wallet.Unknown {
		ret = append(ret,
			abi.ContractName(fmt.Sprintf("wallet_%s", version.String())))
	}

	ifaces, err := s.contractRepo.GetInterfaces(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "get contract interfaces")
	}

	for _, iface := range ifaces {
		if matchByAddress(acc, iface.Address) {
			ret = append(ret, iface.Name)
			continue
		}

		if matchByCode(acc, iface.Code) {
			ret = append(ret, iface.Name)
			continue
		}

		if len(iface.GetMethods) == 0 {
			continue
		}

		var hasMethods = true
		for _, get := range iface.GetMethods {
			if !acc.HasGetMethod(get) {
				hasMethods = false
				break
			}
		}
		if hasMethods {
			ret = append(ret, iface.Name)
		}
	}

	return ret, nil
}

func (s *Service) ParseAccountData(ctx context.Context, b *tlb.BlockInfo, acc *tlb.Account) (*core.AccountData, error) {
	var unknown int

	if acc.State == nil {
		return nil, errors.Wrap(core.ErrNotAvailable, "no account state")
	}
	if acc.State.Address.Type() != address.StdAddress {
		return nil, errors.Wrap(core.ErrNotAvailable, "no account address")
	}

	data := new(core.AccountData)
	data.Address = acc.State.Address.String()
	data.LastTxLT = acc.LastTxLT
	data.LastTxHash = acc.LastTxHash

	types, err := s.DetermineInterfaces(ctx, acc)
	if err != nil {
		return nil, errors.Wrap(err, "get contract interfaces")
	}
	if len(types) == 0 {
		return nil, errors.Wrap(err, "unknown contract interfaces")
	}

	getters := []func(context.Context, *tlb.BlockInfo, *tlb.Account, []abi.ContractName, *core.AccountData) error{
		s.getAccountDataNFT,
	}
	for _, getter := range getters {
		if err := getter(ctx, b, acc, types, data); err != nil && !errors.Is(err, core.ErrNotAvailable) {
			return nil, fmt.Errorf("%s: %w", runtime.FuncForPC(reflect.ValueOf(getter).Pointer()).Name(), err)
		} else if err != nil {
			unknown++
		}
	}
	if unknown == len(types) {
		return nil, errors.Wrap(core.ErrNotAvailable, "no data getters got a contract")
	}

	return data, nil
}
