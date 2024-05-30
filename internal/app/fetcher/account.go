package fetcher

import (
	"context"
	"fmt"
	"sync"
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
	defer core.Timer(time.Now(), "getLastSeenAccountState(%s, %d)", a.String(), lastLT)

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

func (s *Service) makeGetOtherAccountFunc(master *ton.BlockIDExt, lastLT uint64) func(ctx context.Context, a addr.Address) (*core.AccountState, error) {
	getOtherAccountFunc := func(ctx context.Context, a addr.Address) (*core.AccountState, error) {
		defer core.Timer(time.Now(), "getOtherAccount(%s, %d)", a.String(), lastLT)

		itemStateID := core.AccountStateID{Address: a, LastTxLT: lastLT}

		// first attempt is to look into LRU cache, if minter was already fetched for the given id
		if m, ok := s.minterStatesCache.Get(itemStateID); ok {
			return m, nil
		}

		s.minterStatesCacheLocksMx.Lock()
		lock, exists := s.minterStatesCacheLocks.Get(itemStateID)
		if !exists {
			lock = &sync.Once{}
			s.minterStatesCacheLocks.Put(itemStateID, lock)
		}
		s.minterStatesCacheLocksMx.Unlock()

		lock.Do(func() {
			// second attempt is to look for the latest account state in the database
			acc, err := s.getLastSeenAccountState(ctx, a, lastLT)
			if err == nil {
				s.minterStatesCache.Put(itemStateID, acc)
				return
			}
			lvl := log.Warn()
			if errors.Is(err, core.ErrNotFound) || errors.Is(err, core.ErrInvalidArg) {
				lvl = log.Debug()
			}
			lvl.Err(err).Str("addr", a.Base64()).Msg("get latest other account state")

			// third attempt is to get needed contract state from the node
			raw, err := s.API.GetAccount(ctx, master, a.MustToTonutils())
			if err != nil {
				log.Error().Err(err).Str("address", a.Base64()).Msg("cannot get account state")
				return
			}

			s.minterStatesCache.Put(itemStateID, MapAccount(nil, raw))
		})

		if m, ok := s.minterStatesCache.Get(itemStateID); ok {
			return m, nil
		}

		return nil, fmt.Errorf("cannot get account state for (%s, %d)", itemStateID.Address.Base64(), itemStateID.LastTxLT)
	}
	return getOtherAccountFunc
}

func (s *Service) getAccount(ctx context.Context, master, b *ton.BlockIDExt, a addr.Address) (*core.AccountState, error) {
	if core.SkipAddress(a) {
		return nil, errors.Wrap(core.ErrNotFound, "skip account")
	}

	stateID := core.AccountBlockStateID{Address: a, Workchain: b.Workchain, Shard: b.Shard, BlockSeqNo: b.SeqNo}

	res, ok := s.accBlockStatesCache.Get(stateID)
	if ok && res.err == nil {
		return res.acc, nil
	}

	s.accBlockStatesCacheLocksMx.Lock()
	lock, exists := s.accBlockStatesCacheLocks.Get(stateID)
	if !exists || res.err != nil {
		lock = &sync.Once{}
		s.accBlockStatesCacheLocks.Put(stateID, lock)
	}
	s.accBlockStatesCacheLocksMx.Unlock()

	lock.Do(func() {
		defer core.Timer(time.Now(), "getAccount(%d, %d, %d, %s)", b.Workchain, b.Shard, b.SeqNo, a.String())

		var (
			acc *core.AccountState
			err error
		)
		defer func() { s.accBlockStatesCache.Put(stateID, getAccountRes{acc: acc, err: err}) }()

		raw, err := s.API.GetAccount(ctx, b, a.MustToTonutils())
		if err != nil {
			err = errors.Wrapf(err, "get account")
			return
		}

		acc = MapAccount(b, raw)

		if raw.Code != nil { //nolint:nestif // getting get-method hashes from the library
			libs, getErr := s.getAccountLibraries(ctx, raw)
			if getErr != nil {
				err = errors.Wrapf(getErr, "get account libraries")
				return
			}
			if libs != nil {
				acc.Libraries = libs.ToBOC()
			}

			if raw.Code.GetType() == cell.LibraryCellType {
				hash, getErr := getLibraryHash(raw.Code)
				if getErr != nil {
					err = errors.Wrap(getErr, "get library hash")
					return
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
			err = errors.Wrap(core.ErrNotFound, "account does not exists")
			return
		}

		// sometimes, to parse the full account data we need to get other contracts states
		// for example, to get nft item data
		getOtherAccount := s.makeGetOtherAccountFunc(master, acc.LastTxLT)

		err = s.Parser.ParseAccountData(ctx, acc, getOtherAccount)
		if err != nil && !errors.Is(err, app.ErrImpossibleParsing) {
			err = errors.Wrapf(err, "parse account data (%s)", acc.Address.String())
			return
		}

		err = nil
	})

	res, ok = s.accBlockStatesCache.Get(stateID)
	if !ok {
		panic(fmt.Errorf("cannot get %s parsed account result on (%d, %d, %d)", a.String(), b.Workchain, b.Shard, b.SeqNo))
	}

	return res.acc, res.err
}
