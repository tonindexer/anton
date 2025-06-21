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
		Count int
		MaxLT *uint64 `ch:"max_lt"`
	}

	q := r.ch.NewSelect().
		Model((*core.Transaction)(nil)).
		ColumnExpr("count(*) AS count").
		ColumnExpr("(SELECT max(created_lt) FROM transactions) AS max_lt") // unfiltered max

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

	if err := q.Scan(ctx, &result); err != nil {
		return 0, 0, err
	}

	if result.MaxLT == nil {
		return 0, 0, core.ErrNotFound
	}

	return result.Count, *result.MaxLT, nil
}

func (r *Repository) countTxPartialScan(ctx context.Context, req *filter.TransactionsReq, startLt uint64) (partialCount int, maxLt uint64, err error) {
	var result struct {
		Count int
		MaxLT uint64 `bun:"max_lt"`
	}

	q := r.pg.NewSelect().
		Model((*core.Transaction)(nil)).
		ColumnExpr("count(*) AS count").
		ColumnExpr("COALESCE(max(created_lt), ?) AS max_lt", startLt).
		Where("created_lt > ?", startLt)

	q = r.getFilterTxQuery(q, &req.TransactionsFilter)

	if err := q.Scan(ctx, &result); err != nil {
		return 0, 0, err
	}

	if result.MaxLT == 0 {
		result.MaxLT = startLt // no new rows
	}

	return result.Count, result.MaxLT, nil
}

func (r *Repository) countTx(ctx context.Context, req *filter.TransactionsReq) (int, error) {
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

	if len(req.Hash) > 0 || req.BlockID != nil || req.CreatedLT != nil {
		return count, nil // count value cannot change on any of these filters
	}

	partialCount, maxLT, err := r.countTxPartialScan(ctx, req, maxLT)
	if err != nil {
		return 0, err
	}
	if err := r.transactionsFilterCountCache.Set(req.TransactionsFilter, count+partialCount, maxLT); err != nil {
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
