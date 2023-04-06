package tx_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/tonindexer/anton/internal/addr"
	"github.com/tonindexer/anton/internal/core"
	"github.com/tonindexer/anton/internal/core/filter"
	"github.com/tonindexer/anton/internal/core/rndm"
)

func TestFilterRepository(t *testing.T) {
	initdb(t)

	transactions := rndm.Transactions(10)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	t.Run("drop tables", func(t *testing.T) {
		dropTables(t)
	})

	t.Run("create tables", func(t *testing.T) {
		createTables(t)
	})

	t.Run("add transactions", func(t *testing.T) {
		dbtx, err := pg.Begin()
		assert.Nil(t, err)

		err = repo.AddTransactions(ctx, dbtx, transactions)
		assert.Nil(t, err)

		err = dbtx.Commit()
		assert.Nil(t, err)
	})

	t.Run("filter by hash", func(t *testing.T) {
		res, err := repo.FilterTransactions(ctx, &filter.TransactionsReq{
			Hash: transactions[0].Hash,
		})
		assert.Nil(t, err)
		assert.Equal(t, 1, res.Total)
		assert.Equal(t, transactions[0:1], res.Rows)
	})

	t.Run("filter by incoming message hash", func(t *testing.T) {
		res, err := repo.FilterTransactions(ctx, &filter.TransactionsReq{
			InMsgHash: transactions[0].InMsgHash,
		})
		assert.Nil(t, err)
		assert.Equal(t, 1, res.Total)
		assert.Equal(t, transactions[0:1], res.Rows)
	})

	t.Run("filter by addresses", func(t *testing.T) {
		res, err := repo.FilterTransactions(ctx, &filter.TransactionsReq{
			Addresses: []*addr.Address{&transactions[0].Address},
		})
		assert.Nil(t, err)
		assert.Equal(t, 1, res.Total)
		assert.Equal(t, transactions[0:1], res.Rows)
	})

	t.Run("filter by block id", func(t *testing.T) {
		res, err := repo.FilterTransactions(ctx, &filter.TransactionsReq{
			BlockID: &core.BlockID{
				Workchain: transactions[0].BlockWorkchain,
				Shard:     transactions[0].BlockShard,
				SeqNo:     transactions[0].BlockSeqNo,
			},
		})
		assert.Nil(t, err)
		assert.Equal(t, 1, res.Total)
		assert.Equal(t, transactions[0:1], res.Rows)
	})

	t.Run("filter by workchain", func(t *testing.T) {
		res, err := repo.FilterTransactions(ctx, &filter.TransactionsReq{
			Workchain: new(int32),
			Order:     "ASC",
			Limit:     len(transactions),
		})
		assert.Nil(t, err)
		assert.Equal(t, len(transactions), res.Total)
		assert.Equal(t, transactions, res.Rows)
	})

	t.Run("drop tables", func(t *testing.T) {
		dropTables(t)
	})
}
