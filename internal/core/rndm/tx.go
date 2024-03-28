package rndm

import (
	"math/rand"
	"time"

	"github.com/tonindexer/anton/addr"
	"github.com/tonindexer/anton/internal/core"
)

var (
	txTS        = time.Now().UTC()
	txLT uint64 = 80000
)

func BlockTransaction(b core.BlockID) *core.Transaction {
	txTS = txTS.Add(time.Minute)
	txLT++

	return &core.Transaction{
		Address:     *Address(),
		Hash:        Bytes(32),
		Account:     nil,
		Workchain:   b.Workchain,
		Shard:       b.Shard,
		BlockSeqNo:  b.SeqNo,
		PrevTxHash:  Bytes(32),
		PrevTxLT:    rand.Uint64(),
		InMsgHash:   Bytes(32),
		InAmount:    BigInt(),
		OutMsgCount: uint16(rand.Int() % 32),
		OutAmount:   BigInt(),
		TotalFees:   BigInt(),
		Description: Bytes(256),
		OrigStatus:  core.Active,
		EndStatus:   core.Active,
		CreatedAt:   txTS,
		CreatedLT:   txLT,
	}
}

func BlockTransactions(b core.BlockID, n int) (ret []*core.Transaction) {
	for i := 0; i < n; i++ {
		ret = append(ret, BlockTransaction(b))
	}
	return
}

func AddressTransaction(a *addr.Address) *core.Transaction {
	tx := BlockTransaction(BlockID(0))
	tx.Address = *a
	return tx
}

func AddressTransactions(a *addr.Address, n int) (ret []*core.Transaction) {
	for i := 0; i < n; i++ {
		ret = append(ret, AddressTransaction(a))
	}
	return
}

func Transaction() *core.Transaction {
	return BlockTransaction(core.BlockID{
		Workchain: 0,
		Shard:     int64(rand.Uint64()),
		SeqNo:     rand.Uint32(),
	})
}

func Transactions(n int) (ret []*core.Transaction) {
	for i := 0; i < n; i++ {
		ret = append(ret, Transaction())
	}
	return
}
