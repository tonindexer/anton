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

		sd := new(core.AccountData)
		if err := db.CH.NewSelect().Model(sd).Where("address = ?", &accDataItem.Address).Where("last_tx_lt = ?", accDataItem.LastTxLT).Scan(ctx); err != nil {
			t.Fatal(err)
		}
		ad := accDataItem
		ad.TotalSupply = bunbig.FromInt64(0)
		sd.ContentImageData = nil
		sd.Errors = nil
		if !reflect.DeepEqual(sd, &ad) {
			t.Fatalf("wrong account data, expected: %+v, got: %+v", ad, sd)
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

	t.Run("filter latest state by address", func(t *testing.T) {
		ret, err := accountRepo.GetAccountStates(ctx, &core.AccountStateFilter{
			Address:     &accWallet.Address,
			LatestState: true,
		}, 0, 100)
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
			Address:     &accWallet.Address,
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
			t.Fatalf("wrong account, expected: %+v, got: %+v", acc, ret[0])
		}
	})

	t.Run("filter latest item account states by types", func(t *testing.T) {
		ret, err := accountRepo.GetAccountStates(ctx, &core.AccountStateFilter{
			LatestState:   true,
			WithData:      true,
			ContractTypes: []abi.ContractName{"item"},
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
			t.Fatalf("wrong account, expected: %+v, got: %+v", acc, ret[0])
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
			t.Fatalf("wrong account, expected: %+v, got: %+v", acc, ret[0])
		}
	})
}

