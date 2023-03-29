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

	if req.WithPayload {
		q = q.Relation("Payload")
	}

	if len(req.Hash) > 0 {
		q = q.Where("message.hash = ?", req.Hash)
	}
	if len(req.SrcAddresses) > 0 {
		q = q.Where("message.src_address in (?)", bun.In(req.SrcAddresses)).
			Where("length(message.src_address) > 0") // partial index
	}
	if len(req.DstAddresses) > 0 {
		q = q.Where("message.dst_address in (?)", bun.In(req.DstAddresses)).
			Where("length(message.dst_address) > 0") // partial index
	}

	if req.WithPayload {
		if len(req.SrcContracts) > 0 {
			q = q.Where("payload.src_contract IN (?)", bun.In(req.SrcContracts)).
				Where("length(payload.src_contract) > 0") // partial index
		}
		if len(req.DstContracts) > 0 {
			q = q.Where("payload.dst_contract IN (?)", bun.In(req.DstContracts)).
				Where("length(payload.dst_contract) > 0") // partial index
		}
		if len(req.OperationNames) > 0 {
			q = q.Where("payload.operation_name IN (?)", bun.In(req.OperationNames))
		}
		if req.MinterAddress != nil {
			q = q.Where("payload.minter_address = ?", req.MinterAddress).
				Where("length(payload.minter_address) > 0") // partial index
		}
	}

	if req.AfterTxLT != nil {
		if req.Order == "ASC" {
			q = q.Where("message.created_lt > ?", req.AfterTxLT)
		} else {
			q = q.Where("message.created_lt < ?", req.AfterTxLT)
		}
	}

	if req.Order != "" {
		q = q.Order("message.created_lt " + strings.ToUpper(req.Order))
	}

	if req.Limit == 0 {
		req.Limit = 3
	}
	q = q.Limit(req.Limit)

	err = q.Scan(ctx)
	return ret, err
}

func (r *Repository) countMsg(ctx context.Context, req *filter.MessagesReq) (int, error) {
	var payload bool // do we need to count account_data or account_states

	q := r.ch.NewSelect()

	if req.WithPayload {
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
	}

	if len(req.Hash) > 0 {
		q = q.Where("hash = ?", req.Hash)
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
