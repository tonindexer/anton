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

func TestGraphInsert(t *testing.T) { //nolint:gocognit,gocyclo // test master block data insertion
	var insertTx bun.Tx

	initDB(t)

	t.Run("drop tables", func(t *testing.T) {
		dropTables(t)
	})

	t.Run("create tables", func(t *testing.T) {
		err := repository.CreateTablesDB(ctx, db)
		if err != nil {
			t.Fatal(err)
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
		err := accountRepo.AddAccountData(ctx, insertTx, []*core.AccountData{&accDataWallet, &accDataItem})
		if err != nil {
			t.Fatal(err)
		}
		if err := accountRepo.AddAccountData(ctx, insertTx, nil); err != nil {
			t.Fatal(err)
		}
	})

	t.Run("add account states", func(t *testing.T) {
		err := accountRepo.AddAccountStates(ctx, insertTx, []*core.AccountState{&accWalletOlder, &accWalletOld})
		if err != nil {
			t.Fatal(err)
		}
		err = accountRepo.AddAccountStates(ctx, insertTx, []*core.AccountState{&accWallet})
		if err != nil {
			t.Fatal(err)
		}
		err = accountRepo.AddAccountStates(ctx, insertTx, []*core.AccountState{&accItem})
		if err != nil {
			t.Fatal(err)
		}
		err = accountRepo.AddAccountStates(ctx, insertTx, []*core.AccountState{&accNoState})
		if err != nil {
			t.Fatal(err)
		}
		if err := accountRepo.AddAccountStates(ctx, insertTx, nil); err != nil {
			t.Fatal(err)
		}
	})

	t.Run("add message payloads", func(t *testing.T) {
		err := txRepo.AddMessagePayloads(ctx, insertTx, []*core.MessagePayload{&msgInItemPayload})
		if err != nil {
			t.Fatal(err)
		}
		if err := txRepo.AddMessagePayloads(ctx, insertTx, nil); err != nil {
			t.Fatal(err)
		}
	})

	t.Run("add messages", func(t *testing.T) {
		err := txRepo.AddMessages(ctx, insertTx, []*core.Message{&msgExtWallet, &msgOutWallet, &msgInItem})
		if err != nil {
			t.Fatal(err)
		}
		if err := txRepo.AddMessages(ctx, insertTx, nil); err != nil {
			t.Fatal(err)
		}
	})

	t.Run("add transactions", func(t *testing.T) {
		err := txRepo.AddTransactions(ctx, insertTx, []*core.Transaction{&txOutWallet})
		if err != nil {
			t.Fatal(err)
		}
		err = txRepo.AddTransactions(ctx, insertTx, []*core.Transaction{&txInItem})
		if err != nil {
			t.Fatal(err)
		}
		if err := txRepo.AddTransactions(ctx, insertTx, nil); err != nil {
			t.Fatal(err)
		}
	})

	t.Run("add shard blocks", func(t *testing.T) {
		err := blockRepo.AddBlocks(ctx, insertTx, []*core.Block{&shardPrev, &shard})
		if err != nil {
			t.Fatal(err)
		}
		if err := blockRepo.AddBlocks(ctx, insertTx, nil); err != nil {
			t.Fatal(err)
		}
	})

	t.Run("add master blocks", func(t *testing.T) {
		err := blockRepo.AddBlocks(ctx, insertTx, []*core.Block{&master})
		if err != nil {
			t.Fatal(err)
		}
	})

	t.Run("commit insert transaction", func(t *testing.T) {
		err := insertTx.Commit()
		if err != nil {
			t.Fatal(err)
		}
	})
}

func TestGraphFilterBlocks(t *testing.T) {
	initDB(t)

	t.Run("filter last master", func(t *testing.T) {
		b, err := blockRepo.GetLastMasterBlock(ctx)
		if err != nil {
			t.Error(err)
		}

		if !reflect.DeepEqual(&master, b) {
			t.Errorf("wrong master block, expected: %v, got: %v", master, b)
		}
	})

	t.Run("filter master blocks", func(t *testing.T) {
		var wc int32 = -1

		f := &core.BlockFilter{Workchain: &wc, WithShards: true}

		blocks, err := blockRepo.GetBlocks(ctx, f, 0, 100)
		if err != nil {
			t.Error(err)
		}

		m := master
		m.Shards = []*core.Block{&shardPrev, &shard}

		if len(blocks) != 1 {
			t.Fatalf("wrong len, expected: %d, got: %d", 1, len(blocks))
		}
		if !reflect.DeepEqual(&m, blocks[0]) {
			t.Errorf("wrong master block, expected: %v, got: %v", master, blocks[0])
		}
	})

	t.Run("filter shard blocks", func(t *testing.T) {
		var wc int32 = 0

		f := &core.BlockFilter{Workchain: &wc, WithTransactions: true}

		blocks, err := blockRepo.GetBlocks(ctx, f, 0, 100)
		if err != nil {
			t.Error(err)
		}

		s, sv := shard, shardPrev
		s.Transactions = []*core.Transaction{&txOutWallet, &txInItem}

		if len(blocks) != 2 {
			t.Fatalf("wrong len, expected: %d, got: %d", 2, len(blocks))
		}
		if exp := []*core.Block{&s, &sv}; !reflect.DeepEqual(exp, blocks) {
			t.Errorf("wrong shard block, expected: %v, got: %v", exp, blocks)
		}
	})
}

func TestGraphFilterAccounts(t *testing.T) {
	initDB(t)

	t.Run("filter latest state by address", func(t *testing.T) {
		ret, err := accountRepo.GetAccountStates(ctx, &core.AccountStateFilter{
			Address:     accWallet.Address,
			LatestState: true,
		}, 0, 100)
		if err != nil {
			t.Fatal(err)
		}

		if len(ret) != 1 {
			t.Fatalf("wrong len, expected: %d, got: %d", 1, len(ret))
		}
		if !reflect.DeepEqual(&accWallet, ret[0]) {
			t.Errorf("wrong account, expected: %+v, got: %+v", accWallet, ret[0])
		}
	})

	t.Run("filter latest state by address", func(t *testing.T) {
		ret, err := accountRepo.GetAccountStates(ctx, &core.AccountStateFilter{
			Address:     accWallet.Address,
			LatestState: true,
			WithData:    true,
		}, 0, 100)
		if err != nil {
			t.Fatal(err)
		}

		acc := accWallet
		acc.StateData = &accDataWallet

		if len(ret) != 1 {
			t.Fatalf("wrong len, expected: %d, got: %d", 1, len(ret))
		}
		if !reflect.DeepEqual(&acc, ret[0]) {
			t.Errorf("wrong account, expected: %+v, got: %+v", acc, ret[0])
		}
	})

	t.Run("filter latest item account states by owner address", func(t *testing.T) {
		ret, err := accountRepo.GetAccountStates(ctx, &core.AccountStateFilter{
			LatestState:  true,
			WithData:     true,
			OwnerAddress: accDataItem.OwnerAddress,
		}, 0, 100)
		if err != nil {
			t.Fatal(err)
		}

		acc := accItem
		acc.StateData = &accDataItem

		if len(ret) != 1 {
			t.Fatalf("wrong len, expected: %d, got: %d", 1, len(ret))
		}
		if !reflect.DeepEqual(&acc, ret[0]) {
			t.Errorf("wrong account, expected: %+v, got: %+v", acc, ret[0])
		}
	})
}

func TestGraphFilterTransactions(t *testing.T) {
	initDB(t)

	t.Run("filter tx with msg by address", func(t *testing.T) {
		ret, err := txRepo.GetTransactions(ctx, &core.TransactionFilter{
			Address:          accWallet.Address,
			WithAccountState: true,
			WithMessages:     true,
		}, 0, 100)
		if err != nil {
			t.Fatal(err)
		}

		txOut := txOutWallet
		txOut.Account = &accWallet
		txOut.InMsg = &msgExtWallet
		txOut.OutMsg = []*core.Message{&msgOutWallet}

		if len(ret) != 1 {
			t.Fatalf("wrong len, expected: %d, got: %d", 1, len(ret))
		}
		if !reflect.DeepEqual(&txOut, ret[0]) {
			t.Errorf("wrong tx, expected: %+v, got: %+v", txOut, ret[0])
		}
	})

	t.Run("filter tx with msg by address [2]", func(t *testing.T) {
		ret, err := txRepo.GetTransactions(ctx, &core.TransactionFilter{
			Address:          accItem.Address,
			WithAccountState: true,
			WithMessages:     true,
		}, 0, 100)
		if err != nil {
			t.Fatal(err)
		}

		txIn := txInItem
		txIn.Account = &accItem
		txIn.InMsg = &msgInItem

		if len(ret) != 1 {
			t.Fatalf("wrong len, expected: %d, got: %d", 1, len(ret))
		}
		if !reflect.DeepEqual(&txIn, ret[0]) {
			t.Errorf("wrong tx, expected: %+v, got: %+v", txIn, ret[0])
		}
	})
}
