package msg

import (
	"context"
	"strings"

	"github.com/pkg/errors"
	"github.com/uptrace/bun"
	"github.com/uptrace/go-clickhouse/ch"

	"github.com/tonindexer/anton/internal/core"
	"github.com/tonindexer/anton/internal/core/filter"
)

func (r *Repository) getFilterMessageQuery(q *bun.SelectQuery, req *filter.MessagesFilter) *bun.SelectQuery {
	if len(req.Hash) > 0 {
		q = q.Where("hash = ?", req.Hash)
	}
	if len(req.SrcAddresses) > 0 {
		q = q.Where("src_address in (?)", bun.In(req.SrcAddresses))
	}
	if len(req.DstAddresses) > 0 {
		q = q.Where("dst_address in (?)", bun.In(req.DstAddresses))
	}
	if req.SrcWorkchain != nil {
		q = q.Where("src_workchain = ?", *req.SrcWorkchain)
	}
	if req.DstWorkchain != nil {
		q = q.Where("dst_workchain = ?", *req.DstWorkchain)
	}
	if req.OperationID != nil {
		q = q.Where("operation_id = ?", *req.OperationID)
	}

	if len(req.SrcContracts) > 0 {
		q = q.Where("src_contract IN (?)", bun.In(req.SrcContracts))
	}
	if len(req.DstContracts) > 0 {
		q = q.Where("dst_contract IN (?)", bun.In(req.DstContracts))
	}
	if len(req.OperationNames) > 0 {
		q = q.Where("operation_name IN (?)", bun.In(req.OperationNames))
	}

	return q
}

func (r *Repository) filterMsg(ctx context.Context, req *filter.MessagesReq) (ret []*core.Message, err error) {
	q := r.pg.NewSelect()
	if req.DBTx != nil {
		q = req.DBTx.NewSelect()
	}

	q = q.Model(&ret)

	q = r.getFilterMessageQuery(q, &req.MessagesFilter)

	if req.AfterTxLT != nil {
		if req.Order == "ASC" {
			q = q.Where("created_lt > ?", req.AfterTxLT)
		} else {
			q = q.Where("created_lt < ?", req.AfterTxLT)
		}
	}

	if req.Order != "" {
		q = q.Order("created_lt " + strings.ToUpper(req.Order))
	}

	if req.Limit == 0 {
		req.Limit = 3
	}
	q = q.Limit(req.Limit)

	err = q.Scan(ctx)
	return ret, err
}

func (r *Repository) countMsgFullScan(ctx context.Context, req *filter.MessagesReq) (count int, maxLt uint64, err error) {
	var result struct {
		MaxLT        uint64 `ch:"max_lt_value"`
		RoundedMaxLT uint64 `ch:"max_lt_rounded"`
		Count        int
	}

	q := r.ch.NewSelect().
		Model((*core.Message)(nil))

	if len(req.Hash) > 0 {
		q = q.Where("hash = ?", req.Hash)
	}
	if len(req.SrcAddresses) > 0 {
		q = q.Where("src_address in (?)", ch.In(req.SrcAddresses))
	}
	if len(req.DstAddresses) > 0 {
		q = q.Where("dst_address in (?)", ch.In(req.DstAddresses))
	}
	if req.SrcWorkchain != nil {
		q = q.Where("src_workchain = ?", *req.SrcWorkchain)
	}
	if req.DstWorkchain != nil {
		q = q.Where("dst_workchain = ?", *req.DstWorkchain)
	}
	if req.OperationID != nil {
		q = q.Where("operation_id = ?", *req.OperationID)
	}
	if len(req.SrcContracts) > 0 {
		q = q.Where("src_contract IN (?)", ch.In(req.SrcContracts))
	}
	if len(req.DstContracts) > 0 {
		q = q.Where("dst_contract IN (?)", ch.In(req.DstContracts))
	}
	if len(req.OperationNames) > 0 {
		q = q.Where("operation_name IN (?)", ch.In(req.OperationNames))
	}

	q = r.ch.NewSelect().
		With(
			"max_lt",
			r.ch.NewSelect().
				Model((*core.Message)(nil)).
				ColumnExpr("max(created_lt) AS v"),
		).
		With(
			"rounded_count",
			q. // query with filters
				Table("max_lt").
				ColumnExpr("count(*) as v").
				Where("created_lt <= floor(max_lt.v, -7)"), // we round LT as messages in new blocks can have lower LT
		).
		Table("max_lt", "rounded_count").
		ColumnExpr("max_lt.v AS max_lt_value").
		ColumnExpr("floor(max_lt.v, -7) as max_lt_rounded").
		ColumnExpr("rounded_count.v AS count")

	if err := q.Scan(ctx, &result); err != nil {
		return 0, 0, err
	}

	if result.MaxLT == 0 {
		return 0, 0, core.ErrNotFound
	}

	return result.Count, result.RoundedMaxLT, nil
}

