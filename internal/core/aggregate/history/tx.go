package history

import (
	"context"

	"github.com/tonindexer/anton/addr"
)

type TransactionMetric string

var (
	TransactionCount TransactionMetric = "transaction_count"
	// AvgFee
)

type TransactionsReq struct {
	Metric TransactionMetric `form:"metric"`

	Addresses []*addr.Address // `form:"addresses"`
	Workchain *int32          `form:"workchain"`

	ReqParams
}

type TransactionsRes struct {
	CountRes `json:"count_results,omitempty"`
	// BigIntRes `json:"sum_results"`
}

type TransactionRepository interface {
	AggregateTransactionsHistory(ctx context.Context, req *TransactionsReq) (*TransactionsRes, error)
}
