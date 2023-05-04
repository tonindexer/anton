package indexer

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"

	"github.com/xssnick/tonutils-go/address"
	"github.com/xssnick/tonutils-go/ton"

	"github.com/tonindexer/anton/addr"
	"github.com/tonindexer/anton/internal/app"
	"github.com/tonindexer/anton/internal/core"
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
	case "EQA2u5Z5Fn59EUvTI-TIrX8PIGKQzNj3qLixdCPPujfJleXC",
		"EQA2Pnxp0rMB9L6SU2z1VqfMIFIfutiTjQWFEXnwa_zPh0P3",
		"EQDhIloDu1FWY9WFAgQDgw0RjuT5bLkf15Rmd5LCG3-0hyoe": // skip strange heavy testnet address
		return true
	default:
		return false
	}
}

func (s *Service) processAccount(ctx context.Context, b *ton.BlockIDExt, tx *core.Transaction) (*core.AccountState, *core.AccountData, error) {
	a := address.MustParseAddr(tx.Address.Base64())

	if s.skipAccounts(b, a) {
		return nil, nil, nil
	}
	log.Debug().Str("addr", a.String()).Int32("workchain", b.Workchain).Uint32("seq", b.SeqNo).Msg("getting account state")

	defer timeTrack(time.Now(), fmt.Sprintf("processAccount(%d, %d, %s)", b.Workchain, b.SeqNo, a.String()))

	raw, err := s.api.GetAccount(ctx, b, a)
	if err != nil {
		if strings.Contains(err.Error(), "extra currency info is not supported for AccountStorage") { // tonutils-go v1.6.2
			// skip accounts with extra currency info
			return nil, nil, nil
		}
		return nil, nil, errors.Wrapf(err, "get account")
	}

	acc := mapAccount(raw)
	acc.UpdatedAt = tx.CreatedAt
	if acc.Status == core.NonExist {
		return nil, nil, nil
	}

	data, err := s.parser.ParseAccountData(ctx, acc)
	if err != nil && !errors.Is(err, app.ErrImpossibleParsing) {
		return nil, nil, errors.Wrapf(err, "parse account (%s)", a.String())
	}

	return acc, data, nil
}

func (s *Service) processTxAccounts(
	ctx context.Context, b *ton.BlockIDExt,
	transactions []*core.Transaction,
) (accounts map[addr.Address]*core.AccountState, accountsData map[addr.Address]*core.AccountData, err error) {
	accounts = make(map[addr.Address]*core.AccountState)
	accountsData = make(map[addr.Address]*core.AccountData)

	for _, tx := range transactions {
		acc, data, err := s.processAccount(ctx, b, tx)
		if err != nil {
			return nil, nil, errors.Wrapf(err, "process account (%s)", tx.Address.Base64())
		}

		if acc != nil {
			accounts[acc.Address] = acc
		}
		if data != nil {
			accountsData[data.Address] = data
		}
	}

	return accounts, accountsData, nil
}
