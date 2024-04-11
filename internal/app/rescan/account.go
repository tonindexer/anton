package rescan

import (
	"context"
	"reflect"
	"time"

	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"

	"github.com/tonindexer/anton/abi"
	"github.com/tonindexer/anton/abi/known"
	"github.com/tonindexer/anton/addr"
	"github.com/tonindexer/anton/internal/app"
	"github.com/tonindexer/anton/internal/core"
)

func (s *Service) getRecentAccountState(ctx context.Context, a addr.Address, lastLT uint64) (*core.AccountState, error) {
	defer app.TimeTrack(time.Now(), "getRecentAccountState(%s, %d)", a.String(), lastLT)

	if minter, ok := s.minterStateCache.get(a, lastLT); ok {
		return minter, nil
	}

	states, err := s.AccountRepo.GetAllAccountStates(ctx, a, lastLT, maxMinterStates)
	if err != nil {
		return nil, err
	}

	for nextMinterTxLT, state := range states {
		s.minterStateCache.put(a, state, nextMinterTxLT)
	}

	minter, ok := s.minterStateCache.get(a, lastLT)
	if !ok {
		return nil, errors.Wrapf(core.ErrNotFound, "cannot find %s minter state before %d lt", a.Base64(), lastLT)
	}

	return minter, nil
}

func copyAccountState(state *core.AccountState) *core.AccountState {
	update := *state

	update.Types = make([]abi.ContractName, len(state.Types))
	copy(update.Types, state.Types)

	update.ExecutedGetMethods = map[abi.ContractName][]abi.GetMethodExecution{}
	for n, e := range state.ExecutedGetMethods {
		update.ExecutedGetMethods[n] = make([]abi.GetMethodExecution, len(e))
		copy(update.ExecutedGetMethods[n], e)
	}

	return &update
}

func (s *Service) clearParsedAccountsData(task *core.RescanTask, acc *core.AccountState) {
	for it := range acc.Types {
		if acc.Types[it] != task.ContractName {
			continue
		}
		types := acc.Types
		copy(types[it:], types[it+1:])
		acc.Types = types[:len(types)-1]
		break
	}

	_, ok := acc.ExecutedGetMethods[task.ContractName]
	if !ok {
		return
	}

	delete(acc.ExecutedGetMethods, task.ContractName)

	switch task.ContractName {
	case known.NFTCollection, known.NFTItem, known.JettonMinter, known.JettonWallet:
		acc.MinterAddress = nil
		acc.OwnerAddress = nil

		acc.ContentURI = ""
		acc.ContentName = ""
		acc.ContentDescription = ""
		acc.ContentImage = ""
		acc.ContentImageData = nil

		acc.Fake = false

		acc.JettonBalance = nil
	}
}

func (s *Service) parseAccountData(ctx context.Context, task *core.RescanTask, acc *core.AccountState) {
	if len(acc.Types) > 0 && known.IsOnlyWalletInterfaces(acc.Types) {
		// we do not want to emulate wallet get-methods once again,
		// as there are lots of them, so it takes a lot of CPU usage
		return
	}

	getOtherAccountFunc := func(ctx context.Context, a addr.Address) (*core.AccountState, error) {
		return s.getRecentAccountState(ctx, a, acc.LastTxLT)
	}

	err := s.Parser.ParseAccountContractData(ctx, task.Contract, acc, getOtherAccountFunc)
	if err != nil && !errors.Is(err, app.ErrUnmatchedContractInterface) {
		log.Error().Err(err).Str("addr", acc.Address.Base64()).Msg("parse account data")
	}
}

func (s *Service) rescanInterface(ctx context.Context, task *core.RescanTask, acc *core.AccountState) {
	s.clearParsedAccountsData(task, acc)

	if task.Type == core.DelInterface {
		return
	}

	s.parseAccountData(ctx, task, acc)
}

