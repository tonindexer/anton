package repository_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/tonindexer/anton/abi"
	"github.com/tonindexer/anton/addr"
	"github.com/tonindexer/anton/internal/core"
	"github.com/tonindexer/anton/internal/core/aggregate"
	"github.com/tonindexer/anton/internal/core/filter"
	"github.com/tonindexer/anton/internal/core/repository"
	"github.com/tonindexer/anton/internal/core/repository/account"
	"github.com/tonindexer/anton/internal/core/repository/block"
	"github.com/tonindexer/anton/internal/core/repository/contract"
	"github.com/tonindexer/anton/internal/core/repository/msg"
	"github.com/tonindexer/anton/internal/core/repository/tx"
	"github.com/tonindexer/anton/internal/core/rndm"
)

var (
	db *repository.DB

	accountRepo repository.Account
	// abiRepo     repository.Contract
	blockRepo repository.Block
	txRepo    repository.Transaction
	msgRepo   repository.Message
)

func initDB() {
	var err error

	db, err = repository.ConnectDB(context.Background(),
		"clickhouse://user:pass@localhost:9000/default?sslmode=disable",
		"postgres://user:pass@localhost:5432/postgres?sslmode=disable")
	if err != nil {
		panic(err)
	}

	accountRepo = account.NewRepository(db.CH, db.PG)
	// abiRepo = contract.NewRepository(db.PG)
	blockRepo = block.NewRepository(db.CH, db.PG)
	txRepo = tx.NewRepository(db.CH, db.PG)
	msgRepo = msg.NewRepository(db.CH, db.PG)
}

func dropTables(t testing.TB) {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	ck, pg := db.CH, db.PG

	_, err := ck.NewDropTable().Model((*core.Transaction)(nil)).IfExists().Exec(ctx)
	assert.Nil(t, err)
	_, err = pg.NewDropTable().Model((*core.Transaction)(nil)).IfExists().Exec(ctx)
	assert.Nil(t, err)

	_, err = ck.NewDropTable().Model((*core.Message)(nil)).IfExists().Exec(ctx)
	assert.Nil(t, err)
	_, err = pg.NewDropTable().Model((*core.Message)(nil)).IfExists().Exec(ctx)
	assert.Nil(t, err)

	_, err = pg.NewDropTable().Model((*core.LatestAccountState)(nil)).IfExists().Exec(ctx)
	assert.Nil(t, err)

	_, err = ck.NewDropTable().Model((*core.AccountState)(nil)).IfExists().Exec(ctx)
	assert.Nil(t, err)
	_, err = pg.NewDropTable().Model((*core.AccountState)(nil)).IfExists().Exec(ctx)
	assert.Nil(t, err)

	_, err = ck.NewDropTable().Model((*core.AddressLabel)(nil)).IfExists().Exec(ctx)
	assert.Nil(t, err)
	_, err = pg.NewDropTable().Model((*core.AddressLabel)(nil)).IfExists().Exec(ctx)
	assert.Nil(t, err)

	_, err = pg.ExecContext(ctx, "DROP TYPE IF EXISTS account_status")
	assert.Nil(t, err)

	_, err = ck.NewDropTable().Model((*core.Block)(nil)).IfExists().Exec(ctx)
	assert.Nil(t, err)
	_, err = pg.NewDropTable().Model((*core.Block)(nil)).IfExists().Exec(ctx)
	assert.Nil(t, err)

	_, err = pg.NewDropTable().Model((*core.ContractOperation)(nil)).IfExists().Exec(ctx)
	assert.Nil(t, err)
	_, err = pg.NewDropTable().Model((*core.ContractInterface)(nil)).IfExists().Exec(ctx)
	assert.Nil(t, err)

	_, err = pg.ExecContext(ctx, "DROP TYPE IF EXISTS message_type")
	assert.Nil(t, err)
}

func createTables(t testing.TB) {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	err := block.CreateTables(ctx, db.CH, db.PG)
	assert.Nil(t, err)

	err = account.CreateTables(ctx, db.CH, db.PG)
	assert.Nil(t, err)

	err = tx.CreateTables(ctx, db.CH, db.PG)
	assert.Nil(t, err)

	err = msg.CreateTables(ctx, db.CH, db.PG)
	assert.Nil(t, err)

	err = contract.CreateTables(ctx, db.PG)
	assert.Nil(t, err)
}

func TestInsertKnownInterfaces(t *testing.T) {
	initDB()

	t.Run("drop tables", func(t *testing.T) {
		dropTables(t)
	})

	t.Run("create tables", func(t *testing.T) {
		createTables(t)
	})
}

