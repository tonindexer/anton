package repository_test

import (
	"context"
	"reflect"
	"testing"

	"github.com/uptrace/bun"

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
		"postgres://user:pass@localhost:5432/default?sslmode=disable")
	if err != nil {
		t.Fatal(err)
	}

	accountRepo = account.NewRepository(db.CH, db.PG)
	blockRepo = block.NewRepository(db.CH, db.PG)
	txRepo = tx.NewRepository(db.CH, db.PG)
}

func dropTables(t *testing.T) { //nolint:gocyclo // clean database
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

	_, err = db.CH.NewDropTable().Model((*core.ContractOperation)(nil)).IfExists().Exec(ctx)
	if err != nil {
		t.Fatal(err)
	}
	_, err = db.PG.NewDropTable().Model((*core.ContractOperation)(nil)).IfExists().Exec(ctx)
	if err != nil {
		t.Fatal(err)
	}

	_, err = db.CH.NewDropTable().Model((*core.ContractInterface)(nil)).IfExists().Exec(ctx)
	if err != nil {
		t.Fatal(err)
	}
	_, err = db.PG.NewDropTable().Model((*core.ContractInterface)(nil)).IfExists().Exec(ctx)
	if err != nil {
		t.Fatal(err)
	}
}

func TestGraph(t *testing.T) { //nolint:gocognit,gocyclo // test master block data insertion
	var insertTx bun.Tx

	initDB(t)

	t.Run("drop tables", func(t *testing.T) {
		dropTables(t)
	})

	t.Run("create tables", func(t *testing.T) {
		err := repository.CreateTablesDB(ctx, db)
		if err != nil {
			t.Error(err)
		}
	})

	t.Run("create insert transaction", func(t *testing.T) {
		var err error
		insertTx, err = db.PG.Begin()
		if err != nil {
			t.Fatal(err)
		}
	})

	t.Run("add account data", func(t *testing.T) {
		err := accountRepo.AddAccountData(ctx, insertTx, []*core.AccountData{accDataItem})
		if err != nil {
			t.Error(err)
		}
		if err := accountRepo.AddAccountData(ctx, insertTx, nil); err != nil {
			t.Error(err)
		}
	})

	t.Run("add account states", func(t *testing.T) {
		err := accountRepo.AddAccountStates(ctx, insertTx, []*core.AccountState{accWallet})
		if err != nil {
			t.Error(err)
		}
		err = accountRepo.AddAccountStates(ctx, insertTx, []*core.AccountState{accItem})
		if err != nil {
			t.Error(err)
		}
		err = accountRepo.AddAccountStates(ctx, insertTx, []*core.AccountState{accNoState})
		if err != nil {
			t.Error(err)
		}
		if err := accountRepo.AddAccountStates(ctx, insertTx, nil); err != nil {
			t.Error(err)
		}
	})

	t.Run("add message payloads", func(t *testing.T) {
		err := txRepo.AddMessagePayloads(ctx, insertTx, []*core.MessagePayload{msgInItemPayload})
		if err != nil {
			t.Error(err)
		}
		if err := txRepo.AddMessagePayloads(ctx, insertTx, nil); err != nil {
			t.Error(err)
		}
	})

	t.Run("add messages", func(t *testing.T) {
		err := txRepo.AddMessages(ctx, insertTx, []*core.Message{msgExtWallet, msgOutWallet, msgInItem})
		if err != nil {
			t.Error(err)
		}
		if err := txRepo.AddMessages(ctx, insertTx, nil); err != nil {
			t.Error(err)
		}
	})

	t.Run("add transactions", func(t *testing.T) {
		err := txRepo.AddTransactions(ctx, insertTx, []*core.Transaction{txOutWallet})
		if err != nil {
			t.Error(err)
		}
		err = txRepo.AddTransactions(ctx, insertTx, []*core.Transaction{txInItem})
		if err != nil {
			t.Error(err)
		}
		if err := txRepo.AddTransactions(ctx, insertTx, nil); err != nil {
			t.Error(err)
		}
	})

	t.Run("add shard blocks", func(t *testing.T) {
		err := blockRepo.AddBlocks(ctx, insertTx, []*core.Block{shardPrev, shard})
		if err != nil {
			t.Error(err)
		}
		if err := blockRepo.AddBlocks(ctx, insertTx, nil); err != nil {
			t.Error(err)
		}
	})

	t.Run("add master blocks", func(t *testing.T) {
		err := blockRepo.AddBlocks(ctx, insertTx, []*core.Block{master})
		if err != nil {
			t.Error(err)
		}
	})

	t.Run("commit insert transaction", func(t *testing.T) {
		err := insertTx.Commit()
		if err != nil {
			t.Error(err)
		}
	})

	t.Run("filter blocks", func(t *testing.T) {
		var wc int32 = -1

		f := core.BlockFilter{
			ID:        nil,
			Workchain: &wc,
			FileHash:  nil,
		}
		blocks, err := blockRepo.GetBlocks(ctx, &f, 0, 100)
		if err != nil {
			t.Error(err)
		}
		if len(blocks) != 1 || !reflect.DeepEqual(master, blocks[0]) {
			t.Errorf("wrong master block, expected: %v, got: %v", master, blocks[0])
		}
	})

	t.Run("drop tables final", func(t *testing.T) {
		dropTables(t)
	})
}
