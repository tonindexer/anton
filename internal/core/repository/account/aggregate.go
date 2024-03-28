package account

import (
	"context"
	"database/sql"

	"github.com/pkg/errors"
	"github.com/uptrace/go-clickhouse/ch"

	"github.com/tonindexer/anton/abi"
	"github.com/tonindexer/anton/abi/known"
	"github.com/tonindexer/anton/addr"
	"github.com/tonindexer/anton/internal/core"
	"github.com/tonindexer/anton/internal/core/aggregate"
)

func (r *Repository) aggregateAddressStatistics(ctx context.Context, req *aggregate.AccountsReq, res *aggregate.AccountsRes) error {
	var err error

	res.TransactionsCount, err = r.ch.NewSelect().
		Model((*core.Transaction)(nil)).
		Where("address = ?", req.Address).
		Count(ctx)
	if err != nil {
		return errors.Wrap(err, "count transactions")
	}

	var countByInterfaces []struct {
		Types []abi.ContractName
		Count int
	}
	err = r.ch.NewSelect().
		Model((*core.AccountState)(nil)).
		Column("types").
		ColumnExpr("uniqExact(address) as count").
		Where("owner_address = ?", req.Address).
		Group("types").
		Scan(ctx, &countByInterfaces)
	if err != nil {
		return errors.Wrap(err, "count owned nft items")
	}

	for _, x := range countByInterfaces {
		for _, t := range x.Types {
			switch t {
			case known.NFTItem:
				res.OwnedNFTItems += x.Count
			case known.NFTCollection:
				res.OwnedNFTCollections += x.Count
			case known.JettonWallet:
				res.OwnedJettonWallets += x.Count
			}
		}
	}

	return nil
}

func (r *Repository) makeLastItemStateQuery(minter *addr.Address) *ch.SelectQuery {
	return r.ch.NewSelect().
		Model((*core.AccountState)(nil)).
		ColumnExpr("argMax(address, last_tx_lt) as item_address").
		Where("minter_address = ?", minter).
		Where("fake = false").
		Group("address")
}

func (r *Repository) makeLastItemOwnerQuery(minter *addr.Address) *ch.SelectQuery {
	return r.makeLastItemStateQuery(minter).
		ColumnExpr("argMax(owner_address, last_tx_lt) AS owner_address")
}

func (r *Repository) aggregateNFTMinter(ctx context.Context, req *aggregate.AccountsReq, res *aggregate.AccountsRes) error {
	var err error

	res.Items, err = r.makeLastItemStateQuery(req.MinterAddress).Count(ctx)
	if err != nil {
		return errors.Wrap(err, "count nft items")
	}

	// TODO: owners include sale contracts

	err = r.ch.NewSelect().
		ColumnExpr("uniqExact(owner_address)").
		TableExpr("(?) as q", r.makeLastItemOwnerQuery(req.MinterAddress)).
		Scan(ctx, &res.OwnersCount)
	if err != nil {
		return errors.Wrap(err, "count owners of nft minter")
	}

	err = r.ch.NewSelect().
		Model((*core.AccountState)(nil)).
		ColumnExpr("address AS item_address").
		ColumnExpr("uniqExact(owner_address) AS owners_count").
		Where("minter_address = ?", req.MinterAddress).
		Group("item_address").
		Order("owners_count DESC").
		Limit(req.Limit).
		Scan(ctx, &res.UniqueOwners)
	if err != nil {
		return errors.Wrap(err, "count unique owners of nft items")
	}

	err = r.ch.NewSelect().
		ColumnExpr("owner_address").
		ColumnExpr("count(item_address) AS items_count").
		TableExpr("(?) as q", r.makeLastItemOwnerQuery(req.MinterAddress)).
		Group("owner_address").
		Order("items_count DESC").
		Limit(req.Limit).
		Scan(ctx, &res.OwnedItems)
	if err != nil {
		return errors.Wrap(err, "count owned nft items")
	}

	return nil
}

func (r *Repository) aggregateFTMinter(ctx context.Context, req *aggregate.AccountsReq, res *aggregate.AccountsRes) error {
	var err error

	res.Wallets, err = r.makeLastItemStateQuery(req.MinterAddress).Count(ctx)
	if err != nil {
		return errors.Wrap(err, "count jetton wallets")
	}

	err = r.ch.NewSelect().
		ColumnExpr("sum(balance) as total_supply").
		TableExpr("(?) as q",
			r.makeLastItemOwnerQuery(req.MinterAddress).
				ColumnExpr("argMax(jetton_balance, last_tx_lt) AS balance")).
		Scan(ctx, &res.TotalSupply)
	if err != nil {
		return errors.Wrap(err, "count jetton total supply")
	}

	err = r.makeLastItemOwnerQuery(req.MinterAddress).
		ColumnExpr("argMax(jetton_balance, last_tx_lt) AS balance").
		Order("balance DESC").
		Limit(req.Limit).
		Scan(ctx, &res.OwnedBalance)
	if err != nil {
		return errors.Wrap(err, "count jetton holders")
	}

	return err
}

func (r *Repository) aggregateMinterStatistics(ctx context.Context, req *aggregate.AccountsReq, res *aggregate.AccountsRes) error {
	var interfaces []abi.ContractName

	err := r.ch.NewSelect().
		Model((*core.AccountState)(nil)).
		ColumnExpr("argMax(types, last_tx_lt) as interfaces").
		Where("address = ?", req.MinterAddress).
		Group("address").
		Scan(ctx, &interfaces)
	if errors.Is(err, sql.ErrNoRows) {
		return nil
	}
	if err != nil {
		return err
	}

	for _, t := range interfaces {
		switch t {
		case known.NFTCollection:
			if err := r.aggregateNFTMinter(ctx, req, res); err != nil {
				return err
			}

		case known.JettonMinter:
			if err := r.aggregateFTMinter(ctx, req, res); err != nil {
				return err
			}
		}
	}

	return nil
}

func (r *Repository) AggregateAccounts(ctx context.Context, req *aggregate.AccountsReq) (*aggregate.AccountsRes, error) {
	var res aggregate.AccountsRes

	if req.Address == nil && req.MinterAddress == nil {
		return nil, errors.Wrap(core.ErrInvalidArg, "address must be set")
	}
	if req.Address != nil {
		if err := r.aggregateAddressStatistics(ctx, req, &res); err != nil {
			return nil, err
		}
	}
	if req.MinterAddress != nil {
		if err := r.aggregateMinterStatistics(ctx, req, &res); err != nil {
			return nil, err
		}
	}

	return &res, nil
}
