package account

import (
	"context"
	"database/sql"

	"github.com/pkg/errors"

	"github.com/tonindexer/anton/abi"
	"github.com/tonindexer/anton/abi/known"
	"github.com/tonindexer/anton/internal/core"
	"github.com/tonindexer/anton/internal/core/aggregate"
)

func (r *Repository) aggregateAddressStatistics(ctx context.Context, req *aggregate.AccountsReq, res *aggregate.AccountsRes) error {
	var err error

	res.TransactionsCount, err = r.pg.NewSelect().
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
	err = r.pg.NewSelect().
		Model((*core.LatestAccountState)(nil)).
		Column("types").
		ColumnExpr("count(*) as count").
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

func (r *Repository) aggregateNFTMinter(ctx context.Context, req *aggregate.AccountsReq, res *aggregate.AccountsRes) error {
	var err error

	res.Items, err = r.pg.NewSelect().
		Model((*core.LatestAccountState)(nil)).
		Where("minter_address = ?", req.MinterAddress).
		Where("fake = false").
		Count(ctx)
	if err != nil {
		return errors.Wrap(err, "count nft items")
	}

	// TODO: owners include sale contracts

	err = r.pg.NewSelect().
		Model((*core.LatestAccountState)(nil)).
		ColumnExpr("count(owner_address)").
		Where("minter_address = ?", req.MinterAddress).
		Where("fake = false").
		Scan(ctx, &res.OwnersCount)
	if err != nil {
		return errors.Wrap(err, "count owners of nft minter")
	}

	err = r.ch.NewSelect().
		Model((*core.AccountState)(nil)).
		ColumnExpr("address AS item_address").
		ColumnExpr("uniqExact(owner_address) AS owners_count").
		Where("minter_address = ?", req.MinterAddress).
		Where("fake = false").
		Group("item_address").
		Order("owners_count DESC").
		Limit(req.Limit).
		Scan(ctx, &res.UniqueOwners)
	if err != nil {
		return errors.Wrap(err, "count unique owners of nft items")
	}

	err = r.pg.NewSelect().
		Model((*core.LatestAccountState)(nil)).
		ColumnExpr("owner_address").
		ColumnExpr("count(address) as items_count").
		Where("minter_address = ?", req.MinterAddress).
		Where("fake = false").
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

	res.Wallets, err = r.pg.NewSelect().
		Model((*core.LatestAccountState)(nil)).
		Where("latest_account_state.minter_address = ?", req.MinterAddress).
		Where("latest_account_state.fake = false").
		Count(ctx)
	if err != nil {
		return errors.Wrap(err, "count jetton wallets")
	}

	err = r.pg.NewSelect().
		ColumnExpr("sum(jetton_balance)").
		TableExpr("(?) as q",
			r.pg.NewSelect().
				Model((*core.LatestAccountState)(nil)).
				Relation("AccountState").
				ColumnExpr("account_state.jetton_balance").
				Where("latest_account_state.minter_address = ?", req.MinterAddress).
				Where("latest_account_state.fake = false"),
		).
		Scan(ctx, &res.TotalSupply)
	if err != nil {
		return errors.Wrap(err, "count jetton total supply")
	}

	err = r.pg.NewSelect().
		ColumnExpr("wallet_address").
		ColumnExpr("owner_address").
		ColumnExpr("balance").
		TableExpr("(?) as q",
			r.pg.NewSelect().
				Model((*core.LatestAccountState)(nil)).
				Relation("AccountState").
				ColumnExpr("latest_account_state.address as wallet_address").
				ColumnExpr("latest_account_state.owner_address as owner_address").
				ColumnExpr("account_state.jetton_balance as balance").
				Where("latest_account_state.minter_address = ?", req.MinterAddress).
				Where("latest_account_state.fake = false").
				Order("balance DESC").
				Limit(req.Limit),
		).
		Scan(ctx, &res.OwnedBalance)
	if err != nil {
		return errors.Wrap(err, "count jetton holders")
	}

	return err
}

func (r *Repository) aggregateMinterStatistics(ctx context.Context, req *aggregate.AccountsReq, res *aggregate.AccountsRes) error {
	var interfacesRes struct {
		Interfaces []abi.ContractName `bun:"type:text[],array"`
	}

	err := r.pg.NewSelect().
		Model((*core.LatestAccountState)(nil)).
		ColumnExpr("types as interfaces").
		Where("address = ?", req.MinterAddress).
		Group("address").
		Scan(ctx, &interfacesRes)
	if errors.Is(err, sql.ErrNoRows) {
		return nil
	}
	if err != nil {
		return err
	}

	for _, t := range interfacesRes.Interfaces {
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
