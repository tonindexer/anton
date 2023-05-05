package indexer

import (
	"time"

	"github.com/pkg/errors"
	"github.com/uptrace/bun/extra/bunbig"
	"github.com/xssnick/tonutils-go/address"
	"github.com/xssnick/tonutils-go/tlb"
	"github.com/xssnick/tonutils-go/ton"
	"github.com/xssnick/tonutils-go/tvm/cell"

	"github.com/tonindexer/anton/abi"
	"github.com/tonindexer/anton/addr"
	"github.com/tonindexer/anton/internal/core"
)

func getMsgHash(msg *tlb.Message) ([]byte, error) {
	switch raw := msg.Msg.(type) { // TODO: fix ToCell marshal in tonutils-go
	case *tlb.InternalMessage:
		if raw.StateInit != nil {
			raw.StateInit.Lib = nil
		}
	case *tlb.ExternalMessage:
		if raw.StateInit != nil {
			raw.StateInit.Lib = nil
		}
	case *tlb.ExternalMessageOut:
		if raw.StateInit != nil {
			raw.StateInit.Lib = nil
		}
	}

	msgCell, err := tlb.ToCell(msg.Msg)
	if err != nil {
		return nil, errors.Wrap(err, "cannot convert message to cell")
	}

	return msgCell.Hash(), nil
}

func mapMessageInternal(msg *core.Message, raw *tlb.InternalMessage) error {
	msg.Type = core.Internal

	src := addr.MustFromTonutils(raw.SrcAddr)
	dst := addr.MustFromTonutils(raw.DstAddr)
	msg.SrcAddress = *src
	msg.DstAddress = *dst

	msg.Bounce = raw.Bounce
	msg.Bounced = raw.Bounced

	msg.Amount = bunbig.FromMathBig(raw.Amount.NanoTON())

	msg.IHRDisabled = raw.IHRDisabled
	msg.IHRFee = bunbig.FromMathBig(raw.IHRFee.NanoTON())
	msg.FwdFee = bunbig.FromMathBig(raw.FwdFee.NanoTON())

	msg.Body = raw.Body.ToBOC()
	msg.BodyHash = raw.Body.Hash()

	if raw.StateInit != nil && raw.StateInit.Code != nil {
		msg.StateInitCode = raw.StateInit.Code.ToBOC()
	}
	if raw.StateInit != nil && raw.StateInit.Data != nil {
		msg.StateInitData = raw.StateInit.Data.ToBOC()
	}

	msg.CreatedLT = raw.CreatedLT
	msg.CreatedAt = time.Unix(int64(raw.CreatedAt), 0)

	return nil
}

func mapMessageExternal(msg *core.Message, rawTx *tlb.Transaction, rawMsg *tlb.Message) error {
	switch raw := rawMsg.Msg.(type) {
	case *tlb.ExternalMessage:
		msg.Type = core.ExternalIn

		dst := addr.MustFromTonutils(raw.DstAddr)
		msg.DstAddress = *dst

		if raw.StateInit != nil && raw.StateInit.Code != nil {
			msg.StateInitCode = raw.StateInit.Code.ToBOC()
		}
		if raw.StateInit != nil && raw.StateInit.Data != nil {
			msg.StateInitData = raw.StateInit.Data.ToBOC()
		}

		msg.Body = raw.Body.ToBOC()
		msg.BodyHash = raw.Body.Hash()

		msg.CreatedLT = rawTx.LT
		msg.CreatedAt = time.Unix(int64(rawTx.Now), 0)

	case *tlb.ExternalMessageOut:
		msg.Type = core.ExternalOut

		src := addr.MustFromTonutils(raw.SrcAddr)
		msg.SrcAddress = *src

		msg.SourceTxHash = rawTx.Hash
		msg.SourceTxLT = rawTx.LT

		msg.CreatedLT = raw.CreatedLT
		msg.CreatedAt = time.Unix(int64(raw.CreatedAt), 0)

		if raw.StateInit != nil && raw.StateInit.Code != nil {
			msg.StateInitCode = raw.StateInit.Code.ToBOC()
		}
		if raw.StateInit != nil && raw.StateInit.Data != nil {
			msg.StateInitData = raw.StateInit.Data.ToBOC()
		}

		msg.Body = raw.Body.ToBOC()
		msg.BodyHash = raw.Body.Hash()
	}

	return nil
}

