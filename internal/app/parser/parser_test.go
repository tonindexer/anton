package parser

import (
	"context"
	"fmt"
	"testing"

	"github.com/pkg/errors"
	"github.com/xssnick/tonutils-go/address"
	"github.com/xssnick/tonutils-go/tlb"
	"github.com/xssnick/tonutils-go/ton"

	"github.com/tonindexer/anton/internal/app"
	"github.com/tonindexer/anton/internal/core/repository"
)

var _testService *Service

var ctx = context.Background()

func testService(t *testing.T) *Service {
	if _testService != nil {
		return _testService
	}

	bd, err := repository.ConnectDB(ctx,
		"clickhouse://localhost:9000/testing?sslmode=disable",
		"postgres://user:pass@localhost:5432/postgres?sslmode=disable")
	if err != nil {
		t.Fatal(err)
	}

	server := app.ServerAddr{
		IPPort:    "135.181.177.59:53312",
		PubKeyB64: "aF91CuUHuuOv9rm2W5+O/4h38M3sRm40DtSdRxQhmtQ=",
	}
	s, err := NewService(context.Background(), &app.ParserConfig{DB: bd, Servers: []*app.ServerAddr{&server}})
	if err != nil {
		t.Fatal(err)
	}

	_testService = s
	return _testService
}

func getCurrentMaster(t *testing.T) *ton.BlockIDExt {
	s := testService(t)

	master, err := s.api.GetMasterchainInfo(ctx)
	if err != nil {
		t.Fatal(errors.Wrap(err, "cannot get masterchain info"))
	}
	master, err = s.api.LookupBlock(ctx, master.Workchain, master.Shard, master.SeqNo)
	if err != nil {
		t.Fatal(errors.Wrap(err, "lookup block"))
	}

	return master
}

func getTransactionOnce(t *testing.T, addr *address.Address, lt uint64, txHash []byte) *tlb.Transaction {
	transactions, err := testService(t).api.ListTransactions(ctx, addr, 1, lt, txHash)
	if err != nil {
		t.Fatal(err)
	}
	if len(transactions) == 0 {
		t.Fatal(fmt.Errorf("no transactions"))
	}
	return transactions[0]
}

func TestGetMasterchainInfo(t *testing.T) {
	m := getCurrentMaster(t)
	t.Logf("Latest master chain block: %d", m.SeqNo)
}
