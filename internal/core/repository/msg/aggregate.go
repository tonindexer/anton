package msg

import (
	"context"

	"github.com/pkg/errors"
	"github.com/uptrace/go-clickhouse/ch"

	"github.com/tonindexer/anton/internal/core"
	"github.com/tonindexer/anton/internal/core/aggregate"
)

func (r *Repository) AggregateMessages(ctx context.Context, req *aggregate.MessagesReq) (*aggregate.MessagesRes, error) {
	var res aggregate.MessagesRes

	if req.Address == nil {
		return nil, errors.Wrap(core.ErrInvalidArg, "address must be set")
	}

	addTimestampFilter := func(q *ch.SelectQuery) *ch.SelectQuery {
		if !req.From.IsZero() {
			q = q.Where("created_at > ?", req.From)
		}
		if !req.To.IsZero() {
			q = q.Where("created_at < ?", req.To)
		}
		return q
	}

	err := addTimestampFilter(r.ch.NewSelect().Model((*core.Message)(nil)).
		ColumnExpr("count() as recv_count").
		ColumnExpr("sum(amount) as recv_amount").
		Where("dst_address = ?", req.Address)).
		Scan(ctx, &res.RecvCount, &res.RecvAmount)
	if err != nil {
		return nil, errors.Wrap(err, "received total")
	}

	err = addTimestampFilter(r.ch.NewSelect().Model((*core.Message)(nil)).
		ColumnExpr("count() as sent_count").
		ColumnExpr("sum(amount) as sent_amount").
		Where("src_address = ?", req.Address)).
		Scan(ctx, &res.SentCount, &res.SentAmount)
	if err != nil {
		return nil, errors.Wrap(err, "sent total")
	}

	err = r.ch.NewSelect().
		ColumnExpr("src_address as sender").
		ColumnExpr("count() as count").
		ColumnExpr("sum(sent_amount) as amount").
		TableExpr("(?) as q",
			addTimestampFilter(r.ch.NewSelect().Model((*core.Message)(nil)).
				ColumnExpr("src_address").
				ColumnExpr("amount as sent_amount").
				Where("dst_address = ?", req.Address))).
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
			addTimestampFilter(r.ch.NewSelect().Model((*core.Message)(nil)).
				ColumnExpr("dst_address").
				ColumnExpr("amount as sent_amount").
				Where("src_address = ?", req.Address))).
		Group("dst_address").
		Order(req.OrderBy+" DESC").
		Limit(req.Limit).
		Scan(ctx, &res.SentByAddress)
	if err != nil {
		return nil, errors.Wrap(err, "received by address")
	}

	return &res, nil
}
