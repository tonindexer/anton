package rescan

import (
	"context"
	"reflect"
	"sync"
	"time"

	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"

	"github.com/tonindexer/anton/addr"
	"github.com/tonindexer/anton/internal/app"
	"github.com/tonindexer/anton/internal/core"
	"github.com/tonindexer/anton/internal/core/filter"
)

func (s *Service) getRecentAccountState(ctx context.Context, master, b core.BlockID, a addr.Address, afterBlock bool) (*core.AccountState, error) {
	defer app.TimeTrack(time.Now(), "getLastSeenAccountState(%d, %d, %d, %s)", b.Workchain, b.Shard, b.SeqNo, a.String())

	var boundBlock core.BlockID
	switch {
	case b.Workchain == int32(a.Workchain()):
		boundBlock = b
	case master.Workchain == int32(a.Workchain()):
		boundBlock = master
	default:
		return nil, errors.Wrapf(core.ErrInvalidArg, "address is in %d workchain, but the given block is from %d workchain", a.Workchain(), b.Workchain)
	}

	accountReq := filter.AccountsReq{
		Addresses: []*addr.Address{&a},
		Workchain: &boundBlock.Workchain,
		Shard:     &boundBlock.Shard,
		Order:     "DESC",
		Limit:     1,
	}
	if afterBlock {
		accountReq.BlockSeqNoBeq = &boundBlock.SeqNo
	} else {
		accountReq.BlockSeqNoLeq = &boundBlock.SeqNo
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

func (s *Service) rescanAccountsInBlock(master, b *core.Block) (updates []*core.AccountState) {
	for _, tx := range b.Transactions {
		if tx.Account == nil {
			continue
		}

		getOtherAccountFunc := func(ctx context.Context, a addr.Address) (*core.AccountState, error) {
			return s.getRecentAccountState(ctx, master.ID(), b.ID(), a, false)
		}

		update := *tx.Account

		err := s.Parser.ParseAccountData(context.Background(), &update, getOtherAccountFunc)
		if err != nil && !errors.Is(err, app.ErrImpossibleParsing) {
			log.Error().Err(err).Str("addr", update.Address.Base64()).Msg("parse account data")
			continue
		}

		if reflect.DeepEqual(tx.Account, &update) {
			continue
		}
		updates = append(updates, &update)
	}
	return updates
}

func (s *Service) rescanAccountsWorker(b *core.Block) (updates []*core.AccountState) {
	for _, shard := range b.Shards {
		upd := s.rescanAccountsInBlock(b, shard)
		updates = append(updates, upd...)
	}

	upd := s.rescanAccountsInBlock(b, b)
	updates = append(updates, upd...)

	return updates
}

func (s *Service) rescanAccounts(masterBlocks []*core.Block) (lastScanned uint32) {
	var (
		accountUpdates = make(chan []*core.AccountState, len(masterBlocks))
		scanWG         sync.WaitGroup
	)

	scanWG.Add(len(masterBlocks))

	for _, b := range masterBlocks {
		go func(master *core.Block) {
			defer scanWG.Done()
			accountUpdates <- s.rescanAccountsWorker(master)
		}(b)

		if b.SeqNo > lastScanned {
			lastScanned = b.SeqNo
		}
	}

	go func() {
		scanWG.Wait()
		close(accountUpdates)
	}()

	var allUpdates []*core.AccountState
	for upd := range accountUpdates {
		allUpdates = append(allUpdates, upd...)
	}

	if err := s.AccountRepo.UpdateAccountStates(context.Background(), allUpdates); err != nil {
		return 0
	}

	return lastScanned
}
