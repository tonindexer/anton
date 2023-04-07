package block_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/tonindexer/anton/internal/core"
	"github.com/tonindexer/anton/internal/core/filter"
	"github.com/tonindexer/anton/internal/core/rndm"
)

func TestRepository_FilterBlocks(t *testing.T) {
	initdb(t)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	master := rndm.MasterBlock()

	shards := rndm.Blocks(0, 100)
	shard := shards[len(shards)-1]
	shard.MasterID = &core.BlockID{
		Workchain: master.Workchain,
		Shard:     master.Shard,
		SeqNo:     master.SeqNo,
	}

	master.Shards = append(master.Shards, shard)

	nextSeqNo := shard.SeqNo + 1

	t.Run("drop tables", func(t *testing.T) {
		dropTables(t)
	})

	t.Run("create tables", func(t *testing.T) {
		createTables(t)
	})

	t.Run("add block", func(t *testing.T) {
		tx, err := pg.Begin()
		assert.Nil(t, err)

		err = repo.AddBlocks(ctx, tx, shards)
		assert.Nil(t, err)

		err = repo.AddBlocks(ctx, tx, []*core.Block{master})
		assert.Nil(t, err)

		err = tx.Commit()
		assert.Nil(t, err)
	})

	t.Run("filter by workchain", func(t *testing.T) {
		res, err := repo.FilterBlocks(ctx, &filter.BlocksReq{
			Workchain: &shard.Workchain,
			// Shard:     &shard.Shard,
			// SeqNo:     &shard.SeqNo,

			AfterSeqNo: &nextSeqNo, Order: "DESC", Limit: 1,
		})
		assert.Nil(t, err)
		assert.Equal(t, 100, res.Total)
		assert.Equal(t, []*core.Block{shard}, res.Rows)
	})

	t.Run("filter by seq no", func(t *testing.T) {
		res, err := repo.FilterBlocks(ctx, &filter.BlocksReq{
			Workchain: &shard.Workchain,
			SeqNo:     &shard.SeqNo,

			AfterSeqNo: &nextSeqNo, Order: "DESC", Limit: 1,
		})
		assert.Nil(t, err)
		assert.Equal(t, 1, res.Total)
		assert.Equal(t, []*core.Block{shard}, res.Rows)
	})

	t.Run("filter by file hash", func(t *testing.T) {
		res, err := repo.FilterBlocks(ctx, &filter.BlocksReq{
			FileHash: master.FileHash,

			WithShards: true,

			AfterSeqNo: &nextSeqNo, Order: "DESC", Limit: 1,
		})
		assert.Nil(t, err)
		assert.Equal(t, 1, res.Total)
		assert.Equal(t, []*core.Block{master}, res.Rows)
	})

	t.Run("drop tables again", func(t *testing.T) {
		dropTables(t)
	})
}
