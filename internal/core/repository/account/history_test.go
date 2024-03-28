package account_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

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
		require.Nil(t, err)

		for i := 0; i < 10; i++ {
			states := rndm.AccountStates(10)

			err = repo.AddAccountStates(ctx, tx, states)
			require.Nil(t, err)
		}

		for i := 0; i < 10; i++ {
			states := rndm.AccountStatesContract(10, "special", nil)

			err = repo.AddAccountStates(ctx, tx, states)
			require.Nil(t, err)
		}

		err = tx.Commit()
		require.Nil(t, err)
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
		require.Nil(t, err)
		require.Equal(t, 1, len(res.CountRes))
		for _, c := range res.CountRes {
			require.Equal(t, 10, c.Value)
		}
	})

	t.Run("drop tables again", func(t *testing.T) {
		dropTables(t)
	})
}