func TestGraphFilterMessages(t *testing.T) {
	initDB()

	t.Run("filter messages by operation name with source", func(t *testing.T) {
		ret, err := txRepo.GetMessages(ctx, &core.MessageFilter{
			WithPayload:   true,
			OperationName: "item_transfer",
		}, 0, 100)
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
			Address:             &accWallet.Address,
			WithAccountState:    true,
			WithMessages:        true,
			WithMessagePayloads: true,
		}, 0, 100)
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
			Address:             &accItem.Address,
			WithAccountState:    true,
			WithAccountData:     true,
			WithMessages:        true,
			WithMessagePayloads: true,
		}, 0, 100)
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
		}

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
			t.Fatalf("wrong master block, expected: %v, got: %v", master, blocks[0])
		}
	})

	t.Run("filter shard blocks", func(t *testing.T) {
		var wc int32 = 0

		f := &core.BlockFilter{
			Workchain: &wc,
		}

		blocks, err := blockRepo.GetBlocks(ctx, f, 0, 100)
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
	}

	blocks, err := blockRepo.GetBlocks(ctx, f, 0, 100)
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
	// Output: {"workchain":-1,"shard":2222,"seq_no":1234,"file_hash":"Uv38ByGCZU8WP18PmmIdcpVmx00QA3xNe7sEB9Hixkk=","root_hash":"gYVa2GgdDYbR6R4AFnk5y2aU0sQirNIIoAcpOUh/aZk=","master":{},"shards":[{"shard":8888,"seq_no":4320,"file_hash":"YyUlP+xzjdep4ov5IRGcFg8HAkSGFbvaCDE/ao62aNI=","root_hash":"C/UFmHWSHmaKW98sf8SERZLSVyvNBmjS1sUvUFTi0IM=","master":{"workchain":-1,"shard":2222,"seq_no":1234},"transactions":null},{"shard":8888,"seq_no":4321,"file_hash":"650YpEeEBF2H88Z88idG6ZWvWiU2eVG6ov9s1HHEg/E=","root_hash":"X7kLrbN8WCG22VUmpBqVBGgLTnyLdjobHUnUlVyEhiE=","master":{"workchain":-1,"shard":2222,"seq_no":1234},"transactions":[{"hash":"4sr8yuOmH7WGsUMjpryPnn3x2SkzP/mTkzvqb1s69t4=","address":{"hex":"0:6bf84c7174cb7476364cc3dbd968b0f7172ed85794bb358b0c3b525da1786f9f","base64":"AABr-ExxdMt0djZMw9vZaLD3Fy7YV5S7NYsMO1JdoXhvn0mu"},"account":{"address":{"hex":"0:6bf84c7174cb7476364cc3dbd968b0f7172ed85794bb358b0c3b525da1786f9f","base64":"AABr-ExxdMt0djZMw9vZaLD3Fy7YV5S7NYsMO1JdoXhvn0mu"},"latest":true,"is_active":true,"status":"ACTIVE","balance":{},"last_tx_lt":10683692646452564431,"last_tx_hash":"4sr8yuOmH7WGsUMjpryPnn3x2SkzP/mTkzvqb1s69t4=","state_data":{"address":{"hex":"0:6bf84c7174cb7476364cc3dbd968b0f7172ed85794bb358b0c3b525da1786f9f","base64":"AABr-ExxdMt0djZMw9vZaLD3Fy7YV5S7NYsMO1JdoXhvn0mu"},"last_tx_lt":10683692646452564431,"last_tx_hash":"4sr8yuOmH7WGsUMjpryPnn3x2SkzP/mTkzvqb1s69t4=","types":["wallet"]},"state_hash":"A3Q2bEcZ5DobBn2JvH8B8fVzmBZZpE/xekxyFaO1Oes=","code":"HlhJxgd9u1ci9XF6KJomb5dkeYGZjr6onAtLNzlwEV6C7W9BJcj6cxHk1976ki2q53hmZ/fpNs1PJKv334ZrqlYDg2etYUXeHuj0qLCZPr34iDoK2L6cOXiwSIPlahVqjeVjr6Rn1J3sakDpodAH8DPCgjBhvdDqpZ+OTaZDAQU=","code_hash":"Ig0LKWiLc0uOoPPKmTboRh8Q13yW6oCnpmX2BvamO38=","data":"Pf0lZ8GJeeTWDyZobZvy+ybJAf81TN4WB+4pSznzK3x4Irpk+Eq0PKDG5rkcH9O+iZBDQXnTr0SRo2kBLbktGE/DnRc0/1cWQolTu2hl/PkrDDoXyQKL6ZFOt2ScbJNHgAl50YMDVvKlTD3qsqS0R11jr76PtWmHx39YGFJvGBQ=","data_hash":"voIzUOqxOTXzHYRIRRfpJK73iuFRwAdVklg2twdYhWU=","depth":0,"tick":false,"tock":false,"get_method_hashes":null},"block_workchain":0,"block_shard":8888,"block_seq_no":4321,"prev_tx_hash":"nsHgtyewMHLmQVp2HwOrqkCryUSP3eshkdlFwEdnr4Q=","prev_tx_lt":760740741943613320,"in_msg_hash":"1LzpZO1H90qllEaM7TI8t28NP6xHbJ+wP8kij7roj9U=","in_msg":{"type":"EXTERNAL_IN","hash":"1LzpZO1H90qllEaM7TI8t28NP6xHbJ+wP8kij7roj9U=","src_address":{"hex":"0:0000000000000000000000000000000000000000000000000000000000000000","base64":"AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA"},"dst_address":{"hex":"0:6bf84c7174cb7476364cc3dbd968b0f7172ed85794bb358b0c3b525da1786f9f","base64":"AABr-ExxdMt0djZMw9vZaLD3Fy7YV5S7NYsMO1JdoXhvn0mu"},"bounce":false,"bounced":false,"amount":null,"ihr_disabled":false,"ihr_fee":null,"fwd_fee":null,"body":"gGY6BFS2gxIgfwo7WExiMWSStJdTtdUCfOFaTwpYJQ2PtQ538r9PAVLl1JQ1gH+dS5e+b7d5cEZqVib+M0CM+eiOLHl0CKMtKUFrryBqMpz//Up15JgyCYLIWq1wOEhZwFpLE6HVsvW/71pu2S2kgsqpVo5bb+nYqd3Z6wkne5I=","body_hash":"zvkEbvoYUAlEy+gAoLFSfqZHKahh0vZJejI1w39Bknc=","created_at":17941254959206722521,"created_lt":10683692646452564431},"out_msg":[{"type":"INTERNAL","hash":"ev0O211K/6vjA3/+f6aKqK9eOcxBbnNNNzxevryc3MU=","src_address":{"hex":"0:6bf84c7174cb7476364cc3dbd968b0f7172ed85794bb358b0c3b525da1786f9f","base64":"AABr-ExxdMt0djZMw9vZaLD3Fy7YV5S7NYsMO1JdoXhvn0mu"},"dst_address":{"hex":"0:0c30ec29a3703934bf50a28da102975deda77e758579ea3dfe4136abf752b3b8","base64":"AAAMMOwpo3A5NL9Qoo2hApdd7ad-dYV56j3-QTar91KzuH4N"},"source_tx_hash":"4sr8yuOmH7WGsUMjpryPnn3x2SkzP/mTkzvqb1s69t4=","source_tx_lt":10683692646452564431,"bounce":false,"bounced":false,"amount":{},"ihr_disabled":false,"ihr_fee":{},"fwd_fee":{},"body":"lbzOPHvT2N+T+rfhJd3rr+ZaMb1dQeLSzpwrF4kvD+o=","body_hash":"GTGikCIHd6kxQ9/cv6aEBuh3Bz/wiDThl6QDSqSK+j8=","operation_id":16772846,"payload":{"type":"INTERNAL","hash":"ev0O211K/6vjA3/+f6aKqK9eOcxBbnNNNzxevryc3MU=","src_address":{"hex":"0:6bf84c7174cb7476364cc3dbd968b0f7172ed85794bb358b0c3b525da1786f9f","base64":"AABr-ExxdMt0djZMw9vZaLD3Fy7YV5S7NYsMO1JdoXhvn0mu"},"src_contract":"wallet","dst_address":{"hex":"0:0c30ec29a3703934bf50a28da102975deda77e758579ea3dfe4136abf752b3b8","base64":"AAAMMOwpo3A5NL9Qoo2hApdd7ad-dYV56j3-QTar91KzuH4N"},"dst_contract":"item","body_hash":"GTGikCIHd6kxQ9/cv6aEBuh3Bz/wiDThl6QDSqSK+j8=","operation_id":16772846,"operation_name":"item_transfer","data":{"new_owner":"kkkkkk","collection_address":"aaaaaa"},"created_at":17941254959206722521,"created_lt":10683692646452564432},"created_at":17941254959206722521,"created_lt":10683692646452564432}],"total_fees":{},"compute_success":false,"msg_state_used":false,"account_activated":false,"gas_fees":null,"vm_gas_used":null,"vm_gas_limit":null,"vm_gas_credit":null,"vm_mode":0,"vm_exit_code":0,"vm_exit_arg":0,"vm_steps":0,"orig_status":"ACTIVE","end_status":"ACTIVE","created_at":17941254959206722521,"created_lt":10683692646452564431},{"hash":"Jx0D6USzyds2a3UEX479adIq5UEZR8tVPXaUJnrvTrw=","address":{"hex":"0:0c30ec29a3703934bf50a28da102975deda77e758579ea3dfe4136abf752b3b8","base64":"AAAMMOwpo3A5NL9Qoo2hApdd7ad-dYV56j3-QTar91KzuH4N"},"account":{"address":{"hex":"0:0c30ec29a3703934bf50a28da102975deda77e758579ea3dfe4136abf752b3b8","base64":"AAAMMOwpo3A5NL9Qoo2hApdd7ad-dYV56j3-QTar91KzuH4N"},"latest":true,"is_active":true,"status":"ACTIVE","balance":{},"last_tx_lt":10683692646452564441,"last_tx_hash":"Jx0D6USzyds2a3UEX479adIq5UEZR8tVPXaUJnrvTrw=","state_data":{"address":{"hex":"0:0c30ec29a3703934bf50a28da102975deda77e758579ea3dfe4136abf752b3b8","base64":"AAAMMOwpo3A5NL9Qoo2hApdd7ad-dYV56j3-QTar91KzuH4N"},"last_tx_lt":10683692646452564441,"last_tx_hash":"Jx0D6USzyds2a3UEX479adIq5UEZR8tVPXaUJnrvTrw=","types":["item"],"owner_address":{"hex":"0:3e9ac0b00ce73bff706f7ff4b6f44090a32711f3208e4e4b89cb5165ce64002c","base64":"AAA-msCwDOc7_3Bvf_S29ECQoycR8yCOTkuJy1FlzmQALNwa"},"next_item_index":{},"royalty_address":{"hex":"0:bd9c2887aa113df2468928d5a23b9ca740f80c9382d9c6034ad2960c796503e1","base64":"AAC9nCiHqhE98kaJKNWiO5ynQPgMk4LZxgNK0pYMeWUD4SF5"},"content_uri":"git://asdf.t","item_index":42,"collection_address":{"hex":"0:ce221725f50caf1fbfe831b10b7bf5b15c47a53dbf8e7dcafc9e138647a4b44e","base64":"AADOIhcl9QyvH7_oMbELe_WxXEelPb-Ofcr8nhOGR6S0TgpV"},"balance":{}},"state_hash":"6kBrMtYQi9aFhPV+N8qsbjP+qjJjo5lDcCS6nJsUZ4o=","code":"J08BqRCuKV9u+/5fWr9EzN4mO1YGYz4r8ABvKCldfTkGnwGiOcQ2WFTDr39rQdYx+SuajRL0ElcyX/8zL3V2sGIFVjBKPj6uFMKNDOo50pAaUnINqFyh5LOOrz9Exsbvg2Ly9U/ADgnW/CVkCFTBXfysqoos7M5aOrpTq3BbGNs=","code_hash":"lLTTOKUUPmNAjYcksM8/rhej95vhBy+2PDXWBCxBYPM=","data":"juniqfP7T/sAGbRU1SK1/6F2BBk/uJZnEKeWBzLKUs9Tw/UgyIm3m/UEz7V8dgEjLVibrM6p1uJj4lwndB0/bGLLuxXZr7y/f32kGrBAjjlpwuLNzyM0OL8XdKzncJpPCR6ag/3q4OxV6yM6m1OUyzx4VrVG0xPIo7TBwOBUR/Q=","data_hash":"ujcOs228/eyQswLc3Due9SLipvHtCv7B+OIPqr7faxY=","depth":0,"tick":false,"tock":false,"get_method_hashes":[127950]},"block_workchain":0,"block_shard":8888,"block_seq_no":4321,"prev_tx_hash":"hbimJwjK67rIgLW4m5PaU4EBZEAhBOZItiJqG3gCGFE=","prev_tx_lt":9891590185009426703,"in_msg_hash":"ev0O211K/6vjA3/+f6aKqK9eOcxBbnNNNzxevryc3MU=","in_msg":{"type":"INTERNAL","hash":"ev0O211K/6vjA3/+f6aKqK9eOcxBbnNNNzxevryc3MU=","src_address":{"hex":"0:6bf84c7174cb7476364cc3dbd968b0f7172ed85794bb358b0c3b525da1786f9f","base64":"AABr-ExxdMt0djZMw9vZaLD3Fy7YV5S7NYsMO1JdoXhvn0mu"},"dst_address":{"hex":"0:0c30ec29a3703934bf50a28da102975deda77e758579ea3dfe4136abf752b3b8","base64":"AAAMMOwpo3A5NL9Qoo2hApdd7ad-dYV56j3-QTar91KzuH4N"},"source_tx_hash":"4sr8yuOmH7WGsUMjpryPnn3x2SkzP/mTkzvqb1s69t4=","source_tx_lt":10683692646452564431,"bounce":false,"bounced":false,"amount":{},"ihr_disabled":false,"ihr_fee":{},"fwd_fee":{},"body":"lbzOPHvT2N+T+rfhJd3rr+ZaMb1dQeLSzpwrF4kvD+o=","body_hash":"GTGikCIHd6kxQ9/cv6aEBuh3Bz/wiDThl6QDSqSK+j8=","operation_id":16772846,"payload":{"type":"INTERNAL","hash":"ev0O211K/6vjA3/+f6aKqK9eOcxBbnNNNzxevryc3MU=","src_address":{"hex":"0:6bf84c7174cb7476364cc3dbd968b0f7172ed85794bb358b0c3b525da1786f9f","base64":"AABr-ExxdMt0djZMw9vZaLD3Fy7YV5S7NYsMO1JdoXhvn0mu"},"src_contract":"wallet","dst_address":{"hex":"0:0c30ec29a3703934bf50a28da102975deda77e758579ea3dfe4136abf752b3b8","base64":"AAAMMOwpo3A5NL9Qoo2hApdd7ad-dYV56j3-QTar91KzuH4N"},"dst_contract":"item","body_hash":"GTGikCIHd6kxQ9/cv6aEBuh3Bz/wiDThl6QDSqSK+j8=","operation_id":16772846,"operation_name":"item_transfer","data":{"new_owner":"kkkkkk","collection_address":"aaaaaa"},"created_at":17941254959206722521,"created_lt":10683692646452564432},"created_at":17941254959206722521,"created_lt":10683692646452564432},"total_fees":{},"state_update":"9dmsTF+PcqyJs4sZ9TeEwZ6b6sA8h1on2wKd43rjekI=","description":"MYgTSHaFkpNZyoxeuU4VLcGvQuo9FnbBvdGauOKSXG0=","compute_success":false,"msg_state_used":false,"account_activated":false,"gas_fees":null,"vm_gas_used":null,"vm_gas_limit":null,"vm_gas_credit":null,"vm_mode":0,"vm_exit_code":0,"vm_exit_arg":0,"vm_steps":0,"orig_status":"ACTIVE","end_status":"ACTIVE","created_at":17941254959206722521,"created_lt":10683692646452564441}]}],"transactions":null}
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
	}

	blocks, err := blockRepo.GetBlocks(ctx, f, 0, 42)
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
	defer file.Close()

	_, err = file.Write(graph)
	if err != nil {
		panic(err)
	}

	fmt.Println(fn)
}
