package account

import (
	"context"

	"github.com/pkg/errors"
	"github.com/uptrace/go-clickhouse/ch"

	"github.com/tonindexer/anton/abi"
	"github.com/tonindexer/anton/addr"
	"github.com/tonindexer/anton/internal/core"
	"github.com/tonindexer/anton/internal/core/aggregate"
)

func (r *Repository) makeLastItemStateQuery(minter *addr.Address) *ch.SelectQuery {
	return r.ch.NewSelect().
		Model((*core.AccountData)(nil)).
		ColumnExpr("argMax(address, last_tx_lt) as item_address").
		Where("minter_address = ?", minter).
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

	// TODO: in rare cases grouping result is duplicated twice, report to go-ch

	err = r.ch.NewSelect().
		Model((*core.AccountData)(nil)).
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

func (r *Repository) AggregateAccounts(ctx context.Context, req *aggregate.AccountsReq) (*aggregate.AccountsRes, error) {
	var (
		res        aggregate.AccountsRes
		interfaces []abi.ContractName
	)

	if req.MinterAddress == nil {
		return nil, errors.Wrap(core.ErrInvalidArg, "minter address must be set")
	}

	err := r.ch.NewSelect().
		Model((*core.AccountData)(nil)).
		ColumnExpr("argMax(types, last_tx_lt) as interfaces").
		Where("address = ?", req.MinterAddress).
		Group("address").
		Scan(ctx, &interfaces)
	if err != nil {
		return nil, err
	}

	for _, t := range interfaces {
		switch t {
		case abi.NFTCollection:
			if err := r.aggregateNFTMinter(ctx, req, &res); err != nil {
				return nil, err
			}

		case abi.JettonMinter:
			if err := r.aggregateFTMinter(ctx, req, &res); err != nil {
				return nil, err
			}
		}
	}

	return &res, nil
}
