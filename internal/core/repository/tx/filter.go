package tx

import (
	"context"
	"strings"

	"github.com/pkg/errors"
	"github.com/uptrace/bun"
	"github.com/uptrace/go-clickhouse/ch"

	"github.com/tonindexer/anton/internal/core"
	"github.com/tonindexer/anton/internal/core/filter"
)

func (r *Repository) getFilterTxQuery(q *bun.SelectQuery, req *filter.TransactionsFilter) *bun.SelectQuery {
	if len(req.Hash) > 0 {
		q = q.Where("transaction.hash = ?", req.Hash)
	}
	if len(req.InMsgHash) > 0 {
		q = q.Where("transaction.in_msg_hash = ?", req.InMsgHash)
	}
	if len(req.Addresses) > 0 {
		q = q.Where("transaction.address in (?)", bun.In(req.Addresses))
	}
	if req.Workchain != nil {
		q = q.Where("transaction.workchain = ?", req.Workchain)
	}
	if req.BlockID != nil {
		q = q.Where("transaction.workchain = ?", req.BlockID.Workchain).
			Where("transaction.shard = ?", req.BlockID.Shard).
			Where("transaction.block_seq_no = ?", req.BlockID.SeqNo)
	}
	if req.CreatedLT != nil {
		q = q.Where("transaction.created_lt = ?", *req.CreatedLT)
	}
	return q
}

func (r *Repository) filterTx(ctx context.Context, req *filter.TransactionsReq) (ret []*core.Transaction, err error) {
	q := r.pg.NewSelect().Model(&ret)

	if req.WithAccountState {
		q = q.Relation("Account", func(q *bun.SelectQuery) *bun.SelectQuery {
			if len(req.ExcludeColumn) > 0 {
				q = q.ExcludeColumn(req.ExcludeColumn...)
			}
			return q
		})
	}
	if req.WithMessages {
		q = q.
			Relation("InMsg").
			Relation("OutMsg")
	}

	q = r.getFilterTxQuery(q, &req.TransactionsFilter)

	if req.AfterTxLT != nil {
		if req.Order == "ASC" {
			q = q.Where("transaction.created_lt > ?", req.AfterTxLT)
		} else {
			q = q.Where("transaction.created_lt < ?", req.AfterTxLT)
		}
	}

	if req.Order != "" {
		q = q.Order("transaction.created_lt " + strings.ToUpper(req.Order))
	}

	if req.Limit == 0 {
		req.Limit = 3
	}
	q = q.Limit(req.Limit)

	err = q.Scan(ctx)
	return ret, err
}

func (r *Repository) countTxFullScan(ctx context.Context, req *filter.TransactionsReq) (count int, maxLt uint64, err error) {
	var result struct {
		Count        int
		MaxLT        uint64 `ch:"max_lt_value"`
		RoundedMaxLT uint64 `ch:"max_lt_rounded"`
	}

	q := r.ch.NewSelect().
		Model((*core.Transaction)(nil))

	if len(req.Hash) > 0 {
		q = q.Where("hash = ?", req.Hash)
	}
	if len(req.InMsgHash) > 0 {
		q = q.Where("in_msg_hash = ?", req.InMsgHash)
	}
	if len(req.Addresses) > 0 {
		q = q.Where("address in (?)", ch.In(req.Addresses))
	}
	if req.Workchain != nil {
		q = q.Where("workchain = ?", *req.Workchain)
	}
	if req.BlockID != nil {
		q = q.Where("workchain = ?", req.BlockID.Workchain).
			Where("shard = ?", req.BlockID.Shard).
			Where("block_seq_no = ?", req.BlockID.SeqNo)
	}
	if req.CreatedLT != nil {
		q = q.Where("created_lt = ?", *req.CreatedLT)
	}

	q = r.ch.NewSelect().
		With(
			"max_lt",
			r.ch.NewSelect().
				Model((*core.Transaction)(nil)).
				ColumnExpr("max(created_lt) AS v"),
		).
		With(
			"rounded_count",
			q. // query with filters
				Table("max_lt").
				ColumnExpr("count(*) as v").
				Where("created_lt <= floor(max_lt.v, -7) - 1e7"),
		).
		Table("max_lt", "rounded_count").
		ColumnExpr("max_lt.v AS max_lt_value").
		ColumnExpr("max2(floor(max_lt.v, -7) - 1e7, 0) as max_lt_rounded").
		ColumnExpr("rounded_count.v AS count")

	if err := q.Scan(ctx, &result); err != nil {
		return 0, 0, err
	}

	if result.MaxLT == 0 {
		return 0, 0, core.ErrNotFound
	}

	return result.Count, result.RoundedMaxLT, nil
}

