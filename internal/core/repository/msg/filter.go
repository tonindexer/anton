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
		Count int
		MaxLT *uint64 `ch:"max_lt"`
	}

	q := r.ch.NewSelect().
		Model((*core.Message)(nil)).
		ColumnExpr("count(*) AS count").
		ColumnExpr("(SELECT max(created_lt) FROM messages) AS max_lt") // unfiltered max

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

	if err := q.Scan(ctx, &result); err != nil {
		return 0, 0, err
	}

	if result.MaxLT == nil {
		return 0, 0, core.ErrNotFound
	}

	return result.Count, *result.MaxLT, nil
}

func (r *Repository) countMsgPartialScan(ctx context.Context, req *filter.MessagesReq, startLt uint64) (partialCount int, maxLt uint64, err error) {
	var result struct {
		Count int
		MaxLT uint64 `ch:"max_lt"`
	}

	q := r.pg.NewSelect().
		Model((*core.Message)(nil)).
		ColumnExpr("count(*) AS count").
		ColumnExpr("(select max(created_lt) from messages where created_lt > ?) AS max_lt", startLt). // unfiltered max
		Where("created_lt > ?", startLt)

	q = r.getFilterMessageQuery(q, &req.MessagesFilter)

	if err := q.Scan(ctx, &result); err != nil {
		return 0, 0, err
	}

	if result.MaxLT == 0 {
		result.MaxLT = startLt // no new rows
	}

	return result.Count, result.MaxLT, nil
}

func (r *Repository) countMsg(ctx context.Context, req *filter.MessagesReq) (int, error) {
	count, maxLT, err := r.messagesFilterCache.Get(req.MessagesFilter)
	if errors.Is(err, core.ErrNotFound) {
		count, maxLT, err = r.countMsgFullScan(ctx, req)
		if err != nil {
			return 0, err
		}
		if errors.Is(err, core.ErrNotFound) {
			return 0, nil
		}
		if err := r.messagesFilterCache.Set(req.MessagesFilter, count, maxLT); err != nil {
			return 0, err
		}
	}
	if err != nil && !errors.Is(err, core.ErrNotFound) {
		return 0, err
	}

	partialCount, maxLT, err := r.countMsgPartialScan(ctx, req, maxLT)
	if err != nil {
		return 0, err
	}
	if err := r.messagesFilterCache.Set(req.MessagesFilter, count+partialCount, maxLT); err != nil {
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
