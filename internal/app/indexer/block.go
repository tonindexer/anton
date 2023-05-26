package indexer

import (
	"context"
	"strings"
	"time"

	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
	"github.com/xssnick/tonutils-go/ton"

	"github.com/tonindexer/anton/internal/core"
)

func (s *Service) processMaster(ctx context.Context, master *ton.BlockIDExt, shards []*ton.BlockIDExt) error {
	var (
		insertBlocks   []*core.Block
		insertTx       []*core.Transaction
		insertAccounts []*core.AccountState
		insertMsg      []*core.Message
	)

	dbTx, err := s.DB.PG.Begin()
	if err != nil {
		return errors.Wrap(err, "cannot begin db tx")
	}
	defer func() {
		_ = dbTx.Rollback()
	}()

	insertTx, err = s.Fetcher.BlockTransactions(ctx, master)
	if err != nil {
		return errors.Wrap(err, "get block transactions")
	}

	for _, shard := range shards {
		log.Debug().
			Uint32("master_seq", master.SeqNo).
			Int32("shard_workchain", shard.Workchain).Uint32("shard_seq", shard.SeqNo).
			Msg("new shard block")

		shardTx, err := s.Fetcher.BlockTransactions(ctx, master)
		if err != nil {
			return errors.Wrap(err, "get block transactions")
		}

		insertTx = append(insertTx, shardTx...)
		insertBlocks = append(insertBlocks, &core.Block{
			Workchain: shard.Workchain,
			Shard:     shard.Shard,
			SeqNo:     shard.SeqNo,
			RootHash:  shard.RootHash,
			FileHash:  shard.FileHash,
			MasterID: &core.BlockID{
				Workchain: master.Workchain,
				Shard:     master.Shard,
				SeqNo:     master.SeqNo,
			},
		})
	}

	insertBlocks = append(insertBlocks, &core.Block{
		Workchain: master.Workchain,
		Shard:     master.Shard,
		SeqNo:     master.SeqNo,
		FileHash:  master.FileHash,
		RootHash:  master.RootHash,
	})

	for i := range insertTx {
		insertAccounts = append(insertAccounts, insertTx[i].Account)
	}
	for i := range insertTx {
		insertMsg = append(insertMsg, insertTx[i].InMsg)
		insertMsg = append(insertMsg, insertTx[i].OutMsg...)
	}

	if err := s.accountRepo.AddAccountStates(ctx, dbTx, insertAccounts); err != nil {
		return errors.Wrap(err, "add account states")
	}
	if err := s.msgRepo.AddMessages(ctx, dbTx, insertMsg); err != nil {
		return errors.Wrap(err, "add messages")
	}
	if err := s.txRepo.AddTransactions(ctx, dbTx, insertTx); err != nil {
		return errors.Wrap(err, "add transactions")
	}
	if err := s.blockRepo.AddBlocks(ctx, dbTx, insertBlocks); err != nil {
		return errors.Wrap(err, "add shard block")
	}

	if err := dbTx.Commit(); err != nil {
		return errors.Wrap(err, "cannot commit db tx")
	}

	return nil
}

func (s *Service) processMasterSeqNo(ctx context.Context, seq uint32) bool {
	master, shards, err := s.Fetcher.UnseenBlocks(ctx, seq)
	if errors.Is(err, ton.ErrBlockNotFound) || (err != nil && strings.Contains(err.Error(), "block is not applied")) {
		return false
	}
	if err != nil {
		log.Error().Err(err).Uint32("master_seq", seq).Msg("cannot fetch masterchain blocks")
		return false
	}

	if err := s.processMaster(ctx, master, shards); err != nil {
		log.Error().Err(err).Uint32("master_seq", seq).Msg("cannot process masterchain block")
		return false
	}

	return true
}

func (s *Service) fetchBlocksLoop(ctx context.Context, fromBlock uint32) {
	defer s.wg.Done()

	for seq := fromBlock; s.running(); time.Sleep(s.FetchBlockPeriod) {
		if !s.processMasterSeqNo(ctx, seq) {
			continue
		}

		lvl := log.Debug()
		if seq%100 == 0 {
			lvl = log.Info()
		}
		lvl.Uint32("master_seq", seq).Msg("new masterchain block")

		seq++
	}
}
