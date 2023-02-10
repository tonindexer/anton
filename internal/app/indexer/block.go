package indexer

import (
	"context"
	"fmt"
	"time"

	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
	"github.com/xssnick/tonutils-go/tlb"
	"github.com/xssnick/tonutils-go/ton"

	"github.com/iam047801/tonidx/internal/core"
)

func (s *Service) getNotSeenShards(ctx context.Context, shard *tlb.BlockInfo) (ret []*tlb.BlockInfo, err error) {
	if no, ok := s.shardLastSeqno[getShardID(shard)]; ok && no == shard.SeqNo {
		return nil, nil
	}

	b, err := s.api.GetBlockData(ctx, shard)
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

func (s *Service) processShards(ctx context.Context, master *tlb.BlockInfo) ([]*core.Block, error) {
	var dbShards []*core.Block

	currentShards, err := s.api.GetBlockShardsInfo(ctx, master)
	if err != nil {
		return nil, errors.Wrap(err, "get masterchain shards info")
	}
	if len(currentShards) == 0 {
		log.Debug().Uint32("master_seq", master.SeqNo).Msg("master block without shards")
		return nil, nil
	}

	// shards in master block may have holes, e.g. shard seqno 2756461, then 2756463, and no 2756462 in master chain
	// thus we need to scan a bit back in case of discovering a hole, till last seen, to fill the misses.
	var newShards []*tlb.BlockInfo
	for _, shard := range currentShards {
		notSeen, err := s.getNotSeenShards(ctx, shard)
		if err != nil {
			return nil, errors.Wrap(err, "get not seen shards")
		}
		newShards = append(newShards, notSeen...)
	}

	for _, shard := range newShards {
		log.Debug().
			Uint32("master_seq", master.SeqNo).
			Int32("shard_workchain", shard.Workchain).Uint32("shard_seq", shard.SeqNo).
			Msg("new shard block")

		// // TODO: other block data
		// blockInfo, err := s.api.GetBlockData(ctx, shard)
		// if err != nil {
		// 	return nil, errors.Wrap(err, "get block data")
		// }

		if err := s.processBlockTransactions(ctx, master, shard); err != nil {
			return nil, err
		}

		dbShards = append(dbShards, &core.Block{
			BlockID: core.BlockID{
				Workchain: shard.Workchain,
				Shard:     shard.Shard,
				SeqNo:     shard.SeqNo,
			},
			RootHash:       shard.RootHash,
			FileHash:       shard.FileHash,
			MasterFileHash: master.FileHash,
		})
	}

	if err := s.blockRepo.AddBlocks(ctx, dbShards); err != nil {
		return nil, errors.Wrap(err, "add shard block")
	}

	for _, shard := range currentShards {
		s.shardLastSeqno[getShardID(shard)] = shard.SeqNo
	}

	return dbShards, nil
}

func (s *Service) fetchBlocksLoop(workchain int32, shard int64, fromBlock uint32) {
	defer s.wg.Done()

	log.Info().Int32("workchain", workchain).Int64("shard", shard).Uint32("from_block", fromBlock).Msg("starting")

	for seq := fromBlock; s.running(); time.Sleep(s.cfg.FetchBlockPeriod) {
		ctx := context.Background()

		master, err := s.api.LookupBlock(ctx, workchain, shard, seq)
		if errors.Is(err, ton.ErrBlockNotFound) {
			continue
		}
		if err != nil {
			log.Error().Err(err).Uint32("master_seq", seq).Msg("cannot lookup masterchain block")
			continue
		}

		lvl := log.Debug()
		if seq%100 == 0 {
			lvl = log.Info()
		}
		lvl.Uint32("master_seq", seq).Msg("new masterchain block")

		_, err = s.processShards(ctx, master)
		if err != nil {
			log.Error().Err(err).Uint32("master_seq", seq).Msg("cannot process shards")
			continue
		}

		if err := s.processBlockTransactions(ctx, master, master); err != nil {
			log.Error().Err(err).Uint32("master_seq", seq).Msg("cannot process masterchain block transactions")
			continue
		}

		dbMaster := &core.Block{
			BlockID: core.BlockID{
				Workchain: master.Workchain,
				Shard:     master.Shard,
				SeqNo:     master.SeqNo,
			},
			RootHash: master.RootHash,
			FileHash: master.FileHash,
		}
		if err := s.blockRepo.AddBlocks(ctx, []*core.Block{dbMaster}); err != nil {
			log.Error().Err(err).Uint32("master_seq", seq).Msg("cannot add master block")
			continue
		}

		seq++
	}
}
