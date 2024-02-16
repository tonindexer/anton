package rescan

import (
	"context"
	"sync"
	"time"

	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"

	"github.com/tonindexer/anton/internal/app"
	"github.com/tonindexer/anton/internal/core"
	"github.com/tonindexer/anton/internal/core/filter"
)

var _ app.RescanService = (*Service)(nil)

type Service struct {
	*app.RescanConfig

	masterShard int64

	interfacesCache *interfacesCache

	run bool
	mx  sync.RWMutex
	wg  sync.WaitGroup
}

func NewService(cfg *app.RescanConfig) *Service {
	var s = new(Service)

	s.RescanConfig = cfg

	// validate config
	if s.Workers < 1 {
		s.Workers = 1
	}

	s.interfacesCache = newInterfacesCache(16384) // number of addresses

	return s
}

func (s *Service) running() bool {
	s.mx.RLock()
	defer s.mx.RUnlock()

	return s.run
}

func (s *Service) Start() error {
	s.mx.Lock()
	s.run = true
	s.mx.Unlock()

	s.wg.Add(1)
	go s.rescanLoop()

	log.Info().
		Int("workers", s.Workers).
		Msg("rescan started")

	return nil
}

func (s *Service) Stop() {
	s.mx.Lock()
	s.run = false
	s.mx.Unlock()

	s.wg.Wait()
}

func (s *Service) rescanLoop() {
	defer s.wg.Done()

	lastMaster, err := s.BlockRepo.GetLastMasterBlock(context.Background())
	if err != nil {
		log.Error().Err(err).Msg("cannot get last masterchain block")
		return
	}
	toBlock := lastMaster.SeqNo
	s.masterShard = lastMaster.Shard

	for s.running() {
		tx, task, err := s.ContractRepo.GetUnfinishedRescanTask(context.Background())
		if err != nil {
			if !errors.Is(err, core.ErrNotFound) {
				log.Error().Err(err).Msg("get rescan task")
			}
			time.Sleep(time.Second)
			continue
		}

		if err := s.rescanRunTask(task, toBlock); err != nil {
			_ = tx.Rollback()
			log.Error().Err(err).
				Int("id", task.ID).
				Msg("run rescan task")
			time.Sleep(time.Second)
			continue
		}

		if err := s.ContractRepo.SetRescanTask(context.Background(), tx, task); err != nil {
			log.Error().Err(err).Msg("update rescan task")
			time.Sleep(time.Second)
			continue
		}
	}
}

func (s *Service) rescanRunTask(task *core.RescanTask, toBlock uint32) error {
	if task.AccountsRescanDone || task.AccountsLastMaster >= toBlock {
		task.AccountsRescanDone = true
	} else {
		if task.AccountsLastMaster == 0 {
			task.AccountsLastMaster = task.StartFrom - 1
		}
		blocks, err := s.filterBlocksForRescan(task.AccountsLastMaster+1, toBlock, false)
		if err != nil {
			return errors.Wrap(err, "filter blocks for account states rescan")
		}
		if lastScanned := s.rescanAccounts(blocks); lastScanned != 0 {
			task.AccountsLastMaster = lastScanned
		}
	}

	if task.MessagesRescanDone || task.MessagesLastMaster >= toBlock {
		task.MessagesRescanDone = true
	} else if task.AccountsRescanDone {
		if task.MessagesLastMaster == 0 {
			task.MessagesLastMaster = task.StartFrom - 1
		}
		blocks, err := s.filterBlocksForRescan(task.MessagesLastMaster+1, toBlock, true)
		if err != nil {
			return errors.Wrap(err, "filter blocks for messages states rescan")
		}
		if lastScanned := s.rescanMessages(blocks); lastScanned != 0 {
			task.MessagesLastMaster = lastScanned
		}
	}

	if task.AccountsRescanDone && task.MessagesRescanDone {
		task.Finished = true
	}

	return nil
}

func (s *Service) filterBlocksForRescan(fromBlock, toBlock uint32, withMessages bool) ([]*core.Block, error) {
	workers := s.Workers
	if delta := int(toBlock-fromBlock) + 1; delta < workers {
		workers = delta
	}

	req := &filter.BlocksReq{
		Workchain:                   new(int32),
		Shard:                       new(int64),
		WithShards:                  true,
		WithTransactionAccountState: true,
		WithTransactions:            true,
		WithTransactionMessages:     withMessages,
		AfterSeqNo:                  new(uint32),
		Order:                       "ASC",
		Limit:                       workers,
	}
	*req.Workchain = -1
	*req.Shard = s.masterShard
	*req.AfterSeqNo = fromBlock - 1

	defer app.TimeTrack(time.Now(), "filterBlocksForRescan(%d, %d, %t)", fromBlock-1, workers, withMessages)

	res, err := s.BlockRepo.FilterBlocks(context.Background(), req)
	if err != nil {
		return nil, err
	}

	return res.Rows, nil
}
