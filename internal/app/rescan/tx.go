package rescan

import (
	"context"
	"reflect"
	"sync"

	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"

	"github.com/tonindexer/anton/internal/app"
	"github.com/tonindexer/anton/internal/core"
)

func (s *Service) rescanMessage(ctx context.Context, master core.BlockID, msg *core.Message) *core.Message {
	// we must get account state's interfaces to properly determine message operation
	// and to parse message accordingly

	// so for the source of the message we take the account state of the sender,
	// which was updated just before the message was sent

	// for the destination of the given message, we take the account state of receiver,
	// which was update just after the message was received

	if msg.SrcState == nil {
		msg.SrcState, _ = s.getRecentAccountState(ctx,
			master,
			core.BlockID{
				Workchain: msg.SrcWorkchain,
				Shard:     msg.SrcShard,
				SeqNo:     msg.SrcBlockSeqNo,
			},
			msg.SrcAddress,
			false)
	}
	if msg.DstState == nil {
		msg.DstState, _ = s.getRecentAccountState(ctx,
			master,
			core.BlockID{
				Workchain: msg.DstWorkchain,
				Shard:     msg.DstShard,
				SeqNo:     msg.DstBlockSeqNo,
			},
			msg.DstAddress,
			true)
	}

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

func (s *Service) rescanMessagesInBlock(ctx context.Context, master, b *core.Block) (updates []*core.Message) {
	for _, tx := range b.Transactions {
		tx.InMsg.DstState = tx.Account
		if got := s.rescanMessage(ctx, master.ID(), tx.InMsg); got != nil {
			updates = append(updates, got)
		}

		for _, out := range tx.OutMsg {
			out.SrcState = tx.Account
			if got := s.rescanMessage(ctx, master.ID(), out); got != nil {
				updates = append(updates, got)
			}
		}
	}
	return updates
}

func (s *Service) rescanMessagesWorker(m *core.Block) (updates []*core.Message) {
	for _, shard := range m.Shards {
		upd := s.rescanMessagesInBlock(context.Background(), m, shard)
		updates = append(updates, upd...)
	}

	upd := s.rescanMessagesInBlock(context.Background(), m, m)
	updates = append(updates, upd...)

	return updates
}

func (s *Service) rescanMessages(masterBlocks []*core.Block) (lastScanned uint32) {
	var (
		msgUpdates chan []*core.Message
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
