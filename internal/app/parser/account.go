package parser

import (
	"bytes"
	"context"
	"fmt"

	"github.com/pkg/errors"
	"github.com/xssnick/tonutils-go/address"
	"github.com/xssnick/tonutils-go/tlb"
	"github.com/xssnick/tonutils-go/ton/wallet"
	"github.com/xssnick/tonutils-go/tvm/cell"

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

		if iface.Code != nil {
			code, err := cell.FromBOC(iface.Code)
			if err != nil {
				return nil, errors.Wrap(err, "parse contract interface code")
			}
			if acc.Code != nil && bytes.Equal(acc.Code.Hash(), code.Hash()) {
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
		ret.StateHash = acc.State.StateHash
		if acc.State.StateInit != nil {
			ret.Depth = acc.State.StateInit.Depth
			if acc.State.StateInit.TickTock != nil {
				ret.Tick = acc.State.StateInit.TickTock.Tick
				ret.Tock = acc.State.StateInit.TickTock.Tock
			}
		}
	}
	if acc.Data != nil {
		ret.Data = acc.Data.ToBOC()
		ret.DataHash = acc.Data.Hash()
	}
	if acc.Code != nil {
		ret.Code = acc.Code.ToBOC()
		ret.CodeHash = acc.Data.Hash()
	}
	ret.LastTxLT = acc.LastTxLT
	ret.LastTxHash = acc.LastTxHash

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
	var unknown int

	data := new(core.AccountData)
	data.Address = acc.Address
	data.DataHash = acc.DataHash

	for _, t := range acc.Types {
		switch t {
		case core.NFTCollection, core.NFTItem, core.NFTItemSBT:
			// TODO: error "contract exit code: 11"
			err := s.getAccountDataNFT(ctx, master, acc, data)
			if err != nil {
				return nil, errors.Wrap(err, "get nft data")
			}
		default:
			unknown++
		}
	}

	if unknown == len(acc.Types) {
		return data, errors.Wrap(core.ErrNotAvailable, "unknown contract")
	}
	return data, nil
}
