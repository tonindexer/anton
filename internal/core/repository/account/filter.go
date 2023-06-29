package account

import (
	"context"
	"strings"

	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/pgdialect"
	"github.com/uptrace/go-clickhouse/ch"

	"github.com/tonindexer/anton/internal/core"
	"github.com/tonindexer/anton/internal/core/filter"
)

func (r *Repository) filterAddressLabels(ctx context.Context, f *filter.LabelsReq) (ret []*core.AddressLabel, err error) {
	q := r.pg.NewSelect().Model(&ret)

	if f.Name != "" {
		q = q.Where("name ILIKE ?", "%"+f.Name+"%")
	}
	if len(f.Categories) > 0 {
		q = q.Where("categories && ?", pgdialect.Array(f.Categories))
	}

	q = q.Order("name ASC")

	q = q.Offset(f.Offset)

	if f.Limit == 0 {
		f.Limit = 3
	}
	q = q.Limit(f.Limit)

	err = q.Scan(ctx)

	return ret, err
}

func (r *Repository) countAddressLabels(ctx context.Context, f *filter.LabelsReq) (int, error) {
	q := r.ch.NewSelect().Model((*core.AddressLabel)(nil))

	if f.Name != "" {
		q = q.Where("positionCaseInsensitive(name, ?) > 0", f.Name)
	}
	if len(f.Categories) > 0 {
		q = q.Where("hasAny(categories, [?])", ch.In(f.Categories))
	}

	return q.Count(ctx)
}

func (r *Repository) FilterLabels(ctx context.Context, f *filter.LabelsReq) (*filter.LabelsRes, error) {
	var (
		res = new(filter.LabelsRes)
		err error
	)

	res.Rows, err = r.filterAddressLabels(ctx, f)
	if err != nil {
		return res, err
	}
	if len(res.Rows) == 0 {
		return res, nil
	}

	res.Total, err = r.countAddressLabels(ctx, f)
	if err != nil {
		return res, err
	}

	return res, nil
}

func (r *Repository) filterAccountStates(ctx context.Context, f *filter.AccountsReq) (ret []*core.AccountState, err error) {
	var (
		q                   *bun.SelectQuery
		prefix, statesTable string
		latest              []*core.LatestAccountState
	)

	// choose table to filter states by
	// and optionally join account data
	if f.LatestState {
		q = r.pg.NewSelect().Model(&latest).
			Relation("AccountState", func(q *bun.SelectQuery) *bun.SelectQuery {
				return q.ExcludeColumn(f.ExcludeColumn...)
			}).
			Relation("AccountState.Label")
		statesTable = "latest_account_state."
		prefix = "account_state."
	} else {
		q = r.pg.NewSelect().Model(&ret).
			ExcludeColumn(f.ExcludeColumn...).
			Relation("Label")
		statesTable = "account_state."
	}

	if len(f.Addresses) > 0 {
		q = q.Where(statesTable+"address in (?)", bun.In(f.Addresses))
	}
	if len(f.ContractTypes) > 0 {
		q = q.Where(prefix+"types && ?", pgdialect.Array(f.ContractTypes))
	}
	if f.OwnerAddress != nil {
		q = q.Where(prefix+"owner_address = ?", f.OwnerAddress)
	}
	if f.MinterAddress != nil {
		q = q.Where(prefix+"minter_address = ?", f.MinterAddress)
	}

	if f.AfterTxLT != nil {
		if f.Order == "ASC" {
			q = q.Where(statesTable+"last_tx_lt > ?", f.AfterTxLT)
		} else {
			q = q.Where(statesTable+"last_tx_lt < ?", f.AfterTxLT)
		}
	}
	if f.Order != "" {
		q = q.Order(statesTable + "last_tx_lt " + strings.ToUpper(f.Order))
	}

	if f.Limit == 0 {
		f.Limit = 3
	}
	q = q.Limit(f.Limit)

	err = q.Scan(ctx)

	if f.LatestState {
		for _, a := range latest {
			ret = append(ret, a.AccountState)
		}
	}

	return ret, err
}

func (r *Repository) countAccountStates(ctx context.Context, f *filter.AccountsReq) (int, error) {
	q := r.ch.NewSelect().Model((*core.AccountState)(nil))

	if len(f.Addresses) > 0 {
		q = q.Where("address in (?)", ch.In(f.Addresses))
	}
	if len(f.ContractTypes) > 0 {
		q = q.Where("hasAny(types, [?])", ch.In(f.ContractTypes))
	}
	if f.MinterAddress != nil {
		q = q.Where("minter_address = ?", f.MinterAddress)
	}

	if f.LatestState {
		q = q.ColumnExpr("argMax(address, last_tx_lt)")
		if f.OwnerAddress != nil {
			q = q.ColumnExpr("argMax(owner_address, last_tx_lt) as owner_address")
		}
		q = q.Group("address")
	} else {
		q = q.Column("address")
		if f.OwnerAddress != nil {
			q = q.Column("owner_address")
		}
	}

	qCount := r.ch.NewSelect().TableExpr("(?) as q", q)
	if f.OwnerAddress != nil { // that's because owner address can change
		qCount = qCount.Where("owner_address = ?", f.OwnerAddress)
	}
	return qCount.Count(ctx)
}

func (r *Repository) FilterAccounts(ctx context.Context, f *filter.AccountsReq) (*filter.AccountsRes, error) {
	var (
		res = new(filter.AccountsRes)
		err error
	)

	res.Rows, err = r.filterAccountStates(ctx, f)
	if err != nil {
		return res, err
	}
	if len(res.Rows) == 0 {
		return res, nil
	}

	res.Total, err = r.countAccountStates(ctx, f)
	if err != nil {
		return res, err
	}

	return res, nil
}
