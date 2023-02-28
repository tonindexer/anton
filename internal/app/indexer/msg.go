// INSERT INTO "message_payloads" ("type", "hash", "src_address", "src_contract", "dst_address", "dst_contract", "body_hash", "operation_id", "operation_name", "data_json", "created_at", "created_lt") VALUES ('INTERNAL', '\x84ef22fdbc868311d2752721d07d56d8a6090ef68d95b993444a103bfb8c4b74', '\x110080d78a35f955a14b679faa887ff4cd5bfc0f43b4a4eea2a7e6927f3701b273c2', 'telemint_nft_collection', '\x1100590b2fa87461b9c3d8a1fea930d2df468a8600f2bb419b6a99b3b19924b516a1', '', '\x8262fa50cd03cfdb2208934eda38ec61bde7244c1d198251fe07a79adb4ce29d', 697974293, 'teleitem_msg_deploy', '{"Op":{},"SenderAddress":"EQAxHsyesi5hZ0Y57t8C4QO_nXjmMhJ9wepqdXX-zXpTKBR4","Bid":"296000000000","Info":{"Name":{"Len":6,"Text":"geoint"},"Domain":{"Len":5,"Text":"me\u0000t\u0000"}},"Content":{},"AuctionConfig":{"BeneficiaryAddress":"EQBAjaOyi2wGWlk-EDkSabqqnF-MrrwMadnwqrurKpkla9nE","InitialMinBid":"296000000000","MaxBid":"0","MinBidStep":5,"MinExtendTime":3600,"Duration":604800},"RoyaltyParams":{"Numerator":5,"Denominator":100,"Destination":"EQBAjaOyi2wGWlk-EDkSabqqnF-MrrwMadnwqrurKpkla9nE"}}', 1668333738, 32818576000004)
package indexer

import (
	"context"
	"fmt"
	"time"

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

	defer timeTrack(time.Now(), fmt.Sprintf("getLatestAccount(%s)", a.Base64()))

	state, err := s.accountRepo.GetAccountStates(ctx, &core.AccountStateFilter{
		Address:     &a,
		LatestState: true,
		WithData:    true,
	}, 0, 1)
	if err != nil {
		return nil, errors.Wrap(err, "get account data")
	}
	if len(state) > 0 && state[0].StateData != nil {
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
		if err != nil && !errors.Is(err, core.ErrNotFound) {
			log.Error().Err(err).Hex("hash", msg.Hash).Hex("tx_hash", msg.SourceTxHash).
				Str("src_addr", msg.SrcAddress.Base64()).Msg("cannot find src account")
		}
		dst, err := s.getLatestAccount(ctx, msg.DstAddress, accountMap)
		if err != nil && !errors.Is(err, core.ErrNotFound) {
			log.Error().Err(err).Hex("hash", msg.Hash).Hex("tx_hash", msg.SourceTxHash).
				Str("dst_addr", msg.DstAddress.Base64()).Msg("cannot find dst account")
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
