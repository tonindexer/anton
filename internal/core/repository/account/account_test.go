package account_test

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/pgdialect"
	"github.com/uptrace/bun/driver/pgdriver"
	"github.com/uptrace/go-clickhouse/ch"

	"github.com/tonindexer/anton/abi"
	"github.com/tonindexer/anton/internal/core"
	"github.com/tonindexer/anton/internal/core/repository/account"
	"github.com/tonindexer/anton/internal/core/rndm"
)

var (
	ck   *ch.DB
	pg   *bun.DB
	repo *account.Repository
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

	repo = account.NewRepository(ck, pg)
}

func createTables(t testing.TB) {
	err := account.CreateTables(context.Background(), ck, pg)
	require.Nil(t, err)
}

func dropTables(t testing.TB) {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	_, err := pg.NewDropTable().Model((*core.LatestAccountState)(nil)).IfExists().Exec(ctx)
	require.Nil(t, err)

	_, err = ck.NewDropTable().Model((*core.AccountState)(nil)).IfExists().Exec(ctx)
	require.Nil(t, err)
	_, err = pg.NewDropTable().Model((*core.AccountState)(nil)).IfExists().Exec(ctx)
	require.Nil(t, err)

	_, err = ck.NewDropTable().Model((*core.AddressLabel)(nil)).IfExists().Exec(ctx)
	require.Nil(t, err)
	_, err = pg.NewDropTable().Model((*core.AddressLabel)(nil)).IfExists().Exec(ctx)
	require.Nil(t, err)

	_, err = pg.ExecContext(ctx, "DROP TYPE IF EXISTS account_status")
	require.Nil(t, err)
}

func TestRepository_AddAccounts(t *testing.T) {
	initdb(t)

	states := append(rndm.AccountStatesContract(10, "", nil), rndm.AccountStates(10)...)
	a := &states[0].Address

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	tx, err := pg.Begin()
	require.Nil(t, err)

	t.Run("drop tables", func(t *testing.T) {
		dropTables(t)
	})

	t.Run("create tables", func(t *testing.T) {
		createTables(t)
	})

	t.Run("add account states", func(t *testing.T) {
		err := repo.AddAccountStates(ctx, tx, states)
		require.Nil(t, err)

		got := new(core.AccountState)

		err = tx.NewSelect().Model(got).Where("address = ?", a).Where("last_tx_lt = ?", states[0].LastTxLT).Scan(ctx)
		require.Nil(t, err)
		require.Equal(t, states[0], got)

		err = ck.NewSelect().Model(got).Where("address = ?", a).Where("last_tx_lt = ?", states[0].LastTxLT).Scan(ctx)
		require.Nil(t, err)
		got.UpdatedAt = states[0].UpdatedAt // TODO: look at time.Time ch unmarshal
		require.Equal(t, states[0], got)
	})

	t.Run("commit transaction", func(t *testing.T) {
		err := tx.Commit()
		require.Nil(t, err)
	})

	t.Run("update account state", func(t *testing.T) {
		states[0].ExecutedGetMethods = map[abi.ContractName][]abi.GetMethodExecution{
			"nft_item": {{Name: "shiny_method", Error: "failed"}},
		}

		err = repo.UpdateAccountStates(ctx, []*core.AccountState{states[0]})
		require.Nil(t, err)

		got := new(core.AccountState)

		err = pg.NewSelect().Model(got).Where("address = ?", a).Where("last_tx_lt = ?", states[0].LastTxLT).Scan(ctx)
		require.Nil(t, err)
		require.Equal(t, states[0], got)

		err = ck.NewSelect().Model(got).Where("address = ?", a).Where("last_tx_lt = ?", states[0].LastTxLT).Scan(ctx)
		require.Nil(t, err)
		got.UpdatedAt = states[0].UpdatedAt // TODO: look at time.Time ch unmarshal
		require.Equal(t, states[0], got)
	})

	t.Run("drop tables again", func(t *testing.T) {
		dropTables(t)
	})
}

func BenchmarkRepository_AddAccounts(b *testing.B) {
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
			require.Nil(b, err)

			states := rndm.AccountStates(30)
			states = append(states, rndm.AccountStates(30)...)
			states = append(states, rndm.AccountStates(30)...)
			states = append(states, rndm.AccountStatesContract(30, "", nil)...)

			err = repo.AddAccountStates(ctx, tx, states)
			require.Nil(b, err)

			err = tx.Commit()
			require.Nil(b, err)
		}
	})

	b.Run("insert many states", func(b *testing.B) {
		a := rndm.Address()

		for i := 0; i < b.N; i++ {
			tx, err := pg.Begin()
			require.Nil(b, err)

			states := rndm.AddressStates(a, 1)

			err = repo.AddAccountStates(ctx, tx, states)
			require.Nil(b, err)

			err = tx.Commit()
			require.Nil(b, err)
		}
	})

	b.Run("drop tables again", func(b *testing.B) {
		dropTables(b)
	})
}
