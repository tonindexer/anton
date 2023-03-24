package msg

import (
	"context"
	"strings"

	"github.com/uptrace/bun"
	"github.com/uptrace/go-clickhouse/ch"

	"github.com/iam047801/tonidx/internal/core"
	"github.com/iam047801/tonidx/internal/core/filter"
)

func (r *Repository) filterMsg(ctx context.Context, f *filter.MessagesReq) (ret []*core.Message, err error) {
	q := r.pg.NewSelect()
	if f.DBTx != nil {
		q = f.DBTx.NewSelect()
	}

	q = q.Model(&ret)

	if f.WithPayload {
		q = q.Relation("Payload")
	}

	if len(f.Hash) > 0 {
		q = q.Where("message.hash = ?", f.Hash)
	}
	if len(f.SrcAddresses) > 0 {
		q = q.Where("message.src_address in (?)", bun.In(f.SrcAddresses)).
			Where("length(message.src_address) > 0") // partial index
	}
	if len(f.DstAddresses) > 0 {
		q = q.Where("message.dst_address in (?)", bun.In(f.DstAddresses)).
			Where("length(message.dst_address) > 0") // partial index
	}

	if f.WithPayload {
		if len(f.SrcContracts) > 0 {
			q = q.Where("payload.src_contract IN (?)", bun.In(f.SrcContracts)).
				Where("length(payload.src_contract) > 0") // partial index
		}
		if len(f.DstContracts) > 0 {
			q = q.Where("payload.dst_contract IN (?)", bun.In(f.DstContracts)).
				Where("length(payload.dst_contract) > 0") // partial index
		}
		if len(f.OperationNames) > 0 {
			q = q.Where("payload.operation_name IN (?)", bun.In(f.OperationNames))
		}
		if f.MinterAddress != nil {
			q = q.Where("payload.minter_address = ?", f.MinterAddress).
				Where("length(payload.minter_address) > 0") // partial index
		}
	}

	if f.AfterTxLT != nil {
		if f.Order == "ASC" {
			q = q.Where("message.created_lt > ?", f.AfterTxLT)
		} else {
			q = q.Where("message.created_lt < ?", f.AfterTxLT)
		}
	}

	if f.Order != "" {
		q = q.Order("message.created_lt " + strings.ToUpper(f.Order))
	}

	if f.Limit == 0 {
		f.Limit = 3
	}
	q = q.Limit(f.Limit)

	err = q.Scan(ctx)
	return ret, err
}

func (r *Repository) countMsg(ctx context.Context, f *filter.MessagesReq) (int, error) {
	var payload bool // do we need to count account_data or account_states

	q := r.ch.NewSelect()

	if f.WithPayload {
		if len(f.SrcContracts) > 0 {
			q, payload = q.Where("src_contract IN (?)", ch.In(f.SrcContracts)), true
		}
		if len(f.DstContracts) > 0 {
			q, payload = q.Where("dst_contract IN (?)", ch.In(f.DstContracts)), true
		}
		if len(f.OperationNames) > 0 {
			q, payload = q.Where("operation_name IN (?)", ch.In(f.OperationNames)), true
		}
		if f.MinterAddress != nil {
			q, payload = q.Where("minter_address = ?", f.MinterAddress), true
		}
	}

	if len(f.Hash) > 0 {
		q = q.Where("hash = ?", f.Hash)
	}
	if len(f.SrcAddresses) > 0 {
		q = q.Where("src_address in (?)", ch.In(f.SrcAddresses))
	}
	if len(f.DstAddresses) > 0 {
		q = q.Where("dst_address in (?)", ch.In(f.DstAddresses))
	}

	if payload {
		q = q.Model((*core.MessagePayload)(nil))
	} else {
		q = q.Model((*core.Message)(nil))
	}

	return q.Count(ctx)
}

func (r *Repository) FilterMessages(ctx context.Context, f *filter.MessagesReq) (*filter.MessagesRes, error) {
	var (
		res = new(filter.MessagesRes)
		err error
	)

	res.Rows, err = r.filterMsg(ctx, f)
	if err != nil {
		return res, err
	}
	if len(res.Rows) == 0 {
		return res, nil
	}

	res.Total, err = r.countMsg(ctx, f)
	if err != nil {
		return res, err
	}

	return res, nil
}
