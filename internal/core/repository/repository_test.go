package repository_test

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"reflect"
	"testing"
	"time"

	"github.com/uptrace/bun"
	"github.com/uptrace/bun/extra/bunbig"

	"github.com/tonindexer/anton/abi"
	"github.com/tonindexer/anton/internal/addr"
	"github.com/tonindexer/anton/internal/core"
	"github.com/tonindexer/anton/internal/core/aggregate"
	"github.com/tonindexer/anton/internal/core/aggregate/history"
	"github.com/tonindexer/anton/internal/core/filter"
	"github.com/tonindexer/anton/internal/core/repository"
	"github.com/tonindexer/anton/internal/core/repository/account"
	"github.com/tonindexer/anton/internal/core/repository/block"
	"github.com/tonindexer/anton/internal/core/repository/contract"
	"github.com/tonindexer/anton/internal/core/repository/msg"
	"github.com/tonindexer/anton/internal/core/repository/tx"
)

var (
	ctx = context.Background()

	_db *repository.DB

	accountRepo repository.Account
	abiRepo     repository.Contract
	blockRepo   repository.Block
	txRepo      repository.Transaction
	msgRepo     repository.Message
)

func initDB() {
	var err error

	_db, err = repository.ConnectDB(context.Background(),
		"clickhouse://localhost:9000/default?sslmode=disable",
		"postgres://user:pass@localhost:5432/default?sslmode=disable")
	if err != nil {
		panic(err)
	}

	accountRepo = account.NewRepository(_db.CH, _db.PG)
	abiRepo = contract.NewRepository(_db.PG)
	blockRepo = block.NewRepository(_db.CH, _db.PG)
	txRepo = tx.NewRepository(_db.CH, _db.PG)
	msgRepo = msg.NewRepository(_db.CH, _db.PG)
}

func dropTables(t *testing.T) { //nolint:gocyclo // clean database
	var err error

	// TODO: drop pg enums

	_, err = _db.CH.NewDropTable().Model((*core.Transaction)(nil)).IfExists().Exec(ctx)
	if err != nil {
		t.Fatal(err)
	}
	_, err = _db.PG.NewDropTable().Model((*core.Transaction)(nil)).IfExists().Exec(ctx)
	if err != nil {
		t.Fatal(err)
	}

	_, err = _db.CH.NewDropTable().Model((*core.Message)(nil)).IfExists().Exec(ctx)
	if err != nil {
		t.Fatal(err)
	}
	_, err = _db.PG.NewDropTable().Model((*core.Message)(nil)).IfExists().Exec(ctx)
	if err != nil {
		t.Fatal(err)
	}

	_, err = _db.CH.NewDropTable().Model((*core.MessagePayload)(nil)).IfExists().Exec(ctx)
	if err != nil {
		t.Fatal(err)
	}
	_, err = _db.PG.NewDropTable().Model((*core.MessagePayload)(nil)).IfExists().Exec(ctx)
	if err != nil {
		t.Fatal(err)
	}

	_, err = _db.PG.NewDropTable().Model((*core.LatestAccountState)(nil)).IfExists().Exec(ctx)
	if err != nil {
		t.Fatal(err)
	}

	_, err = _db.CH.NewDropTable().Model((*core.AccountState)(nil)).IfExists().Exec(ctx)
	if err != nil {
		t.Fatal(err)
	}
	_, err = _db.PG.NewDropTable().Model((*core.AccountState)(nil)).IfExists().Exec(ctx)
	if err != nil {
		t.Fatal(err)
	}
	_, err = _db.CH.NewDropTable().Model((*core.AccountData)(nil)).IfExists().Exec(ctx)
	if err != nil {
		t.Fatal(err)
	}
	_, err = _db.PG.NewDropTable().Model((*core.AccountData)(nil)).IfExists().Exec(ctx)
	if err != nil {
		t.Fatal(err)
	}

	_, err = _db.CH.NewDropTable().Model((*core.Block)(nil)).IfExists().Exec(ctx)
	if err != nil {
		t.Fatal(err)
	}
	_, err = _db.PG.NewDropTable().Model((*core.Block)(nil)).IfExists().Exec(ctx)
	if err != nil {
		t.Fatal(err)
	}

	_, err = _db.CH.NewDropTable().Model((*core.ContractOperation)(nil)).IfExists().Exec(ctx)
	if err != nil {
		t.Fatal(err)
	}
	_, err = _db.PG.NewDropTable().Model((*core.ContractOperation)(nil)).IfExists().Exec(ctx)
	if err != nil {
		t.Fatal(err)
	}

	_, err = _db.CH.NewDropTable().Model((*core.ContractInterface)(nil)).IfExists().Exec(ctx)
	if err != nil {
		t.Fatal(err)
	}
	_, err = _db.PG.NewDropTable().Model((*core.ContractInterface)(nil)).IfExists().Exec(ctx)
	if err != nil {
		t.Fatal(err)
	}
}

