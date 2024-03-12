package rescan

import (
	"context"
	"database/sql"
	"strings"
	"testing"
	"time"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/require"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/pgdialect"
	"github.com/uptrace/bun/driver/pgdriver"

	"github.com/tonindexer/anton/abi"
	"github.com/tonindexer/anton/abi/known"
	"github.com/tonindexer/anton/addr"
	"github.com/tonindexer/anton/internal/core"
	"github.com/tonindexer/anton/internal/core/repository/contract"
	"github.com/tonindexer/anton/internal/core/rndm"
)

var (
	pg   *bun.DB
	repo *Repository
)

func initdb(t testing.TB) {
	var (
		dsnPG = "postgres://user:pass@localhost:5432/postgres?sslmode=disable"
		err   error
	)

	pg = bun.NewDB(sql.OpenDB(pgdriver.NewConnector(pgdriver.WithDSN(dsnPG))), pgdialect.New())
	err = pg.Ping()
	require.Nil(t, err)

	repo = NewRepository(pg)
}

func createTables(t testing.TB) {
	_, err := pg.ExecContext(context.Background(), "CREATE TYPE message_type AS ENUM (?, ?, ?)", core.ExternalIn, core.ExternalOut, core.Internal)
	require.Nil(t, err)
	err = contract.CreateTables(context.Background(), pg)
	require.Nil(t, err)
	err = CreateTables(context.Background(), pg)
	require.Nil(t, err)
}

func dropTables(t testing.TB) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	_, err := pg.NewDropTable().Model((*core.RescanTask)(nil)).IfExists().Exec(ctx)
	require.Nil(t, err)
	_, err = pg.NewDropTable().Model((*core.ContractOperation)(nil)).IfExists().Exec(ctx)
	require.Nil(t, err)
	_, err = pg.NewDropTable().Model((*core.ContractInterface)(nil)).IfExists().Exec(ctx)
	require.Nil(t, err)
	_, err = pg.NewDropTable().Model((*core.ContractDefinition)(nil)).IfExists().Exec(ctx)
	require.Nil(t, err)

	_, err = pg.ExecContext(context.Background(), "DROP TYPE message_type")
	if err != nil && !strings.Contains(err.Error(), "does not exist") {
		t.Fatal(err)
	}

	_, err = pg.ExecContext(context.Background(), "DROP TYPE rescan_task_type")
	if err != nil && !strings.Contains(err.Error(), "does not exist") {
		t.Fatal(err)
	}
}

func TestRepository_CreateNewRescanTask(t *testing.T) {
	initdb(t)

	i := &core.ContractInterface{
		Name:      known.NFTItem,
		Addresses: []*addr.Address{rndm.Address()},
		Code:      rndm.Bytes(128),
		GetMethodsDesc: []abi.GetMethodDesc{
			{
				Name: "get_nft_content",
				Arguments: []abi.VmValueDesc{
					{
						Name:      "index",
						StackType: "int",
					}, {
						Name:      "individual_content",
						StackType: "cell",
					},
				},
				ReturnValues: []abi.VmValueDesc{
					{
						Name:      "full_content",
						StackType: "cell",
						Format:    "content",
					},
				},
			},
		},
		GetMethodHashes: rndm.GetMethodHashes(),
	}

	task := core.RescanTask{
		Type:         core.AddInterface,
		ContractName: known.NFTItem,
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	t.Run("drop tables", func(t *testing.T) {
		dropTables(t)
	})

	t.Run("create tables", func(t *testing.T) {
		createTables(t)
	})

	t.Run("insert interface", func(t *testing.T) {
		err := contract.NewRepository(pg).AddInterface(ctx, i)
		require.Nil(t, err)
	})

	t.Run("create new task", func(t *testing.T) {
		err := repo.AddRescanTask(ctx, &task)
		require.NoError(t, err)
	})

	t.Run("update unfinished task", func(t *testing.T) {
		tx, task, err := repo.GetUnfinishedRescanTask(ctx)
		require.NoError(t, err)

		task.LastAddress = i.Addresses[0]
		task.LastTxLt = 10

		err = repo.SetRescanTask(ctx, tx, task)
		require.NoError(t, err)
	})

	t.Run("finish task", func(t *testing.T) {
		tx, task, err := repo.GetUnfinishedRescanTask(ctx)
		require.NoError(t, err)
		require.Equal(t, i.Addresses[0], task.LastAddress)
		require.Equal(t, uint64(10), task.LastTxLt)

		task.LastAddress = i.Addresses[0]
		task.LastTxLt = 20
		task.Finished = true

		err = repo.SetRescanTask(ctx, tx, task)
		require.NoError(t, err)
	})

	t.Run("get 'not found' error on choosing unfinished task", func(t *testing.T) {
		_, _, err := repo.GetUnfinishedRescanTask(ctx)
		require.Error(t, err)
		require.True(t, errors.Is(err, core.ErrNotFound))
	})

	t.Run("create second task", func(t *testing.T) {
		err := repo.AddRescanTask(ctx, &task)
		require.NoError(t, err)

		tx, task, err := repo.GetUnfinishedRescanTask(ctx)
		require.NoError(t, err)
		require.Equal(t, 2, task.ID)

		task.Finished = true

		err = repo.SetRescanTask(ctx, tx, task)
		require.NoError(t, err)
	})

	t.Run("drop tables", func(t *testing.T) {
		dropTables(t)
	})
}
