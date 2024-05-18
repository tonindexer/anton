package fetcher

import (
	"time"

	"github.com/pkg/errors"
	"github.com/uptrace/bun/extra/bunbig"
	"github.com/xssnick/tonutils-go/address"
	"github.com/xssnick/tonutils-go/tlb"
	"github.com/xssnick/tonutils-go/ton"
	"github.com/xssnick/tonutils-go/tvm/cell"

	"github.com/tonindexer/anton/addr"
	"github.com/tonindexer/anton/internal/core"
)

func MapAccount(b *ton.BlockIDExt, acc *tlb.Account) *core.AccountState {
	ret := new(core.AccountState)

	if b != nil {
		ret.Workchain = b.Workchain
		ret.Shard = b.Shard
		ret.BlockSeqNo = b.SeqNo
	}

	ret.IsActive = acc.IsActive
	ret.Status = core.NonExist
	if acc.State != nil {
		if acc.State.Address != nil {
			ret.Address = *addr.MustFromTonutils(acc.State.Address)
		}
		ret.Status = core.AccountStatus(acc.State.Status)
		ret.Balance = bunbig.FromMathBig(acc.State.Balance.Nano())
		ret.StateHash = acc.State.StateHash
	}
	if acc.Data != nil {
		ret.Data = acc.Data.ToBOC()
		ret.DataHash = acc.Data.Hash()
	}
	if acc.Code != nil {
		ret.Code = acc.Code.ToBOC()
		ret.CodeHash = acc.Code.Hash()
	}
	ret.LastTxLT = acc.LastTxLT
	ret.LastTxHash = acc.LastTxHash

	return ret
}

