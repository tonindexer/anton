package rescan

import (
	"context"
	"reflect"

	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"

	"github.com/tonindexer/anton/abi"
	"github.com/tonindexer/anton/addr"
	"github.com/tonindexer/anton/internal/app"
	"github.com/tonindexer/anton/internal/core"
)

func (s *Service) chooseInterfaces(updates map[uint64][]abi.ContractName, txLT uint64) (ret []abi.ContractName) {
	if len(updates) == 0 {
		return ret
	}

	var maxLT uint64
	for updLT, types := range updates {
		if updLT <= txLT && updLT > maxLT {
			maxLT = updLT
			ret = types
		}
	}
	if ret != nil {
		return ret
	}

	minLT := uint64(1 << 63)
	for updLT, types := range updates {
		if updLT < minLT {
			minLT = updLT
			ret = types
		}
	}
	return ret
}

func (s *Service) getAccountStateForMessage(ctx context.Context, a addr.Address, txLT uint64) *core.AccountState {
	interfaceUpdates, ok := s.interfacesCache.Get(a)
	if ok {
		return &core.AccountState{Address: a, Types: s.chooseInterfaces(interfaceUpdates, txLT)}
	}

	interfaceUpdates, err := s.AccountRepo.GetAllAccountInterfaces(ctx, a)
	if err != nil {
		log.Error().Err(err).Str("addr", a.Base64()).Msg("get all account interfaces")
		return nil
	}

	s.interfacesCache.Put(a, interfaceUpdates)

	return &core.AccountState{Address: a, Types: s.chooseInterfaces(interfaceUpdates, txLT)}
}

func (s *Service) rescanMessage(ctx context.Context, task *core.RescanTask, update *core.Message) error {
	// we must get account state's interfaces to properly determine message operation
	// and to parse message accordingly

	// so for the source of the message we take the account state of the sender,
	// which is updated just before the message is sent

	// for the destination of the given message, we take the account state of receiver,
	// which is updated just after the message is received

	if task.Outgoing {
		update.SrcState = &core.AccountState{Address: update.SrcAddress, Types: []abi.ContractName{task.ContractName}}
		update.DstState = s.getAccountStateForMessage(ctx, update.DstAddress, update.DstTxLT)
	} else {
		update.SrcState = s.getAccountStateForMessage(ctx, update.SrcAddress, update.SrcTxLT)
		update.DstState = &core.AccountState{Address: update.DstAddress, Types: []abi.ContractName{task.ContractName}}
	}

	err := s.Parser.ParseMessagePayload(ctx, update)
	if err != nil {
		if !errors.Is(err, app.ErrImpossibleParsing) {
			log.Error().Err(err).
				Hex("msg_hash", update.Hash).
				Hex("src_tx_hash", update.SrcTxHash).
				Str("src_addr", update.SrcAddress.String()).
				Hex("dst_tx_hash", update.DstTxHash).
				Str("dst_addr", update.DstAddress.String()).
				Uint32("op_id", update.OperationID).
				Msg("parse message payload")
		}
		return err
	}

	return nil
}

func (s *Service) rescanMessagesWorker(ctx context.Context, task *core.RescanTask, messages []*core.Message) (updates []*core.Message) {
	for _, msg := range messages {
		upd := *msg

		switch task.Type {
		case core.DelOperation:
			upd.SrcContract, upd.DstContract, upd.OperationName, upd.DataJSON, upd.Error = "", "", "", nil, ""

		case core.UpdOperation:
			if err := s.rescanMessage(ctx, task, &upd); err != nil {
				continue
			}
		}

		if !reflect.DeepEqual(msg, &upd) {
			updates = append(updates, &upd)
		}
	}

	return updates
}
