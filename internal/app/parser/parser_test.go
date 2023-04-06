package parser

import (
	"context"
	"testing"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
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
	assert.Nil(t, err)

	server := app.ServerAddr{
		IPPort:    "135.181.177.59:53312",
		PubKeyB64: "aF91CuUHuuOv9rm2W5+O/4h38M3sRm40DtSdRxQhmtQ=",
	}
	s, err := NewService(context.Background(), &app.ParserConfig{DB: bd, Servers: []*app.ServerAddr{&server}})
	assert.Nil(t, err)

	_testService = s
	return _testService
}

func getCurrentMaster(t *testing.T) *ton.BlockIDExt {
	s := testService(t)

	master, err := s.api.GetMasterchainInfo(ctx)
	assert.Nil(t, errors.Wrap(err, "cannot get masterchain info"))

	master, err = s.api.LookupBlock(ctx, master.Workchain, master.Shard, master.SeqNo)
	assert.Nil(t, errors.Wrap(err, "lookup block"))

	return master
}

func getTransactionOnce(t *testing.T, addr *address.Address, lt uint64, txHash []byte) *tlb.Transaction {
	transactions, err := testService(t).api.ListTransactions(ctx, addr, 1, lt, txHash)
	assert.Nil(t, err)
	assert.NotEqual(t, 0, len(transactions))
	return transactions[0]
}

func TestGetMasterchainInfo(t *testing.T) {
	m := getCurrentMaster(t)
	t.Logf("Latest master chain block: %d", m.SeqNo)
}
