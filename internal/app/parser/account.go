package parser

import (
	"bytes"
	"context"

	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
	"github.com/xssnick/tonutils-go/address"
	"github.com/xssnick/tonutils-go/ton"
	"github.com/xssnick/tonutils-go/tvm/cell"

	"github.com/iam047801/tonidx/abi"
	"github.com/iam047801/tonidx/internal/addr"
	"github.com/iam047801/tonidx/internal/core"
)

func matchByAddress(acc *core.AccountState, addresses []*addr.Address) bool {
	for _, a := range addresses {
		if addr.Equal(a, &acc.Address) {
			return true
		}
	}
	return false
}

func matchByCode(acc *core.AccountState, code []byte) bool {
	if len(acc.Code) == 0 || len(code) == 0 {
		return false
	}

	codeCell, err := cell.FromBOC(code)
	if err != nil {
		log.Error().Err(err).Msg("parse contract interface code")
		return false
	}

	accCodeCell, err := cell.FromBOC(acc.Code)
	if err != nil {
		log.Error().Err(err).Str("addr", acc.Address.Base64()).Msg("parse account code cell")
		return false
	}

	return bytes.Equal(accCodeCell.Hash(), codeCell.Hash())
}

func matchByGetMethods(acc *core.AccountState, getMethodHashes []uint32) bool {
	if len(acc.GetMethodHashes) == 0 || len(getMethodHashes) == 0 {
		return false
	}
	for _, x := range getMethodHashes {
		var found bool
		for _, y := range acc.GetMethodHashes {
			if x == y {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}
	return true
}

func (s *Service) DetermineInterfaces(ctx context.Context, acc *core.AccountState) ([]abi.ContractName, error) {
	var ret []abi.ContractName

	ifaces, err := s.contractRepo.GetInterfaces(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "get contract interfaces")
	}

	for _, iface := range ifaces {
		if matchByAddress(acc, iface.Addresses) {
			ret = append(ret, iface.Name)
			continue
		}

		if matchByCode(acc, iface.Code) {
			ret = append(ret, iface.Name)
			continue
		}

		if matchByGetMethods(acc, iface.GetMethodHashes) {
			ret = append(ret, iface.Name)
			continue
		}
	}

	return ret, nil
}

func (s *Service) ParseAccountData(ctx context.Context, b *ton.BlockIDExt, acc *core.AccountState, types []abi.ContractName) (*core.AccountData, error) {
	var unknown int

	if len(types) == 0 {
		return nil, errors.Wrap(core.ErrNotAvailable, "unknown contract interfaces")
	}

	a, err := acc.Address.ToTU()
	if err != nil {
		return nil, errors.Wrapf(err, "address to TU (%s)", acc.Address.Base64())
	}

	data := new(core.AccountData)
	data.Address = acc.Address
	data.LastTxLT = acc.LastTxLT
	data.LastTxHash = acc.LastTxHash
	data.Balance = acc.Balance
	data.Types = types

	getters := []func(context.Context, *ton.BlockIDExt, *address.Address, []abi.ContractName, *core.AccountData) bool{
		s.getAccountDataNFT,
		s.getAccountDataFT,
	}
	for _, getter := range getters {
		if !getter(ctx, b, a, types, data) {
			unknown++
			continue
		}
	}

	if data.Errors != nil {
		log.Warn().Str("address", acc.Address.Base64()).Strs("errors", data.Errors).Msg("parse account data")
	}
	return data, nil
}
