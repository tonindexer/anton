package account

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
	"github.com/tonindexer/anton/internal/core/rndm"
)

var (
	ck   *ch.DB
	pg   *bun.DB
	repo *Repository
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

	repo = NewRepository(ck, pg)
}

func createTables(t testing.TB) {
	err := CreateTables(context.Background(), ck, pg)
	if err != nil {
		t.Fatal(err)
	}
}

func dropTables(t testing.TB) {
	var err error

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	_, err = pg.NewDropTable().Model((*core.LatestAccountState)(nil)).IfExists().Exec(ctx)
	assert.Nil(t, err)

	_, err = ck.NewDropTable().Model((*core.AccountState)(nil)).IfExists().Exec(ctx)
	assert.Nil(t, err)
	_, err = pg.NewDropTable().Model((*core.AccountState)(nil)).IfExists().Exec(ctx)
	assert.Nil(t, err)

	_, err = ck.NewDropTable().Model((*core.AccountData)(nil)).IfExists().Exec(ctx)
	assert.Nil(t, err)
	_, err = pg.NewDropTable().Model((*core.AccountData)(nil)).IfExists().Exec(ctx)
	assert.Nil(t, err)

	_, err = pg.ExecContext(ctx, "DROP TYPE IF EXISTS account_status")
	assert.Nil(t, err)
}

func TestCoreRepository(t *testing.T) {
	initdb(t)

	states := rndm.AccountStates(10)
	data := rndm.AccountData(states)
	a := &states[0].Address

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

	t.Run("add account data", func(t *testing.T) {
		err := repo.AddAccountData(ctx, tx, data)
		assert.Nil(t, err)

		got := new(core.AccountData)

		err = tx.NewSelect().Model(got).Where("address = ?", a).Where("last_tx_lt = ?", data[0].LastTxLT).Scan(ctx)
		assert.Nil(t, err)
		assert.Equal(t, data[0], got)

		err = ck.NewSelect().Model(got).Where("address = ?", a).Where("last_tx_lt = ?", data[0].LastTxLT).Scan(ctx)
		assert.Nil(t, err)
		got.UpdatedAt = data[0].UpdatedAt // TODO: look at time.Time ch unmarshal
		assert.Equal(t, data[0], got)
	})

	t.Run("add account states", func(t *testing.T) {
		err := repo.AddAccountStates(ctx, tx, states)
		assert.Nil(t, err)

		got := new(core.AccountState)

		err = tx.NewSelect().Model(got).Where("address = ?", a).Where("last_tx_lt = ?", states[0].LastTxLT).Scan(ctx)
		assert.Nil(t, err)
		assert.Equal(t, states[0], got)

		err = ck.NewSelect().Model(got).Where("address = ?", a).Where("last_tx_lt = ?", states[0].LastTxLT).Scan(ctx)
		assert.Nil(t, err)
		got.UpdatedAt = states[0].UpdatedAt // TODO: look at time.Time ch unmarshal
		assert.Equal(t, states[0], got)
	})

	t.Run("commit transaction", func(t *testing.T) {
		err := tx.Commit()
		assert.Nil(t, err)
	})

	t.Run("drop tables again", func(t *testing.T) {
		dropTables(t)
	})
}

func BenchmarkCoreRepository(b *testing.B) {
	ctx := context.Background()

	initdb(b)

	b.Run("drop tables", func(t *testing.B) {
		dropTables(t)
	})

	b.Run("create tables", func(t *testing.B) {
		createTables(t)
	})

	b.Run("insert many addresses", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			tx, err := pg.Begin()
			assert.Nil(b, err)

			states := rndm.AccountStates(30)
			states = append(states, rndm.AccountStates(30)...)
			states = append(states, rndm.AccountStates(30)...)
			data := rndm.AccountData(states)

			err = repo.AddAccountData(ctx, tx, data)
			assert.Nil(b, err)
			err = repo.AddAccountStates(ctx, tx, states)
			assert.Nil(b, err)

			err = tx.Commit()
			assert.Nil(b, err)
		}
	})

	b.Run("insert many states", func(b *testing.B) {
		a := rndm.Address()

		for i := 0; i < b.N; i++ {
			tx, err := pg.Begin()
			assert.Nil(b, err)

			states := rndm.AddressStates(a, 1)
			data := rndm.AccountData(states)

			err = repo.AddAccountData(ctx, tx, data)
			assert.Nil(b, err)
			err = repo.AddAccountStates(ctx, tx, states)
			assert.Nil(b, err)

			err = tx.Commit()
			assert.Nil(b, err)
		}
	})

	b.Run("drop tables again", func(b *testing.B) {
		dropTables(b)
	})
}
