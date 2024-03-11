package rescan

import (
	"context"
	"sync"
	"time"

	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
	"github.com/xssnick/tonutils-go/tvm/cell"

	"github.com/tonindexer/anton/internal/app"
	"github.com/tonindexer/anton/internal/core"
)

var _ app.RescanService = (*Service)(nil)

type Service struct {
	*app.RescanConfig

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

	for s.running() {
		tx, task, err := s.ContractRepo.GetUnfinishedRescanTask(context.Background())
		if err != nil {
			if !errors.Is(err, core.ErrNotFound) {
				log.Error().Err(err).Msg("get rescan task")
			}
			time.Sleep(time.Second)
			continue
		}

		if err := s.rescanRunTask(context.Background(), task); err != nil {
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

func (s *Service) rescanRunTask(ctx context.Context, task *core.RescanTask) error {
	switch task.Type {
	case core.AddInterface:
		var codeHash []byte
		if task.Contract.Code != nil {
			codeCell, err := cell.FromBOC(task.Contract.Code)
			if err != nil {
				return errors.Wrapf(err, "making %s code cell from boc", task.Contract.Name)
			}
			codeHash = codeCell.Hash()
		}

		ids, err := s.AccountRepo.MatchStatesByInterfaceDesc(ctx, "", task.Contract.Addresses, codeHash, task.Contract.GetMethodHashes, task.LastAddress, task.LastTxLt, s.SelectLimit)
		if err != nil {
			return errors.Wrapf(err, "match states by interface description")
		}
		if len(ids) == 0 {
			task.Finished = true
			return nil
		}

		if err := s.rescanAccounts(ctx, task, ids); err != nil {
			return errors.Wrapf(err, "rescan accounts")
		}

		return nil

	case core.UpdInterface, core.DelInterface:
		ids, err := s.AccountRepo.MatchStatesByInterfaceDesc(ctx, task.ContractName, nil, nil, nil, task.LastAddress, task.LastTxLt, s.SelectLimit)
		if err != nil {
			return errors.Wrapf(err, "match states by interface description")
		}
		if len(ids) == 0 {
			task.Finished = true
			return nil
		}

		if err := s.rescanAccounts(ctx, task, ids); err != nil {
			return errors.Wrapf(err, "rescan accounts")
		}

		return nil

	case core.AddGetMethod, core.DelGetMethod, core.UpdGetMethod:
		var codeHash []byte
		if task.Contract.Code != nil {
			codeCell, err := cell.FromBOC(task.Contract.Code)
			if err != nil {
				return errors.Wrapf(err, "making %s code cell from boc", task.Contract.Name)
			}
			codeHash = codeCell.Hash()
		}

		ids, err := s.AccountRepo.MatchStatesByInterfaceDesc(ctx, task.ContractName, task.Contract.Addresses, codeHash, task.Contract.GetMethodHashes, task.LastAddress, task.LastTxLt, s.SelectLimit)
		if err != nil {
			return errors.Wrapf(err, "match states by interface description")
		}

		if err := s.rescanAccounts(ctx, task, ids); err != nil {
			return errors.Wrapf(err, "rescan accounts")
		}

		return nil

	case core.DelOperation, core.UpdOperation:
		hashes, err := s.MessageRepo.MatchMessagesByOperationDesc(ctx, task.ContractName, task.MessageType, task.Outgoing, task.OperationID, task.LastAddress, task.LastTxLt, s.SelectLimit)
		if err != nil {
			return errors.Wrapf(err, "get addresses by contract name")
		}

		if err := s.rescanMessages(ctx, task, hashes); err != nil {
			return errors.Wrapf(err, "rescan messages")
		}
	}

	return errors.Wrapf(core.ErrInvalidArg, "unknown rescan task type %s", task.Type)
}
