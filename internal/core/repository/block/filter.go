package block

import (
	"context"
	"strings"

	"github.com/pkg/errors"
	"github.com/uptrace/bun"

	"github.com/tonindexer/anton/internal/core"
	"github.com/tonindexer/anton/internal/core/filter"
)

func loadTransactions(q *bun.SelectQuery, prefix string, f *filter.BlocksReq) *bun.SelectQuery {
	q = q.Relation(prefix + "Transactions")
	if f.WithTransactionAccountState {
		q = q.Relation(prefix+"Transactions.Account", func(q *bun.SelectQuery) *bun.SelectQuery {
			if len(f.ExcludeColumn) > 0 {
				return q.ExcludeColumn(f.ExcludeColumn...)
			}
			return q
		})
	}
	if f.WithTransactionMessages {
		q = q.
			Relation(prefix + "Transactions.InMsg").
			Relation(prefix + "Transactions.OutMsg")
	}
	return q
}

func (r *Repository) countTransactions(ctx context.Context, ret []*core.Block) error {
	if len(ret) == 0 {
		return nil
	}

	var blockIDs [][]int64
	for _, m := range ret {
		for _, s := range m.Shards {
			blockIDs = append(blockIDs, []int64{int64(s.Workchain), s.Shard, int64(s.SeqNo)})
		}
		blockIDs = append(blockIDs, []int64{int64(m.Workchain), m.Shard, int64(m.SeqNo)})
	}

	var res []struct {
		Workchain         int32
		Shard             int64
		SeqNo             uint32 `bun:"block_seq_no"`
		TransactionsCount int
	}
	err := r.pg.NewSelect().
		TableExpr("transactions AS tx").
		Column("workchain", "shard", "block_seq_no").
		ColumnExpr("COUNT(*) as transactions_count").
		Where("(workchain, shard, block_seq_no) IN (?)", bun.In(blockIDs)).
		Group("workchain", "shard", "block_seq_no").
		Scan(ctx, &res)
	if err != nil {
		return errors.Wrap(err, "count transactions for each block")
	}

	var counts = make(map[int32]map[int64]map[uint32]int)
	for _, r := range res {
		if counts[r.Workchain] == nil {
			counts[r.Workchain] = map[int64]map[uint32]int{}
		}
		if counts[r.Workchain][r.Shard] == nil {
			counts[r.Workchain][r.Shard] = map[uint32]int{}
		}
		counts[r.Workchain][r.Shard][r.SeqNo] = r.TransactionsCount
	}

	getCount := func(workchain int32, shard int64, seqNo uint32) int {
		if counts[workchain] == nil {
			return 0
		}
		if counts[workchain][shard] == nil {
			return 0
		}
		return counts[workchain][shard][seqNo]
	}

	for _, m := range ret {
		for _, s := range m.Shards {
			s.TransactionsCount = getCount(s.Workchain, s.Shard, s.SeqNo)
		}
		m.TransactionsCount = getCount(m.Workchain, m.Shard, m.SeqNo)
	}

	return nil
}

func (r *Repository) filterBlocks(ctx context.Context, f *filter.BlocksReq) (ret []*core.Block, err error) {
	q := r.pg.NewSelect().Model(&ret)

	if f.WithShards {
		q = q.Relation("Shards")
		if f.WithTransactions {
			q = loadTransactions(q, "Shards.", f)
		}
		if f.WithAccountStates {
			q = q.Relation("Shards.Accounts")
		}
	}
	if f.WithTransactions {
		q = loadTransactions(q, "", f)
	}
	if f.WithAccountStates {
		q = q.Relation("Accounts")
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

	if err := q.Scan(ctx); err != nil {
		return nil, err
	}

	if err := r.countTransactions(ctx, ret); err != nil {
		return nil, err
	}

	return ret, nil
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

	if f.Count {
		res.Total, err = r.countBlocks(ctx, f)
		if err != nil {
			return res, err
		}
	}

	return res, nil
}
