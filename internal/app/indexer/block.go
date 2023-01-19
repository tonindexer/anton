package indexer

import (
	"context"
	"fmt"
	"time"

	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
	"github.com/xssnick/tonutils-go/address"
	"github.com/xssnick/tonutils-go/tlb"
	"github.com/xssnick/tonutils-go/ton"

	"github.com/iam047801/tonidx/internal/core"
)

//nolint // TODO: simplify this
func (s *Service) processShardTransactions(ctx context.Context, master, shard *tlb.BlockInfo, blockTransactions []*tlb.Transaction) error {
	var (
		accounts     []*core.Account
		accountMap   = make(map[string]*core.Account)
		accountsData []*core.AccountData
		payloads     []*core.MessagePayload
	)

	transactions, err := s.parser.ParseBlockTransactions(ctx, shard, blockTransactions)
	if err != nil {
		return errors.Wrap(err, "parse block transactions")
	}

	for _, tx := range transactions {
		addr := address.MustParseAddr(tx.AccountAddr)

		acc, err := s.parser.ParseAccount(ctx, master, addr)
		if err != nil {
			return errors.Wrapf(err, "parse account (addr = %s)", tx.AccountAddr)
		}
		accounts = append(accounts, acc)
		if addr.Type() == address.StdAddress {
			accountMap[acc.Address] = acc
		}
		tx.AccountBalance = acc.Balance

		data, err := s.parser.ParseAccountData(ctx, master, acc)
		if err != nil && !errors.Is(err, core.ErrNotAvailable) {
			return errors.Wrapf(err, "parse account data (addr = %s)", tx.AccountAddr)
		}
		if err == nil {
			accountsData = append(accountsData, data)
		}
	}

	messages, err := s.parser.ParseBlockMessages(ctx, shard, blockTransactions)
	if err != nil {
		return errors.Wrap(err, "parse block messages")
	}

	for _, msg := range messages {
		if !msg.Incoming {
			continue // parse only incoming messages
		}
		acc, ok := accountMap[msg.DstAddr]
		if !ok {
			continue
		}
		payload, err := s.parser.ParseMessagePayload(ctx, acc, msg)
		if errors.Is(err, core.ErrNotAvailable) {
			continue
		}
		if err != nil {
			return errors.Wrapf(err, "parse message payload (msg_hash = %x, tx_hash = %x)", msg.Hash, msg.TxHash)
		}
		payloads = append(payloads, payload)
	}

	// TODO: do not insert duplicated accounts and account data
	if err := s.accountRepo.AddAccounts(ctx, accounts); err != nil {
		return errors.Wrap(err, "add accounts")
	}
	if err := s.accountRepo.AddAccountData(ctx, accountsData); err != nil {
		return errors.Wrap(err, "add account data")
	}
	if err := s.txRepo.AddTransactions(ctx, transactions); err != nil {
		return errors.Wrap(err, "add transactions")
	}
	if err := s.txRepo.AddMessages(ctx, messages); err != nil {
		return errors.Wrap(err, "add messages")
	}
	if err := s.txRepo.AddMessagePayloads(ctx, payloads); err != nil {
		return errors.Wrap(err, "add message payloads")
	}

	return nil
}

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

func (s *Service) processShards(ctx context.Context, master *tlb.BlockInfo) ([]*core.ShardBlockInfo, error) {
	var dbShards []*core.ShardBlockInfo

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
		s.shardLastSeqno[getShardID(shard)] = shard.SeqNo
		newShards = append(newShards, notSeen...)
	}

	for _, shard := range newShards {
		log.Debug().
			Uint32("master_seq", master.SeqNo).
			Int32("shard_workchain", shard.Workchain).Uint32("shard_seq", shard.SeqNo).
			Msg("new shard block")

		blockTx, err := s.parser.GetBlockTransactions(ctx, shard)
		if err != nil {
			return nil, err
		}

		if err := s.processShardTransactions(ctx, master, shard, blockTx); err != nil {
			return nil, err
		}

		dbShards = append(dbShards, &core.ShardBlockInfo{
			Workchain:      shard.Workchain,
			Shard:          shard.Shard,
			SeqNo:          shard.SeqNo,
			RootHash:       shard.RootHash,
			FileHash:       shard.FileHash,
			MasterFileHash: master.FileHash,
		})
	}

	if err := s.blockRepo.AddShardBlocksInfo(ctx, dbShards); err != nil {
		return nil, errors.Wrap(err, "add shard block")
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

		shards, err := s.processShards(ctx, master)
		if err != nil {
			log.Error().Err(err).Uint32("master_seq", seq).Msg("cannot process shards")
			continue
		}

		// TODO: do we need to parse transactions on master chain (?)
		// if err := s.processBlockTransactions(ctx, master); err != nil {
		// 	log.Error().Err(err).Uint32("master_seq", seq).Msg("cannot process masterchain block transactions")
		// 	continue
		// }

		dbMaster := &core.MasterBlockInfo{
			Workchain: master.Workchain,
			Shard:     master.Shard,
			SeqNo:     master.SeqNo,
			RootHash:  master.RootHash,
			FileHash:  master.FileHash,
		}
		for _, shardBlock := range shards {
			dbMaster.ShardFileHashes = append(dbMaster.ShardFileHashes, shardBlock.FileHash)
		}
		if err := s.blockRepo.AddMasterBlockInfo(ctx, dbMaster); err != nil {
			log.Error().Err(err).Uint32("master_seq", seq).Msg("cannot add master block")
			continue
		}

		seq++
	}
}
