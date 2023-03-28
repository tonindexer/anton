package tx

import (
	"context"
	"strings"

	"github.com/uptrace/bun"
	"github.com/uptrace/go-clickhouse/ch"

	"github.com/iam047801/tonidx/internal/core"
	"github.com/iam047801/tonidx/internal/core/filter"
)

func (r *Repository) filterTx(ctx context.Context, f *filter.TransactionsReq) (ret []*core.Transaction, err error) {
	q := r.pg.NewSelect().Model(&ret)

	if f.WithAccountState {
		q = q.Relation("Account", func(q *bun.SelectQuery) *bun.SelectQuery {
			return q.ExcludeColumn("code", "data") // TODO: optional
		})
		if f.WithAccountData {
			q = q.Relation("Account.StateData")
		}
	}
	if f.WithMessages {
		q = q.
			Relation("InMsg").
			Relation("InMsg.Payload").
			Relation("OutMsg").
			Relation("OutMsg.Payload")
	}

	if len(f.Hash) > 0 {
		q = q.Where("transaction.hash = ?", f.Hash)
	}
	if len(f.InMsgHash) > 0 {
		q = q.Where("transaction.in_msg_hash = ?", f.InMsgHash)
	}
	if len(f.Addresses) > 0 {
		q = q.Where("transaction.address in (?)", bun.In(f.Addresses))
	}
	if f.Workchain != nil {
		q = q.Where("transaction.block_workchain = ?", f.Workchain)
	}
	if f.BlockID != nil {
		q = q.Where("transaction.block_workchain = ?", f.BlockID.Workchain).
			Where("transaction.block_shard = ?", f.BlockID.Shard).
			Where("transaction.block_seq_no = ?", f.BlockID.SeqNo)
	}

	if f.AfterTxLT != nil {
		if f.Order == "ASC" {
			q = q.Where("transaction.created_lt > ?", f.AfterTxLT)
		} else {
			q = q.Where("transaction.created_lt < ?", f.AfterTxLT)
		}
	}

	if f.Order != "" {
		q = q.Order("transaction.created_lt " + strings.ToUpper(f.Order))
	}

	if f.Limit == 0 {
		f.Limit = 3
	}
	q = q.Limit(f.Limit)

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
		q = q.Where("block_workchain = ?", *req.Workchain)
	}
	if req.BlockID != nil {
		q = q.Where("block_workchain = ?", req.BlockID.Workchain).
			Where("block_shard = ?", req.BlockID.Shard).
			Where("block_seq_no = ?", req.BlockID.SeqNo)
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
