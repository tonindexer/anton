package tx

import (
	"context"

	"github.com/pkg/errors"

	"github.com/iam047801/tonidx/internal/core"
)

func (r *Repository) AggregateMessages(ctx context.Context, req *core.MessageAggregate) (*core.MessageAggregated, error) {
	var res core.MessageAggregated

	if req.Address == nil {
		return nil, errors.Wrap(core.ErrInvalidArg, "address must be set")
	}

	err := r.ch.NewSelect().Model((*core.Message)(nil)).
		ColumnExpr("count() as recv_count").
		ColumnExpr("sum(amount) as recv_amount").
		Where("dst_address = ?", req.Address).
		Scan(ctx, &res.RecvCount, &res.RecvAmount)
	if err != nil {
		return nil, errors.Wrap(err, "received total")
	}

	err = r.ch.NewSelect().Model((*core.Message)(nil)).
		ColumnExpr("count() as sent_count").
		ColumnExpr("sum(amount) as sent_amount").
		Where("src_address = ?", req.Address).
		Scan(ctx, &res.SentCount, &res.SentAmount)
	if err != nil {
		return nil, errors.Wrap(err, "sent total")
	}

	err = r.ch.NewSelect().
		ColumnExpr("src_address as sender").
		ColumnExpr("count() as count").
		ColumnExpr("sum(sent_amount) as amount").
		TableExpr("(?) as q",
			r.ch.NewSelect().Model((*core.Message)(nil)).
				ColumnExpr("src_address").
				ColumnExpr("amount as sent_amount").
				Where("dst_address = ?", req.Address)).
		Group("src_address").
		Order(req.OrderBy+" DESC").
		Limit(req.Limit).
		Scan(ctx, &res.RecvByAddress)
	if err != nil {
		return nil, errors.Wrap(err, "received by address")
	}

	err = r.ch.NewSelect().
		ColumnExpr("dst_address as receiver").
		ColumnExpr("count() as count").
		ColumnExpr("sum(sent_amount) as amount").
		TableExpr("(?) as q",
			r.ch.NewSelect().Model((*core.Message)(nil)).
				ColumnExpr("dst_address").
				ColumnExpr("amount as sent_amount").
				Where("src_address = ?", req.Address)).
		Group("dst_address").
		Order(req.OrderBy+" DESC").
		Limit(req.Limit).
		Scan(ctx, &res.SentByAddress)
	if err != nil {
		return nil, errors.Wrap(err, "received by address")
	}

	return &res, nil
}
