package history

import (
	"context"

	"github.com/iam047801/tonidx/abi"
)

type AccountMetric string

const (
	UniqueAddresses AccountMetric = "unique_addresses"
)

type AccountsReq struct {
	Metric AccountMetric `form:"metric"`

	ContractTypes []abi.ContractName `form:"interface"`

	ReqParams
}

type AccountsRes struct {
	CountRes `json:"count_results"`
}

type AccountRepository interface {
	AggregateAccountsHistory(ctx context.Context, req *AccountsReq) (*AccountsRes, error)
}
