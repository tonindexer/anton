package block

import (
	"context"
	"strings"

	"github.com/uptrace/bun"

	"github.com/tonindexer/anton/internal/core"
	"github.com/tonindexer/anton/internal/core/filter"
)

func loadTransactions(q *bun.SelectQuery, prefix string, f *filter.BlocksReq) *bun.SelectQuery {
	q = q.Relation(prefix + "Transactions")
	if f.WithTransactionAccountState {
		q = q.Relation(prefix+"Transactions.Account", func(q *bun.SelectQuery) *bun.SelectQuery {
			if len(f.ExcludeColumn) > 0 {
				return q.ExcludeColumn(f.ExcludeColumn...) // TODO: optional
			}
			return q
		})
		if f.WithTransactionAccountData {
			q = q.Relation(prefix + "Transactions.Account.StateData")
		}
	}
	if f.WithTransactionMessages {
		q = q.
			Relation(prefix + "Transactions.InMsg").
			Relation(prefix + "Transactions.OutMsg")
		if f.WithTransactionMessagePayloads {
			q = q.
				Relation(prefix + "Transactions.InMsg.Payload").
				Relation(prefix + "Transactions.OutMsg.Payload")
		}
	}
	return q
}

func (r *Repository) filterBlocks(ctx context.Context, f *filter.BlocksReq) (ret []*core.Block, err error) {
	q := r.pg.NewSelect().Model(&ret)

	if f.WithShards {
		q = q.Relation("Shards")
		if f.WithTransactions {
			q = loadTransactions(q, "Shards.", f)
		}
	}
	if f.WithTransactions {
		q = loadTransactions(q, "", f)
	}

	if f.Workchain != nil {
		q = q.Where("workchain = ?", *f.Workchain)
	}
	if f.Shard != nil {
		q = q.Where("shard = ?", *f.Shard)
	}
	if f.SeqNo != nil {
		q = q.Where("seq_no = ?", *f.SeqNo)
	}

	if len(f.FileHash) > 0 {
		q = q.Where("file_hash = ?", f.FileHash)
	}

	if f.AfterSeqNo != nil {
		if f.Order == "ASC" {
			q = q.Where("seq_no > ?", f.AfterSeqNo)
		} else {
			q = q.Where("seq_no < ?", f.AfterSeqNo)
		}
	}

	if f.Order != "" {
		q = q.Order("seq_no " + strings.ToUpper(f.Order))
	}

	if f.Limit == 0 {
		f.Limit = 3
	}
	q = q.Limit(f.Limit)

	err = q.Scan(ctx)
	return ret, err
}

func (r *Repository) countBlocks(ctx context.Context, f *filter.BlocksReq) (int, error) {
	q := r.ch.NewSelect().
		Model((*core.Block)(nil))

	if f.Workchain != nil {
		q = q.Where("workchain = ?", *f.Workchain)
	}
	if f.Shard != nil {
		q = q.Where("shard = ?", *f.Shard)
	}
	if f.SeqNo != nil {
		q = q.Where("seq_no = ?", *f.SeqNo)
	}

	if len(f.FileHash) > 0 {
		q = q.Where("file_hash = ?", f.FileHash)
	}

	return q.Count(ctx)
}

func (r *Repository) FilterBlocks(ctx context.Context, f *filter.BlocksReq) (*filter.BlocksRes, error) {
	var (
		res = new(filter.BlocksRes)
		err error
	)

	res.Rows, err = r.filterBlocks(ctx, f)
	if err != nil {
		return res, err
	}
	if len(res.Rows) == 0 {
		return res, nil
	}

	res.Total, err = r.countBlocks(ctx, f)
	if err != nil {
		return res, err
	}

	return res, nil
}
