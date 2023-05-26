package fetcher

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/pkg/errors"
	"github.com/xssnick/tonutils-go/ton"

	"github.com/tonindexer/anton/addr"
	"github.com/tonindexer/anton/internal/app"
	"github.com/tonindexer/anton/internal/core"
)

func (s *Service) skipAccounts(_ *ton.BlockIDExt, a addr.Address) bool {
	switch a.Base64() {
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

func (s *Service) getAccount(ctx context.Context, b *ton.BlockIDExt, a addr.Address) (*core.AccountState, error) {
	acc, ok := s.accounts.get(b, a)
	if ok {
		return acc, nil
	}

	if s.skipAccounts(b, a) {
		return nil, errors.Wrap(core.ErrNotFound, "skip account")
	}

	defer app.TimeTrack(time.Now(), fmt.Sprintf("getAccount(%d, %d, %s)", b.Workchain, b.SeqNo, a.String()))

	raw, err := s.API.GetAccount(ctx, b, a.MustToTonutils())
	if err != nil {
		return nil, errors.Wrapf(err, "get account")
	}

	acc = mapAccount(raw)
	if acc.Status != core.Active {
		return nil, errors.Wrap(core.ErrNotFound, "account is not active")
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
			return mapAccount(raw), nil
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

func (s *Service) getTxAccounts(ctx context.Context, b *ton.BlockIDExt, transactions []*core.Transaction) error {
	var wg sync.WaitGroup

	type ret struct {
		addr  addr.Address
		state *core.AccountState
		err   error
	}

	defer app.TimeTrack(time.Now(), fmt.Sprintf("getAccounts(%d, %d)", b.Workchain, b.SeqNo))

	ch := make(chan ret, len(transactions))

	wg.Add(len(transactions))

	for i := range transactions {
		go func(tx *core.Transaction) {
			defer wg.Done()
			account, err := s.getAccount(ctx, b, tx.Address)
			ch <- ret{addr: tx.Address, state: account, err: err}
		}(transactions[i])
	}

	go func() {
		wg.Wait()
		close(ch)
	}()

	results := make(map[addr.Address]*core.AccountState)
	for r := range ch {
		if errors.Is(r.err, core.ErrNotFound) {
			continue
		}
		if r.err != nil {
			return errors.Wrapf(r.err, "get %s account", r.addr.Base64())
		}
		results[r.addr] = r.state
	}

	for _, tx := range transactions {
		tx.Account = results[tx.Address]
		if tx.Account != nil {
			tx.Account.UpdatedAt = tx.CreatedAt
		}
	}

	return nil
}
