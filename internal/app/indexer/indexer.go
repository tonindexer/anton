package indexer

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
	"github.com/xssnick/tonutils-go/ton"

	"github.com/iam047801/tonidx/internal/app"
	"github.com/iam047801/tonidx/internal/core"
	"github.com/iam047801/tonidx/internal/core/repository/account"
	"github.com/iam047801/tonidx/internal/core/repository/block"
	"github.com/iam047801/tonidx/internal/core/repository/contract"
	"github.com/iam047801/tonidx/internal/core/repository/tx"
)

var _ app.IndexerService = (*Service)(nil)

type Service struct {
	cfg *app.IndexerConfig

	abiRepo     core.ContractRepository
	blockRepo   core.BlockRepository
	txRepo      core.TxRepository
	accountRepo core.AccountRepository

	parser app.ParserService
	api    *ton.APIClient

	shardLastSeqno map[string]uint32

	run bool
	mx  sync.RWMutex
	wg  sync.WaitGroup
}

func NewService(_ context.Context, cfg *app.IndexerConfig) (*Service, error) {
	var s = new(Service)

	s.cfg = cfg
	ch, pg := cfg.DB.CH, cfg.DB.PG
	s.abiRepo = contract.NewRepository(pg)
	s.blockRepo = block.NewRepository(ch, pg)
	s.txRepo = tx.NewRepository(ch, pg)
	s.accountRepo = account.NewRepository(ch, pg)

	s.parser = cfg.Parser
	s.api = s.parser.API()

	s.shardLastSeqno = make(map[string]uint32)

	return s, nil
}

func (s *Service) running() bool {
	s.mx.RLock()
	defer s.mx.RUnlock()

	return s.run
}

// func to get storage map key
func getShardID(shard *ton.BlockIDExt) string {
	return fmt.Sprintf("%d|%d", shard.Workchain, shard.Shard)
}

func timeTrack(start time.Time, name string) {
	elapsed := float64(time.Since(start)) / 1e9
	if elapsed < 0.05 {
		return
	}
	log.Debug().Str("func", name).Float64("elapsed", elapsed).Msg("")
}

func (s *Service) Start() error {
	var fromBlock uint32

	ctx := context.Background()

	master, err := s.api.GetMasterchainInfo(ctx)
	if err != nil {
		return errors.Wrap(err, "cannot get masterchain info")
	}

	lastMaster, err := s.blockRepo.GetLastMasterBlock(ctx)
	switch {
	case err == nil:
		fromBlock = lastMaster.SeqNo + 1
	case errors.Is(err, core.ErrNotFound):
		fromBlock = master.SeqNo
		if s.cfg.FromBlock != 0 {
			fromBlock = s.cfg.FromBlock
		}
	case err != nil && !errors.Is(err, core.ErrNotFound):
		return errors.Wrap(err, "cannot get last masterchain block")
	}

	master, err = s.api.LookupBlock(ctx, master.Workchain, master.Shard, fromBlock-1)
	if err != nil {
		return errors.Wrap(err, "lookup master")
	}

	// getting information about other work-chains and shards of first master block
	// to init storage of last seen shard seq numbers
	firstShards, err := s.api.GetBlockShardsInfo(ctx, master)
	if err != nil {
		return errors.Wrapf(err, "get block shards info (master seq no = %d)", master.SeqNo)
	}
	for _, shard := range firstShards {
		s.shardLastSeqno[getShardID(shard)] = shard.SeqNo
	}

	s.wg.Add(1)
	go s.fetchBlocksLoop(master.Workchain, master.Shard, fromBlock)

	s.mx.Lock()
	s.run = true
	s.mx.Unlock()

	return nil
}

func (s *Service) Stop() {
	s.mx.Lock()
	s.run = false
	s.mx.Unlock()

	s.wg.Wait()

	s.cfg.DB.Close()
}