func (s *Service) clearExecutedGetMethod(task *core.RescanTask, acc *core.AccountState, gm string) {
	_, ok := acc.ExecutedGetMethods[task.ContractName]
	if !ok {
		return
	}

	for it := range acc.ExecutedGetMethods[task.ContractName] {
		if acc.ExecutedGetMethods[task.ContractName][it].Name != gm {
			continue
		}
		executions := acc.ExecutedGetMethods[task.ContractName]
		copy(executions[it:], executions[it+1:])
		acc.ExecutedGetMethods[task.ContractName] = executions[:len(executions)-1]
		break
	}

	switch task.ContractName {
	case known.NFTCollection, known.NFTItem, known.JettonMinter, known.JettonWallet:
	default:
		return
	}

	switch gm {
	case "get_nft_content", "get_collection_data", "get_jetton_data":
		acc.ContentURI = ""
		acc.ContentName = ""
		acc.ContentDescription = ""
		acc.ContentImage = ""
		acc.ContentImageData = nil
	}

	switch gm {
	case "get_collection_data":
		acc.OwnerAddress = nil
	case "get_nft_data", "get_wallet_data":
		acc.OwnerAddress = nil
		acc.MinterAddress = nil
		acc.Fake = false

		acc.JettonBalance = nil
	}

	switch gm {
	case "get_wallet_address", "get_nft_address_by_index":
		acc.Fake = false
	}
}

func (s *Service) executeGetMethod(ctx context.Context, task *core.RescanTask, acc *core.AccountState, gm string) {
	getOtherAccountFunc := func(ctx context.Context, a addr.Address) (*core.AccountState, error) {
		return s.getRecentAccountState(ctx, a, acc.LastTxLT)
	}

	err := s.Parser.ExecuteAccountGetMethod(ctx, task.ContractName, gm, acc, getOtherAccountFunc)
	if err != nil && !errors.Is(err, app.ErrImpossibleParsing) {
		log.Error().Err(err).
			Str("contract_name", string(task.ContractName)).
			Str("get_method", gm).
			Str("addr", acc.Address.Base64()).
			Msg("parse account data")
	}
}

func (s *Service) rescanGetMethod(ctx context.Context, task *core.RescanTask, acc *core.AccountState, gm string) {
	s.clearExecutedGetMethod(task, acc, gm)

	matchedByGetMethod := func() (matchedByGM, hasGM bool) {
		if len(task.Contract.Code) > 0 || len(task.Contract.Addresses) > 0 {
			return false, false
		}

		changed := abi.MethodNameHash(gm)
		for _, gmh := range task.Contract.GetMethodHashes {
			if gmh == changed {
				return true, true
			}
		}
		return true, false
	}

	switch task.Type {
	case core.AddGetMethod:
		m, h := matchedByGetMethod()
		if m && !h {
			// clear all parsed data in account states lacking the new get method
			s.clearParsedAccountsData(task, acc)
			return
		}

		s.executeGetMethod(ctx, task, acc, gm)

	case core.DelGetMethod:
		m, h := matchedByGetMethod()
		if m && !h {
			// include all account states that match the contract interface description,
			// minus the deleted get method
			s.parseAccountData(ctx, task, acc)
			return
		}

	case core.UpdGetMethod:
		s.executeGetMethod(ctx, task, acc, gm)
	}
}

func (s *Service) rescanAccountsWorker(ctx context.Context, task *core.RescanTask, batch []*core.AccountState) (updates []*core.AccountState) {
	for _, acc := range batch {
		update := copyAccountState(acc)

		switch task.Type {
		case core.AddInterface, core.UpdInterface, core.DelInterface:
			s.rescanInterface(ctx, task, update)
		case core.AddGetMethod, core.UpdGetMethod, core.DelGetMethod:
			for _, gm := range task.ChangedGetMethods {
				s.rescanGetMethod(ctx, task, update, gm)
			}
		}

		if reflect.DeepEqual(acc, update) {
			continue
		}

		updates = append(updates, update)
	}

	return updates
}
