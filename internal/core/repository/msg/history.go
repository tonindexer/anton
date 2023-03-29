package msg

import (
	"context"
	"fmt"

	"github.com/pkg/errors"
	"github.com/uptrace/go-clickhouse/ch"

	"github.com/tonindexer/anton/internal/core"
	"github.com/tonindexer/anton/internal/core/aggregate/history"
)

func (r *Repository) AggregateMessagesHistory(ctx context.Context, req *history.MessagesReq) (*history.MessagesRes, error) {
	var res history.MessagesRes
	var payload, bigIntRes bool // do we need to count account_data or account_states

	q := r.ch.NewSelect()

	if len(req.SrcContracts) > 0 {
		q, payload = q.Where("src_contract IN (?)", ch.In(req.SrcContracts)), true
	}
	if len(req.DstContracts) > 0 {
		q, payload = q.Where("dst_contract IN (?)", ch.In(req.DstContracts)), true
	}
	if len(req.OperationNames) > 0 {
		q, payload = q.Where("operation_name IN (?)", ch.In(req.OperationNames)), true
	}
	if req.MinterAddress != nil {
		q, payload = q.Where("minter_address = ?", req.MinterAddress), true
	}

	if len(req.SrcAddresses) > 0 {
		q = q.Where("src_address in (?)", ch.In(req.SrcAddresses))
	}
	if len(req.DstAddresses) > 0 {
		q = q.Where("dst_address in (?)", ch.In(req.DstAddresses))
	}

	if payload {
		q = q.Model((*core.MessagePayload)(nil))
	} else {
		q = q.Model((*core.Message)(nil))
	}

	switch req.Metric {
	case history.MessageCount:
		q = q.ColumnExpr("count() as value")
	case history.MessageAmountSum:
		q, bigIntRes = q.ColumnExpr("sum(amount) as value"), true
	default:
		return nil, errors.Wrapf(core.ErrInvalidArg, "invalid message metric %s", req.Metric)
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

	if bigIntRes {
		if err := q.Scan(ctx, &res.BigIntRes); err != nil {
			return nil, err
		}
	} else {
		if err := q.Scan(ctx, &res.CountRes); err != nil {
			return nil, err
		}
	}

	return &res, nil
}
