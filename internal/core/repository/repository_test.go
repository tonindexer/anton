package repository_test

import (
	"context"
	"testing"

	"github.com/uptrace/bun"
	"github.com/uptrace/go-clickhouse/ch"

	"github.com/iam047801/tonidx/internal/core"
	"github.com/iam047801/tonidx/internal/core/db"
	"github.com/iam047801/tonidx/internal/core/repository/account"
	"github.com/iam047801/tonidx/internal/core/repository/block"
	"github.com/iam047801/tonidx/internal/core/repository/tx"
)

var (
	ctx = context.Background()

	_ch *ch.DB
	_pg *bun.DB

	_account core.AccountRepository
	_block   core.BlockRepository
	_tx      core.TxRepository
)

func _initDB(t *testing.T) {
	var err error

	_ch, _pg, err = db.Connect(context.Background(),
		"clickhouse://localhost:9000/default?sslmode=disable",
		"postgres://postgres:postgres@localhost:5432/postgres?sslmode=disable")
	if err != nil {
		t.Fatal(err)
	}
}

func chdb(t *testing.T) *ch.DB {
	if _ch != nil {
		return _ch
	}
	_initDB(t)
	return _ch
}

func pgdb(t *testing.T) *bun.DB {
	if _pg != nil {
		return _pg
	}
	_initDB(t)
	return _pg
}

func accountRepo(t *testing.T) core.AccountRepository {
	if _account != nil {
		return _account
	}
	_account = account.NewRepository(chdb(t), pgdb(t))
	return _account
}

func blockRepo(t *testing.T) core.BlockRepository {
	if _block != nil {
		return _block
	}
	_block = block.NewRepository(chdb(t), pgdb(t))
	return _block
}

func txRepo(t *testing.T) core.TxRepository {
	if _tx != nil {
		return _tx
	}
	_tx = tx.NewRepository(chdb(t), pgdb(t))
	return _tx
}

func createTables(t *testing.T) {
	err := block.CreateTables(ctx, chdb(t), pgdb(t))
	if err != nil {
		t.Fatal(err)
	}
	err = account.CreateTables(ctx, chdb(t), pgdb(t))
	if err != nil {
		t.Fatal(err)
	}
	err = tx.CreateTables(ctx, chdb(t), pgdb(t))
	if err != nil {
		t.Fatal(err)
	}
}

func dropTables(t *testing.T) {
	var err error

	// TODO: drop pg enums

	_, err = chdb(t).NewDropTable().Model((*core.Transaction)(nil)).IfExists().Exec(ctx)
	if err != nil {
		t.Fatal(err)
	}
	_, err = pgdb(t).NewDropTable().Model((*core.Transaction)(nil)).IfExists().Exec(ctx)
	if err != nil {
		t.Fatal(err)
	}

	_, err = chdb(t).NewDropTable().Model((*core.Message)(nil)).IfExists().Exec(ctx)
	if err != nil {
		t.Fatal(err)
	}
	_, err = pgdb(t).NewDropTable().Model((*core.Message)(nil)).IfExists().Exec(ctx)
	if err != nil {
		t.Fatal(err)
	}

	_, err = chdb(t).NewDropTable().Model((*core.MessagePayload)(nil)).IfExists().Exec(ctx)
	if err != nil {
		t.Fatal(err)
	}
	_, err = pgdb(t).NewDropTable().Model((*core.MessagePayload)(nil)).IfExists().Exec(ctx)
	if err != nil {
		t.Fatal(err)
	}

	_, err = chdb(t).NewDropTable().Model((*core.AccountState)(nil)).IfExists().Exec(ctx)
	if err != nil {
		t.Fatal(err)
	}
	_, err = pgdb(t).NewDropTable().Model((*core.AccountState)(nil)).IfExists().Exec(ctx)
	if err != nil {
		t.Fatal(err)
	}
	_, err = chdb(t).NewDropTable().Model((*core.AccountData)(nil)).IfExists().Exec(ctx)
	if err != nil {
		t.Fatal(err)
	}
	_, err = pgdb(t).NewDropTable().Model((*core.AccountData)(nil)).IfExists().Exec(ctx)
	if err != nil {
		t.Fatal(err)
	}

	_, err = chdb(t).NewDropTable().Model((*core.Block)(nil)).IfExists().Exec(ctx)
	if err != nil {
		t.Fatal(err)
	}
	_, err = pgdb(t).NewDropTable().Model((*core.Block)(nil)).IfExists().Exec(ctx)
	if err != nil {
		t.Fatal(err)
	}
}

func TestGraph(t *testing.T) {
	t.Run("drop tables", func(t *testing.T) {
		dropTables(t)
	})

	t.Run("create tables", func(t *testing.T) {
		createTables(t)
	})

	t.Run("add account data", func(t *testing.T) {
		err := accountRepo(t).AddAccountData(ctx, []*core.AccountData{accDataItem})
		if err != nil {
			t.Fatal(err)
		}
	})

	t.Run("add account states", func(t *testing.T) {
		err := accountRepo(t).AddAccountStates(ctx, []*core.AccountState{accWallet})
		if err != nil {
			t.Fatal(err)
		}
		err = accountRepo(t).AddAccountStates(ctx, []*core.AccountState{accItem})
		if err != nil {
			t.Fatal(err)
		}
	})

	t.Run("add message payloads", func(t *testing.T) {
		err := txRepo(t).AddMessagePayloads(ctx, []*core.MessagePayload{msgInItemPayload})
		if err != nil {
			t.Fatal(err)
		}
	})

	t.Run("add messages", func(t *testing.T) {
		err := txRepo(t).AddMessages(ctx, []*core.Message{msgExtWallet, msgOutWallet, msgInItem})
		if err != nil {
			t.Fatal(err)
		}
	})

	t.Run("add transactions", func(t *testing.T) {
		err := txRepo(t).AddTransactions(ctx, []*core.Transaction{txOutWallet})
		if err != nil {
			t.Fatal(err)
		}
		err = txRepo(t).AddTransactions(ctx, []*core.Transaction{txInItem})
		if err != nil {
			t.Fatal(err)
		}
	})

	t.Run("add shard blocks", func(t *testing.T) {
		err := blockRepo(t).AddBlocks(ctx, []*core.Block{shard})
		if err != nil {
			t.Fatal(err)
		}
	})

	t.Run("add master blocks", func(t *testing.T) {
		err := blockRepo(t).AddBlocks(ctx, []*core.Block{master})
		if err != nil {
			t.Fatal(err)
		}
	})
}
