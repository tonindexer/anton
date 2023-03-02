package repository_test

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"reflect"
	"testing"

	"github.com/uptrace/bun"
	"github.com/uptrace/bun/extra/bunbig"

	"github.com/iam047801/tonidx/abi"
	"github.com/iam047801/tonidx/internal/addr"
	"github.com/iam047801/tonidx/internal/core"
	"github.com/iam047801/tonidx/internal/core/repository"
	"github.com/iam047801/tonidx/internal/core/repository/account"
	"github.com/iam047801/tonidx/internal/core/repository/block"
	"github.com/iam047801/tonidx/internal/core/repository/contract"
	"github.com/iam047801/tonidx/internal/core/repository/tx"
)

var (
	ctx = context.Background()

	db *repository.DB

	accountRepo core.AccountRepository
	abiRepo     core.ContractRepository
	blockRepo   core.BlockRepository
	txRepo      core.TxRepository
)

func initDB() {
	var err error

	db, err = repository.ConnectDB(context.Background(),
		"clickhouse://localhost:9000/default?sslmode=disable",
		"postgres://user:pass@localhost:5432/default?sslmode=disable")
	if err != nil {
		panic(err)
	}

	accountRepo = account.NewRepository(db.CH, db.PG)
	abiRepo = contract.NewRepository(db.PG)
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

func TestInsertKnownInterfaces(t *testing.T) {
	initDB()

	t.Run("drop tables", func(t *testing.T) {
		dropTables(t)
	})

	t.Run("create tables", func(t *testing.T) {
		err := repository.CreateTablesDB(ctx, db)
		if err != nil {
			t.Fatal(err)
		}
	})

	t.Run("insert known interfaces", func(t *testing.T) {
		err := repository.InsertKnownInterfaces(ctx, db.PG)
		if err != nil {
			t.Fatal(err)
		}
	})

	t.Run("get contact operation", func(t *testing.T) {
		op, err := abiRepo.GetOperationByID(ctx, []abi.ContractName{abi.NFTItem}, false, 0x5fcc3d14)
		if err != nil {
			t.Fatal(err)
		}
		_, err = abi.UnmarshalSchema(op.Schema)
		if err != nil {
			t.Fatal(err)
		}
	})
}

func TestGraphInsert(t *testing.T) { //nolint:gocognit,gocyclo // test master block data insertion
	var insertTx bun.Tx

	initDB()

	t.Run("drop tables", func(t *testing.T) {
		dropTables(t)
	})

	t.Run("create tables", func(t *testing.T) {
		err := repository.CreateTablesDB(ctx, db)
		if err != nil {
			t.Fatal(err)
		}
	})

	t.Run("insert interfaces", func(t *testing.T) {
		_, err := db.PG.NewInsert().Model(&ifaceItem).Exec(ctx)
		if err != nil {
			t.Fatal(err)
		}
		_, err = db.PG.NewInsert().Model(&opItemTransfer).Exec(ctx)
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

		sd := new(core.AccountData)
		if err := db.CH.NewSelect().Model(sd).Where("address = ?", &accDataItem.Address).Where("last_tx_lt = ?", accDataItem.LastTxLT).Scan(ctx); err != nil {
			t.Fatal(err)
		}
		ad := accDataItem
		ad.TotalSupply, ad.TotalSupply, sd.ContentImageData, sd.Errors =
			bunbig.FromInt64(0), bunbig.FromInt64(0), nil, nil
		if !reflect.DeepEqual(sd, &ad) {
			t.Fatalf("wrong account data, expected: %+v, got: %+v", ad, sd)
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

		s := new(core.AccountState)
		if err := db.CH.NewSelect().Model(s).Where("address = ?", &accWallet.Address).Where("last_tx_lt = ?", accWallet.LastTxLT).Scan(ctx); err != nil {
			t.Fatal(err)
		}
		acc := accWallet
		acc.Latest = false
		s.GetMethodHashes = nil
		if !reflect.DeepEqual(s, &acc) {
			t.Fatalf("wrong account, expected: %+v, got: %+v", acc, s)
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
		err := txRepo.AddMessages(ctx, insertTx, []*core.Message{&msgExtWallet, &msgOutWallet})
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

func TestGraphFilterAccounts(t *testing.T) {
	initDB()

	// TODO: optional fields
	accWallet.Code, accWallet.Data = nil, nil
	accItem.Code, accItem.Data = nil, nil

	t.Run("filter latest state by address", func(t *testing.T) {
		ret, err := accountRepo.GetAccountStates(ctx, &core.AccountStateFilter{
			Addresses:   []*addr.Address{&accWallet.Address},
			LatestState: true,
			Order:       "DESC",
			Limit:       1,
		})
		if err != nil {
			t.Fatal(err)
		}

		if len(ret) != 1 {
			t.Fatalf("wrong len, expected: %d, got: %d", 1, len(ret))
		}
		if !reflect.DeepEqual(&accWallet, ret[0]) {
			t.Fatalf("wrong account, expected: %+v, got: %+v", accWallet, ret[0])
		}
	})

	t.Run("filter latest state by address", func(t *testing.T) {
		ret, err := accountRepo.GetAccountStates(ctx, &core.AccountStateFilter{
			Addresses:   []*addr.Address{&accWallet.Address},
			LatestState: true,
			WithData:    true,
			Order:       "DESC",
			Limit:       1,
		})
		if err != nil {
			t.Fatal(err)
		}

		acc := accWallet
		acc.StateData = &accDataWallet

		if len(ret) != 1 {
			t.Fatalf("wrong len, expected: %d, got: %d", 1, len(ret))
		}
		if !reflect.DeepEqual(&acc, ret[0]) {
			t.Fatalf("wrong account, expected: %+v, got: %+v", acc, ret[0])
		}
	})

	t.Run("filter latest item account states by types", func(t *testing.T) {
		ret, err := accountRepo.GetAccountStates(ctx, &core.AccountStateFilter{
			LatestState:   true,
			WithData:      true,
			ContractTypes: []abi.ContractName{"item"},
			Order:         "DESC",
			Limit:         1,
		})
		if err != nil {
			t.Fatal(err)
		}

		acc := accItem
		acc.StateData = &accDataItem

		if len(ret) != 1 {
			t.Fatalf("wrong len, expected: %d, got: %d", 1, len(ret))
		}
		if !reflect.DeepEqual(&acc, ret[0]) {
			t.Fatalf("wrong account, expected: %+v, got: %+v", acc, ret[0])
		}
	})

	t.Run("filter latest item account states by owner address", func(t *testing.T) {
		ret, err := accountRepo.GetAccountStates(ctx, &core.AccountStateFilter{
			LatestState:  true,
			WithData:     true,
			OwnerAddress: accDataItem.OwnerAddress,
			Order:        "DESC",
			Limit:        1,
		})
		if err != nil {
			t.Fatal(err)
		}

		acc := accItem
		acc.StateData = &accDataItem

		if len(ret) != 1 {
			t.Fatalf("wrong len, expected: %d, got: %d", 1, len(ret))
		}
		if !reflect.DeepEqual(&acc, ret[0]) {
			t.Fatalf("wrong account, expected: %+v, got: %+v", acc, ret[0])
		}
	})
}

func TestGraphFilterMessages(t *testing.T) {
	initDB()

	t.Run("filter messages by operation name with source", func(t *testing.T) {
		ret, err := txRepo.GetMessages(ctx, &core.MessageFilter{
			WithPayload:    true,
			OperationNames: []string{"item_transfer"},
			Order:          "DESC",
			Limit:          10,
		})
		if err != nil {
			t.Fatal(err)
		}

		msgIn := msgOutWallet
		msgIn.Payload = &msgInItemPayload

		if len(ret) != 1 {
			t.Fatalf("wrong len, expected: %d, got: %d", 1, len(ret))
		}
		if !reflect.DeepEqual(msgIn.Payload.DataJSON, ret[0].Payload.DataJSON) {
			t.Fatalf("wrong msg payload data json, expected: %s, got: %s", msgIn.Payload.DataJSON, ret[0].Payload.DataJSON)
		}
		if !reflect.DeepEqual(msgIn.Payload, ret[0].Payload) {
			t.Fatalf("wrong msg payload, expected: %+v, got: %+v", msgIn.Payload, ret[0].Payload)
		}
		if !reflect.DeepEqual(&msgIn, ret[0]) {
			t.Fatalf("wrong msg, expected: %+v, got: %+v", msgIn, ret[0])
		}
	})
}

func TestGraphFilterTransactions(t *testing.T) {
	initDB()

	t.Run("filter tx with msg by address", func(t *testing.T) {
		ret, err := txRepo.GetTransactions(ctx, &core.TransactionFilter{
			Addresses:           []*addr.Address{&accWallet.Address},
			WithAccountState:    true,
			WithMessages:        true,
			WithMessagePayloads: true,
			Order:               "DESC",
			Limit:               10,
		})
		if err != nil {
			t.Fatal(err)
		}

		txOut := txOutWallet
		txOut.Account = &accWallet
		txOut.InMsg = &msgExtWallet
		msgOut := msgOutWallet
		msgOut.Payload = &msgInItemPayload
		txOut.OutMsg = []*core.Message{&msgOut}

		if len(ret) != 1 {
			t.Fatalf("wrong len, expected: %d, got: %d", 1, len(ret))
		}
		if !reflect.DeepEqual(txOut.InMsg, ret[0].InMsg) {
			t.Fatalf("wrong tx in msg, expected: %+v, got: %+v", txOut.InMsg, ret[0].InMsg)
		}
		if len(ret[0].OutMsg) != 1 || !reflect.DeepEqual(txOut.OutMsg[0], ret[0].OutMsg[0]) {
			t.Fatalf("wrong tx out msg, expected: %+v, got: %+v", txOut.OutMsg, ret[0].OutMsg)
		}
		if !reflect.DeepEqual(&txOut, ret[0]) {
			t.Fatalf("wrong tx, expected: %+v, got: %+v", txOut, ret[0])
		}
	})

	t.Run("filter tx with msg by address __item", func(t *testing.T) {
		ret, err := txRepo.GetTransactions(ctx, &core.TransactionFilter{
			Addresses:           []*addr.Address{&accItem.Address},
			WithAccountState:    true,
			WithAccountData:     true,
			WithMessages:        true,
			WithMessagePayloads: true,
			Order:               "DESC",
			Limit:               8,
		})
		if err != nil {
			t.Fatal(err)
		}

		txIn, acc := txInItem, accItem
		txIn.Account = &acc
		txIn.Account.StateData = &accDataItem
		txIn.InMsg = &msgOutWallet
		txIn.InMsg.Payload = &msgInItemPayload

		if len(ret) != 1 {
			t.Fatalf("wrong len, expected: %d, got: %d", 1, len(ret))
		}
		if !reflect.DeepEqual(&txIn, ret[0]) {
			t.Fatalf("wrong tx, expected: %+v, got: %+v", txIn, ret[0])
		}
	})
}

func TestGraphFilterBlocks(t *testing.T) {
	initDB()

	t.Run("filter last master", func(t *testing.T) {
		b, err := blockRepo.GetLastMasterBlock(ctx)
		if err != nil {
			t.Error(err)
		}

		if !reflect.DeepEqual(&master, b) {
			t.Fatalf("wrong master block, expected: %v, got: %v", master, b)
		}
	})

	t.Run("filter master blocks", func(t *testing.T) {
		var wc int32 = -1

		f := &core.BlockFilter{
			Workchain:  &wc,
			WithShards: true,
			Order:      "DESC",
			Limit:      100,
		}

		blocks, err := blockRepo.GetBlocks(ctx, f)
		if err != nil {
			t.Error(err)
		}

		m := master
		m.Shards = []*core.Block{&shardPrev, &shard}

		if len(blocks) != 1 {
			t.Fatalf("wrong len, expected: %d, got: %d", 1, len(blocks))
		}
		if !reflect.DeepEqual(&m, blocks[0]) {
			t.Fatalf("wrong master block, expected: %v, got: %v", master, blocks[0])
		}
	})

	t.Run("filter shard blocks", func(t *testing.T) {
		var wc int32 = 0

		f := &core.BlockFilter{
			Workchain: &wc,

			Order: "DESC",
			Limit: 100,
		}

		blocks, err := blockRepo.GetBlocks(ctx, f)
		if err != nil {
			t.Error(err)
		}

		if len(blocks) != 2 {
			t.Fatalf("wrong len, expected: %d, got: %d", 2, len(blocks))
		}
		if exp := []*core.Block{&shard, &shardPrev}; !reflect.DeepEqual(exp, blocks) {
			t.Fatalf("wrong shard block, expected: %v, got: %v", exp, blocks)
		}
	})
}

func Example_blockRepo_GetBlocks() {
	var wc int32 = -1

	initDB()

	f := &core.BlockFilter{
		Workchain:                      &wc,
		WithShards:                     true,
		WithTransactionAccountState:    true,
		WithTransactionAccountData:     true,
		WithTransactions:               true,
		WithTransactionMessages:        true,
		WithTransactionMessagePayloads: true,

		Order: "DESC",
		Limit: 100,
	}

	blocks, err := blockRepo.GetBlocks(ctx, f)
	if err != nil {
		panic(err)
	}

	s, sv := shard, shardPrev

	txOut, msgOut := txOutWallet, msgOutWallet
	txOut.Account = &accWallet
	txOut.Account.StateData = &accDataWallet
	txOut.InMsg = &msgExtWallet
	msgOut.Payload = &msgInItemPayload
	txOut.OutMsg = []*core.Message{&msgOut}

	txIn := txInItem
	txIn.Account = &accItem
	txIn.Account.StateData = &accDataItem
	txIn.InMsg = &msgOut

	s.Transactions = []*core.Transaction{&txOut, &txIn}

	m := master
	m.Shards = []*core.Block{&sv, &s}

	if len(blocks) != 1 {
		panic(fmt.Errorf("wrong len, expected: %d, got: %d", 2, len(blocks)))
	}
	if !reflect.DeepEqual(&m, blocks[0]) {
		panic(fmt.Errorf("expected: %v, got: %v", m, blocks[0]))
	}

	graph, err := json.Marshal(blocks[0])
	if err != nil {
		panic(err)
	}

	fmt.Printf("%s", graph)
}

func Example_blockRepo_GetBlocks_writeFile() {
	var wc int32 = -1

	initDB()

	f := &core.BlockFilter{
		Workchain:                      &wc,
		WithShards:                     true,
		WithTransactionAccountState:    true,
		WithTransactionAccountData:     true,
		WithTransactions:               true,
		WithTransactionMessages:        true,
		WithTransactionMessagePayloads: true,

		Order: "DESC",
		Limit: 12,
	}

	blocks, err := blockRepo.GetBlocks(ctx, f)
	if err != nil {
		panic(err)
	}

	graph, err := json.Marshal(blocks)
	if err != nil {
		panic(err)
	}

	fn := fmt.Sprintf("/tmp/%d-%d-%d.graph", wc, blocks[0].SeqNo, blocks[len(blocks)-1].SeqNo)
	file, err := os.Create(fn)
	if err != nil {
		panic(err)
	}
	defer func() { _ = file.Close() }()

	_, err = file.Write(graph)
	if err != nil {
		panic(err)
	}

	fmt.Println(fn)
}
