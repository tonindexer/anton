package history

import (
	"context"

	"github.com/iam047801/tonidx/internal/addr"
)

type TransactionMetric string

var (
	TransactionCount TransactionMetric = "transaction_count"
	// AvgFee
)

type TransactionsReq struct {
	Metric TransactionMetric

	Addresses []*addr.Address
	Workchain *int32

	ReqParams
}

type TransactionsRes struct {
	CountRes `json:"count_results"`
	// BigIntRes `json:"sum_results"`
}

type TransactionRepository interface {
	AggregateTransactionsHistory(ctx context.Context, req *TransactionsReq) (*TransactionsRes, error)
}
