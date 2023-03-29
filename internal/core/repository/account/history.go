package account

import (
	"context"
	"fmt"

	"github.com/pkg/errors"
	"github.com/uptrace/go-clickhouse/ch"

	"github.com/tonindexer/anton/abi"
	"github.com/tonindexer/anton/internal/core"
	"github.com/tonindexer/anton/internal/core/aggregate/history"
)

func getContractTypes(types []abi.ContractName) (ret []abi.ContractName) {
	for _, t := range types {
		ret = append(ret, t)
		if t == "wallet" {
			ret = append(ret, abi.GetAllWalletNames()...)
		}
	}
	return
}

func (r *Repository) AggregateAccountsHistory(ctx context.Context, req *history.AccountsReq) (*history.AccountsRes, error) {
	var res history.AccountsRes
	var data bool // do we need to count account_data or account_states

	q := r.ch.NewSelect()

	if len(req.ContractTypes) > 0 {
		q, data = q.Where("hasAny(types, [?])", ch.In(getContractTypes(req.ContractTypes))), true
	}
	if req.MinterAddress != nil {
		q, data = q.Where("minter_address = ?", req.MinterAddress), true
	}

	if data {
		q = q.Model((*core.AccountData)(nil))
	} else {
		q = q.Model((*core.AccountState)(nil))
	}

	switch req.Metric {
	case history.ActiveAddresses:
		q = q.ColumnExpr("uniq(address) as value")
	default:
		return nil, errors.Wrapf(core.ErrInvalidArg, "invalid account metric %s", req.Metric)
	}

	rounding, err := history.GetRoundingFunction(req.Interval)
	if err != nil {
		return nil, err
	}
	q = q.ColumnExpr(fmt.Sprintf(rounding, "updated_at") + " as timestamp")
	q = q.Group("timestamp")

	if !req.From.IsZero() {
		q = q.Where("updated_at > ?", req.From)
	}
	if !req.To.IsZero() {
		q = q.Where("updated_at < ?", req.To)
	}

	q = q.Order("timestamp ASC")

	if err := q.Scan(ctx, &res.CountRes); err != nil {
		return nil, err
	}

	return &res, nil
}
