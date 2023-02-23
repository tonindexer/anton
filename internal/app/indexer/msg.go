package indexer

import (
	"context"

	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
	"github.com/uptrace/bun"
	"github.com/xssnick/tonutils-go/tlb"

	"github.com/iam047801/tonidx/internal/core"
)

func (s *Service) messageAlreadyKnown(ctx context.Context, tx bun.Tx, in *core.Message, outMsgMap map[uint64]*core.Message) (bool, error) {
	// do not duplicate messages in a database

	if _, ok := outMsgMap[in.CreatedLT]; in.Type == core.Internal && ok { // found outgoing internal message in the block
		return true, nil
	}

	res, err := s.txRepo.GetMessages(ctx, &core.MessageFilter{DBTx: &tx, Hash: in.Hash}, 0, 1)
	if err != nil {
		return false, errors.Wrap(err, "get messages")
	}
	if len(res) == 1 {
		return true, nil
	}

	return false, nil
}

func (s *Service) processBlockMessages(ctx context.Context, dbtx bun.Tx, b *tlb.BlockInfo, blockTx []*tlb.Transaction) ([]*core.Message, error) {
	var (
		inMessages  []*core.Message
		outMessages []*core.Message
		outMsgMap   = make(map[uint64]*core.Message)
	)

	for _, tx := range blockTx {
		for _, outMsg := range tx.IO.Out {
			msg, err := mapMessage(tx, outMsg)
			if err != nil {
				return nil, errors.Wrap(err, "map outcoming message")
			}

			if msg.Source, err = mapTransaction(b, tx); err != nil {
				return nil, errors.Wrapf(err, "map source transaction (tx_hash = %x, msg_hash = %x)", tx.Hash, msg.BodyHash)
			}
			msg.SourceTxHash = tx.Hash
			msg.SourceTxLT = tx.LT

			outMessages = append(outMessages, msg)
			if msg.Type == core.Internal {
				outMsgMap[msg.CreatedLT] = msg
			}
		}
	}

	for _, tx := range blockTx {
		if tx.IO.In == nil {
			continue
		}

		msg, err := mapMessage(tx, tx.IO.In)
		if err != nil {
			return nil, errors.Wrap(err, "map incoming message")
		}

		known, err := s.messageAlreadyKnown(ctx, dbtx, msg, outMsgMap)
		if err != nil {
			return nil, errors.Wrap(err, "is message already known")
		}
		if known {
			continue
		}

		inMessages = append(inMessages, msg)
	}

	return append(outMessages, inMessages...), nil
}

func (s *Service) parseMessagePayloads(ctx context.Context, messages []*core.Message, accountMap map[string]*core.AccountState) (ret []*core.MessagePayload) {
	for _, msg := range messages {
		if msg.Type != core.Internal {
			continue // TODO: external message parsing
		}

		if msg.SrcAddress == "Ef8AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAADAU" {
			continue
		}

		src, ok := accountMap[msg.SrcAddress]
		if !ok {
			log.Error().Str("src_addr", msg.SrcAddress).Msg("cannot find src account")
			continue
		}
		dst, ok := accountMap[msg.DstAddress]
		if !ok {
			log.Error().Str("dst_addr", msg.SrcAddress).Msg("cannot find dst account")
			continue
		}

		payload, err := s.parser.ParseMessagePayload(ctx, src, dst, msg)
		if errors.Is(err, core.ErrNotAvailable) {
			continue
		}
		if err != nil {
			log.Error().Err(err).Hex("msg_hash", msg.BodyHash).Hex("tx_hash", msg.SourceTxHash).Msg("parse message payload")
			continue
		}
		ret = append(ret, payload)
	}

	return ret
}
