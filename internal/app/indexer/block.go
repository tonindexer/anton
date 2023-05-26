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
	var insertBlocks []*core.Block

	insertTx, err := s.DB.PG.Begin()
	if err != nil {
		return errors.Wrap(err, "cannot begin db tx")
	}
	defer func() {
		_ = insertTx.Rollback()
	}()

	if err := s.processBlockTransactions(ctx, insertTx, master); err != nil {
		return errors.Wrap(err, "cannot process masterchain block transactions")
	}

	for _, shard := range shards {
		log.Debug().
			Uint32("master_seq", master.SeqNo).
			Int32("shard_workchain", shard.Workchain).Uint32("shard_seq", shard.SeqNo).
			Msg("new shard block")

		if err := s.processBlockTransactions(ctx, insertTx, shard); err != nil {
			return err
		}

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

	if err := s.blockRepo.AddBlocks(ctx, insertTx, insertBlocks); err != nil {
		return errors.Wrap(err, "add shard block")
	}

	if err := insertTx.Commit(); err != nil {
		return errors.Wrap(err, "cannot commit db tx")
	}

	return nil
}

func (s *Service) processMasterSeqNo(ctx context.Context, seq uint32) bool {
	master, shards, err := s.Fetcher.FetchBlocksInMaster(ctx, seq)
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
