package repository_test

import (
	"context"
	"testing"

	"github.com/iam047801/tonidx/internal/core"
	"github.com/iam047801/tonidx/internal/core/repository"
	"github.com/iam047801/tonidx/internal/core/repository/account"
	"github.com/iam047801/tonidx/internal/core/repository/block"
	"github.com/iam047801/tonidx/internal/core/repository/tx"
)

var (
	ctx = context.Background()

	db *repository.DB

	accountRepo core.AccountRepository
	blockRepo   core.BlockRepository
	txRepo      core.TxRepository
)

func initDB(t *testing.T) {
	var err error

	db, err = repository.ConnectDB(context.Background(),
		"clickhouse://localhost:9000/default?sslmode=disable",
		"postgres://postgres:postgres@localhost:5432/postgres?sslmode=disable")
	if err != nil {
		t.Fatal(err)
	}

	accountRepo = account.NewRepository(db.CH, db.PG)
	blockRepo = block.NewRepository(db.CH, db.PG)
	txRepo = tx.NewRepository(db.CH, db.PG)
}

func dropTables(t *testing.T) {
	var err error

	// TODO: drop pg enums

	_, err = db.CH.NewDropTable().Model((*core.Transaction)(nil)).IfExists().Exec(ctx)
	if err != nil {
		t.Fatal(err)
	}
	_, err = db.PG.NewDropTable().Model((*core.Transaction)(nil)).IfExists().Exec(ctx)
	if err != nil {
		t.Fatal(err)
	}

	_, err = db.CH.NewDropTable().Model((*core.Message)(nil)).IfExists().Exec(ctx)
	if err != nil {
		t.Fatal(err)
	}
	_, err = db.PG.NewDropTable().Model((*core.Message)(nil)).IfExists().Exec(ctx)
	if err != nil {
		t.Fatal(err)
	}

	_, err = db.CH.NewDropTable().Model((*core.MessagePayload)(nil)).IfExists().Exec(ctx)
	if err != nil {
		t.Fatal(err)
	}
	_, err = db.PG.NewDropTable().Model((*core.MessagePayload)(nil)).IfExists().Exec(ctx)
	if err != nil {
		t.Fatal(err)
	}

	_, err = db.CH.NewDropTable().Model((*core.AccountState)(nil)).IfExists().Exec(ctx)
	if err != nil {
		t.Fatal(err)
	}
	_, err = db.PG.NewDropTable().Model((*core.AccountState)(nil)).IfExists().Exec(ctx)
	if err != nil {
		t.Fatal(err)
	}
	_, err = db.CH.NewDropTable().Model((*core.AccountData)(nil)).IfExists().Exec(ctx)
	if err != nil {
		t.Fatal(err)
	}
	_, err = db.PG.NewDropTable().Model((*core.AccountData)(nil)).IfExists().Exec(ctx)
	if err != nil {
		t.Fatal(err)
	}

	_, err = db.CH.NewDropTable().Model((*core.Block)(nil)).IfExists().Exec(ctx)
	if err != nil {
		t.Fatal(err)
	}
	_, err = db.PG.NewDropTable().Model((*core.Block)(nil)).IfExists().Exec(ctx)
	if err != nil {
		t.Fatal(err)
	}
}

func TestGraph(t *testing.T) {
	t.Run("init db", func(t *testing.T) {
		initDB(t)
	})

	t.Run("drop tables", func(t *testing.T) {
		dropTables(t)
	})

	t.Run("create tables", func(t *testing.T) {
		err := repository.CreateTablesDB(ctx, db)
		if err != nil {
			t.Fatal(err)
		}
	})

	t.Run("add account data", func(t *testing.T) {
		err := accountRepo.AddAccountData(ctx, []*core.AccountData{accDataItem})
		if err != nil {
			t.Fatal(err)
		}
	})

	t.Run("add account states", func(t *testing.T) {
		err := accountRepo.AddAccountStates(ctx, []*core.AccountState{accWallet})
		if err != nil {
			t.Fatal(err)
		}
		err = accountRepo.AddAccountStates(ctx, []*core.AccountState{accItem})
		if err != nil {
			t.Fatal(err)
		}
	})

	t.Run("add message payloads", func(t *testing.T) {
		err := txRepo.AddMessagePayloads(ctx, []*core.MessagePayload{msgInItemPayload})
		if err != nil {
			t.Fatal(err)
		}
	})

	t.Run("add messages", func(t *testing.T) {
		err := txRepo.AddMessages(ctx, []*core.Message{msgExtWallet, msgOutWallet, msgInItem})
		if err != nil {
			t.Fatal(err)
		}
	})

	t.Run("add transactions", func(t *testing.T) {
		err := txRepo.AddTransactions(ctx, []*core.Transaction{txOutWallet})
		if err != nil {
			t.Fatal(err)
		}
		err = txRepo.AddTransactions(ctx, []*core.Transaction{txInItem})
		if err != nil {
			t.Fatal(err)
		}
	})

	t.Run("add shard blocks", func(t *testing.T) {
		err := blockRepo.AddBlocks(ctx, []*core.Block{shard})
		if err != nil {
			t.Fatal(err)
		}
	})

	t.Run("add master blocks", func(t *testing.T) {
		err := blockRepo.AddBlocks(ctx, []*core.Block{master})
		if err != nil {
			t.Fatal(err)
		}
	})

	t.Run("filter blocks", func(t *testing.T) {
		var wc int32 = -1

		f := core.BlockFilter{
			ID:        nil,
			Workchain: &wc,
			FileHash:  nil,

			// WithMaster: true,
			// WithShards: true,
			// WithTransactions:        true,
			// WithTransactionMessages: true,
		}
		blocks, err := blockRepo.GetBlocks(ctx, &f, 0, 100)
		if err != nil {
			t.Fatal(err)
		}
		if len(blocks) != 1 {
			t.Fatal("wrong blocks length")
		}
	})
}
