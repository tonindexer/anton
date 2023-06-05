package indexer

import (
	"context"
	"sync"

	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"

	"github.com/tonindexer/anton/internal/app"
	"github.com/tonindexer/anton/internal/core"
	"github.com/tonindexer/anton/internal/core/repository"
	"github.com/tonindexer/anton/internal/core/repository/account"
	"github.com/tonindexer/anton/internal/core/repository/block"
	"github.com/tonindexer/anton/internal/core/repository/msg"
	"github.com/tonindexer/anton/internal/core/repository/tx"
)

var _ app.IndexerService = (*Service)(nil)

type pendingMaster struct {
	Info []*core.Block
	Tx   []*core.Transaction
	Acc  []*core.AccountState
	Msg  []*core.Message
}

type Service struct {
	*app.IndexerConfig

	blockRepo   core.BlockRepository
	txRepo      core.TransactionRepository
	msgRepo     repository.Message
	accountRepo core.AccountRepository

	threaded bool

	unknownDstMsg map[string]*core.Message

	masterShardsCache map[core.BlockID][]core.BlockID
	shardsMasterMap   map[core.BlockID]core.BlockID

	pendingMasters map[uint32]pendingMaster

	run bool
	mx  sync.RWMutex
	wg  sync.WaitGroup
}

func NewService(cfg *app.IndexerConfig) *Service {
	var s = new(Service)

	s.IndexerConfig = cfg

	ch, pg := s.DB.CH, s.DB.PG
	s.txRepo = tx.NewRepository(ch, pg)
	s.msgRepo = msg.NewRepository(ch, pg)
	s.blockRepo = block.NewRepository(ch, pg)
	s.accountRepo = account.NewRepository(ch, pg)

	s.threaded = true

	s.unknownDstMsg = make(map[string]*core.Message)

	s.masterShardsCache = make(map[core.BlockID][]core.BlockID)
	s.shardsMasterMap = make(map[core.BlockID]core.BlockID)

	s.pendingMasters = make(map[uint32]pendingMaster)

	return s
}

func (s *Service) running() bool {
	s.mx.RLock()
	defer s.mx.RUnlock()

	return s.run
}

func (s *Service) Start() error {
	ctx := context.Background()

	fromBlock := s.FromBlock

	lastMaster, err := s.blockRepo.GetLastMasterBlock(ctx)
	switch {
	case err == nil:
		fromBlock = lastMaster.SeqNo + 1
	case err != nil && !errors.Is(err, core.ErrNotFound):
		return errors.Wrap(err, "cannot get last masterchain block")
	}

	s.mx.Lock()
	s.run = true
	s.mx.Unlock()

	blocksChan := make(chan *core.Block, s.Workers*2)

	s.wg.Add(1)
	go s.fetchMasterLoop(fromBlock, blocksChan)

	s.wg.Add(1)
	go s.saveBlocksLoop(blocksChan)

	log.Info().
		Uint32("from_block", fromBlock).
		Int("workers", s.Workers).
		Int("insert_block_batch", s.InsertBlockBatch).
		Msg("started")

	return nil
}

func (s *Service) Stop() {
	s.mx.Lock()
	s.run = false
	s.mx.Unlock()

	s.wg.Wait()
}