func TestRelations(t *testing.T) {
	initDB()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// prepare data
	contractName := abi.ContractName("special")
	operation := "special_op"

	address := rndm.Address()

	state := rndm.AddressStateContract(address, contractName, nil)

	messageTo := rndm.MessageTo(address)
	messagesFrom := rndm.MessagesFrom(address, 1)
	messageTo.OperationName, messageTo.DstContract = operation, contractName

	transaction := rndm.AddressTransaction(address)

	shard := rndm.Block(0)
	master := rndm.Block(-1)

	// make related graph
	for _, m := range messagesFrom {
		m.SrcTxLT = transaction.CreatedLT
	}
	transaction.InMsgHash = messageTo.Hash
	transaction.InMsg = messageTo
	transaction.OutMsg = messagesFrom

	state.LastTxHash = transaction.Hash
	state.LastTxLT = transaction.CreatedLT
	transaction.Account = state

	transaction.Workchain = shard.Workchain
	transaction.Shard = shard.Shard
	transaction.BlockSeqNo = shard.SeqNo

	shard.Transactions = []*core.Transaction{transaction}
	shard.MasterID = &core.BlockID{Workchain: master.Workchain, Shard: master.Shard, SeqNo: master.SeqNo}
	master.Shards = []*core.Block{shard}

	// make slices
	addresses := []*addr.Address{address}
	states := []*core.AccountState{state}
	messagesTo := []*core.Message{messageTo}
	messages := append(messagesFrom, messageTo) //nolint:gocritic // append result not assigned to the same slice
	transactions := []*core.Transaction{transaction}
	blocks := []*core.Block{shard, master}

	t.Run("drop tables", func(t *testing.T) {
		dropTables(t)
	})

	t.Run("create tables", func(t *testing.T) {
		createTables(t)
	})

	t.Run("insert related data", func(t *testing.T) {
		dbtx, err := db.PG.Begin()
		assert.Nil(t, err)

		err = accountRepo.AddAccountStates(ctx, dbtx, states)
		assert.Nil(t, err)
		err = msgRepo.AddMessages(ctx, dbtx, messages)
		assert.Nil(t, err)
		err = txRepo.AddTransactions(ctx, dbtx, transactions)
		assert.Nil(t, err)
		err = blockRepo.AddBlocks(ctx, dbtx, blocks)
		assert.Nil(t, err)

		err = dbtx.Commit()
		assert.Nil(t, err)
	})

	t.Run("get account states with data", func(t *testing.T) {
		res, err := accountRepo.FilterAccounts(ctx, &filter.AccountsReq{
			Addresses:   addresses,
			LatestState: true,
		})
		assert.Nil(t, err)
		assert.Equal(t, 1, res.Total)
		assert.Equal(t, states, res.Rows)
	})

	t.Run("get messages with payloads", func(t *testing.T) {
		res, err := msgRepo.FilterMessages(ctx, &filter.MessagesReq{
			DstAddresses: addresses,
		})
		assert.Nil(t, err)
		assert.Equal(t, 1, res.Total)
		assert.Equal(t, len(messagesTo), len(res.Rows))
		for i := range messagesTo {
			assert.JSONEq(t, string(messagesTo[i].DataJSON), string(res.Rows[i].DataJSON))
			res.Rows[i].DataJSON = messagesTo[i].DataJSON
		}
		assert.Equal(t, messagesTo, res.Rows)
	})

	t.Run("get transactions with states and messages", func(t *testing.T) {
		res, err := txRepo.FilterTransactions(ctx, &filter.TransactionsReq{
			Addresses:        addresses,
			WithAccountState: true,
			WithMessages:     true,
		})
		assert.Nil(t, err)
		assert.Equal(t, 1, res.Total)
		assert.Equal(t, 1, len(res.Rows))
		assert.NotNil(t, res.Rows[0].InMsg)
		assert.JSONEq(t, string(messagesTo[0].DataJSON), string(res.Rows[0].InMsg.DataJSON))
		res.Rows[0].InMsg.DataJSON = messagesTo[0].DataJSON
		assert.Equal(t, transactions, res.Rows)
	})

	t.Run("get master block with shards and transactions", func(t *testing.T) {
		var workchain int32 = -1
		res, err := blockRepo.FilterBlocks(ctx, &filter.BlocksReq{
			Workchain:                   &workchain,
			WithShards:                  true,
			WithTransactions:            true,
			WithTransactionAccountState: true,
			WithTransactionMessages:     true,
		})
		assert.Nil(t, err)
		assert.Equal(t, 1, res.Total)
		assert.Equal(t, 1, len(res.Rows))
		assert.Equal(t, 1, len(res.Rows[0].Shards))
		assert.Equal(t, 1, len(res.Rows[0].Shards[0].Transactions))
		assert.NotNil(t, res.Rows[0].Shards[0].Transactions[0].InMsg)
		assert.JSONEq(t, string(messagesTo[0].DataJSON), string(res.Rows[0].Shards[0].Transactions[0].InMsg.DataJSON))
		res.Rows[0].Shards[0].Transactions[0].InMsg.DataJSON = messagesTo[0].DataJSON
		assert.Equal(t, blocks[1:], res.Rows)
	})

	t.Run("get statistics", func(t *testing.T) {
		stats, err := aggregate.GetStatistics(ctx, db.CH, db.PG)
		assert.Nil(t, err)
		assert.Equal(t, int(master.SeqNo), stats.FirstBlock)
		assert.Equal(t, int(master.SeqNo), stats.LastBlock)
		assert.Equal(t, 1, stats.MasterBlockCount)
		assert.Equal(t, 1, stats.AddressCount)
		assert.Equal(t, 1, stats.ParsedAddressCount)
		assert.Equal(t, 1, stats.AccountCount)
		assert.Equal(t, 1, stats.ParsedAccountCount)
		assert.Equal(t, 1, stats.TransactionCount)
		assert.Equal(t, 2, stats.MessageCount)
		assert.Equal(t, 1, stats.ParsedMessageCount)
		assert.Equal(t, 1, len(stats.AddressStatusCount))
		assert.Equal(t, core.Active, stats.AddressStatusCount[0].Status)
		assert.Equal(t, 1, stats.AddressStatusCount[0].Count)
		assert.Equal(t, state.Types, stats.AddressTypesCount[0].Interfaces)
		assert.Equal(t, 1, stats.AddressTypesCount[0].Count)
		assert.Equal(t, operation, stats.MessageTypesCount[0].Operation)
		assert.Equal(t, 1, stats.MessageTypesCount[0].Count)
	})

	t.Run("drop tables again", func(t *testing.T) {
		dropTables(t)
	})
}
