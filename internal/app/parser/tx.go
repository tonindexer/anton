package parser

import (
	"context"
	"encoding/json"
	"reflect"

	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
	"github.com/xssnick/tonutils-go/address"
	"github.com/xssnick/tonutils-go/tlb"
	"github.com/xssnick/tonutils-go/tvm/cell"

	"github.com/iam047801/tonidx/internal/core"
)

func (s *Service) parseTransaction(_ context.Context, b *tlb.BlockInfo, raw *tlb.Transaction) *core.Transaction {
	addr := address.NewAddress(0, byte(b.Workchain), raw.AccountAddr)

	tx := &core.Transaction{
		Hash:    raw.Hash,
		Address: addr.String(),

		BlockWorkchain: b.Workchain,
		BlockShard:     b.Shard,
		BlockSeqNo:     b.SeqNo,

		PrevTxHash: raw.PrevTxHash,
		PrevTxLT:   raw.PrevTxLT,

		InMsgBodyHash:    raw.IO.In.Msg.Payload().Hash(),
		OutMsgBodyHashes: nil,

		TotalFees: raw.TotalFees.Coins.NanoTON().Uint64(),

		OrigStatus: core.AccountStatus(raw.OrigStatus),
		EndStatus:  core.AccountStatus(raw.EndStatus),

		CreatedLT: raw.LT,
		CreatedAT: uint64(raw.Now),
	}
	for _, out := range raw.IO.Out {
		tx.OutMsgBodyHashes = append(tx.OutMsgBodyHashes, out.Msg.Payload().Hash())
	}
	if raw.StateUpdate != nil {
		tx.StateUpdate = raw.StateUpdate.ToBOC()
	}
	if raw.Description != nil {
		tx.Description = raw.Description.ToBOC()
	}

	return tx
}

func (s *Service) ParseBlockTransactions(ctx context.Context, b *tlb.BlockInfo, blockTx []*tlb.Transaction) ([]*core.Transaction, error) {
	var transactions []*core.Transaction

	for _, raw := range blockTx {
		tx := s.parseTransaction(ctx, b, raw)
		transactions = append(transactions, tx)
	}

	return transactions, nil
}

func (s *Service) parseMessage(incoming bool, tx *tlb.Transaction, message *tlb.Message) *core.Message {
	msg := new(core.Message)

	switch raw := message.Msg.(type) {
	case *tlb.InternalMessage:
		msg.Type = core.Internal

		msg.Incoming = incoming
		msg.SrcAddress = raw.SrcAddr.String()
		msg.DstAddress = raw.DstAddr.String()

		msg.Bounce = raw.Bounce
		msg.Bounced = raw.Bounced

		msg.Amount = raw.Amount.NanoTON().Uint64()

		msg.IHRDisabled = raw.IHRDisabled
		msg.IHRFee = raw.IHRFee.NanoTON().Uint64()
		msg.FwdFee = raw.FwdFee.NanoTON().Uint64()

		msg.Body = raw.Body.ToBOC()
		msg.BodyHash = raw.Body.Hash()

		if raw.StateInit != nil {
			msg.StateInitCode = raw.StateInit.Code.ToBOC()
			msg.StateInitData = raw.StateInit.Data.ToBOC()
		}

		msg.CreatedLT = raw.CreatedLT
		msg.CreatedAt = uint64(raw.CreatedAt)

	case *tlb.ExternalMessage:
		msg.Type = core.ExternalIn

		msg.Incoming = true
		msg.DstAddress = raw.DstAddr.String()

		if raw.StateInit != nil {
			msg.StateInitCode = raw.StateInit.Code.ToBOC()
			msg.StateInitData = raw.StateInit.Data.ToBOC()
		}

		msg.Body = raw.Body.ToBOC()
		msg.BodyHash = raw.Body.Hash()

		msg.CreatedLT = tx.LT
		msg.CreatedAt = uint64(tx.Now)

	case *tlb.ExternalMessageOut:
		msg.Type = core.ExternalOut

		msg.Incoming = false
		msg.SrcAddress = raw.SrcAddr.String()

		msg.CreatedLT = raw.CreatedLT
		msg.CreatedAt = uint64(raw.CreatedAt)

		if raw.StateInit != nil {
			msg.StateInitCode = raw.StateInit.Code.ToBOC()
			msg.StateInitData = raw.StateInit.Data.ToBOC()
		}

		msg.Body = raw.Body.ToBOC()
		msg.BodyHash = raw.Body.Hash()
	}

	if msg.Incoming {
		msg.TxAddress = msg.DstAddress
	} else {
		msg.TxAddress = msg.SrcAddress
	}
	msg.TxHash = tx.Hash

	return msg
}

func (s *Service) getMsgSourceHash(ctx context.Context, in *core.Message, outMsgMap map[string]*core.Message) ([]byte, error) {
	if !in.Incoming || in.Type != core.Internal {
		return nil, errors.Wrap(core.ErrNotAvailable, "msg is not incoming or internal")
	}

	out, ok := outMsgMap[string(in.BodyHash)]
	if ok {
		return out.TxHash, nil
	}

	// TODO: that's totally wrong, it's possible to match message only by source tx hash
	sourceMsg, err := s.txRepo.GetMessageByHash(ctx, in.BodyHash) // TODO: batch request (?)
	if err != nil {
		log.Error().Err(err).Hex("tx_hash", in.TxHash).Hex("body_hash", in.BodyHash).Msg("get source msg")
		return nil, errors.Wrap(core.ErrNotAvailable, err.Error()) // TODO: fail on this err
	}

	return sourceMsg.TxHash, nil
}

