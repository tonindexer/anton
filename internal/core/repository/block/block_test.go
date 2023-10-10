package block_test

import (
	"context"
	"database/sql"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/pgdialect"
	"github.com/uptrace/bun/driver/pgdriver"
	"github.com/uptrace/go-clickhouse/ch"

	"github.com/tonindexer/anton/internal/core"
	"github.com/tonindexer/anton/internal/core/repository/block"
	"github.com/tonindexer/anton/internal/core/repository/tx"
	"github.com/tonindexer/anton/internal/core/rndm"
)

var (
	ck   *ch.DB
	pg   *bun.DB
	repo *block.Repository
)

func initdb(t testing.TB) {
	var (
		dsnCH = "clickhouse://user:pass@localhost:9000/default?sslmode=disable"
		dsnPG = "postgres://user:pass@localhost:5432/postgres?sslmode=disable"
		err   error
	)

	ctx := context.Background()

	ck = ch.Connect(ch.WithDSN(dsnCH), ch.WithAutoCreateDatabase(true), ch.WithPoolSize(16))
	err = ck.Ping(ctx)
	require.Nil(t, err)

	pg = bun.NewDB(sql.OpenDB(pgdriver.NewConnector(pgdriver.WithDSN(dsnPG))), pgdialect.New())
	err = pg.Ping()
	require.Nil(t, err)

	repo = block.NewRepository(ck, pg)
}

func createTables(t testing.TB) {
	err := block.CreateTables(context.Background(), ck, pg)
	require.Nil(t, err)

	_, err = pg.ExecContext(context.Background(), "CREATE TYPE account_status AS ENUM (?, ?, ?, ?)",
		core.Uninit, core.Active, core.Frozen, core.NonExist)
	require.False(t, err != nil && !strings.Contains(err.Error(), "already exists"))

	err = tx.CreateTables(context.Background(), ck, pg)
	require.Nil(t, err)
}

func dropTables(t testing.TB) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	_, err := ck.NewDropTable().Model((*core.Transaction)(nil)).IfExists().Exec(ctx)
	require.Nil(t, err)
	_, err = pg.NewDropTable().Model((*core.Transaction)(nil)).IfExists().Exec(ctx)
	require.Nil(t, err)

	_, err = pg.ExecContext(ctx, "DROP TYPE IF EXISTS account_status")
	require.Nil(t, err)

	_, err = ck.NewDropTable().Model((*core.Block)(nil)).IfExists().Exec(ctx)
	require.Nil(t, err)
	_, err = pg.NewDropTable().Model((*core.Block)(nil)).IfExists().Exec(ctx)
	require.Nil(t, err)
}

func TestRepository_AddBlocks(t *testing.T) {
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
		dbTx, err := pg.Begin()
		require.Nil(t, err)

		err = repo.AddBlocks(ctx, dbTx, []*core.Block{master, shard})
		require.Nil(t, err)

		got := new(core.Block)

		err = dbTx.NewSelect().Model(got).Where("workchain = 0").Scan(ctx)
		require.Nil(t, err)
		require.Equal(t, shard, got)

		err = ck.NewSelect().Model(got).Where("workchain = 0").Scan(ctx)
		require.Nil(t, err)
		require.Equal(t, shard.ScannedAt.Truncate(time.Second), got.ScannedAt.UTC()) // TODO: debug ch timestamps
		got.ScannedAt = shard.ScannedAt
		require.Equal(t, shard, got)

		err = dbTx.Commit()
		require.Nil(t, err)
	})

	t.Run("get last masterchain block", func(t *testing.T) {
		b, err := repo.GetLastMasterBlock(ctx)
		require.Nil(t, err)
		require.Equal(t, master, b)
	})

	t.Run("drop tables again", func(t *testing.T) {
		dropTables(t)
	})
}
