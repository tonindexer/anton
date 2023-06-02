package indexer

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
	"github.com/xssnick/tonutils-go/ton"

	"github.com/tonindexer/anton/internal/app"
	"github.com/tonindexer/anton/internal/core"
)

type processedBlock struct {
	block *core.Block
	err   error
}

type processedMasterBlock struct {
	master processedBlock
	shards []processedBlock
}

func (s *Service) fetchMaster(seq uint32) (ret processedMasterBlock) {
	defer app.TimeTrack(time.Now(), fmt.Sprintf("fetchMaster(%d)", seq))

	for {
		ctx := context.Background()

		master, shards, err := s.Fetcher.UnseenBlocks(ctx, seq)
		if err != nil {
			lvl := log.Error()
			if errors.Is(err, ton.ErrBlockNotFound) || (err != nil && strings.Contains(err.Error(), "block is not applied")) {
				lvl = log.Debug()
			}
			lvl.Err(err).Uint32("master_seq", seq).Msg("cannot fetch unseen blocks")
			time.Sleep(300 * time.Millisecond)
			continue
		}
		ret.master, ret.shards = processedBlock{}, nil

		var wg sync.WaitGroup
		wg.Add(len(shards) + 1)

		ch := make(chan processedBlock, len(shards)+1)

		go func() {
			defer wg.Done()

			tx, err := s.Fetcher.BlockTransactions(ctx, master)

			ch <- processedBlock{
				block: &core.Block{
					Workchain:    master.Workchain,
					Shard:        master.Shard,
					SeqNo:        master.SeqNo,
					FileHash:     master.FileHash,
					RootHash:     master.RootHash,
					Transactions: tx,
					ScannedAt:    time.Now(),
				},
				err: err,
			}
		}()

		for i := range shards {
			go func(shard *ton.BlockIDExt) {
				defer wg.Done()

				tx, err := s.Fetcher.BlockTransactions(ctx, shard)

				ch <- processedBlock{
					block: &core.Block{
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
						Transactions: tx,
						ScannedAt:    time.Now(),
					},
					err: err,
				}
			}(shards[i])
		}

		wg.Wait()
		close(ch)

		var errBlock processedBlock
		for i := range ch {
			if i.err != nil {
				errBlock = i
			}
			if i.block.Workchain == master.Workchain {
				ret.master = i
			} else {
				ret.shards = append(ret.shards, i)
			}
		}
		if errBlock.err != nil {
			log.Error().
				Err(errBlock.err).
				Int32("workchain", errBlock.block.Workchain).
				Uint64("shard", uint64(errBlock.block.Shard)).
				Uint32("seq", errBlock.block.SeqNo).
				Msg("cannot process block")
		} else {
			return ret
		}
	}
}

func (s *Service) fetchMastersConcurrent(fromBlock uint32) []processedMasterBlock {
	var blocks []processedMasterBlock
	var wg sync.WaitGroup

	wg.Add(s.Workers)

	ch := make(chan processedMasterBlock, s.Workers)

	for i := 0; i < s.Workers; i++ {
		go func(seq uint32) {
			defer wg.Done()
			r := s.fetchMaster(seq)
			ch <- r
		}(fromBlock + uint32(i))
	}

	wg.Wait()
	close(ch)

	for b := range ch {
		blocks = append(blocks, b)
	}

	sort.Slice(blocks, func(i, j int) bool {
		return blocks[i].master.block.SeqNo < blocks[j].master.block.SeqNo
	})

	return blocks
}

func (s *Service) needThreads(latestProcessedBlock uint32) bool {
	if !s.threaded {
		return s.threaded
	}

	latest, err := s.API.GetMasterchainInfo(context.Background())
	if err != nil {
		log.Error().Err(err).Msg("cannot get masterchain info")
		return false
	}

	if latestProcessedBlock > latest.SeqNo-16 {
		s.threaded = false
	}
	return s.threaded
}

func (s *Service) fetchMasterLoop(fromBlock uint32, results chan<- processedMasterBlock) {
	defer s.wg.Done()

	for s.running() {
		if !s.needThreads(fromBlock) || s.Workers <= 1 {
			block := s.fetchMaster(fromBlock)
			results <- block
			fromBlock = block.master.block.SeqNo + 1
			continue
		}

		blocks := s.fetchMastersConcurrent(fromBlock)
		for i := range blocks {
			results <- blocks[i]
		}
		fromBlock = blocks[len(blocks)-1].master.block.SeqNo + 1
	}
}
