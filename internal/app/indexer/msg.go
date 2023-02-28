package indexer

import (
	"context"

	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
	"github.com/uptrace/bun"
	"github.com/xssnick/tonutils-go/tlb"

	"github.com/iam047801/tonidx/internal/addr"
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

		msg.Known, err = s.messageAlreadyKnown(ctx, dbtx, msg, outMsgMap)
		if err != nil {
			return nil, errors.Wrap(err, "is message already known")
		}

		inMessages = append(inMessages, msg)
	}

	return append(outMessages, inMessages...), nil
}

func (s *Service) messagePayloadAlreadyKnown(ctx context.Context, tx bun.Tx, in *core.Message, msgMap map[string]*core.MessagePayload) (bool, error) {
	// do not duplicate message payloads in a database

	if _, ok := msgMap[string(in.Hash)]; ok { // already parsed msg payload
		return true, nil
	}

	res, err := s.txRepo.GetMessages(ctx, &core.MessageFilter{DBTx: &tx, Hash: in.Hash}, 0, 1)
	if err != nil {
		return false, errors.Wrap(err, "get messages")
	}
	if len(res) > 0 && res[0].Payload != nil {
		return true, nil
	}

	return false, nil
}

func (s *Service) getLatestAccount(ctx context.Context, a addr.Address, accountMap map[addr.Address]*core.AccountData) (*core.AccountData, error) {
	src, ok := accountMap[a]
	if ok {
		return src, nil
	}

	state, err := s.accountRepo.GetAccountStates(ctx, &core.AccountStateFilter{
		Address:     &a,
		LatestState: true,
		WithData:    true,
	}, 0, 1)
	if err != nil {
		return nil, errors.Wrap(err, "get account data")
	}
	if len(state) > 0 && state[0].Data != nil {
		return state[0].StateData, nil
	}

	return nil, errors.Wrap(core.ErrNotFound, "no account data found")
}

func (s *Service) parseMessagePayloads(ctx context.Context, tx bun.Tx, messages []*core.Message, accountMap map[addr.Address]*core.AccountData) (ret []*core.MessagePayload) {
	msgMap := make(map[string]*core.MessagePayload)

	for _, msg := range messages {
		if msg.Type != core.Internal {
			continue // TODO: external message parsing
		}
		if msg.SrcAddress.Base64() == "Ef8AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAADAU" {
			continue
		}

		known, err := s.messagePayloadAlreadyKnown(ctx, tx, msg, msgMap)
		if err != nil {
			log.Error().Err(err).Hex("tx_hash", msg.SourceTxHash).Str("src_addr", msg.SrcAddress.Base64()).Msg("message payload already known")
			continue
		}
		if known {
			continue
		}

		src, err := s.getLatestAccount(ctx, msg.SrcAddress, accountMap)
		if err != nil {
			log.Warn().Err(err).Hex("hash", msg.Hash).Hex("tx_hash", msg.SourceTxHash).
				Str("src_addr", msg.SrcAddress.Base64()).Msg("cannot find src account")
			continue
		}
		dst, err := s.getLatestAccount(ctx, msg.DstAddress, accountMap)
		if err != nil {
			log.Warn().Err(err).Hex("hash", msg.Hash).Hex("tx_hash", msg.SourceTxHash).
				Str("dst_addr", msg.DstAddress.Base64()).Msg("cannot find dst account")
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

		msgMap[string(payload.Hash)] = payload
		ret = append(ret, payload)
	}

	return ret
}
