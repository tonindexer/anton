package indexer

import (
	"context"
	"fmt"
	"time"

	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
	"github.com/xssnick/tonutils-go/address"
	"github.com/xssnick/tonutils-go/tlb"

	"github.com/iam047801/tonidx/internal/core"
)

func (s *Service) skipAccounts(_ *tlb.BlockInfo, addr *address.Address) bool {
	switch addr.String() {
	case "Ef8zMzMzMzMzMzMzMzMzMzMzMzMzMzMzMzMzMzMzMzMzM0vF": // skip elector contract
		return true

	case "Ef80UXx731GHxVr0-LYf3DIViMerdo3uJLAG3ykQZFjXz2kW": // skip log tests contract
		return true

	case "Ef9VVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVbxn": // skip config contract
		return true

	default:
		return false
	}
}

func (s *Service) processAccount(ctx context.Context, b *tlb.BlockInfo, addr *address.Address) (*core.AccountState, *core.AccountData, error) {
	if s.skipAccounts(b, addr) {
		return nil, nil, nil
	}

	defer timeTrack(time.Now(), fmt.Sprintf("processAccount(%d, %d, %s)", b.Workchain, b.SeqNo, addr.String()))

	raw, err := s.api.GetAccount(ctx, b, addr)
	if err != nil {
		return nil, nil, errors.Wrapf(err, "get account")
	}

	acc := mapAccount(raw)
	if acc.Status == core.NonExist {
		return acc, nil, nil
	}

	accTypes, err := s.parser.DetermineInterfaces(ctx, raw)
	if err != nil {
		return nil, nil, errors.Wrapf(err, "determine contract interfaces")
	}
	for _, t := range accTypes { // TODO: remove this
		acc.Types = append(acc.Types, string(t))
	}

	data, err := s.parser.ParseAccountData(ctx, b, raw)
	if err != nil && !errors.Is(err, core.ErrNotAvailable) {
		log.Error().Err(err).Str("addr", addr.String()).Msg("parse account data")
		// return nil, nil, errors.Wrapf(err, "get account (%s)", addr.String())
	}

	return acc, data, nil
}

func (s *Service) processTxAccounts(
	ctx context.Context, b *tlb.BlockInfo,
	transactions []*core.Transaction,
) (accounts map[string]*core.AccountState, accountsData map[string]*core.AccountData, err error) {
	accounts = make(map[string]*core.AccountState)
	accountsData = make(map[string]*core.AccountData)

	for _, tx := range transactions {
		addr := address.MustParseAddr(tx.Address)

		acc, data, err := s.processAccount(ctx, b, addr)
		if err != nil {
			return nil, nil, errors.Wrapf(err, "process account (%s)", addr.String())
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
