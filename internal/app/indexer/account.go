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
) (accounts map[string]*core.AccountState, accountsData []*core.AccountData, err error) {
	accounts = make(map[string]*core.AccountState)

	for _, tx := range transactions {
		addr := address.MustParseAddr(tx.Address)

		raw, err := s.api.GetAccount(ctx, master, addr)
		if err != nil {
			return nil, nil, errors.Wrapf(err, "get account (%s)", addr.String())
		}
		acc := mapAccount(raw)

		accTypes, err := s.abiRepo.DetermineContractInterfaces(ctx, raw)
		if err != nil {
			return nil, nil, errors.Wrapf(err, "determine contract interfaces")
		}
		for _, t := range accTypes {
			acc.Types = append(acc.Types, string(t))
		}

		if st, ok := accounts[acc.Address]; !ok || st.LastTxLT <= acc.LastTxLT {
			accounts[acc.Address] = acc
		} else {
			continue
		}

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
