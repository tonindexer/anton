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
		AccountAddr: addr.String(),
		Hash:        raw.Hash,
		LT:          raw.LT,
		PrevTxHash:  raw.PrevTxHash,
		PrevTxLT:    raw.PrevTxLT,
		Now:         raw.Now,
		OutMsgCount: raw.OutMsgCount,
		TotalFees:   raw.TotalFees.Coins.NanoTON().Uint64(),
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

func (s *Service) parseMessage(incoming bool, txHash []byte, message *tlb.Message) *core.Message {
	msg := new(core.Message)

	switch raw := message.Msg.(type) {
	case *tlb.InternalMessage:
		msg.Type = core.Internal
		msg.TxHash = txHash
		msg.Incoming = incoming
		msg.Hash = raw.Body.Hash()
		msg.SrcAddr = raw.SrcAddr.String()
		msg.DstAddr = raw.DstAddr.String()
		msg.Bounce = raw.Bounce
		msg.Bounced = raw.Bounced
		msg.Amount = raw.Amount.NanoTON().Uint64()
		msg.IHRDisabled = raw.IHRDisabled
		msg.IHRFee = raw.IHRFee.NanoTON().Uint64()
		msg.FwdFee = raw.FwdFee.NanoTON().Uint64()
		msg.CreatedLT = raw.CreatedLT
		msg.CreatedAt = raw.CreatedAt
		if raw.StateInit != nil {
			msg.StateInitCode = raw.StateInit.Code.ToBOC()
			msg.StateInitData = raw.StateInit.Data.ToBOC()
		}
		msg.Body = raw.Body.ToBOC()
		msg.BodyHash = raw.Body.Hash()

	case *tlb.ExternalMessage:
		msg.Type = core.ExternalIn
		msg.TxHash = txHash
		msg.Incoming = true
		msg.Hash = raw.Body.Hash()
		msg.DstAddr = raw.DstAddr.String()
		if raw.StateInit != nil {
			msg.StateInitCode = raw.StateInit.Code.ToBOC()
			msg.StateInitData = raw.StateInit.Data.ToBOC()
		}
		msg.Body = raw.Body.ToBOC()
		msg.BodyHash = raw.Body.Hash()

	case *tlb.ExternalMessageOut:
		msg.Type = core.ExternalOut
		msg.TxHash = txHash
		msg.Incoming = false
		msg.Hash = raw.Body.Hash()
		msg.SrcAddr = raw.SrcAddr.String()
		msg.CreatedLT = raw.CreatedLT
		msg.CreatedAt = raw.CreatedAt
		if raw.StateInit != nil {
			msg.StateInitCode = raw.StateInit.Code.ToBOC()
			msg.StateInitData = raw.StateInit.Data.ToBOC()
		}
		msg.Body = raw.Body.ToBOC()
		msg.BodyHash = raw.Body.Hash()
	}

	return msg
}

func (s *Service) getMsgSourceHash(ctx context.Context, in *core.Message, outMsgMap map[string]*core.Message) ([]byte, error) {
	if !in.Incoming || in.Type != core.Internal {
		return nil, errors.Wrap(core.ErrNotAvailable, "msg is not incoming or internal")
	}

	out, ok := outMsgMap[string(in.Hash)]
	if ok {
		return out.TxHash, nil
	}

	sourceMsg, err := s.txRepo.GetMessageByHash(ctx, in.Hash) // TODO: batch request (?)
	if err != nil {
		// TODO: error
		// parse incoming message (tx_hash = b1d2d32a650b3a39fea2526f5247f00bed961a98bec9befb565f68a54883336f)
		// get source message by hash (msg_hash = dbe1c797d6d2d914c463173b1da531103d2efe747471212d1fad170fcf0c75d2)
		// master_seq: 23516077
		return nil, errors.Wrapf(err, "get source msg (hash = %x)", in.Hash)
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
			msg := s.parseMessage(false, tx.Hash, outMsg)
			if err = s.parseOperation(msg); err != nil {
				return nil, errors.Wrapf(err, "parse operation (tx_hash = %x, msg_hash = %x)", tx.Hash, msg.Hash)
			}
			outMessages = append(outMessages, msg)
			outMsgMap[string(msg.Hash)] = msg
		}
	}

	for _, tx := range blockTx {
		if tx.IO.In == nil {
			continue
		}
		msg := s.parseMessage(true, tx.Hash, tx.IO.In)
		msg.SourceTxHash, err = s.getMsgSourceHash(ctx, msg, outMsgMap)
		if err != nil && !errors.Is(err, core.ErrNotAvailable) {
			if !errors.Is(err, core.ErrNotFound) {
				return nil, errors.Wrapf(err, "get source hash (tx_hash = %x)", tx.Hash)
			}
			log.Error().Err(err).Hex("tx_hash", tx.Hash).Msg("cannot get msg source hash")
		}
		if err = s.parseOperation(msg); err != nil {
			return nil, errors.Wrapf(err, "parse operation (tx_hash = %x, msg_hash = %x)", tx.Hash, msg.Hash)
		}
		inMessages = append(inMessages, msg)
	}

	return append(outMessages, inMessages...), nil
}

func (s *Service) ParseMessagePayload(ctx context.Context, toAcc *core.Account, message *core.Message) (*core.MessagePayload, error) {
	var ifaces []core.ContractType

	ret := core.MessagePayload{
		TxHash:      message.TxHash,
		MsgHash:     message.Hash,
		DstAddr:     message.DstAddr,
		OperationID: message.OperationID,
	}
	if ret.OperationID == 0 {
		return nil, errors.Wrap(core.ErrNotAvailable, "no operation")
	}

	payloadCell, err := cell.FromBOC(message.Body)
	if err != nil {
		return nil, errors.Wrap(err, "msg body from boc")
	}
	payloadSlice := payloadCell.BeginParse()

	for _, t := range toAcc.Types {
		ifaces = append(ifaces, core.ContractType(t))
	}
	operation, err := s.accountRepo.GetContractOperationByID(ctx, ifaces, message.OperationID)
	if errors.Is(err, core.ErrNotFound) {
		return nil, errors.Wrap(core.ErrNotAvailable, "unknown operation")
	}
	if err != nil {
		return nil, errors.Wrap(err, "get contract operations")
	}
	ret.OperationName = operation.Name
	ret.ContractName = operation.ContractName

	parsed := reflect.New(reflect.StructOf(operation.StructSchema)).Interface()
	if err = tlb.LoadFromCell(parsed, payloadSlice); err != nil {
		return nil, errors.Wrapf(core.ErrNotAvailable, "load from cell (%s)", err.Error())
	}

	parsedJSON, err := json.Marshal(parsed)
	if err != nil {
		return nil, errors.Wrap(err, "json marshal parsed payload")
	}
	ret.DataJSON = string(parsedJSON)

	return &ret, nil
}
