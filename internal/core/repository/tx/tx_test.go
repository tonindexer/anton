package tx_test

import (
	"context"
	"database/sql"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/pgdialect"
	"github.com/uptrace/bun/driver/pgdriver"
	"github.com/uptrace/go-clickhouse/ch"

	"github.com/tonindexer/anton/internal/core"
	"github.com/tonindexer/anton/internal/core/repository/tx"
	"github.com/tonindexer/anton/internal/core/rndm"
)

var (
	ck   *ch.DB
	pg   *bun.DB
	repo *tx.Repository
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
	assert.Nil(t, err)

	pg = bun.NewDB(sql.OpenDB(pgdriver.NewConnector(pgdriver.WithDSN(dsnPG))), pgdialect.New())
	err = pg.Ping()
	assert.Nil(t, err)

	repo = tx.NewRepository(ck, pg)
}

func createTables(t testing.TB) {
	_, err := pg.ExecContext(context.Background(), "CREATE TYPE account_status AS ENUM (?, ?, ?, ?)",
		core.Uninit, core.Active, core.Frozen, core.NonExist)
	assert.False(t, err != nil && !strings.Contains(err.Error(), "already exists"))

	err = tx.CreateTables(context.Background(), ck, pg)
	assert.Nil(t, err)
}

func dropTables(t testing.TB) {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	_, err := ck.NewDropTable().Model((*core.Transaction)(nil)).IfExists().Exec(ctx)
	assert.Nil(t, err)
	_, err = pg.NewDropTable().Model((*core.Transaction)(nil)).IfExists().Exec(ctx)
	assert.Nil(t, err)

	_, err = pg.ExecContext(ctx, "DROP TYPE IF EXISTS account_status")
	assert.Nil(t, err)
}

func TestRepository_AddTransactions(t *testing.T) {
	initdb(t)

	transactions := rndm.Transactions(10)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	dbtx, err := pg.Begin()
	assert.Nil(t, err)

	t.Run("drop tables", func(t *testing.T) {
		dropTables(t)
	})

	t.Run("create tables", func(t *testing.T) {
		createTables(t)
	})

	t.Run("add transactions", func(t *testing.T) {
		err := repo.AddTransactions(ctx, dbtx, transactions)
		assert.Nil(t, err)

		got := new(core.Transaction)

		err = dbtx.NewSelect().Model(got).Where("hash = ?", transactions[0].Hash).Scan(ctx)
		assert.Nil(t, err)
		assert.Equal(t, transactions[0], got)

		err = ck.NewSelect().Model(got).Where("hash = ?", transactions[0].Hash).Scan(ctx)
		assert.Nil(t, err)
		got.CreatedAt = transactions[0].CreatedAt // TODO: look at time.Time ch unmarshal
		assert.Equal(t, transactions[0], got)
	})

	t.Run("commit transaction", func(t *testing.T) {
		err := dbtx.Commit()
		assert.Nil(t, err)
	})

	t.Run("drop tables again", func(t *testing.T) {
		dropTables(t)
	})
}
