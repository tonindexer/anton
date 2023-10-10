package account_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/uptrace/bun/extra/bunbig"

	"github.com/tonindexer/anton/abi/known"
	"github.com/tonindexer/anton/addr"
	"github.com/tonindexer/anton/internal/core"
	"github.com/tonindexer/anton/internal/core/aggregate"
	"github.com/tonindexer/anton/internal/core/rndm"
)

func TestRepository_AggregateAccounts_NFTCollection(t *testing.T) {
	var (
		itemCount        = 15
		collectionStates []*core.AccountState
		itemsStates      []*core.AccountState
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

	t.Run("insert test collection data", func(t *testing.T) {
		tx, err := pg.Begin()
		require.Nil(t, err)

		collectionStates = rndm.AccountStatesContract(100, known.NFTCollection, nil)

		for i := 0; i < itemCount; i++ {
			itemStates := rndm.AccountStatesContract(100/itemCount, known.NFTItem, &collectionStates[0].Address)
			itemsStates = append(itemsStates, itemStates...)
		}

		err = repo.AddAccountStates(ctx, tx, append(itemsStates, collectionStates...))
		require.Nil(t, err)

		err = tx.Commit()
		require.Nil(t, err)
	})

	t.Run("aggregate collections info", func(t *testing.T) {
		res, err := repo.AggregateAccounts(ctx, &aggregate.AccountsReq{
			MinterAddress: &collectionStates[0].Address,
			Limit:         25,
		})
		require.Nil(t, err)
		require.Equal(t, itemCount, res.Items)
		require.Equal(t, itemCount, res.OwnersCount)
		require.Equal(t, itemCount, len(res.OwnedItems))
		for _, c := range res.OwnedItems {
			require.Equal(t, 1, c.ItemsCount)
		}
		require.Equal(t, itemCount, len(res.UniqueOwners))
		for _, c := range res.UniqueOwners {
			require.Equal(t, 100/itemCount, c.OwnersCount)
		}
	})

	t.Run("drop tables again", func(t *testing.T) {
		dropTables(t)
	})
}

func TestRepository_AggregateAccounts_JettonMinter(t *testing.T) {
	var (
		walletsCount = 15

		totalSupply  = new(bunbig.Int)
		ownedBalance = make(map[addr.Address]*bunbig.Int)

		minterStates  []*core.AccountState
		walletsStates []*core.AccountState
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

	t.Run("insert test jetton data", func(t *testing.T) {
		tx, err := pg.Begin()
		require.Nil(t, err)

		minterStates = rndm.AccountStatesContract(100, known.JettonMinter, nil)

		for i := 0; i < walletsCount; i++ {
			walletStates := rndm.AccountStatesContract(100/walletsCount, known.JettonWallet, &minterStates[0].Address)
			walletsStates = append(walletsStates, walletStates...)

			walletLatestData := walletsStates[len(walletsStates)-1]
			totalSupply = totalSupply.Add(walletLatestData.JettonBalance)
			ownedBalance[*walletLatestData.OwnerAddress] = walletLatestData.JettonBalance
		}

		err = repo.AddAccountStates(ctx, tx, append(walletsStates, minterStates...))
		require.Nil(t, err)

		err = tx.Commit()
		require.Nil(t, err)
	})

	t.Run("aggregate jetton data", func(t *testing.T) {
		res, err := repo.AggregateAccounts(ctx, &aggregate.AccountsReq{
			MinterAddress: &minterStates[0].Address,
			Limit:         25,
		})
		require.Nil(t, err)
		require.Equal(t, walletsCount, res.Wallets)
		require.Equal(t, totalSupply, res.TotalSupply)
		require.Equal(t, walletsCount, len(res.OwnedBalance))
		for _, b := range res.OwnedBalance {
			require.Equal(t, ownedBalance[*b.OwnerAddress], b.Balance)
		}
	})

	t.Run("drop tables again", func(t *testing.T) {
		dropTables(t)
	})
}
