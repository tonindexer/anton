package account_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/tonindexer/anton/abi"
	"github.com/tonindexer/anton/internal/core/aggregate/history"
	"github.com/tonindexer/anton/internal/core/rndm"
)

func TestRepository_AggregateAccountsHistory(t *testing.T) {
	initdb(t)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	t.Run("drop tables", func(t *testing.T) {
		dropTables(t)
	})

	t.Run("create tables", func(t *testing.T) {
		createTables(t)
	})

	t.Run("insert test data", func(t *testing.T) {
		tx, err := pg.Begin()
		assert.Nil(t, err)

		for i := 0; i < 10; i++ {
			states := rndm.AccountStates(10)
			data := rndm.AccountData(states)

			err = repo.AddAccountData(ctx, tx, data)
			assert.Nil(t, err)
			err = repo.AddAccountStates(ctx, tx, states)
			assert.Nil(t, err)
		}

		for i := 0; i < 10; i++ {
			states := rndm.AccountStates(10)
			data := rndm.ContractsData(states, "special", nil)

			err = repo.AddAccountData(ctx, tx, data)
			assert.Nil(t, err)
			err = repo.AddAccountStates(ctx, tx, states)
			assert.Nil(t, err)
		}

		err = tx.Commit()
		assert.Nil(t, err)
	})

	t.Run("count active addresses", func(t *testing.T) {
		res, err := repo.AggregateAccountsHistory(ctx, &history.AccountsReq{
			Metric:        history.ActiveAddresses,
			ContractTypes: []abi.ContractName{"special"},
			ReqParams: history.ReqParams{
				From:     time.Now().UTC(),
				Interval: 24 * time.Hour,
			},
		})
		assert.Nil(t, err)
		assert.Equal(t, 1, len(res.CountRes))
		for _, c := range res.CountRes {
			assert.Equal(t, 10, c.Value)
		}
	})
}
