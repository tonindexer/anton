package block_test

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/pgdialect"
	"github.com/uptrace/bun/driver/pgdriver"
	"github.com/uptrace/go-clickhouse/ch"

	"github.com/tonindexer/anton/internal/core"
	"github.com/tonindexer/anton/internal/core/repository/block"
	"github.com/tonindexer/anton/internal/core/rndm"
)

var (
	ck   *ch.DB
	pg   *bun.DB
	repo *block.Repository
)

func initdb(t testing.TB) {
	var (
		dsnCH = "clickhouse://localhost:9000/testing?sslmode=disable"
		dsnPG = "postgres://user:pass@localhost:5432/postgres?sslmode=disable"
		err   error
	)

	ctx := context.Background()

	ck = ch.Connect(ch.WithDSN(dsnCH), ch.WithAutoCreateDatabase(true), ch.WithPoolSize(16))
	err = ck.Ping(ctx)
	assert.Nil(t, err)

	pg = bun.NewDB(sql.OpenDB(pgdriver.NewConnector(pgdriver.WithDSN(dsnPG))), pgdialect.New())
	err = pg.Ping()
	assert.Nil(t, err)

	repo = block.NewRepository(ck, pg)
}

func createTables(t testing.TB) {
	err := block.CreateTables(context.Background(), ck, pg)
	if err != nil {
		t.Fatal(err)
	}
}

func dropTables(t testing.TB) {
	var err error

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	_, err = ck.NewDropTable().Model((*core.Block)(nil)).IfExists().Exec(ctx)
	assert.Nil(t, err)
	_, err = pg.NewDropTable().Model((*core.Block)(nil)).IfExists().Exec(ctx)
	assert.Nil(t, err)
}

func TestCoreRepository(t *testing.T) {
	initdb(t)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	master := rndm.MasterBlock()
	shard := rndm.Block(0)
	shard.MasterID = &core.BlockID{
		Workchain: master.Workchain,
		Shard:     master.Shard,
		SeqNo:     master.SeqNo,
	}

	t.Run("drop tables", func(t *testing.T) {
		dropTables(t)
	})

	t.Run("create tables", func(t *testing.T) {
		createTables(t)
	})

	t.Run("add block", func(t *testing.T) {
		tx, err := pg.Begin()
		assert.Nil(t, err)

		err = repo.AddBlocks(ctx, tx, []*core.Block{master, shard})
		assert.Nil(t, err)

		got := new(core.Block)

		err = tx.NewSelect().Model(got).Where("workchain = 0").Scan(ctx)
		assert.Nil(t, err)
		assert.Equal(t, shard, got)

		err = ck.NewSelect().Model(got).Where("workchain = 0").Scan(ctx)
		assert.Nil(t, err)
		assert.Equal(t, shard, got)

		err = tx.Commit()
		assert.Nil(t, err)
	})

	t.Run("get last masterchain block", func(t *testing.T) {
		b, err := repo.GetLastMasterBlock(ctx)
		assert.Nil(t, err)
		assert.Equal(t, master, b)
	})

	t.Run("drop tables again", func(t *testing.T) {
		dropTables(t)
	})
}
