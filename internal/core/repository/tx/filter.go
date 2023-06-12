package tx

import (
	"context"
	"strings"

	"github.com/uptrace/bun"
	"github.com/uptrace/go-clickhouse/ch"

	"github.com/tonindexer/anton/internal/core"
	"github.com/tonindexer/anton/internal/core/filter"
)

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
		q = q.Where("transaction.created_lt = ?", req.CreatedLT)
	}

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

func (r *Repository) countTx(ctx context.Context, req *filter.TransactionsReq) (int, error) {
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
		q = q.Where("created_lt = ?", req.CreatedLT)
	}

	return q.Count(ctx)
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
	if len(res.Rows) == 0 {
		return res, nil
	}

	res.Total, err = r.countTx(ctx, req)
	if err != nil {
		return res, err
	}

	return res, nil
}
