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
	ctx context.Context, b *tlb.BlockInfo,
	transactions []*core.Transaction,
) (accounts map[string]*core.AccountState, accountsData map[string]*core.AccountData, err error) {
	accounts = make(map[string]*core.AccountState)
	accountsData = make(map[string]*core.AccountData)

	for _, tx := range transactions {
		addr := address.MustParseAddr(tx.Address)

		raw, err := s.api.GetAccount(ctx, b, addr)
		if err != nil {
			return nil, nil, errors.Wrapf(err, "get account (%s)", addr.String())
		}

		acc := mapAccount(raw)
		if acc.Status == core.NonExist {
			continue
		}

		accTypes, err := s.parser.DetermineInterfaces(ctx, raw)
		if err != nil {
			return nil, nil, errors.Wrapf(err, "determine contract interfaces")
		}

		for _, t := range accTypes {
			acc.Types = append(acc.Types, string(t))
		}
		accounts[acc.Address] = acc

		data, err := s.parser.ParseAccountData(ctx, b, raw)
		if err != nil && !errors.Is(err, core.ErrNotAvailable) {
			log.Error().Err(err).Str("addr", tx.Address).Msg("parse account data")
			continue
		}

		accountsData[data.Address] = data
	}

	return accounts, accountsData, nil
}
