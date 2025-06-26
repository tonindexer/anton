package tx_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/tonindexer/anton/addr"
	"github.com/tonindexer/anton/internal/core"
	"github.com/tonindexer/anton/internal/core/filter"
	"github.com/tonindexer/anton/internal/core/rndm"
)

func TestRepository_FilterTransactions(t *testing.T) {
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
		require.Nil(t, err)

		err = repo.AddTransactions(ctx, dbtx, transactions)
		require.Nil(t, err)

		err = dbtx.Commit()
		require.Nil(t, err)
	})

	t.Run("filter by hash", func(t *testing.T) {
		res, err := repo.FilterTransactions(ctx, &filter.TransactionsReq{
			TransactionsFilter: filter.TransactionsFilter{
				Hash: transactions[0].Hash,
			},
			Count: true,
		})
		require.Nil(t, err)
		require.Equal(t, 1, res.Total)
		require.Equal(t, transactions[0:1], res.Rows)
	})

	t.Run("filter by incoming message hash", func(t *testing.T) {
		res, err := repo.FilterTransactions(ctx, &filter.TransactionsReq{
			TransactionsFilter: filter.TransactionsFilter{
				InMsgHash: transactions[0].InMsgHash,
			},
			Count: true,
		})
		require.Nil(t, err)
		require.Equal(t, 1, res.Total)
		require.Equal(t, transactions[0:1], res.Rows)
	})

	t.Run("filter by addresses", func(t *testing.T) {
		res, err := repo.FilterTransactions(ctx, &filter.TransactionsReq{
			TransactionsFilter: filter.TransactionsFilter{
				Addresses: []*addr.Address{&transactions[0].Address},
			},
			Count: true,
		})
		require.Nil(t, err)
		require.Equal(t, 1, res.Total)
		require.Equal(t, transactions[0:1], res.Rows)
	})

	t.Run("filter by block id", func(t *testing.T) {
		res, err := repo.FilterTransactions(ctx, &filter.TransactionsReq{
			TransactionsFilter: filter.TransactionsFilter{
				BlockID: &core.BlockID{
					Workchain: transactions[0].Workchain,
					Shard:     transactions[0].Shard,
					SeqNo:     transactions[0].BlockSeqNo,
				},
			},
			Count: true,
		})
		require.Nil(t, err)
		require.Equal(t, 1, res.Total)
		require.Equal(t, transactions[0:1], res.Rows)
	})

	t.Run("filter by workchain", func(t *testing.T) {
		res, err := repo.FilterTransactions(ctx, &filter.TransactionsReq{
			TransactionsFilter: filter.TransactionsFilter{
				Workchain: new(int32),
			},
			Order: "ASC",
			Limit: len(transactions),
			Count: true,
		})
		require.Nil(t, err)
		require.Equal(t, len(transactions), res.Total)
		require.Equal(t, transactions, res.Rows)
	})

	t.Run("drop tables again", func(t *testing.T) {
		dropTables(t)
	})
}
