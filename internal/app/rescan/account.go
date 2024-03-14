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
	"github.com/tonindexer/anton/internal/core/filter"
)

func (s *Service) getRecentAccountState(ctx context.Context, b core.BlockID, a addr.Address) (*core.AccountState, error) {
	defer app.TimeTrack(time.Now(), "getLastSeenAccountState(%d, %d, %d, %s)", b.Workchain, b.Shard, b.SeqNo, a.String())

	var boundBlock core.BlockID
	switch {
	case b.Workchain == int32(a.Workchain()):
		boundBlock = b
	default:
		return nil, errors.Wrapf(core.ErrInvalidArg, "address is in %d workchain, but the given block is from %d workchain", a.Workchain(), b.Workchain)
	}

	accountReq := filter.AccountsReq{
		Addresses:     []*addr.Address{&a},
		Workchain:     &boundBlock.Workchain,
		Shard:         &boundBlock.Shard,
		BlockSeqNoLeq: &boundBlock.SeqNo,
		Order:         "DESC",
		Limit:         1,
	}
	accountRes, err := s.AccountRepo.FilterAccounts(ctx, &accountReq)
	if err != nil {
		return nil, errors.Wrap(err, "filter accounts")
	}
	if len(accountRes.Rows) < 1 {
		return nil, errors.Wrap(core.ErrNotFound, "could not find needed account state")
	}

	return accountRes.Rows[0], nil
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
		return s.getRecentAccountState(ctx, acc.BlockID(), a)
	}

	err := s.Parser.ParseAccountContractData(ctx, task.Contract, acc, getOtherAccountFunc)
	if err != nil && !errors.Is(err, app.ErrImpossibleParsing) {
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

func (s *Service) clearExecutedGetMethod(task *core.RescanTask, acc *core.AccountState) {
	_, ok := acc.ExecutedGetMethods[task.ContractName]
	if !ok {
		return
	}

	for it := range acc.ExecutedGetMethods[task.ContractName] {
		if acc.ExecutedGetMethods[task.ContractName][it].Name != task.ChangedGetMethod {
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

	switch task.ChangedGetMethod {
	case "get_nft_content", "get_collection_data", "get_jetton_data":
		acc.ContentURI = ""
		acc.ContentName = ""
		acc.ContentDescription = ""
		acc.ContentImage = ""
		acc.ContentImageData = nil
	}

	switch task.ChangedGetMethod {
	case "get_collection_data":
		acc.OwnerAddress = nil
	case "get_nft_data", "get_wallet_data":
		acc.OwnerAddress = nil
		acc.MinterAddress = nil
		acc.Fake = false

		acc.JettonBalance = nil
	}

	switch task.ChangedGetMethod {
	case "get_wallet_address", "get_nft_address_by_index":
		acc.Fake = false
	}
}

func (s *Service) executeGetMethod(ctx context.Context, task *core.RescanTask, acc *core.AccountState) {
	getOtherAccountFunc := func(ctx context.Context, a addr.Address) (*core.AccountState, error) {
		return s.getRecentAccountState(ctx, acc.BlockID(), a)
	}

	err := s.Parser.ExecuteAccountGetMethod(ctx, task.ContractName, task.ChangedGetMethod, acc, getOtherAccountFunc)
	if err != nil && !errors.Is(err, app.ErrImpossibleParsing) {
		log.Error().Err(err).
			Str("contract_name", string(task.ContractName)).
			Str("get_method", task.ChangedGetMethod).
			Str("addr", acc.Address.Base64()).
			Msg("parse account data")
	}
}

func (s *Service) rescanGetMethod(ctx context.Context, task *core.RescanTask, acc *core.AccountState) {
	s.clearExecutedGetMethod(task, acc)

	matchedByGetMethod := func() (matchedByGM, hasGM bool) {
		if len(task.Contract.Code) > 0 || len(task.Contract.Addresses) > 0 {
			return false, false
		}

		changed := abi.MethodNameHash(task.ChangedGetMethod)
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

		s.executeGetMethod(ctx, task, acc)

	case core.DelGetMethod:
		m, h := matchedByGetMethod()
		if m && !h {
			// include all account states that match the contract interface description,
			// minus the deleted get method
			s.parseAccountData(ctx, task, acc)
			return
		}

	case core.UpdGetMethod:
		s.executeGetMethod(ctx, task, acc)
	}
}

func (s *Service) rescanAccountsWorker(ctx context.Context, task *core.RescanTask, batch []*core.AccountState) (updates []*core.AccountState) {
	for _, acc := range batch {
		update := copyAccountState(acc)

		switch task.Type {
		case core.AddInterface, core.UpdInterface, core.DelInterface:
			s.rescanInterface(ctx, task, update)
		case core.AddGetMethod, core.UpdGetMethod, core.DelGetMethod:
			s.rescanGetMethod(ctx, task, update)
		}

		if reflect.DeepEqual(acc, update) {
			continue
		}

		updates = append(updates, update)
	}

	return updates
}
