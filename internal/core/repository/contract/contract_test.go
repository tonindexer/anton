package contract_test

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/pgdialect"
	"github.com/uptrace/bun/driver/pgdriver"

	"github.com/tonindexer/anton/abi"
	"github.com/tonindexer/anton/internal/addr"
	"github.com/tonindexer/anton/internal/core"
	"github.com/tonindexer/anton/internal/core/repository/contract"
	"github.com/tonindexer/anton/internal/core/rndm"
)

var (
	pg   *bun.DB
	repo *contract.Repository
)

func initdb(t testing.TB) {
	var (
		dsnPG = "postgres://user:pass@localhost:5432/postgres?sslmode=disable"
		err   error
	)

	pg = bun.NewDB(sql.OpenDB(pgdriver.NewConnector(pgdriver.WithDSN(dsnPG))), pgdialect.New())
	err = pg.Ping()
	assert.Nil(t, err)

	repo = contract.NewRepository(pg)
}

func createTables(t testing.TB) {
	err := contract.CreateTables(context.Background(), pg)
	if err != nil {
		t.Fatal(err)
	}
}

func dropTables(t testing.TB) {
	var err error

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	_, err = pg.NewDropTable().Model((*core.ContractOperation)(nil)).IfExists().Exec(ctx)
	assert.Nil(t, err)
	_, err = pg.NewDropTable().Model((*core.ContractInterface)(nil)).IfExists().Exec(ctx)
	assert.Nil(t, err)
}

func TestRepository_AddContracts(t *testing.T) {
	initdb(t)

	i := &core.ContractInterface{
		Name:            abi.NFTItem,
		Addresses:       []*addr.Address{rndm.Address()},
		Code:            rndm.Bytes(128),
		CodeHash:        rndm.Bytes(32),
		GetMethods:      []string{rndm.String(16)},
		GetMethodHashes: rndm.GetMethodHashes(),
	}

	schema, err := abi.MarshalSchema((*abi.NFTItemTransfer)(nil))
	assert.Nil(t, err)

	op := &core.ContractOperation{
		Name:         "nft_item_transfer",
		ContractName: abi.NFTItem,
		Outgoing:     false,
		OperationID:  0xdeadbeed,
		Schema:       schema,
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
		err := repo.AddInterface(ctx, i)
		assert.Nil(t, err)
	})

	t.Run("insert operation", func(t *testing.T) {
		err := repo.AddOperation(ctx, op)
		assert.Nil(t, err)
	})

	t.Run("get interfaces", func(t *testing.T) {
		ret, err := repo.GetInterfaces(ctx)
		assert.Nil(t, err)
		assert.Equal(t, []*core.ContractInterface{i}, ret)
	})

	t.Run("get operations", func(t *testing.T) {
		ret, err := repo.GetOperations(ctx)
		assert.Nil(t, err)
		assert.Equal(t, 1, len(ret))
		assert.JSONEq(t, string(schema), string(ret[0].Schema))
		ret[0].Schema = schema
		assert.Equal(t, []*core.ContractOperation{op}, ret)
	})

	t.Run("get operation by id", func(t *testing.T) {
		ret, err := repo.GetOperationByID(
			ctx,
			[]abi.ContractName{op.ContractName},
			op.Outgoing,
			op.OperationID,
		)
		assert.Nil(t, err)
		assert.JSONEq(t, string(schema), string(ret.Schema))
		ret.Schema = schema
		assert.Equal(t, op, ret)
	})
}
