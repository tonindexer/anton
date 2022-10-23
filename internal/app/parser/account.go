package parser

import (
	"context"
	"fmt"

	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
	"github.com/xssnick/tonutils-go/address"
	"github.com/xssnick/tonutils-go/tlb"
	"github.com/xssnick/tonutils-go/ton/wallet"

	"github.com/iam047801/tonidx/internal/core"
)

func (s *Service) ContractInterfaces(ctx context.Context, acc *tlb.Account) ([]core.ContractType, error) {
	var ret []core.ContractType

	version := wallet.GetWalletVersion(acc)
	if version != wallet.Unknown {
		ret = append(ret,
			core.ContractType(fmt.Sprintf("wallet_%s", version.String())))
	}

	ifaces, err := s.accountRepo.GetContractInterfaces(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "get contract interfaces")
	}

	for _, iface := range ifaces {
		if iface.Address != "" {
			addr, err := address.ParseAddr(iface.Address)
			if err != nil {
				return nil, errors.Wrap(err, "parse contract interface address")
			}
			if acc.State != nil && addr.String() == acc.State.Address.String() {
				ret = append(ret, iface.Name)
			}
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

func (s *Service) ParseAccount(ctx context.Context, master *tlb.BlockInfo, addr *address.Address) (*core.Account, error) {
	ret := new(core.Account)

	acc, err := s.api.GetAccount(ctx, master, addr)
	if err != nil {
		return nil, errors.Wrapf(err, "get account (%s)", addr.String())
	}

	ret.Address = addr.String()
	ret.IsActive = acc.IsActive
	if acc.State != nil {
		ret.Status = core.AccountStatus(acc.State.Status)
		ret.Balance = acc.State.Balance.NanoTON().Uint64()
	}
	if acc.Data != nil {
		ret.Data = acc.Data.ToBOC()
		ret.DataHash = acc.Data.Hash()
	}
	if acc.Code != nil {
		ret.Code = acc.Code.ToBOC()
	}
	ret.LastTxHash = acc.LastTxHash
	ret.LastTxLT = acc.LastTxLT

	ifaces, err := s.ContractInterfaces(ctx, acc)
	if err != nil {
		return nil, errors.Wrap(err, "get contract interfaces")
	}
	for _, iface := range ifaces {
		ret.Types = append(ret.Types, string(iface))
	}

	return ret, nil
}

func (s *Service) ParseAccountData(ctx context.Context, master *tlb.BlockInfo, acc *core.Account) (*core.AccountData, error) {
	data := new(core.AccountData)
	data.Address = acc.Address
	data.DataHash = acc.DataHash

	for _, t := range acc.Types {
		switch t {
		case core.NFTCollection, core.NFTItem, core.NFTItemSBT, core.NFTItemEditable, core.NFTItemEditableSBT:
			// TODO: error "contract exit code: 11"
			err := s.getAccountDataNFT(ctx, master, acc, data)
			if err != nil {
				log.Error().Err(err).Str("address", acc.Address).Str("type", t).Msg("get nft data")
			}
		}
	}

	return data, nil
}