func (s *Service) parseOperation(msg *core.Message) error {
	payload, err := cell.FromBOC(msg.Body)
	if err != nil {
		return errors.Wrap(err, "msg body from boc")
	}

	slice := payload.BeginParse()

	op, _ := slice.LoadUInt(32)
	msg.OperationID = uint32(op)

	if msg.OperationID == 0 {
		// simple transfer with comment
		msg.TransferComment, _ = slice.LoadStringSnake()
	}

	return nil
}

func (s *Service) ParseBlockMessages(ctx context.Context, _ *tlb.BlockInfo, blockTx []*tlb.Transaction) ([]*core.Message, error) {
	var (
		inMessages  []*core.Message
		outMessages []*core.Message
		outMsgMap   = make(map[string]*core.Message)
		err         error
	)

	for _, tx := range blockTx {
		for _, outMsg := range tx.IO.Out {
			msg := s.parseMessage(false, tx, outMsg)
			if err = s.parseOperation(msg); err != nil {
				return nil, errors.Wrapf(err, "parse operation (tx_hash = %x, msg_hash = %x)", tx.Hash, msg.BodyHash)
			}
			outMessages = append(outMessages, msg)
			outMsgMap[string(msg.BodyHash)] = msg
		}
	}

	for _, tx := range blockTx {
		if tx.IO.In == nil {
			continue
		}
		msg := s.parseMessage(true, tx, tx.IO.In)
		msg.SourceTxHash, err = s.getMsgSourceHash(ctx, msg, outMsgMap)
		if err != nil && !errors.Is(err, core.ErrNotAvailable) {
			if !errors.Is(err, core.ErrNotFound) {
				return nil, errors.Wrapf(err, "get source hash (tx_hash = %x)", tx.Hash)
			}
			log.Error().Err(err).Hex("tx_hash", tx.Hash).Msg("cannot get msg source hash")
		}
		if err = s.parseOperation(msg); err != nil {
			return nil, errors.Wrapf(err, "parse operation (tx_hash = %x, msg_hash = %x)", tx.Hash, msg.BodyHash)
		}
		inMessages = append(inMessages, msg)
	}

	return append(outMessages, inMessages...), nil
}

func (s *Service) parseDirectedMessage(ctx context.Context, acc *core.Account, message *core.Message, ret *core.MessagePayload) error {
	if len(acc.Types) == 0 {
		return errors.Wrap(core.ErrNotAvailable, "no interfaces")
	}

	operation, err := s.accountRepo.GetContractOperationByID(ctx, acc, acc.Address == message.SrcAddress, message.OperationID)
	if errors.Is(err, core.ErrNotFound) {
		return errors.Wrap(core.ErrNotAvailable, "unknown operation")
	}
	if err != nil {
		return errors.Wrap(err, "get contract operations")
	}
	ret.OperationName = operation.Name

	if acc.Address == message.SrcAddress {
		ret.SrcContract = operation.ContractName
	} else {
		ret.DstContract = operation.ContractName
	}

	payloadCell, err := cell.FromBOC(message.Body)
	if err != nil {
		return errors.Wrap(err, "msg body from boc")
	}
	payloadSlice := payloadCell.BeginParse()

	parsed := reflect.New(reflect.StructOf(operation.StructSchema)).Interface()
	if err = tlb.LoadFromCell(parsed, payloadSlice); err != nil {
		// return errors.Wrapf(core.ErrNotAvailable, "load from cell (%s)", err.Error())
		return errors.Wrap(err, "load from cell")
	}

	parsedJSON, err := json.Marshal(parsed)
	if err != nil {
		return errors.Wrap(err, "json marshal parsed payload")
	}
	ret.DataJSON = string(parsedJSON)

	return nil
}

func (s *Service) ParseMessagePayload(ctx context.Context, src, dst *core.Account, message *core.Message) (*core.MessagePayload, error) {
	// you can parse separately incoming messages to known contracts and outgoing message from them

	ret := &core.MessagePayload{
		TxHash:      message.TxHash,
		BodyHash:    message.BodyHash,
		SrcAddress:  message.SrcAddress,
		DstAddress:  message.DstAddress,
		OperationID: message.OperationID,
	}
	if len(message.Body) == 0 {
		return nil, errors.Wrap(core.ErrNotAvailable, "no message body")
	}

	err := s.parseDirectedMessage(ctx, dst, message, ret)
	if err != nil && !errors.Is(err, core.ErrNotAvailable) {
		log.Warn().
			Err(err).
			Hex("tx_hash", message.TxHash).
			Str("dst_addr", dst.Address).
			Strs("dst_types", dst.Types).
			Uint32("op_id", message.OperationID).Msg("parse dst message")
	}
	if err == nil {
		return ret, nil
	}

	err = s.parseDirectedMessage(ctx, src, message, ret)
	if err != nil && !errors.Is(err, core.ErrNotAvailable) {
		log.Warn().
			Err(err).
			Hex("tx_hash", message.TxHash).
			Str("src_addr", src.Address).
			Strs("src_types", src.Types).
			Uint32("op_id", message.OperationID).Msg("parse src message")
	}
	if err == nil {
		return ret, nil
	}

	return nil, err
}
