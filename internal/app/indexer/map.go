package indexer

import (
	"context"

	"github.com/pkg/errors"
	"github.com/xssnick/tonutils-go/address"
	"github.com/xssnick/tonutils-go/tlb"

	"github.com/iam047801/tonidx/internal/core"
)

func getMsgHash(msg *tlb.Message) ([]byte, error) {
	msgCell, err := tlb.ToCell(msg.Msg)
	if err != nil {
		return nil, errors.Wrap(err, "cannot convert message to cell")
	}
	return msgCell.Hash(), nil
}

func mapTransaction(_ context.Context, b *tlb.BlockInfo, raw *tlb.Transaction) (*core.Transaction, error) {
	var err error

	tx := &core.Transaction{
		Hash:    raw.Hash,
		Address: address.NewAddress(0, byte(b.Workchain), raw.AccountAddr).String(),

		BlockWorkchain: b.Workchain,
		BlockShard:     b.Shard,
		BlockSeqNo:     b.SeqNo,
		BlockFileHash:  b.FileHash,

		PrevTxHash: raw.PrevTxHash,
		PrevTxLT:   raw.PrevTxLT,

		TotalFees: raw.TotalFees.Coins.NanoTON().Uint64(),

		OrigStatus: core.AccountStatus(raw.OrigStatus),
		EndStatus:  core.AccountStatus(raw.EndStatus),

		CreatedLT: raw.LT,
		CreatedAt: uint64(raw.Now),
	}
	if raw.IO.In != nil {
		tx.InMsgHash, err = getMsgHash(raw.IO.In)
		if err != nil {
			return nil, err
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

func mapTransactions(ctx context.Context, b *tlb.BlockInfo, blockTx []*tlb.Transaction) ([]*core.Transaction, error) {
	var transactions []*core.Transaction

	for _, raw := range blockTx {
		tx, err := mapTransaction(ctx, b, raw)
		if err != nil {
			return nil, err
		}
		transactions = append(transactions, tx)
	}

	return transactions, nil
}

func mapMessage(incoming bool, tx *tlb.Transaction, message *tlb.Message) (*core.Message, error) {
	msg := new(core.Message)

	msgCell, err := tlb.ToCell(message.Msg)
	if err != nil {
		return nil, errors.Wrap(err, "cannot convert message to cell")
	}
	msg.Hash = msgCell.Hash()

	msg.Incoming = incoming
	msg.TxHash = tx.Hash
	if msg.Incoming {
		msg.TxAddress = msg.DstAddress
	} else {
		msg.TxAddress = msg.SrcAddress
	}

	switch raw := message.Msg.(type) {
	case *tlb.InternalMessage:
		msg.Type = core.Internal

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

	return msg, nil
}

func mapAccount(acc *tlb.Account) *core.AccountState {
	ret := new(core.AccountState)

	ret.IsActive = acc.IsActive
	if acc.State != nil {
		if acc.State.Address != nil {
			ret.Address = acc.State.Address.String()
		}
		ret.Status = core.AccountStatus(acc.State.Status)
		ret.Balance = acc.State.Balance.NanoTON().Uint64()
		ret.StateHash = acc.State.StateHash
		if acc.State.StateInit != nil {
			ret.Depth = acc.State.StateInit.Depth
			if acc.State.StateInit.TickTock != nil {
				ret.Tick = acc.State.StateInit.TickTock.Tick
				ret.Tock = acc.State.StateInit.TickTock.Tock
			}
		}
	}
	if acc.Data != nil {
		ret.Data = acc.Data.ToBOC()
		ret.DataHash = acc.Data.Hash()
	}
	if acc.Code != nil {
		ret.Code = acc.Code.ToBOC()
		ret.CodeHash = acc.Data.Hash()
	}
	ret.LastTxLT = acc.LastTxLT
	ret.LastTxHash = acc.LastTxHash

	return ret
}