func mapMessageInternal(msg *core.Message, raw *tlb.InternalMessage) error {
	msg.Type = core.Internal

	msg.SrcAddress = *addr.MustFromTonutils(raw.SrcAddr)
	msg.SrcWorkchain = int32(msg.SrcAddress.Workchain())
	msg.DstAddress = *addr.MustFromTonutils(raw.DstAddr)
	msg.DstWorkchain = int32(msg.DstAddress.Workchain())

	msg.Bounce = raw.Bounce
	msg.Bounced = raw.Bounced

	msg.Amount = bunbig.FromMathBig(raw.Amount.Nano())

	msg.IHRDisabled = raw.IHRDisabled
	msg.IHRFee = bunbig.FromMathBig(raw.IHRFee.Nano())
	msg.FwdFee = bunbig.FromMathBig(raw.FwdFee.Nano())

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

func mapMessageExternal(msg *core.Message, rawTx *tlb.Transaction, rawMsg tlb.Message) error {
	switch raw := rawMsg.Msg.(type) {
	case *tlb.ExternalMessage:
		msg.Type = core.ExternalIn

		msg.DstAddress = *addr.MustFromTonutils(raw.DstAddr)
		msg.DstWorkchain = int32(msg.DstAddress.Workchain())
		msg.DstTxLT, msg.DstTxHash = rawTx.LT, rawTx.Hash

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

		msg.SrcAddress = *addr.MustFromTonutils(raw.SrcAddr)
		msg.SrcWorkchain = int32(msg.SrcAddress.Workchain())
		msg.SrcTxLT, msg.SrcTxHash = rawTx.LT, rawTx.Hash

		if raw.StateInit != nil && raw.StateInit.Code != nil {
			msg.StateInitCode = raw.StateInit.Code.ToBOC()
		}
		if raw.StateInit != nil && raw.StateInit.Data != nil {
			msg.StateInitData = raw.StateInit.Data.ToBOC()
		}

		msg.Body = raw.Body.ToBOC()
		msg.BodyHash = raw.Body.Hash()

		msg.CreatedLT = raw.CreatedLT
		msg.CreatedAt = time.Unix(int64(raw.CreatedAt), 0)
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

func mapMessage(tx *tlb.Transaction, message tlb.Message) (*core.Message, error) {
	var (
		msg = new(core.Message)
		err error
	)

	msgCell, err := tlb.ToCell(message.Msg)
	if err != nil {
		return nil, errors.Wrap(err, "cannot convert message to cell")
	}
	msg.Hash = msgCell.Hash()

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

func mapTransactionComputePhase(phase tlb.ComputePhase, tx *core.Transaction) {
	if p, ok := phase.Phase.(tlb.ComputePhaseVM); ok {
		tx.ComputePhaseExitCode = p.Details.ExitCode
	}
}

func mapTransactionDescription(desc any, tx *core.Transaction) {
	switch d := desc.(type) {
	case tlb.TransactionDescriptionOrdinary:
		if d.ActionPhase != nil {
			tx.ActionPhaseResultCode = d.ActionPhase.ResultCode
		}
		mapTransactionComputePhase(d.ComputePhase, tx)

	case tlb.TransactionDescriptionTickTock:
		if d.ActionPhase != nil {
			tx.ActionPhaseResultCode = d.ActionPhase.ResultCode
		}
		mapTransactionComputePhase(d.ComputePhase, tx)

	case tlb.TransactionDescriptionSplitPrepare:
		if d.ActionPhase != nil {
			tx.ActionPhaseResultCode = d.ActionPhase.ResultCode
		}
		mapTransactionComputePhase(d.ComputePhase, tx)

	case tlb.TransactionDescriptionMergeInstall:
		if d.ActionPhase != nil {
			tx.ActionPhaseResultCode = d.ActionPhase.ResultCode
		}
		mapTransactionComputePhase(d.ComputePhase, tx)
	}
}

func mapTransaction(b *ton.BlockIDExt, raw *tlb.Transaction) (*core.Transaction, error) {
	tx := &core.Transaction{
		Hash: raw.Hash,

		PrevTxHash: raw.PrevTxHash,
		PrevTxLT:   raw.PrevTxLT,

		OutMsgCount: raw.OutMsgCount,

		InAmount:  bunbig.NewInt(),
		OutAmount: bunbig.NewInt(),

		TotalFees: bunbig.FromMathBig(raw.TotalFees.Coins.Nano()),

		OrigStatus: core.AccountStatus(raw.OrigStatus),
		EndStatus:  core.AccountStatus(raw.EndStatus),

		CreatedLT: raw.LT,
		CreatedAt: time.Unix(int64(raw.Now), 0),
	}
	if b != nil {
		tx.Address = *addr.MustFromTonutils(address.NewAddress(0, byte(b.Workchain), raw.AccountAddr))
		tx.Workchain = b.Workchain
		tx.Shard = b.Shard
		tx.BlockSeqNo = b.SeqNo
	}
	if raw.IO.In != nil && raw.IO.In.Msg != nil {
		in, err := mapMessage(raw, *raw.IO.In)
		if err != nil {
			return nil, errors.Wrap(err, "map incoming message")
		}
		in.DstTxLT, in.DstTxHash = tx.CreatedLT, tx.Hash
		in.DstWorkchain, in.DstShard, in.DstBlockSeqNo = b.Workchain, b.Shard, b.SeqNo
		tx.InMsg, tx.InMsgHash = in, in.Hash
		if in.Type == core.Internal {
			tx.InAmount = in.Amount
		}
	}
	if raw.IO.Out != nil {
		messages, err := raw.IO.Out.ToSlice()
		if err != nil {
			return nil, errors.Wrap(err, "getting outgoing tx messages")
		}
		for _, m := range messages {
			out, err := mapMessage(raw, m)
			if err != nil {
				return nil, errors.Wrap(err, "map outgoing message")
			}
			out.SrcTxLT, out.SrcTxHash = tx.CreatedLT, tx.Hash
			out.SrcWorkchain, out.SrcShard, out.SrcBlockSeqNo = b.Workchain, b.Shard, b.SeqNo
			tx.OutMsg = append(tx.OutMsg, out)
			if out.Type == core.Internal {
				tx.OutAmount = tx.OutAmount.Add(out.Amount)
			}
		}
	}
	if raw.Description.Description != nil {
		c, err := tlb.ToCell(raw.Description.Description)
		if err != nil {
			return nil, errors.Wrap(err, "tx description to cell")
		}
		tx.Description = c.ToBOC()
		mapTransactionDescription(raw.Description.Description, tx)
	}

	return tx, nil
}
