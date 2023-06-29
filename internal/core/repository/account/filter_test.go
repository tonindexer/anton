package account_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/tonindexer/anton/abi"
	"github.com/tonindexer/anton/addr"
	"github.com/tonindexer/anton/internal/core"
	"github.com/tonindexer/anton/internal/core/filter"
	"github.com/tonindexer/anton/internal/core/rndm"
)

func TestRepository_FilterLabels(t *testing.T) {
	initdb(t)

	dead := &core.AddressLabel{
		Address:    *rndm.Address(),
		Name:       "dead",
		Categories: []core.LabelCategory{core.CentralizedExchange},
	}
	beef := &core.AddressLabel{
		Address:    *rndm.Address(),
		Name:       "beef",
		Categories: []core.LabelCategory{core.CentralizedExchange},
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	t.Run("drop tables", func(t *testing.T) {
		dropTables(t)
	})

	t.Run("create tables", func(t *testing.T) {
		createTables(t)
	})

	t.Run("insert test data", func(t *testing.T) {
		err := repo.AddAddressLabel(ctx, dead)
		require.Nil(t, err)

		err = repo.AddAddressLabel(ctx, beef)
		require.Nil(t, err)
	})

	t.Run("filter by name", func(t *testing.T) {
		res, err := repo.FilterLabels(ctx, &filter.LabelsReq{Name: "be"})
		require.Nil(t, err)
		require.Equal(t, 1, res.Total)
		require.Equal(t, 1, len(res.Rows))
		require.Equal(t, beef, res.Rows[0])

		res, err = repo.FilterLabels(ctx, &filter.LabelsReq{Name: "E", Categories: []core.LabelCategory{core.CentralizedExchange}})
		require.Nil(t, err)
		require.Equal(t, 2, res.Total)
		require.Equal(t, 2, len(res.Rows))
		require.Equal(t, []*core.AddressLabel{beef, dead}, res.Rows)

		res, err = repo.FilterLabels(ctx, &filter.LabelsReq{Name: "feeb"})
		require.Nil(t, err)
		require.Equal(t, 0, res.Total)
		require.Equal(t, 0, len(res.Rows))
	})

	t.Run("filter by categories", func(t *testing.T) {
		res, err := repo.FilterLabels(ctx, &filter.LabelsReq{Categories: []core.LabelCategory{core.CentralizedExchange}})
		require.Nil(t, err)
		require.Equal(t, 2, res.Total)
		require.Equal(t, 2, len(res.Rows))
		require.Equal(t, []*core.AddressLabel{beef, dead}, res.Rows)

		res, err = repo.FilterLabels(ctx, &filter.LabelsReq{Categories: []core.LabelCategory{core.Scam}})
		require.Nil(t, err)
		require.Equal(t, 0, res.Total)
		require.Equal(t, 0, len(res.Rows))
	})

	t.Run("drop tables again", func(t *testing.T) {
		dropTables(t)
	})
}

func TestRepository_FilterAccounts(t *testing.T) {
	var (
		// filter by address
		address       *addr.Address
		addressStates []*core.AccountState

		// filter by latest
		latestState *core.AccountState

		// filter by contract type
		specialState *core.AccountState
	)

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

		for i := 0; i < 10; i++ { // 10 account states on 100 addresses
			var states []*core.AccountState

			for j := 0; j < 10; j++ {
				states = append(states, rndm.AccountStates(10)...)
			}

			// filter by address
			address = &states[len(states)-10].Address
			addressStates = states[len(states)-10:]

			err = repo.AddAccountStates(ctx, tx, states)
			assert.Nil(t, err)
		}

		err = tx.Commit()
		assert.Nil(t, err)
	})

	t.Run("insert states with special contract type", func(t *testing.T) {
		tx, err := pg.Begin()
		assert.Nil(t, err)

		// filter by contract interfaces
		for i := 0; i < 15; i++ { // add 15 addresses with 10 states
			states := rndm.AccountStatesContract(10, "special", nil)

			specialState = states[len(states)-1]

			err = repo.AddAccountStates(ctx, tx, states)
			assert.Nil(t, err)
		}

		err = tx.Commit()
		assert.Nil(t, err)
	})

	t.Run("insert many states on some address", func(t *testing.T) {
		tx, err := pg.Begin()
		assert.Nil(t, err)

		for i := 0; i < 5; i++ {
			latestState = rndm.AddressStateContract(address, "", nil)
			err = repo.AddAccountStates(ctx, tx, []*core.AccountState{latestState})
			assert.Nil(t, err)
		}

		err = tx.Commit()
		assert.Nil(t, err)
	})

	t.Run("filter states by address", func(t *testing.T) {
		results, err := repo.FilterAccounts(ctx, &filter.AccountsReq{
			Addresses: []*addr.Address{address},
			Order:     "ASC", Limit: len(addressStates),
		})
		assert.Nil(t, err)
		assert.Equal(t, 15, results.Total)
		assert.Equal(t, addressStates, results.Rows)
	})

	t.Run("filter latest state by address and exclude columns", func(t *testing.T) {
		latest := *latestState
		latest.Code = nil

		results, err := repo.FilterAccounts(ctx, &filter.AccountsReq{
			Addresses:     []*addr.Address{&latest.Address},
			LatestState:   true,
			ExcludeColumn: []string{"code"},
		})
		assert.Nil(t, err)
		assert.Equal(t, 1, results.Total)
		assert.Equal(t, []*core.AccountState{&latest}, results.Rows)
	})

	t.Run("filter latest state with data by address and exclude columns", func(t *testing.T) {
		latest := *latestState
		latest.Code = nil

		results, err := repo.FilterAccounts(ctx, &filter.AccountsReq{
			Addresses:     []*addr.Address{&latest.Address},
			LatestState:   true,
			ExcludeColumn: []string{"code"},
		})
		assert.Nil(t, err)
		assert.Equal(t, 1, results.Total)
		assert.Equal(t, []*core.AccountState{&latest}, results.Rows)
	})

	t.Run("filter latest state with data by contract types", func(t *testing.T) {
		results, err := repo.FilterAccounts(ctx, &filter.AccountsReq{
			ContractTypes: []abi.ContractName{"special"},
			LatestState:   true,
			Order:         "DESC", Limit: 1,
		})
		assert.Nil(t, err)
		assert.Equal(t, 15, results.Total)
		assert.Equal(t, []*core.AccountState{specialState}, results.Rows)
	})

	t.Run("filter states by minter", func(t *testing.T) {
		results, err := repo.FilterAccounts(ctx, &filter.AccountsReq{
			MinterAddress: latestState.MinterAddress,
			Order:         "DESC", Limit: 1,
		})
		assert.Nil(t, err)
		assert.Equal(t, 5, results.Total)
		assert.Equal(t, []*core.AccountState{latestState}, results.Rows)
	})

	t.Run("filter states by owner", func(t *testing.T) {
		results, err := repo.FilterAccounts(ctx, &filter.AccountsReq{
			OwnerAddress: latestState.OwnerAddress,
			Order:        "DESC", Limit: 1,
		})
		assert.Nil(t, err)
		assert.Equal(t, 1, results.Total)
		assert.Equal(t, []*core.AccountState{latestState}, results.Rows)
	})

	t.Run("filter latest states by owner", func(t *testing.T) {
		results, err := repo.FilterAccounts(ctx, &filter.AccountsReq{
			LatestState:  true,
			OwnerAddress: latestState.OwnerAddress,
			Order:        "DESC", Limit: 1,
		})
		assert.Nil(t, err)
		assert.Equal(t, 1, results.Total)
		assert.Equal(t, []*core.AccountState{latestState}, results.Rows)
	})

	t.Run("drop tables again", func(t *testing.T) {
		dropTables(t)
	})
}

