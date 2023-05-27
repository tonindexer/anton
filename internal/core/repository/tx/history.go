package tx

import (
	"context"
	"fmt"

	"github.com/pkg/errors"
	"github.com/uptrace/go-clickhouse/ch"

	"github.com/tonindexer/anton/internal/core"
	"github.com/tonindexer/anton/internal/core/aggregate/history"
)

func (r *Repository) AggregateTransactionsHistory(ctx context.Context, req *history.TransactionsReq) (*history.TransactionsRes, error) {
	var res history.TransactionsRes

	q := r.ch.NewSelect().
		Model((*core.Transaction)(nil))

	if len(req.Addresses) > 0 {
		q = q.Where("address in (?)", ch.In(req.Addresses))
	}
	if req.Workchain != nil {
		q = q.Where("workchain = ?", *req.Workchain)
	}

	switch req.Metric {
	case history.TransactionCount:
		q = q.ColumnExpr("count() as value")
	default:
		return nil, errors.Wrapf(core.ErrInvalidArg, "invalid transaction metric %s", req.Metric)
	}

	rounding, err := history.GetRoundingFunction(req.Interval)
	if err != nil {
		return nil, err
	}
	q = q.ColumnExpr(fmt.Sprintf(rounding, "created_at") + " as timestamp")
	q = q.Group("timestamp")

	if !req.From.IsZero() {
		q = q.Where("created_at > ?", req.From)
	}
	if !req.To.IsZero() {
		q = q.Where("created_at < ?", req.To)
	}

	q = q.Order("timestamp ASC")

	if err := q.Scan(ctx, &res.CountRes); err != nil {
		return nil, err
	}

	return &res, nil
}
