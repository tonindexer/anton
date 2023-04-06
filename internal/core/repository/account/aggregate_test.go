package account

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/uptrace/bun/extra/bunbig"

	"github.com/tonindexer/anton/abi"
	"github.com/tonindexer/anton/internal/addr"
	"github.com/tonindexer/anton/internal/core"
	"github.com/tonindexer/anton/internal/core/aggregate"
)

func TestAggregateRepository_NFTCollection(t *testing.T) {
	var (
		itemCount = 15

		collectionStates []*core.AccountState
		collectionData   []*core.AccountData

		itemsStates []*core.AccountState
		itemsData   []*core.AccountData
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
		assert.Nil(t, err)

		collectionStates = randAccountStates(100)
		collectionData = randContractsData(collectionStates, abi.NFTCollection, nil)

		for i := 0; i < itemCount; i++ {
			itemStates := randAccountStates(100 / itemCount)
			itemsStates = append(itemsStates, itemStates...)
			itemsData = append(itemsData, randContractsData(itemStates, abi.NFTItem, &collectionStates[0].Address)...)
		}

		err = repo.AddAccountData(ctx, tx, append(itemsData, collectionData...))
		assert.Nil(t, err)
		err = repo.AddAccountStates(ctx, tx, append(itemsStates, collectionStates...))
		assert.Nil(t, err)

		err = tx.Commit()
		assert.Nil(t, err)
	})

	t.Run("aggregate collections info", func(t *testing.T) {
		res, err := repo.AggregateAccounts(ctx, &aggregate.AccountsReq{
			MinterAddress: &collectionStates[0].Address,
			Limit:         25,
		})
		assert.Nil(t, err)
		assert.Equal(t, itemCount, res.Items)
		assert.Equal(t, itemCount, res.OwnersCount)
		assert.Equal(t, itemCount, len(res.OwnedItems))
		for _, c := range res.OwnedItems {
			assert.Equal(t, 1, c.ItemsCount)
		}
		assert.Equal(t, itemCount, len(res.UniqueOwners))
		for _, c := range res.UniqueOwners {
			assert.Equal(t, 100/itemCount, c.OwnersCount)
		}
	})
}

func TestAggregateRepository_JettonMinter(t *testing.T) {
	var (
		walletsCount = 15

		totalSupply  = new(bunbig.Int)
		ownedBalance = make(map[addr.Address]*bunbig.Int)

		minterStates []*core.AccountState
		minterData   []*core.AccountData

		walletsStates []*core.AccountState
		walletsData   []*core.AccountData
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
		assert.Nil(t, err)

		minterStates = randAccountStates(100)
		minterData = randContractsData(minterStates, abi.JettonMinter, nil)

		for i := 0; i < walletsCount; i++ {
			walletStates := randAccountStates(100 / walletsCount)

			walletsStates = append(walletsStates, walletStates...)
			walletsData = append(walletsData, randContractsData(walletStates, abi.JettonWallet, &minterStates[0].Address)...)

			walletLatestData := walletsData[len(walletsData)-1]
			totalSupply = totalSupply.Add(walletLatestData.JettonBalance)
			ownedBalance[*walletLatestData.OwnerAddress] = walletLatestData.JettonBalance
		}

		err = repo.AddAccountData(ctx, tx, append(walletsData, minterData...))
		assert.Nil(t, err)
		err = repo.AddAccountStates(ctx, tx, append(walletsStates, minterStates...))
		assert.Nil(t, err)

		err = tx.Commit()
		assert.Nil(t, err)
	})

	t.Run("aggregate jetton data", func(t *testing.T) {
		res, err := repo.AggregateAccounts(ctx, &aggregate.AccountsReq{
			MinterAddress: &minterStates[0].Address,
			Limit:         25,
		})
		assert.Nil(t, err)
		assert.Equal(t, walletsCount, res.Wallets)
		assert.Equal(t, totalSupply, res.TotalSupply)
		assert.Equal(t, walletsCount, len(res.OwnedBalance))
		for _, b := range res.OwnedBalance {
			assert.Equal(t, ownedBalance[*b.OwnerAddress], b.Balance)
		}
	})
}
