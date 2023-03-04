package indexer

import (
	"context"
	"fmt"
	"time"

	"github.com/pkg/errors"
	"github.com/xssnick/tonutils-go/address"
	"github.com/xssnick/tonutils-go/ton"

	"github.com/iam047801/tonidx/internal/addr"
	"github.com/iam047801/tonidx/internal/core"
)

func (s *Service) skipAccounts(_ *ton.BlockIDExt, a *address.Address) bool {
	switch a.String() {
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
	default:
		return false
	}
}

func (s *Service) processAccount(ctx context.Context, b *ton.BlockIDExt, a *address.Address) (*core.AccountState, *core.AccountData, error) {
	if s.skipAccounts(b, a) {
		return nil, nil, nil
	}

	defer timeTrack(time.Now(), fmt.Sprintf("processAccount(%d, %d, %s)", b.Workchain, b.SeqNo, a.String()))

	raw, err := s.api.GetAccount(ctx, b, a)
	if err != nil {
		return nil, nil, errors.Wrapf(err, "get account")
	}

	acc := mapAccount(raw)
	if acc.Status == core.NonExist {
		return nil, nil, nil
	}

	types, err := s.parser.DetermineInterfaces(ctx, acc)
	if err != nil {
		return nil, nil, errors.Wrapf(err, "determine contract interfaces")
	}

	data, err := s.parser.ParseAccountData(ctx, b, acc, types)
	if err != nil && !errors.Is(err, core.ErrNotAvailable) {
		return nil, nil, errors.Wrapf(err, "get account (%s)", a.String())
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
		a := address.MustParseAddr(tx.Address.Base64())

		acc, data, err := s.processAccount(ctx, b, a)
		if err != nil {
			return nil, nil, errors.Wrapf(err, "process account (%s)", a.String())
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
