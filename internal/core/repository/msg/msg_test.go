package msg_test

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
	"github.com/tonindexer/anton/internal/core/repository/msg"
	"github.com/tonindexer/anton/internal/core/rndm"
)

var (
	ck   *ch.DB
	pg   *bun.DB
	repo *msg.Repository
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

	repo = msg.NewRepository(ck, pg)
}

func createTables(t testing.TB) {
	err := msg.CreateTables(context.Background(), ck, pg)
	assert.Nil(t, err)
}

func dropTables(t testing.TB) {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	_, err := ck.NewDropTable().Model((*core.Message)(nil)).IfExists().Exec(ctx)
	assert.Nil(t, err)
	_, err = pg.NewDropTable().Model((*core.Message)(nil)).IfExists().Exec(ctx)
	assert.Nil(t, err)

	_, err = pg.ExecContext(ctx, "DROP TYPE IF EXISTS message_type")
	assert.Nil(t, err)
}

func TestRepository_AddMessages(t *testing.T) {
	initdb(t)

	messages := rndm.Messages(10)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	tx, err := pg.Begin()
	assert.Nil(t, err)

	t.Run("drop tables", func(t *testing.T) {
		dropTables(t)
	})

	t.Run("create tables", func(t *testing.T) {
		createTables(t)
	})

	t.Run("add messages", func(t *testing.T) {
		err := repo.AddMessages(ctx, tx, messages)
		assert.Nil(t, err)

		got := new(core.Message)

		err = tx.NewSelect().Model(got).Where("hash = ?", messages[0].Hash).Scan(ctx)
		assert.Nil(t, err)
		assert.Equal(t, messages[0], got)

		err = ck.NewSelect().Model(got).Where("hash = ?", messages[0].Hash).Scan(ctx)
		assert.Nil(t, err)
		got.CreatedAt = messages[0].CreatedAt // TODO: look at time.Time ch unmarshal
		assert.Equal(t, messages[0], got)
	})

	t.Run("commit transaction", func(t *testing.T) {
		err := tx.Commit()
		assert.Nil(t, err)
	})
}
