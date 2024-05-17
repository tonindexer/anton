package fetcher

import (
	"context"
	"time"

	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
	"github.com/xssnick/tonutils-go/ton"
	"github.com/xssnick/tonutils-go/tvm/cell"

	"github.com/tonindexer/anton/abi"
	"github.com/tonindexer/anton/addr"
	"github.com/tonindexer/anton/internal/app"
	"github.com/tonindexer/anton/internal/core"
	"github.com/tonindexer/anton/internal/core/filter"
)

func (s *Service) getLastSeenAccountState(ctx context.Context, a addr.Address, lastLT uint64) (*core.AccountState, error) {
	defer app.TimeTrack(time.Now(), "getLastSeenAccountState(%s, %d)", a.String(), lastLT)

	lastLT++

	accountReq := filter.AccountsReq{
		WithCodeData: true,
		Addresses:    []*addr.Address{&a},
		Order:        "DESC",
		AfterTxLT:    &lastLT,
		NoCount:      true,
		Limit:        1,
	}
	accountRes, err := s.AccountRepo.FilterAccounts(ctx, &accountReq)
	if err != nil {
		return nil, errors.Wrap(err, "filter accounts")
	}
	if len(accountRes.Rows) < 1 {
		return nil, errors.Wrap(core.ErrNotFound, "could not find needed account state")
	}

	return accountRes.Rows[0], nil
}

func (s *Service) makeGetOtherAccountFunc(master, b *ton.BlockIDExt, lastLT uint64) func(ctx context.Context, a addr.Address) (*core.AccountState, error) {
	getOtherAccountFunc := func(ctx context.Context, a addr.Address) (*core.AccountState, error) {
		// first attempt is to look for an account in this given block
		acc, ok := s.accounts.get(b, a)
		if ok {
			return acc, nil
		}

		// second attempt is to look for the latest account state in the database
		acc, err := s.getLastSeenAccountState(ctx, a, lastLT)
		if err == nil {
			return acc, nil
		}
		lvl := log.Warn()
		if errors.Is(err, core.ErrNotFound) || errors.Is(err, core.ErrInvalidArg) {
			lvl = log.Debug()
		}
		lvl.Err(err).Str("addr", a.Base64()).Msg("get latest other account state")

		// third attempt is to get needed contract state from the node
		raw, err := s.API.GetAccount(ctx, master, a.MustToTonutils())
		if err != nil {
			return nil, errors.Wrapf(err, "cannot get %s account state", a.Base64())
		}
		return MapAccount(b, raw), nil
	}
	return getOtherAccountFunc
}

func (s *Service) getAccount(ctx context.Context, master, b *ton.BlockIDExt, a addr.Address) (*core.AccountState, error) {
	acc, ok := s.accounts.get(b, a)
	if ok {
		return acc, nil
	}

	if core.SkipAddress(a) {
		return nil, errors.Wrap(core.ErrNotFound, "skip account")
	}

	defer app.TimeTrack(time.Now(), "getAccount(%d, %d, %d, %s)", b.Workchain, b.Shard, b.SeqNo, a.String())

	raw, err := s.API.GetAccount(ctx, b, a.MustToTonutils())
	if err != nil {
		return nil, errors.Wrapf(err, "get account")
	}

	acc = MapAccount(b, raw)

	if raw.Code != nil { //nolint:nestif // getting get method hashes from the library
		libs, err := s.getAccountLibraries(ctx, raw)
		if err != nil {
			return nil, errors.Wrapf(err, "get account libraries")
		}
		if libs != nil {
			acc.Libraries = libs.ToBOC()
		}

		if raw.Code.GetType() == cell.LibraryCellType {
			hash, err := getLibraryHash(raw.Code)
			if err != nil {
				return nil, errors.Wrap(err, "get library hash")
			}

			lib := s.libraries.get(hash)
			if lib != nil && lib.Lib != nil {
				acc.GetMethodHashes, _ = abi.GetMethodHashes(lib.Lib)
			}
		} else {
			acc.GetMethodHashes, _ = abi.GetMethodHashes(raw.Code)
		}
	}

	if acc.Status == core.NonExist {
		return nil, errors.Wrap(core.ErrNotFound, "account does not exists")
	}

	// sometimes, to parse the full account data we need to get other contracts states
	// for example, to get nft item data
	getOtherAccount := s.makeGetOtherAccountFunc(master, b, acc.LastTxLT)

	err = s.Parser.ParseAccountData(ctx, acc, getOtherAccount)
	if err != nil && !errors.Is(err, app.ErrImpossibleParsing) {
		return nil, errors.Wrapf(err, "parse account data (%s)", acc.Address.String())
	}

	s.accounts.set(b, acc)
	return acc, nil
}
