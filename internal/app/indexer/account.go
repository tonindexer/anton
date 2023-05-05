package indexer

import (
	"context"
	"fmt"
	"time"

	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"

	"github.com/xssnick/tonutils-go/address"
	"github.com/xssnick/tonutils-go/ton"

	"github.com/tonindexer/anton/addr"
	"github.com/tonindexer/anton/internal/app"
	"github.com/tonindexer/anton/internal/core"
	"github.com/tonindexer/anton/internal/core/filter"
)

func (s *Service) skipAccounts(_ *ton.BlockIDExt, a *address.Address) bool {
	switch a.String() {
	case "Ef8AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAADAU": // skip system contract
		return true
	case "Ef8zMzMzMzMzMzMzMzMzMzMzMzMzMzMzMzMzMzMzMzMzM0vF": // skip elector contract
		return true
	case "Ef80UXx731GHxVr0-LYf3DIViMerdo3uJLAG3ykQZFjXz2kW": // skip log tests contract
		return true
	case "Ef9VVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVbxn": // skip config contract
		return true
	case "EQAHI1vGuw7d4WG-CtfDrWqEPNtmUuKjKFEFeJmZaqqfWTvW": // skip BSC Bridge Collector
		return true
	case "EQCuzvIOXLjH2tv35gY4tzhIvXCqZWDuK9kUhFGXKLImgxT5": // skip ETH Bridge Collector
		return true
	case "EQA3Hv0IobJQMf-6YZsFQt8ZzuTGEy2Nfngby7_rMDnstiVe": // skip Sedov's contract
		return true
	case "EQA2u5Z5Fn59EUvTI-TIrX8PIGKQzNj3qLixdCPPujfJleXC",
		"EQA2Pnxp0rMB9L6SU2z1VqfMIFIfutiTjQWFEXnwa_zPh0P3",
		"EQDhIloDu1FWY9WFAgQDgw0RjuT5bLkf15Rmd5LCG3-0hyoe": // skip strange heavy testnet address
		return true
	default:
		return false
	}
}

func (s *Service) processAccount(ctx context.Context, b *ton.BlockIDExt, tx *core.Transaction) (*core.AccountState, error) {
	a := address.MustParseAddr(tx.Address.Base64())

	if s.skipAccounts(b, a) {
		return nil, errors.Wrap(core.ErrNotFound, "skip account")
	}
	log.Debug().Str("addr", a.String()).Int32("workchain", b.Workchain).Uint32("seq", b.SeqNo).Msg("getting account state")

	defer timeTrack(time.Now(), fmt.Sprintf("processAccount(%d, %d, %s)", b.Workchain, b.SeqNo, a.String()))

	raw, err := s.api.GetAccount(ctx, b, a)
	if err != nil {
		return nil, errors.Wrapf(err, "get account")
	}

	acc := mapAccount(raw)
	acc.UpdatedAt = tx.CreatedAt
	if acc.Status != core.Active {
		return nil, errors.Wrap(core.ErrNotFound, "account is not active")
	}

	return acc, nil
}

func (s *Service) processTxAccounts(
	ctx context.Context, b *ton.BlockIDExt, transactions []*core.Transaction,
) (
	map[addr.Address]*core.AccountState, map[addr.Address]*core.AccountData, error,
) {
	accounts := make(map[addr.Address]*core.AccountState)
	for _, tx := range transactions {
		acc, err := s.processAccount(ctx, b, tx)
		if err != nil && !errors.Is(err, core.ErrNotFound) {
			return nil, nil, errors.Wrapf(err, "process account (%s)", tx.Address.Base64())
		}
		if acc != nil {
			accounts[acc.Address] = acc
		}
	}

	// sometimes, to fetch the full account data we need to get other contracts states
	getOtherAccount := func(ctx context.Context, a *addr.Address) (*core.AccountState, error) {
		// first attempt is to look for an account in this given block
		acc, ok := accounts[*a]
		if ok {
			return acc, nil
		}

		// second attempt is to look for an account states in our database
		got, err := s.accountRepo.FilterAccounts(ctx, &filter.AccountsReq{
			Addresses:   []*addr.Address{a},
			LatestState: true,
		})
		if err == nil && len(got.Rows) > 0 {
			return got.Rows[0], nil
		}

		// final attempt is take an account from the liteserver
		log.Warn().Str("address", a.String()).Msg("account state is not found locally")

		raw, err := s.api.GetAccount(ctx, b, a.MustToTonutils())
		if err == nil {
			return mapAccount(raw), nil
		}

		return nil, errors.Wrapf(core.ErrNotFound, "cannot find %s account state", a.Base64())
	}

	accountsData := make(map[addr.Address]*core.AccountData)
	for _, acc := range accounts {
		data, err := s.parser.ParseAccountData(ctx, acc, getOtherAccount)
		if err != nil && !errors.Is(err, app.ErrImpossibleParsing) {
			return nil, nil, errors.Wrapf(err, "parse account data (%s)", acc.Address.String())
		}
		if data != nil {
			accountsData[data.Address] = data
		}
	}

	return accounts, accountsData, nil
}
