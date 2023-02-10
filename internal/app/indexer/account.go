package indexer

import (
	"context"

	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
	"github.com/xssnick/tonutils-go/address"
	"github.com/xssnick/tonutils-go/tlb"

	"github.com/iam047801/tonidx/internal/core"
)

func (s *Service) processTxAccounts(
	ctx context.Context, master *tlb.BlockInfo,
	transactions []*core.Transaction,
) (accounts []*core.AccountState, accountsData []*core.AccountData, err error) {
	for _, tx := range transactions {
		addr := address.MustParseAddr(tx.Address)

		raw, err := s.api.GetAccount(ctx, master, addr)
		if err != nil {
			return nil, nil, errors.Wrapf(err, "get account (%s)", addr.String())
		}
		acc := mapAccount(raw)
		acc.Raw = raw

		accTypes, err := s.abiRepo.DetermineContractInterfaces(ctx, raw)
		if err != nil {
			return nil, nil, errors.Wrapf(err, "determine contract interfaces")
		}
		for _, t := range accTypes {
			acc.Types = append(acc.Types, string(t))
		}

		accounts = append(accounts, acc)

		data, err := s.parser.ParseAccountData(ctx, master, raw)
		if err != nil && !errors.Is(err, core.ErrNotAvailable) {
			log.Error().Err(err).Str("addr", tx.Address).Msg("parse account data")
			continue
		}
		if err == nil {
			accountsData = append(accountsData, data)
		}
	}

	return accounts, accountsData, nil
}
