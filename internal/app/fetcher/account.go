package fetcher

import (
	"context"
	"time"

	"github.com/pkg/errors"
	"github.com/xssnick/tonutils-go/ton"

	"github.com/tonindexer/anton/addr"
	"github.com/tonindexer/anton/internal/app"
	"github.com/tonindexer/anton/internal/core"
)

func (s *Service) getAccount(ctx context.Context, b *ton.BlockIDExt, a addr.Address) (*core.AccountState, error) {
	acc, ok := s.accounts.get(b, a)
	if ok {
		return acc, nil
	}

	if core.SkipAddress(a) {
		return nil, errors.Wrap(core.ErrNotFound, "skip account")
	}

	defer app.TimeTrack(time.Now(), "getAccount(%d, %d, %s)", b.Workchain, b.SeqNo, a.String())

	raw, err := s.API.GetAccount(ctx, b, a.MustToTonutils())
	if err != nil {
		return nil, errors.Wrapf(err, "get account")
	}

	acc = MapAccount(b, raw)
	if acc.Status == core.NonExist {
		return nil, errors.Wrap(core.ErrNotFound, "account does not exists")
	}

	// sometimes, to parse the full account data we need to get other contracts states
	// for example, to get nft item data
	getOtherAccount := func(ctx context.Context, a addr.Address) (*core.AccountState, error) {
		// first attempt is to look for an account in this given block
		acc, ok := s.accounts.get(b, a)
		if ok {
			return acc, nil
		}
		raw, err := s.API.GetAccount(ctx, b, a.MustToTonutils())
		if err == nil {
			return MapAccount(b, raw), nil
		}
		return nil, errors.Wrapf(core.ErrNotFound, "cannot find %s account state", a.Base64())
	}

	err = s.Parser.ParseAccountData(ctx, acc, getOtherAccount)
	if err != nil && !errors.Is(err, app.ErrImpossibleParsing) {
		return nil, errors.Wrapf(err, "parse account data (%s)", acc.Address.String())
	}

	s.accounts.set(b, acc)
	return acc, nil
}