func TestRepository_FilterAccounts_Heavy(t *testing.T) {
	const (
		totalStates   = 1000000
		specialStates = 100000
	)

	var (
		address      *addr.Address
		specialState *core.AccountState
	)

	// uncomment it
	t.Skip("skipping heavy tests")

	initdb(t)

	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Minute)
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

		for i := 0; i < totalStates/100; i++ {
			states := rndm.AccountStates(100)

			err = repo.AddAccountStates(ctx, tx, states)
			assert.Nil(t, err)

			if i%100 == 0 {
				t.Logf("%s: add %d states on %s address", time.Now().UTC(), 100*(i+1), states[0].Address.String())
			}
		}

		err = tx.Commit()
		assert.Nil(t, err)
	})

	t.Run("insert states with special contract type", func(t *testing.T) {
		tx, err := pg.Begin()
		assert.Nil(t, err)

		address = rndm.Address()

		for i := 0; i < specialStates/100; i++ {
			specialState = rndm.AddressStateContract(address, "special", nil)

			err = repo.AddAccountStates(ctx, tx, []*core.AccountState{specialState})
			assert.Nil(t, err)

			if i%100 == 0 {
				t.Logf("%s: add %d special states", time.Now().UTC(), 100*(i+1))
			}
		}

		err = tx.Commit()
		assert.Nil(t, err)
	})

	t.Run("filter latest state with data by contract types", func(t *testing.T) {
		start := time.Now()

		results, err := repo.FilterAccounts(ctx, &filter.AccountsReq{
			ContractTypes: []abi.ContractName{"special"},
			LatestState:   true,
			Order:         "DESC", Limit: 1,
		})
		assert.Nil(t, err)
		assert.Equal(t, 1, results.Total)
		assert.Equal(t, []*core.AccountState{specialState}, results.Rows)
		assert.Less(t, time.Since(start), 2*time.Second)
	})

	t.Run("drop tables again", func(t *testing.T) {
		dropTables(t)
	})
}