func (r *Repository) countMsgPartialScan(ctx context.Context, req *filter.MessagesReq, startLt uint64) (partialCount, roundedCount int, roundedMaxLt uint64, err error) {
	var result struct {
		Since        int    `bun:"since_rounded_count"`
		Until        int    `bun:"until_rounded_count"`
		RoundedMaxLT uint64 `bun:"rounded_max_lt_value"`
	}

	q := r.pg.NewSelect().
		With(
			"rounded_max_lt",
			r.pg.NewSelect().
				Model((*core.Message)(nil)).
				ColumnExpr("floor(max(created_lt) / 1e7) * 1e7 AS v"), // we round LT as messages in new blocks can have lower LT
		).
		With(
			"until_rounded_count",
			r.getFilterMessageQuery(
				r.pg.NewSelect().Model((*core.Message)(nil)),
				&req.MessagesFilter,
			).
				Table("rounded_max_lt").
				ColumnExpr("count(*) as v").
				Where("created_lt > ?", startLt).
				Where("created_lt <= rounded_max_lt.v"),
		).
		With("since_rounded_count",
			r.getFilterMessageQuery(
				r.pg.NewSelect().Model((*core.Message)(nil)),
				&req.MessagesFilter,
			).
				Table("rounded_max_lt").
				ColumnExpr("count(*) as v").
				Where("created_lt >= rounded_max_lt.v")).
		Table("rounded_max_lt", "until_rounded_count", "since_rounded_count").
		ColumnExpr("since_rounded_count.v AS since_rounded_count").
		ColumnExpr("until_rounded_count.v AS until_rounded_count").
		ColumnExpr("rounded_max_lt.v as rounded_max_lt_value")

	if err := q.Scan(ctx, &result); err != nil {
		return 0, 0, 0, err
	}

	return result.Since + result.Until, result.Until, result.RoundedMaxLT, nil
}

func (r *Repository) countMsg(ctx context.Context, req *filter.MessagesReq) (int, error) {
	count, maxLT, err := r.messagesFilterCountCache.Get(req.MessagesFilter)
	if errors.Is(err, core.ErrNotFound) {
		count, maxLT, err = r.countMsgFullScan(ctx, req)
		if errors.Is(err, core.ErrNotFound) {
			return 0, nil
		}
		if err != nil {
			return 0, err
		}
		if err := r.messagesFilterCountCache.Set(req.MessagesFilter, count, maxLT); err != nil {
			return 0, err
		}
	}
	if err != nil && !errors.Is(err, core.ErrNotFound) {
		return 0, err
	}

	partialCount, roundedPartialCount, roundedMaxLT, err := r.countMsgPartialScan(ctx, req, maxLT)
	if err != nil {
		return 0, err
	}
	if err := r.messagesFilterCountCache.Set(req.MessagesFilter, count+roundedPartialCount, roundedMaxLT); err != nil {
		return 0, err
	}

	return count + partialCount, nil
}

func (r *Repository) FilterMessages(ctx context.Context, req *filter.MessagesReq) (*filter.MessagesRes, error) {
	var (
		res = new(filter.MessagesRes)
		err error
	)

	res.Rows, err = r.filterMsg(ctx, req)
	if err != nil {
		return res, err
	}

	switch {
	case len(res.Rows) == 0:

	case len(req.Hash) > 0:
		res.Total = len(res.Rows)

	case req.Count:
		res.Total, err = r.countMsg(ctx, req)
		if err != nil {
			return res, err
		}
	}

	return res, nil
}
