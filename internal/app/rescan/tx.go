package rescan

import (
	"context"
	"reflect"
	"sync"

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
	interfaceUpdates, ok := s.interfacesCache.get(a)
	if ok {
		return &core.AccountState{Address: a, Types: s.chooseInterfaces(interfaceUpdates, txLT)}
	}

	interfaceUpdates, err := s.AccountRepo.GetAllAccountInterfaces(ctx, a)
	if err != nil {
		log.Error().Err(err).Str("addr", a.Base64()).Msg("get all account interfaces")
		return nil
	}

	s.interfacesCache.put(a, interfaceUpdates)

	return &core.AccountState{Address: a, Types: s.chooseInterfaces(interfaceUpdates, txLT)}
}

func (s *Service) rescanMessage(ctx context.Context, msg *core.Message) *core.Message {
	// we must get account state's interfaces to properly determine message operation
	// and to parse message accordingly

	// so for the source of the message we take the account state of the sender,
	// which was updated just before the message was sent

	// for the destination of the given message, we take the account state of receiver,
	// which was update just after the message was received

	msg.SrcState = s.getAccountStateForMessage(ctx, msg.SrcAddress, msg.SrcTxLT)
	msg.DstState = s.getAccountStateForMessage(ctx, msg.DstAddress, msg.DstTxLT)

	update := *msg

	err := s.Parser.ParseMessagePayload(ctx, &update)
	if err != nil {
		if !errors.Is(err, app.ErrImpossibleParsing) {
			log.Error().Err(err).
				Hex("msg_hash", msg.Hash).
				Hex("src_tx_hash", msg.SrcTxHash).
				Str("src_addr", msg.SrcAddress.String()).
				Hex("dst_tx_hash", msg.DstTxHash).
				Str("dst_addr", msg.DstAddress.String()).
				Uint32("op_id", msg.OperationID).
				Msg("parse message payload")
		}
		return nil
	}

	if reflect.DeepEqual(msg, &update) {
		return nil
	}

	return &update
}

func (s *Service) rescanMessagesInBlock(ctx context.Context, b *core.Block) (updates []*core.Message) {
	for _, tx := range b.Transactions {
		if tx.InMsg != nil {
			if got := s.rescanMessage(ctx, tx.InMsg); got != nil {
				updates = append(updates, got)
			}
		}
		for _, out := range tx.OutMsg {
			if got := s.rescanMessage(ctx, out); got != nil {
				updates = append(updates, got)
			}
		}
	}
	return updates
}

func (s *Service) rescanMessagesWorker(m *core.Block) (updates []*core.Message) {
	for _, shard := range m.Shards {
		upd := s.rescanMessagesInBlock(context.Background(), shard)
		updates = append(updates, upd...)
	}

	upd := s.rescanMessagesInBlock(context.Background(), m)
	updates = append(updates, upd...)

	return updates
}

func (s *Service) rescanMessages(masterBlocks []*core.Block) (lastScanned uint32) {
	var (
		msgUpdates = make(chan []*core.Message, len(masterBlocks))
		scanWG     sync.WaitGroup
	)

	scanWG.Add(len(masterBlocks))

	for _, b := range masterBlocks {
		go func(master *core.Block) {
			defer scanWG.Done()
			msgUpdates <- s.rescanMessagesWorker(master)
		}(b)

		if b.SeqNo > lastScanned {
			lastScanned = b.SeqNo
		}
	}

	go func() {
		scanWG.Wait()
		close(msgUpdates)
	}()

	var allUpdates []*core.Message
	for upd := range msgUpdates {
		allUpdates = append(allUpdates, upd...)
	}

	if err := s.MessageRepo.UpdateMessages(context.Background(), allUpdates); err != nil {
		return 0
	}

	return lastScanned
}
