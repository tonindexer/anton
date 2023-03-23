package indexer

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
	"github.com/uptrace/bun"
	"github.com/xssnick/tonutils-go/ton"

	"github.com/iam047801/tonidx/internal/core"
)

func (s *Service) getNotSeenShards(ctx context.Context, shard *ton.BlockIDExt) (ret []*ton.BlockIDExt, err error) {
	defer timeTrack(time.Now(), fmt.Sprintf("getNotSeenShards(%d, %d)", shard.Workchain, shard.SeqNo))

	if no, ok := s.shardLastSeqno[getShardID(shard)]; ok && no == shard.SeqNo {
		return nil, nil
	}

	b, err := s.api.GetBlockData(ctx, shard) // TODO: save this block data to a database
	if err != nil {
		return nil, fmt.Errorf("get block data: %w", err)
	}

	parents, err := b.BlockInfo.GetParentBlocks()
	if err != nil {
		return nil, fmt.Errorf("get parent blocks (%d:%x:%d): %w", shard.Workchain, uint64(shard.Shard), shard.Shard, err)
	}

	for _, parent := range parents {
		ext, err := s.getNotSeenShards(ctx, parent)
		if err != nil {
			return nil, err
		}
		ret = append(ret, ext...)
	}

	ret = append(ret, shard)
	return ret, nil
}

func (s *Service) processShards(ctx context.Context, tx bun.Tx, master *ton.BlockIDExt) error {
	var dbShards []*core.Block

	currentShards, err := s.api.GetBlockShardsInfo(ctx, master)
	if err != nil {
		return errors.Wrap(err, "get masterchain shards info")
	}
	if len(currentShards) == 0 {
		log.Debug().Uint32("master_seq", master.SeqNo).Msg("master block without shards")
		return nil
	}

	// shards in master block may have holes, e.g. shard seqno 2756461, then 2756463, and no 2756462 in master chain
	// thus we need to scan a bit back in case of discovering a hole, till last seen, to fill the misses.
	var newShards []*ton.BlockIDExt
	for _, shard := range currentShards {
		notSeen, err := s.getNotSeenShards(ctx, shard)
		if err != nil {
			return errors.Wrap(err, "get not seen shards")
		}
		newShards = append(newShards, notSeen...)
	}
	if len(newShards) == 0 {
		return nil
	}

	for _, shard := range newShards {
		log.Debug().
			Uint32("master_seq", master.SeqNo).
			Int32("shard_workchain", shard.Workchain).Uint32("shard_seq", shard.SeqNo).
			Msg("new shard block")

		if err := s.processBlockTransactions(ctx, tx, shard); err != nil {
			return err
		}

		dbShards = append(dbShards, &core.Block{
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

	if err := s.blockRepo.AddBlocks(ctx, tx, dbShards); err != nil {
		return errors.Wrap(err, "add shard block")
	}

	for _, shard := range currentShards {
		s.shardLastSeqno[getShardID(shard)] = shard.SeqNo
	}

	return nil
}

func (s *Service) processMaster(ctx context.Context, master *ton.BlockIDExt) error {
	insertTx, err := s.cfg.DB.PG.Begin()
	if err != nil {
		return errors.Wrap(err, "cannot begin db tx")
	}
	defer func() {
		_ = insertTx.Rollback()
	}()

	if err := s.processShards(ctx, insertTx, master); err != nil {
		return errors.Wrap(err, "cannot process shards")
	}

	if err := s.processBlockTransactions(ctx, insertTx, master); err != nil {
		return errors.Wrap(err, "cannot process masterchain block transactions")
	}

	dbMaster := &core.Block{
		Workchain: master.Workchain,
		Shard:     master.Shard,
		SeqNo:     master.SeqNo,
		RootHash:  master.RootHash,
		FileHash:  master.FileHash,
	}
	if err := s.blockRepo.AddBlocks(ctx, insertTx, []*core.Block{dbMaster}); err != nil {
		return errors.Wrap(err, "cannot add master block")
	}

	if err := insertTx.Commit(); err != nil {
		return errors.Wrap(err, "cannot commit db tx")
	}

	return nil
}

func (s *Service) fetchBlocksLoop(workchain int32, shard int64, fromBlock uint32) {
	defer s.wg.Done()

	log.Info().Int32("workchain", workchain).Int64("shard", shard).Uint32("from_block", fromBlock).Msg("starting")

	for seq := fromBlock; s.running(); time.Sleep(s.cfg.FetchBlockPeriod) {
		ctx := context.Background()

		master, err := s.api.LookupBlock(ctx, workchain, shard, seq)
		if errors.Is(err, ton.ErrBlockNotFound) || strings.Contains(err.Error(), "block is not applied") {
			continue
		}
		if err != nil {
			log.Error().Err(err).Uint32("master_seq", seq).Msg("cannot lookup masterchain block")
			continue
		}

		if err := s.processMaster(ctx, master); err != nil {
			if strings.Contains(err.Error(), "block is not applied") {
				continue
			}
			log.Error().Err(err).Uint32("master_seq", seq).Msg("cannot process masterchain block")
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
