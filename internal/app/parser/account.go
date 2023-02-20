package parser

import (
	"context"

	"github.com/pkg/errors"
	"github.com/xssnick/tonutils-go/address"
	"github.com/xssnick/tonutils-go/tlb"

	"github.com/iam047801/tonidx/internal/core"
)

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

	types, err := s.abiRepo.DetermineInterfaces(ctx, acc)
	if err != nil {
		return nil, errors.Wrap(err, "get contract interfaces")
	}

	for _, t := range types {
		switch t {
		case core.NFTCollection, core.NFTItem, core.NFTItemSBT:
			// TODO: error "contract exit code: 11"
			err := s.getAccountDataNFT(ctx, b, acc, types, data)
			if err != nil {
				return nil, errors.Wrap(err, "get nft data")
			}
		default:
			unknown++
		}
	}

	if unknown == len(types) {
		return data, errors.Wrap(core.ErrNotAvailable, "unknown contract")
	}
	return data, nil
}