func parseOperationID(body []byte) (opId uint32, comment string, err error) {
	payload, err := cell.FromBOC(body)
	if err != nil {
		return 0, "", errors.Wrap(err, "msg body from boc")
	}
	slice := payload.BeginParse()

	op, err := slice.LoadUInt(32)
	if err != nil {
		return 0, "", errors.Wrap(err, "load uint")
	}

	if opId = uint32(op); opId != 0 {
		return opId, "", nil
	}

	// simple transfer with comment
	if comment, err = slice.LoadStringSnake(); err != nil {
		return 0, "", errors.Wrap(err, "load transfer comment")
	}

	return opId, comment, nil
}

func mapMessage(tx *tlb.Transaction, message *tlb.Message) (*core.Message, error) {
	var (
		msg = new(core.Message)
		err error
	)

	msg.Hash, err = getMsgHash(message)
	if err != nil {
		return nil, err
	}

	switch raw := message.Msg.(type) {
	case *tlb.InternalMessage:
		if err := mapMessageInternal(msg, raw); err != nil {
			return nil, err
		}

	case *tlb.ExternalMessage, *tlb.ExternalMessageOut:
		if err := mapMessageExternal(msg, tx, message); err != nil {
			return nil, err
		}
	}

	if msg.Body == nil {
		return msg, nil
	}

	msg.OperationID, msg.TransferComment, _ = parseOperationID(msg.Body)

	return msg, nil
}

func mapAccount(acc *tlb.Account) *core.AccountState {
	ret := new(core.AccountState)

	ret.IsActive = acc.IsActive
	ret.Status = core.NonExist
	if acc.State != nil {
		if acc.State.Address != nil {
			ret.Address = *addr.MustFromTonutils(acc.State.Address)
		}
		ret.Status = core.AccountStatus(acc.State.Status)
		ret.Balance = bunbig.FromMathBig(acc.State.Balance.NanoTON())
		ret.StateHash = acc.State.StateHash
	}
	if acc.Data != nil {
		ret.Data = acc.Data.ToBOC()
		ret.DataHash = acc.Data.Hash()
	}
	if acc.Code != nil {
		ret.Code = acc.Code.ToBOC()
		ret.CodeHash = acc.Code.Hash()
		ret.GetMethodHashes, _ = abi.GetMethodHashes(acc.Code)
	}
	ret.LastTxLT = acc.LastTxLT
	ret.LastTxHash = acc.LastTxHash

	return ret
}

func mapTransaction(b *ton.BlockIDExt, raw *tlb.Transaction) (*core.Transaction, error) {
	var err error

	tx := &core.Transaction{
		Hash: raw.Hash,

		PrevTxHash: raw.PrevTxHash,
		PrevTxLT:   raw.PrevTxLT,

		OutMsgCount: raw.OutMsgCount,

		InAmount:  bunbig.NewInt(),
		OutAmount: bunbig.NewInt(),

		TotalFees: bunbig.FromMathBig(raw.TotalFees.Coins.NanoTON()),

		OrigStatus: core.AccountStatus(raw.OrigStatus),
		EndStatus:  core.AccountStatus(raw.EndStatus),

		CreatedLT: raw.LT,
		CreatedAt: time.Unix(int64(raw.Now), 0),
	}
	if b != nil {
		tx.Address = *addr.MustFromTonutils(address.NewAddress(0, byte(b.Workchain), raw.AccountAddr))
		tx.BlockWorkchain = b.Workchain
		tx.BlockShard = b.Shard
		tx.BlockSeqNo = b.SeqNo
	}
	if raw.IO.In != nil && raw.IO.In.Msg != nil {
		tx.InMsgHash, err = getMsgHash(raw.IO.In)
		if err != nil {
			return nil, err
		}
		if in, ok := raw.IO.In.Msg.(*tlb.InternalMessage); ok {
			tx.InAmount = bunbig.FromMathBig(in.Amount.NanoTON())
		}
	}
	for _, m := range raw.IO.Out {
		if out, ok := m.Msg.(*tlb.InternalMessage); ok {
			tx.OutAmount = tx.OutAmount.Add(bunbig.FromMathBig(out.Amount.NanoTON()))
		}
	}
	if raw.StateUpdate != nil {
		tx.StateUpdate = raw.StateUpdate.ToBOC()
	}
	if raw.Description != nil {
		tx.Description = raw.Description.ToBOC()
	}

	return tx, nil
}

func mapTransactions(b *ton.BlockIDExt, blockTx []*tlb.Transaction) ([]*core.Transaction, error) {
	var transactions []*core.Transaction

	for _, raw := range blockTx {
		tx, err := mapTransaction(b, raw)
		if err != nil {
			return nil, err
		}
		transactions = append(transactions, tx)
	}

	return transactions, nil
}
