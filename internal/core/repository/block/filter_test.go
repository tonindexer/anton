package block_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/tonindexer/anton/internal/core"
	"github.com/tonindexer/anton/internal/core/filter"
	"github.com/tonindexer/anton/internal/core/repository/tx"
	"github.com/tonindexer/anton/internal/core/rndm"
)

func TestRepository_FilterBlocks(t *testing.T) {
	initdb(t)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	master := rndm.MasterBlock()
	master.TransactionsCount = 10

	shards := rndm.Blocks(0, 100)
	shard := shards[len(shards)-1]
	shard.MasterID = &core.BlockID{
		Workchain: master.Workchain,
		Shard:     master.Shard,
		SeqNo:     master.SeqNo,
	}
	shard.TransactionsCount = 20

	master.Shards = append(master.Shards, shard)

	nextSeqNo := shard.SeqNo + 1

	t.Run("drop tables", func(t *testing.T) {
		dropTables(t)
	})

	t.Run("create tables", func(t *testing.T) {
		createTables(t)
	})

	t.Run("add block", func(t *testing.T) {
		dbTx, err := pg.Begin()
		require.Nil(t, err)

		err = repo.AddBlocks(ctx, dbTx, shards)
		require.Nil(t, err)

		err = repo.AddBlocks(ctx, dbTx, []*core.Block{master})
		require.Nil(t, err)

		err = dbTx.Commit()
		require.Nil(t, err)
	})

	t.Run("add some transactions", func(t *testing.T) {
		txRepo := tx.NewRepository(ck, pg)

		dbTx, err := pg.Begin()
		require.Nil(t, err)

		masterTx := rndm.BlockTransactions(*shard.MasterID, master.TransactionsCount)
		shardTx := rndm.BlockTransactions(shard.ID(), shard.TransactionsCount)

		err = txRepo.AddTransactions(ctx, dbTx, masterTx)
		require.Nil(t, err)

		err = txRepo.AddTransactions(ctx, dbTx, shardTx)
		require.Nil(t, err)

		err = dbTx.Commit()
		require.Nil(t, err)
	})

	t.Run("filter by workchain", func(t *testing.T) {
		res, err := repo.FilterBlocks(ctx, &filter.BlocksReq{
			Workchain: &shard.Workchain,
			// Shard:     &shard.Shard,
			// SeqNo:     &shard.SeqNo,

			AfterSeqNo: &nextSeqNo, Order: "DESC", Limit: 1,
		})
		require.Nil(t, err)
		require.Equal(t, 100, res.Total)
		require.Equal(t, []*core.Block{shard}, res.Rows)
	})

	t.Run("filter by seq no", func(t *testing.T) {
		res, err := repo.FilterBlocks(ctx, &filter.BlocksReq{
			Workchain: &shard.Workchain,
			SeqNo:     &shard.SeqNo,

			AfterSeqNo: &nextSeqNo, Order: "DESC", Limit: 1,
		})
		require.Nil(t, err)
		require.Equal(t, 1, res.Total)
		require.Equal(t, []*core.Block{shard}, res.Rows)
	})

	t.Run("filter by file hash", func(t *testing.T) {
		res, err := repo.FilterBlocks(ctx, &filter.BlocksReq{
			FileHash: master.FileHash,

			WithShards: true,

			AfterSeqNo: &nextSeqNo, Order: "DESC", Limit: 1,
		})
		require.Nil(t, err)
		require.Equal(t, 1, res.Total)
		require.Equal(t, []*core.Block{master}, res.Rows)
	})

	t.Run("drop tables again", func(t *testing.T) {
		dropTables(t)
	})
}
