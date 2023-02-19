package repository_test

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
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

func initDB() {
	var err error

	db, err = repository.ConnectDB(context.Background(),
		"clickhouse://localhost:9000/default?sslmode=disable",
		"postgres://user:pass@localhost:5432/default?sslmode=disable")
	if err != nil {
		panic(err)
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
			t.Fatalf("wrong account, expected: %+v, got: %+v", accWallet, ret[0])
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
		if !reflect.DeepEqual(&msgIn, ret[0]) {
			t.Fatalf("wrong tx, expected: %+v, got: %+v", msgIn, ret[0])
		}
	})
}

func TestGraphFilterTransactions(t *testing.T) {
	initDB()

	t.Run("filter tx with msg by address", func(t *testing.T) {
		ret, err := txRepo.GetTransactions(ctx, &core.TransactionFilter{
			Address:             accWallet.Address,
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
			Address:             accItem.Address,
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
	// Output: {"workchain":-1,"shard":2222,"seq_no":1234,"FileHash":"Uv38ByGCZU8WP18PmmIdcpVmx00QA3xNe7sEB9Hixkk=","RootHash":"gYVa2GgdDYbR6R4AFnk5y2aU0sQirNIIoAcpOUh/aZk=","MasterID":{"workchain":0,"shard":0,"seq_no":0},"Shards":[{"workchain":0,"shard":8888,"seq_no":4320,"FileHash":"YyUlP+xzjdep4ov5IRGcFg8HAkSGFbvaCDE/ao62aNI=","RootHash":"C/UFmHWSHmaKW98sf8SERZLSVyvNBmjS1sUvUFTi0IM=","MasterID":{"workchain":-1,"shard":2222,"seq_no":1234},"Shards":null,"Transactions":null},{"workchain":0,"shard":8888,"seq_no":4321,"FileHash":"650YpEeEBF2H88Z88idG6ZWvWiU2eVG6ov9s1HHEg/E=","RootHash":"X7kLrbN8WCG22VUmpBqVBGgLTnyLdjobHUnUlVyEhiE=","MasterID":{"workchain":-1,"shard":2222,"seq_no":1234},"Shards":null,"Transactions":[{"Hash":"4sr8yuOmH7WGsUMjpryPnn3x2SkzP/mTkzvqb1s69t4=","Address":"0:6bf84c7174cb7476364cc3dbd968b0f7172ed85794bb358b0c3b525da1786f9f","Account":{"Latest":true,"Address":"0:6bf84c7174cb7476364cc3dbd968b0f7172ed85794bb358b0c3b525da1786f9f","IsActive":true,"Status":"ACTIVE","Balance":1000000000,"LastTxLT":10683692646452564431,"LastTxHash":"4sr8yuOmH7WGsUMjpryPnn3x2SkzP/mTkzvqb1s69t4=","StateData":{"Address":"0:6bf84c7174cb7476364cc3dbd968b0f7172ed85794bb358b0c3b525da1786f9f","LastTxLT":10683692646452564431,"LastTxHash":"4sr8yuOmH7WGsUMjpryPnn3x2SkzP/mTkzvqb1s69t4=","Types":["wallet"],"OwnerAddress":"","NextItemIndex":0,"RoyaltyAddress":"","RoyaltyFactor":0,"RoyaltyBase":0,"ContentURI":"","ContentName":"","ContentDescription":"","ContentImage":"","ContentImageData":null,"Initialized":false,"ItemIndex":0,"CollectionAddress":"","EditorAddress":""},"StateHash":"A3Q2bEcZ5DobBn2JvH8B8fVzmBZZpE/xekxyFaO1Oes=","Code":"HlhJxgd9u1ci9XF6KJomb5dkeYGZjr6onAtLNzlwEV6C7W9BJcj6cxHk1976ki2q53hmZ/fpNs1PJKv334ZrqlYDg2etYUXeHuj0qLCZPr34iDoK2L6cOXiwSIPlahVqjeVjr6Rn1J3sakDpodAH8DPCgjBhvdDqpZ+OTaZDAQU=","CodeHash":"Ig0LKWiLc0uOoPPKmTboRh8Q13yW6oCnpmX2BvamO38=","Data":"Pf0lZ8GJeeTWDyZobZvy+ybJAf81TN4WB+4pSznzK3x4Irpk+Eq0PKDG5rkcH9O+iZBDQXnTr0SRo2kBLbktGE/DnRc0/1cWQolTu2hl/PkrDDoXyQKL6ZFOt2ScbJNHgAl50YMDVvKlTD3qsqS0R11jr76PtWmHx39YGFJvGBQ=","DataHash":"voIzUOqxOTXzHYRIRRfpJK73iuFRwAdVklg2twdYhWU=","Depth":0,"Tick":false,"Tock":false,"Types":["wallet"]},"BlockWorkchain":0,"BlockShard":8888,"BlockSeqNo":4321,"PrevTxHash":"MjXZazscVCT84LcnsDBy5kFadh8Dq6pAq8lEj93rIZE=","PrevTxLT":13717360897469088943,"InMsgHash":"nhOGR6S0TtS86WTtR/dKpZRGjO0yPLdvDT+sR2yfsD8=","InMsg":{"Type":"EXTERNAL_IN","Hash":"nhOGR6S0TtS86WTtR/dKpZRGjO0yPLdvDT+sR2yfsD8=","SrcAddress":"","DstAddress":"0:6bf84c7174cb7476364cc3dbd968b0f7172ed85794bb358b0c3b525da1786f9f","SourceTxHash":null,"SourceTxLT":0,"Source":null,"Bounce":false,"Bounced":false,"Amount":0,"IHRDisabled":false,"IHRFee":0,"FwdFee":0,"Body":"ySKPuuiP1YBmOgRUtoMSIH8KO1hMYjFkkrSXU7XVAnzhWk8KWCUNj7UOd/K/TwFS5dSUNYB/nUuXvm+3eXBGalYm/jNAjPnojix5dAijLSlBa68gajKc//1KdeSYMgmCyFqtcDhIWcBaSxOh1bL1v+9abtktpILKqVaOW2/p2Kk=","BodyHash":"3dnrCSd7ks75BG76GFAJRMvoAKCxUn6mRymoYdL2SXo=","OperationID":0,"TransferComment":"","Payload":null,"StateInitCode":null,"StateInitData":null,"CreatedAt":198614094973075395,"CreatedLT":10683692646452564431},"OutMsg":[{"Type":"INTERNAL","Hash":"2UXAR2eIV7eZrLGOSv+r4wN//n+miqivXjnMQW5zTTc=","SrcAddress":"0:6bf84c7174cb7476364cc3dbd968b0f7172ed85794bb358b0c3b525da1786f9f","DstAddress":"0:0c30ec29a3703934bf50a28da102975deda77e758579ea3dfe4136abf752b3b8","SourceTxHash":"4sr8yuOmH7WGsUMjpryPnn3x2SkzP/mTkzvqb1s69t4=","SourceTxLT":10683692646452564431,"Source":null,"Bounce":false,"Bounced":false,"Amount":100000,"IHRDisabled":false,"IHRFee":0,"FwdFee":0,"Body":"PF6+vJzcxZW8zjx709jfk/q34SXd66/mWjG9XUHi0s4=","BodyHash":"nCsXiS8P6hkxopAiB3epMUPf3L+mhAbodwc/8Ig04Zc=","OperationID":16772846,"TransferComment":"","Payload":{"Type":"INTERNAL","Hash":"2UXAR2eIV7eZrLGOSv+r4wN//n+miqivXjnMQW5zTTc=","SrcAddress":"0:6bf84c7174cb7476364cc3dbd968b0f7172ed85794bb358b0c3b525da1786f9f","SrcContract":"wallet","DstAddress":"0:0c30ec29a3703934bf50a28da102975deda77e758579ea3dfe4136abf752b3b8","DstContract":"item","BodyHash":"nCsXiS8P6hkxopAiB3epMUPf3L+mhAbodwc/8Ig04Zc=","OperationID":16772846,"OperationName":"item_transfer","DataJSON":"{\"new_owner\": \"kkkkkk\"}","CreatedAt":198614094973075395,"CreatedLT":10683692646452564432},"StateInitCode":null,"StateInitData":null,"CreatedAt":198614094973075395,"CreatedLT":10683692646452564432}],"TotalFees":100000,"StateUpdate":null,"Description":null,"OrigStatus":"ACTIVE","EndStatus":"ACTIVE","CreatedAt":198614094973075395,"CreatedLT":10683692646452564431},{"Hash":"Jx0D6USzyds2a3UEX479adIq5UEZR8tVPXaUJnrvTrw=","Address":"0:0c30ec29a3703934bf50a28da102975deda77e758579ea3dfe4136abf752b3b8","Account":{"Latest":true,"Address":"0:0c30ec29a3703934bf50a28da102975deda77e758579ea3dfe4136abf752b3b8","IsActive":true,"Status":"ACTIVE","Balance":1000000000,"LastTxLT":10683692646452564441,"LastTxHash":"Jx0D6USzyds2a3UEX479adIq5UEZR8tVPXaUJnrvTrw=","StateData":{"Address":"0:0c30ec29a3703934bf50a28da102975deda77e758579ea3dfe4136abf752b3b8","LastTxLT":10683692646452564441,"LastTxHash":"Jx0D6USzyds2a3UEX479adIq5UEZR8tVPXaUJnrvTrw=","Types":["item"],"OwnerAddress":"0:3e9ac0b7413ef110bd58b00ce73bff706f7ff4b6f44090a32711f3208e4e4b89","NextItemIndex":43,"RoyaltyAddress":"0:cb5165ce64002cbd9c2887aa113df2468928d5a23b9ca740f80c9382d9c6034a","RoyaltyFactor":0,"RoyaltyBase":0,"ContentURI":"git://asdf.t","ContentName":"","ContentDescription":"","ContentImage":"","ContentImageData":null,"Initialized":false,"ItemIndex":42,"CollectionAddress":"0:d2960c796503e1ce221725f50caf1fbfe831b10b7bf5b15c47a53dbf8e7dcafc","EditorAddress":""},"StateHash":"6kBrMtYQi9aFhPV+N8qsbjP+qjJjo5lDcCS6nJsUZ4o=","Code":"J08BqRCuKV9u+/5fWr9EzN4mO1YGYz4r8ABvKCldfTkGnwGiOcQ2WFTDr39rQdYx+SuajRL0ElcyX/8zL3V2sGIFVjBKPj6uFMKNDOo50pAaUnINqFyh5LOOrz9Exsbvg2Ly9U/ADgnW/CVkCFTBXfysqoos7M5aOrpTq3BbGNs=","CodeHash":"lLTTOKUUPmNAjYcksM8/rhej95vhBy+2PDXWBCxBYPM=","Data":"juniqfP7T/sAGbRU1SK1/6F2BBk/uJZnEKeWBzLKUs9Tw/UgyIm3m/UEz7V8dgEjLVibrM6p1uJj4lwndB0/bGLLuxXZr7y/f32kGrBAjjlpwuLNzyM0OL8XdKzncJpPCR6ag/3q4OxV6yM6m1OUyzx4VrVG0xPIo7TBwOBUR/Q=","DataHash":"ujcOs228/eyQswLc3Due9SLipvHtCv7B+OIPqr7faxY=","Depth":0,"Tick":false,"Tock":false,"Types":["item"]},"BlockWorkchain":0,"BlockShard":8888,"BlockSeqNo":4321,"PrevTxHash":"pANKpIr6P4W4picIyuu6yIC1uJuT2lOBAWRAIQTmSLY=","PrevTxLT":13955768992965067384,"InMsgHash":"2UXAR2eIV7eZrLGOSv+r4wN//n+miqivXjnMQW5zTTc=","InMsg":{"Type":"INTERNAL","Hash":"2UXAR2eIV7eZrLGOSv+r4wN//n+miqivXjnMQW5zTTc=","SrcAddress":"0:6bf84c7174cb7476364cc3dbd968b0f7172ed85794bb358b0c3b525da1786f9f","DstAddress":"0:0c30ec29a3703934bf50a28da102975deda77e758579ea3dfe4136abf752b3b8","SourceTxHash":"4sr8yuOmH7WGsUMjpryPnn3x2SkzP/mTkzvqb1s69t4=","SourceTxLT":10683692646452564431,"Source":null,"Bounce":false,"Bounced":false,"Amount":100000,"IHRDisabled":false,"IHRFee":0,"FwdFee":0,"Body":"PF6+vJzcxZW8zjx709jfk/q34SXd66/mWjG9XUHi0s4=","BodyHash":"nCsXiS8P6hkxopAiB3epMUPf3L+mhAbodwc/8Ig04Zc=","OperationID":16772846,"TransferComment":"","Payload":{"Type":"INTERNAL","Hash":"2UXAR2eIV7eZrLGOSv+r4wN//n+miqivXjnMQW5zTTc=","SrcAddress":"0:6bf84c7174cb7476364cc3dbd968b0f7172ed85794bb358b0c3b525da1786f9f","SrcContract":"wallet","DstAddress":"0:0c30ec29a3703934bf50a28da102975deda77e758579ea3dfe4136abf752b3b8","DstContract":"item","BodyHash":"nCsXiS8P6hkxopAiB3epMUPf3L+mhAbodwc/8Ig04Zc=","OperationID":16772846,"OperationName":"item_transfer","DataJSON":"{\"new_owner\": \"kkkkkk\"}","CreatedAt":198614094973075395,"CreatedLT":10683692646452564432},"StateInitCode":null,"StateInitData":null,"CreatedAt":198614094973075395,"CreatedLT":10683692646452564432},"OutMsg":null,"TotalFees":1000,"StateUpdate":"ImobDzE6id38RUxfj3KsibOLGfU3hMGem+rAPIdaJ9s=","Description":"Ap3jeuN6QjGIE0h2hZKTWcqMXrlOFS3Br0LqPRZ2wb0=","OrigStatus":"ACTIVE","EndStatus":"ACTIVE","CreatedAt":198614094973075395,"CreatedLT":10683692646452564441}]}],"Transactions":null}
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
