package rescan

import (
	"bytes"
	"context"
	"strings"
	"sync"
	"time"

	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
	"github.com/xssnick/tonutils-go/tvm/cell"

	"github.com/tonindexer/anton/internal/app"
	"github.com/tonindexer/anton/internal/core"
	"github.com/tonindexer/anton/internal/core/filter"
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
		tx, task, err := s.RescanRepo.GetUnfinishedRescanTask(context.Background())
		if err != nil {
			if !(errors.Is(err, core.ErrNotFound) && strings.Contains(err.Error(), "no unfinished tasks")) {
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

		if err := s.RescanRepo.SetRescanTask(context.Background(), tx, task); err != nil {
			log.Error().Err(err).Msg("update rescan task")
			time.Sleep(time.Second)
			continue
		}
	}
}

func (s *Service) rescanRunTask(ctx context.Context, task *core.RescanTask) error {
	var codeHash []byte
	if task.Contract != nil && task.Contract.Code != nil {
		codeCell, err := cell.FromBOC(task.Contract.Code)
		if err != nil {
			return errors.Wrapf(err, "making %s code cell from boc", task.Contract.Name)
		}
		codeHash = codeCell.Hash()
	}

	switch task.Type {
	case core.AddInterface:
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
		ids, err := s.AccountRepo.MatchStatesByInterfaceDesc(ctx, task.ContractName, task.Contract.Addresses, codeHash, task.Contract.GetMethodHashes, task.LastAddress, task.LastTxLt, s.SelectLimit)
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

	case core.DelOperation, core.UpdOperation:
		hashes, err := s.MessageRepo.MatchMessagesByOperationDesc(ctx, task.ContractName, task.MessageType, task.Outgoing, task.OperationID, task.LastAddress, task.LastTxLt, s.SelectLimit)
		if err != nil {
			return errors.Wrapf(err, "get addresses by contract name")
		}
		if len(hashes) == 0 {
			task.Finished = true
			return nil
		}

		if err := s.rescanMessages(ctx, task, hashes); err != nil {
			return errors.Wrapf(err, "rescan messages")
		}

		return nil
	}

	return errors.Wrapf(core.ErrInvalidArg, "unknown rescan task type %s", task.Type)
}

func (s *Service) rescanAccounts(ctx context.Context, task *core.RescanTask, ids []*core.AccountStateID) error {
	accRet, err := s.AccountRepo.FilterAccounts(ctx, &filter.AccountsReq{StateIDs: ids})
	if err != nil {
		return errors.Wrapf(err, "filter accounts")
	}

	updates, lastScanned := rescanStartWorkers(
		ctx, task, accRet.Rows,
		func(v *core.AccountState) core.AccountStateID {
			return core.AccountStateID{Address: v.Address, LastTxLT: v.LastTxLT}
		},
		s.rescanAccountsWorker, s.Workers)

	if len(updates) > 0 {
		if err := s.AccountRepo.UpdateAccountStates(ctx, updates); err != nil {
			return errors.Wrapf(err, "update account states")
		}
	}

	task.LastAddress = &lastScanned.Address
	task.LastTxLt = lastScanned.LastTxLT

	return nil
}

func (s *Service) rescanMessages(ctx context.Context, task *core.RescanTask, hashes [][]byte) error {
	messages, err := s.MessageRepo.GetMessages(ctx, hashes)
	if err != nil {
		return err
	}

	updates, lastScanned := rescanStartWorkers(
		ctx, task, messages,
		func(v *core.Message) core.AccountStateID {
			msgID := core.AccountStateID{Address: v.DstAddress, LastTxLT: v.DstTxLT}
			if task.Outgoing {
				msgID = core.AccountStateID{Address: v.SrcAddress, LastTxLT: v.SrcTxLT}
			}
			return msgID
		},
		s.rescanMessagesWorker, s.Workers)

	if len(updates) > 0 {
		if err := s.MessageRepo.UpdateMessages(context.Background(), updates); err != nil {
			return errors.Wrap(err, "update messages")
		}
	}

	task.LastAddress = &lastScanned.Address
	task.LastTxLt = lastScanned.LastTxLT

	return nil
}

func rescanStartWorkers[V any](ctx context.Context,
	task *core.RescanTask,
	slice []V,
	getID func(V) core.AccountStateID,
	workerFunc func(context.Context, *core.RescanTask, []V) []V,
	workers int,
) (updatesAll []V, lastParsed core.AccountStateID) {
	var (
		updatesChan = make(chan []V)
		scanWG      sync.WaitGroup
	)

	if len(slice) < workers {
		workers = len(slice)
	}

	for i := 0; i < len(slice); {
		batchLen := (len(slice) - i) / workers
		if (len(slice)-i)%workers != 0 {
			batchLen++
		}

		scanWG.Add(1)
		go func(batch []V) {
			defer scanWG.Done()
			updatesChan <- workerFunc(ctx, task, batch)
		}(slice[i : i+batchLen])

		i += batchLen
		workers--
	}

	go func() {
		scanWG.Wait()
		close(updatesChan)
	}()

	for updates := range updatesChan {
		for _, upd := range updates {
			updID := getID(upd)
			if bytes.Compare(lastParsed.Address[:], updID.Address[:]) <= 0 || lastParsed.LastTxLT <= updID.LastTxLT {
				lastParsed = updID
			}
		}
		updatesAll = append(updatesAll, updates...)
	}

	return updatesAll, lastParsed
}