func createTables(t *testing.T) {
	err := block.CreateTables(ctx, _db.CH, _db.PG)
	if err != nil {
		t.Fatal(err)
	}
	err = account.CreateTables(ctx, _db.CH, _db.PG)
	if err != nil {
		t.Fatal(err)
	}
	err = tx.CreateTables(ctx, _db.CH, _db.PG)
	if err != nil {
		t.Fatal(err)
	}
	err = msg.CreateTables(ctx, _db.CH, _db.PG)
	if err != nil {
		t.Fatal(err)
	}
	err = contract.CreateTables(ctx, _db.PG)
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
		createTables(t)
	})

	t.Run("insert known interfaces", func(t *testing.T) {
		err := repository.InsertKnownInterfaces(ctx, _db.PG)
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
		createTables(t)
	})

	t.Run("insert interfaces", func(t *testing.T) {
		_, err := _db.PG.NewInsert().Model(&ifaceItem).Exec(ctx)
		if err != nil {
			t.Fatal(err)
		}
		_, err = _db.PG.NewInsert().Model(&opItemTransfer).Exec(ctx)
		if err != nil {
			t.Fatal(err)
		}
	})

	t.Run("create insert transaction", func(t *testing.T) {
		var err error
		insertTx, err = _db.PG.Begin()
		if err != nil {
			t.Fatal(err)
		}
	})

	t.Run("add account data", func(t *testing.T) {
		err := accountRepo.AddAccountData(ctx, insertTx, []*core.AccountData{&accDataWallet, &accDataItem, &accDataMinter})
		if err != nil {
			t.Fatal(err)
		}
		if err := accountRepo.AddAccountData(ctx, insertTx, nil); err != nil {
			t.Fatal(err)
		}

		sd := new(core.AccountData)
		if err := _db.CH.NewSelect().Model(sd).Where("address = ?", &accDataItem.Address).Where("last_tx_lt = ?", accDataItem.LastTxLT).Scan(ctx); err != nil {
			t.Fatal(err)
		}
		ad := accDataItem
		ad.TotalSupply, ad.TotalSupply, sd.ContentImageData, sd.Errors, ad.UpdatedAt =
			bunbig.FromInt64(0), bunbig.FromInt64(0), nil, nil, ad.UpdatedAt.Local()
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
		err = accountRepo.AddAccountStates(ctx, insertTx, []*core.AccountState{&accItem, &accItemMinter})
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
		if err := _db.CH.NewSelect().Model(s).Where("address = ?", &accWallet.Address).Where("last_tx_lt = ?", accWallet.LastTxLT).Scan(ctx); err != nil {
			t.Fatal(err)
		}
		acc := accWallet
		s.GetMethodHashes, acc.UpdatedAt = nil, acc.UpdatedAt.Local()
		if !reflect.DeepEqual(s, &acc) {
			t.Fatalf("wrong account, expected: %+v, got: %+v", acc, s)
		}
	})

	t.Run("add message payloads", func(t *testing.T) {
		err := msgRepo.AddMessagePayloads(ctx, insertTx, []*core.MessagePayload{&msgInItemPayload})
		if err != nil {
			t.Fatal(err)
		}
		if err := msgRepo.AddMessagePayloads(ctx, insertTx, nil); err != nil {
			t.Fatal(err)
		}
	})

	t.Run("add messages", func(t *testing.T) {
		err := msgRepo.AddMessages(ctx, insertTx, []*core.Message{&msgExtWallet, &msgOutWallet})
		if err != nil {
			t.Fatal(err)
		}
		if err := msgRepo.AddMessages(ctx, insertTx, nil); err != nil {
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

func TestGraphFilterAccounts(t *testing.T) { //nolint:gocognit,gocyclo // a lot of functions
	initDB()

	// TODO: optional fields
	accWalletOlder.Code, accWalletOlder.Data = nil, nil
	accWallet.Code, accWallet.Data = nil, nil
	accItem.Code, accItem.Data = nil, nil

	t.Run("filter state by address", func(t *testing.T) {
		results, err := accountRepo.FilterAccounts(ctx, &filter.AccountsReq{
			Addresses:   []*addr.Address{&accWallet.Address},
			LatestState: false,
			WithData:    true,
			Order:       "ASC",
			Limit:       1,
		})
		if err != nil {
			t.Fatal(err)
		}
		ret := results.Rows

		if results.Total != 3 {
			t.Fatalf("wrong len, expected: %d, got: %d", 3, results.Total)
		}
		if len(ret) != 1 {
			t.Fatalf("wrong len, expected: %d, got: %d", 1, len(ret))
		}
		if !reflect.DeepEqual(&accWalletOlder, ret[0]) {
			t.Fatalf("wrong account, expected: %+v, got: %+v", accWalletOlder, ret[0])
		}
	})

	t.Run("filter latest state by address", func(t *testing.T) {
		results, err := accountRepo.FilterAccounts(ctx, &filter.AccountsReq{
			Addresses:   []*addr.Address{&accWallet.Address},
			LatestState: true,
			Order:       "DESC",
			Limit:       1,
		})
		if err != nil {
			t.Fatal(err)
		}
		ret := results.Rows

		if results.Total != 1 {
			t.Fatalf("wrong len, expected: %d, got: %d", 1, results.Total)
		}
		if len(ret) != 1 {
			t.Fatalf("wrong len, expected: %d, got: %d", 1, len(ret))
		}
		if !reflect.DeepEqual(&accWallet, ret[0]) {
			t.Fatalf("wrong account, expected: %+v, got: %+v", accWallet, ret[0])
		}
	})

	t.Run("filter latest state by address", func(t *testing.T) {
		results, err := accountRepo.FilterAccounts(ctx, &filter.AccountsReq{
			Addresses:   []*addr.Address{&accWallet.Address},
			LatestState: true,
			WithData:    true,
			Order:       "DESC",
			Limit:       1,
		})
		if err != nil {
			t.Fatal(err)
		}
		ret := results.Rows

		acc := accWallet
		acc.StateData = &accDataWallet

		if results.Total != 1 {
			t.Fatalf("wrong len, expected: %d, got: %d", 1, results.Total)
		}
		if len(ret) != 1 {
			t.Fatalf("wrong len, expected: %d, got: %d", 1, len(ret))
		}
		if !reflect.DeepEqual(acc.StateData, ret[0].StateData) {
			t.Fatalf("wrong account data, expected: %+v, got: %+v", acc.StateData, ret[0].StateData)
		}
		if !reflect.DeepEqual(&acc, ret[0]) {
			t.Fatalf("wrong account, expected: %+v, got: %+v", acc, ret[0])
		}
	})

	t.Run("filter latest item account states by types", func(t *testing.T) {
		results, err := accountRepo.FilterAccounts(ctx, &filter.AccountsReq{
			LatestState:   true,
			WithData:      true,
			ContractTypes: []abi.ContractName{abi.NFTItem},
			Order:         "DESC",
			Limit:         1,
		})
		if err != nil {
			t.Fatal(err)
		}
		ret := results.Rows

		acc := accItem
		acc.StateData = &accDataItem

		if results.Total != 1 {
			t.Fatalf("wrong len, expected: %d, got: %d", 1, results.Total)
		}
		if len(ret) != 1 {
			t.Fatalf("wrong len, expected: %d, got: %d", 1, len(ret))
		}
		if !reflect.DeepEqual(&acc, ret[0]) {
			t.Fatalf("wrong account, expected: %+v, got: %+v", acc, ret[0])
		}
	})

	t.Run("filter latest item account states by minter address", func(t *testing.T) {
		results, err := accountRepo.FilterAccounts(ctx, &filter.AccountsReq{
			LatestState:   true,
			WithData:      true,
			MinterAddress: accDataItem.MinterAddress,
			Order:         "DESC",
			Limit:         1,
		})
		if err != nil {
			t.Fatal(err)
		}
		ret := results.Rows

		acc := accItem
		acc.StateData = &accDataItem

		if results.Total != 1 {
			t.Fatalf("wrong len, expected: %d, got: %d", 1, results.Total)
		}
		if len(ret) != 1 {
			t.Fatalf("wrong len, expected: %d, got: %d", 1, len(ret))
		}
		if !reflect.DeepEqual(&acc, ret[0]) {
			t.Fatalf("wrong account, expected: %+v, got: %+v", acc, ret[0])
		}
	})
}

func TestGraphAggregateAccounts(t *testing.T) { //nolint:gocognit,gocyclo // a lot of functions
	initDB()

	t.Run("aggregate nft collection", func(t *testing.T) {
		res, err := accountRepo.AggregateAccounts(ctx, &aggregate.AccountsReq{
			MinterAddress: accDataItem.MinterAddress,
			Limit:         1000,
		})
		if err != nil {
			t.Fatal(err)
		}

		if res.Items != 1 {
			t.Fatalf("expected: %d, got: %d", 1, res.Items)
		}
		if res.OwnersCount != 1 {
			t.Fatalf("expected: %d, got: %d", 1, res.OwnersCount)
		}
		if len(res.UniqueOwners) != 1 || res.UniqueOwners[0].ItemAddress.String() != accDataItem.Address.String() || res.UniqueOwners[0].OwnersCount != 1 {
			t.Fatalf("wrong owned items: %+v", res.OwnedItems)
		}
		if len(res.OwnedItems) != 1 || res.OwnedItems[0].OwnerAddress.String() != accDataItem.OwnerAddress.String() || res.OwnedItems[0].ItemsCount != 1 {
			t.Fatalf("wrong owned items: %+v", res.OwnedItems)
		}
	})

	t.Run("aggregate jetton wallets", func(t *testing.T) {
		res, err := accountRepo.AggregateAccounts(ctx, &aggregate.AccountsReq{
			MinterAddress: accDataItem.MinterAddress,
			Limit:         1000,
		})
		if err != nil {
			t.Fatal(err)
		}

		if res.Wallets != 1 {
			t.Fatalf("expected: %d, got: %d", 1, res.Items)
		}
		if res.TotalSupply == nil || res.TotalSupply.ToUInt64() != accDataItem.JettonBalance.ToUInt64() {
			t.Fatalf("expected: %s, got: %s", accDataItem.JettonBalance, res.TotalSupply)
		}
		if len(res.OwnedBalance) != 1 || res.OwnedBalance[0].OwnerAddress.String() != accDataItem.OwnerAddress.String() || res.OwnedBalance[0].Balance.ToUInt64() != accDataItem.JettonBalance.ToUInt64() {
			t.Fatalf("wrong owned balance: %+v", res.OwnedItems)
		}
	})
}

func TestGraphAggregateAccountsHistory(t *testing.T) {
	initDB()

	t.Run("count uniq wallet addresses", func(t *testing.T) {
		res, err := accountRepo.AggregateAccountsHistory(ctx, &history.AccountsReq{
			Metric:        history.ActiveAddresses,
			ContractTypes: []abi.ContractName{"wallet"},
			ReqParams:     history.ReqParams{Interval: time.Hour},
		})
		if err != nil {
			t.Fatal(err)
		}

		t.Logf("%+v", res)
	})
}

func TestGraphFilterMessages(t *testing.T) {
	initDB()

	t.Run("filter messages by operation name with source", func(t *testing.T) {
		res, err := msgRepo.FilterMessages(ctx, &filter.MessagesReq{
			WithPayload:    true,
			OperationNames: []string{"item_transfer"},
			Order:          "DESC",
			Limit:          10,
		})
		if err != nil {
			t.Fatal(err)
		}

		ret := res.Rows
		if res.Total != 1 {
			t.Fatalf("wrong len, expected: %d, got: %d", 1, res.Total)
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

	t.Run("filter messages by minter address", func(t *testing.T) {
		res, err := msgRepo.FilterMessages(ctx, &filter.MessagesReq{
			WithPayload:   true,
			MinterAddress: accDataItem.MinterAddress,
			Order:         "DESC",
			Limit:         10,
		})
		if err != nil {
			t.Fatal(err)
		}

		ret := res.Rows
		if res.Total != 1 {
			t.Fatalf("wrong len, expected: %d, got: %d", 1, res.Total)
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

func TestGraphAggregateMessages(t *testing.T) {
	initDB()

	t.Run("aggregate sender and receivers", func(t *testing.T) {
		res, err := msgRepo.AggregateMessages(ctx, &aggregate.MessagesReq{
			Address: &accWallet.Address,
			OrderBy: "count",
			Limit:   25,
		})
		if err != nil {
			t.Fatal(err)
		}

		if res.RecvCount != 1 {
			t.Fatalf("expected: %d, got: %d", 1, res.RecvCount)
		}
		if res.RecvAmount.ToUInt64() != 0 {
			t.Fatalf("expected: %d, got: %d", 0, res.RecvAmount.ToUInt64())
		}
		if res.SentAmount.ToUInt64() != msgOutWallet.Amount.ToUInt64() {
			t.Fatalf("expected: %d, got: %d", msgOutWallet.Amount.ToUInt64(), res.SentAmount.ToUInt64())
		}
		if res.SentCount != 1 {
			t.Fatalf("expected: %d, got: %d", 1, res.SentCount)
		}

		j, err := json.Marshal(res)
		if err != nil {
			t.Fatal(err)
		}

		t.Logf("%s", string(j))
	})
}

func TestGraphAggregateMessagesHistory(t *testing.T) {
	initDB()

	t.Run("count nft item transfers", func(t *testing.T) {
		res, err := msgRepo.AggregateMessagesHistory(ctx, &history.MessagesReq{
			Metric:         history.MessageCount,
			OperationNames: []string{"nft_item_transfer"},
			ReqParams:      history.ReqParams{Interval: time.Hour},
		})
		if err != nil {
			t.Fatal(err)
		}

		t.Logf("%+v", res)
	})

	t.Run("sum message amount", func(t *testing.T) {
		res, err := msgRepo.AggregateMessagesHistory(ctx, &history.MessagesReq{
			Metric:         history.MessageAmountSum,
			OperationNames: []string{"nft_item_transfer"},
			ReqParams:      history.ReqParams{Interval: time.Hour},
		})
		if err != nil {
			t.Fatal(err)
		}

		t.Logf("%+v", res)
	})
}

func TestGraphFilterTransactions(t *testing.T) {
	initDB()

	t.Run("filter tx by in_msg_hash", func(t *testing.T) {
		res, err := txRepo.FilterTransactions(ctx, &filter.TransactionsReq{
			InMsgHash: txInItem.InMsgHash,
			Workchain: new(int32),
			Order:     "DESC",
			Limit:     10,
		})
		if err != nil {
			t.Fatal(err)
		}

		ret := res.Rows
		if res.Total != 1 {
			t.Fatalf("wrong len, expected: %d, got: %d", 1, res.Total)
		}

		txIn := txInItem

		if len(ret) != 1 {
			t.Fatalf("wrong len, expected: %d, got: %d", 1, len(ret))
		}
		if !reflect.DeepEqual(&txIn, ret[0]) {
			t.Fatalf("wrong tx, expected: %+v, got: %+v", txIn, ret[0])
		}
	})

	t.Run("filter tx with msg by address", func(t *testing.T) {
		res, err := txRepo.FilterTransactions(ctx, &filter.TransactionsReq{
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

		ret := res.Rows
		if res.Total != 1 {
			t.Fatalf("wrong len, expected: %d, got: %d", 1, res.Total)
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
		res, err := txRepo.FilterTransactions(ctx, &filter.TransactionsReq{
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

		ret := res.Rows
		if res.Total != 1 {
			t.Fatalf("wrong len, expected: %d, got: %d", 1, res.Total)
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

func TestGraphAggregateTransactionsHistory(t *testing.T) {
	initDB()

	t.Run("count transactions", func(t *testing.T) {
		res, err := txRepo.AggregateTransactionsHistory(ctx, &history.TransactionsReq{
			Metric:    history.TransactionCount,
			ReqParams: history.ReqParams{Interval: time.Hour},
		})
		if err != nil {
			t.Fatal(err)
		}

		t.Logf("%+v", res)
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

		f := &filter.BlocksReq{
			Workchain:  &wc,
			WithShards: true,
			Order:      "DESC",
			Limit:      100,
		}

		res, err := blockRepo.FilterBlocks(ctx, f)
		if err != nil {
			t.Error(err)
		}

		m := master
		m.Shards = []*core.Block{&shardPrev, &shard}

		if res.Total != 1 {
			t.Fatalf("wrong len, expected: %d, got: %d", 1, res.Total)
		}
		if len(res.Rows) != 1 {
			t.Fatalf("wrong len, expected: %d, got: %d", 1, len(res.Rows))
		}
		if !reflect.DeepEqual(&m, res.Rows[0]) {
			t.Fatalf("wrong master block, expected: %v, got: %v", master, res.Rows[0])
		}
	})

	t.Run("filter shard blocks", func(t *testing.T) {
		var wc int32 = 0

		f := &filter.BlocksReq{
			Workchain: &wc,

			Order: "DESC",
			Limit: 100,
		}

		res, err := blockRepo.FilterBlocks(ctx, f)
		if err != nil {
			t.Error(err)
		}

		if res.Total != 2 {
			t.Fatalf("wrong len, expected: %d, got: %d", 2, res.Total)
		}
		if len(res.Rows) != 2 {
			t.Fatalf("wrong len, expected: %d, got: %d", 2, len(res.Rows))
		}
		if exp := []*core.Block{&shard, &shardPrev}; !reflect.DeepEqual(exp, res.Rows) {
			t.Fatalf("wrong shard block, expected: %v, got: %v", exp, res.Rows)
		}
	})
}

func TestGetStatistics(t *testing.T) {
	initDB()

	stats, err := aggregate.GetStatistics(ctx, _db.CH, _db.PG)
	if err != nil {
		t.Fatal(err)
	}

	j, err := json.Marshal(stats)
	if err != nil {
		t.Fatal(err)
	}

	t.Logf("%s", string(j))
}

func Example_blockRepo_GetBlocks() {
	var wc int32 = -1

	initDB()

	f := &filter.BlocksReq{
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

	res, err := blockRepo.FilterBlocks(ctx, f)
	if err != nil {
		panic(err)
	}
	blocks := res.Rows

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

	f := &filter.BlocksReq{
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

	res, err := blockRepo.FilterBlocks(ctx, f)
	if err != nil {
		panic(err)
	}
	blocks := res.Rows

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
