package history

import (
	"context"

	"github.com/iam047801/tonidx/abi"
	"github.com/iam047801/tonidx/internal/addr"
)

type AccountMetric string

const (
	ActiveAddresses AccountMetric = "active_addresses"
)

type AccountsReq struct {
	Metric AccountMetric `form:"metric"`

	ContractTypes []abi.ContractName `form:"interface"`
	MinterAddress *addr.Address      // NFT or FT minter

	ReqParams
}

type AccountsRes struct {
	CountRes `json:"count_results,omitempty"`
}

type AccountRepository interface {
	AggregateAccountsHistory(ctx context.Context, req *AccountsReq) (*AccountsRes, error)
}
