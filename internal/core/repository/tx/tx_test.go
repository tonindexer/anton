package tx

import (
	"context"
	"math/big"
	"testing"

	"github.com/pkg/errors"
	"github.com/uptrace/go-clickhouse/ch"

	"github.com/iam047801/tonidx/internal/core"
	"github.com/iam047801/tonidx/internal/core/db"
)

var ctx = context.Background()

var _db *ch.DB

func chdb(t *testing.T) *ch.DB {
	if _db != nil {
		return _db
	}

	database, err := db.Connect(context.Background(), "clickhouse://localhost:9000/test?sslmode=disable")
	if err != nil {
		t.Fatal(err)
	}

	_db = database
	return _db
}

var _txRepo *Repository

func txRepo(t *testing.T) *Repository {
	if _txRepo != nil {
		return _txRepo
	}

	_txRepo = NewRepository(chdb(t))
	return _txRepo
}

func TestTxRepository_AddMessage(t *testing.T) {
	txHash := core.RandBytes()
	addr, from, to1, to2 := "asdf", "sdfg", "dfgh", "fghj"

	inMsg := &core.Message{
		TxHash:        txHash,
		BodyHash:      core.RandBytes(),
		SrcAddr:       from,
		DstAddr:       addr,
		Amount:        big.NewInt(1).Uint64(),
		FwdFee:        big.NewInt(3).Uint64(),
		CreatedLT:     4,
		CreatedAt:     5,
		StateInitCode: []byte("a"),
		StateInitData: []byte("a"),
		Body:          []byte("b"),
		SourceTxHash:  core.RandBytes(),
	}

	outMsg1 := &core.Message{
		TxHash:    txHash,
		BodyHash:  core.RandBytes(),
		SrcAddr:   addr,
		DstAddr:   to1,
		Amount:    big.NewInt(1).Uint64(),
		FwdFee:    big.NewInt(3).Uint64(),
		CreatedLT: 4,
		CreatedAt: 5,
		Body:      []byte("b"),
	}
	outMsg2 := &core.Message{
		TxHash:    txHash,
		BodyHash:  core.RandBytes(),
		SrcAddr:   addr,
		DstAddr:   to2,
		Amount:    big.NewInt(1).Uint64(),
		FwdFee:    big.NewInt(3).Uint64(),
		CreatedLT: 5,
		CreatedAt: 6,
		Body:      []byte("b"),
	}

	tx := &core.Transaction{
		AccountAddr: addr,
		Hash:        txHash,
		CreatedLT:   4,
		TotalFees:   big.NewInt(3).Uint64(),
	}

	if err := txRepo(t).AddMessages(ctx, []*core.Message{outMsg1, outMsg2}); err != nil {
		t.Fatal(err)
	}
	if err := txRepo(t).AddMessages(ctx, []*core.Message{inMsg}); err != nil {
		t.Fatal(err)
	}
	if err := txRepo(t).AddTransactions(ctx, []*core.Transaction{tx}); err != nil {
		t.Fatal(err)
	}

	_, err := txRepo(t).GetMessageByHash(ctx, outMsg1.BodyHash)
	if err != nil {
		t.Fatal(errors.Wrapf(err, "cannot find message with hash %s", outMsg1.BodyHash))
	}

	_, err = txRepo(t).GetInMessageByTxHash(ctx, txHash)
	if err != nil {
		t.Fatal(err)
	}

	_, err = txRepo(t).GetOutMessagesByTxHash(ctx, txHash)
	if err != nil {
		t.Fatal(err)
	}
}
