package msg

import (
	"context"
	"strings"

	"github.com/uptrace/bun"
	"github.com/uptrace/go-clickhouse/ch"

	"github.com/tonindexer/anton/internal/core"
	"github.com/tonindexer/anton/internal/core/filter"
)

func (r *Repository) filterMsg(ctx context.Context, req *filter.MessagesReq) (ret []*core.Message, err error) {
	q := r.pg.NewSelect()
	if req.DBTx != nil {
		q = req.DBTx.NewSelect()
	}

	q = q.Model(&ret)

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

func (r *Repository) countMsg(ctx context.Context, req *filter.MessagesReq) (int, error) {
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

	return q.Count(ctx)
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
	if len(res.Rows) == 0 {
		return res, nil
	}

	res.Total, err = r.countMsg(ctx, req)
	if err != nil {
		return res, err
	}

	return res, nil
}
