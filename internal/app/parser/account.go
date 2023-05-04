package parser

import (
	"bytes"
	"context"

	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
	"github.com/xssnick/tonutils-go/tvm/cell"

	"github.com/tonindexer/anton/addr"
	"github.com/tonindexer/anton/internal/app"
	"github.com/tonindexer/anton/internal/core"
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
	codeHash := codeCell.Hash()

	accCodeCell, err := cell.FromBOC(acc.Code)
	if err != nil {
		log.Error().Err(err).Str("addr", acc.Address.Base64()).Msg("parse account code cell")
		return false
	}

	return bytes.Equal(accCodeCell.Hash(), codeHash)
}

func matchByGetMethods(acc *core.AccountState, getMethodHashes []int32) bool {
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

func (s *Service) determineInterfaces(ctx context.Context, acc *core.AccountState) ([]*core.ContractInterface, error) {
	var ret []*core.ContractInterface

	interfaces, err := s.contractRepo.GetInterfaces(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "get contract interfaces")
	}

	for _, i := range interfaces {
		if matchByAddress(acc, i.Addresses) {
			ret = append(ret, i)
			continue
		}

		if matchByCode(acc, i.Code) {
			ret = append(ret, i)
			continue
		}

		if len(i.Addresses) != 0 || len(i.Code) != 0 {
			continue // match by get methods only if code and addresses are not set
		}
		if matchByGetMethods(acc, i.GetMethodHashes) {
			ret = append(ret, i)
			continue
		}
	}

	return ret, nil
}

func (s *Service) ParseAccountData(
	ctx context.Context,
	acc *core.AccountState,
	others func(context.Context, *addr.Address) (*core.AccountState, error),
) (*core.AccountData, error) {
	interfaces, err := s.determineInterfaces(ctx, acc)
	if err != nil {
		return nil, errors.Wrapf(err, "determine contract interfaces")
	}
	if len(interfaces) == 0 {
		return nil, errors.Wrap(app.ErrImpossibleParsing, "unknown contract interfaces")
	}

	data := new(core.AccountData)
	data.Address = acc.Address
	data.LastTxLT = acc.LastTxLT
	data.LastTxHash = acc.LastTxHash
	data.Balance = acc.Balance
	for _, i := range interfaces {
		data.Types = append(data.Types, i.Name)
	}
	data.UpdatedAt = acc.UpdatedAt

	getters := []func(context.Context, *core.AccountState, func(context.Context, *addr.Address) (*core.AccountState, error), []*core.ContractInterface, *core.AccountData){
		s.getAccountDataNFT,
		s.getAccountDataFT,
		s.getAccountDataWallet,
	}
	for _, getter := range getters {
		getter(ctx, acc, others, interfaces, data)
	}

	if data.Errors != nil {
		log.Warn().Str("address", acc.Address.Base64()).Strs("errors", data.Errors).Msg("parse account data")
	}

	return data, nil
}