func (r *Repository) countTxPartialScan(ctx context.Context, req *filter.TransactionsReq, startLt uint64) (partialCount, roundedCount int, roundedMaxLt uint64, err error) {
	var result struct {
		SinceStartCount int    `bun:"since_start_count"`
		RoundedCount    int    `bun:"until_rounded_count"`
		RoundedMaxLT    uint64 `bun:"rounded_max_lt_value"`
	}

	q := r.pg.NewSelect().
		With(
			"rounded_max_lt",
			r.pg.NewSelect().
				Model((*core.Transaction)(nil)).
				ColumnExpr("greatest(floor(max(created_lt) / 1e7) * 1e7 - 1e7, 0) AS v"), // we round LT as transactions in new blocks can have lower LT
		).
		With(
			"until_rounded_count",
			r.getFilterTxQuery(
				r.pg.NewSelect().Model((*core.Transaction)(nil)),
				&req.TransactionsFilter,
			).
				Table("rounded_max_lt").
				ColumnExpr("count(*) as v").
				Where("created_lt > ?", startLt).
				Where("created_lt <= rounded_max_lt.v"),
		).
		With("since_start_count",
			r.getFilterTxQuery(
				r.pg.NewSelect().Model((*core.Transaction)(nil)),
				&req.TransactionsFilter,
			).
				ColumnExpr("count(*) as v").
				Where("created_lt > ?", startLt)).
		Table("rounded_max_lt", "until_rounded_count", "since_start_count").
		ColumnExpr("since_start_count.v AS since_start_count").
		ColumnExpr("until_rounded_count.v AS until_rounded_count").
		ColumnExpr("rounded_max_lt.v as rounded_max_lt_value")

	if err := q.Scan(ctx, &result); err != nil {
		return 0, 0, 0, err
	}

	return result.SinceStartCount, result.RoundedCount, result.RoundedMaxLT, nil
}

func (r *Repository) countTx(ctx context.Context, req *filter.TransactionsReq) (int, error) {
	if len(req.Hash) > 0 || req.BlockID != nil || req.CreatedLT != nil { // count value cannot change on any of these filters
		count, _, _, err := r.countTxPartialScan(ctx, req, 0)
		return count, err
	}

	count, maxLT, err := r.transactionsFilterCountCache.Get(req.TransactionsFilter)
	if errors.Is(err, core.ErrNotFound) {
		count, maxLT, err = r.countTxFullScan(ctx, req)
		if err != nil {
			return 0, err
		}
		if errors.Is(err, core.ErrNotFound) {
			return 0, nil
		}
		if err := r.transactionsFilterCountCache.Set(req.TransactionsFilter, count, maxLT); err != nil {
			return 0, err
		}
	} else if err != nil {
		return 0, err
	}

	partialCount, roundedPartialCount, roundedMaxLT, err := r.countTxPartialScan(ctx, req, maxLT)
	if err != nil {
		return 0, err
	}
	if err := r.transactionsFilterCountCache.Set(req.TransactionsFilter, count+roundedPartialCount, roundedMaxLT); err != nil {
		return 0, err
	}

	return count + partialCount, nil
}

func (r *Repository) FilterTransactions(ctx context.Context, req *filter.TransactionsReq) (*filter.TransactionsRes, error) {
	var (
		res = new(filter.TransactionsRes)
		err error
	)

	res.Rows, err = r.filterTx(ctx, req)
	if err != nil {
		return res, err
	}

	switch {
	case len(res.Rows) == 0:

	case len(req.Hash) > 0:
		res.Total = len(res.Rows)

	case req.Count:
		res.Total, err = r.countTx(ctx, req)
		if err != nil {
			return res, err
		}
	}

	return res, nil
}
