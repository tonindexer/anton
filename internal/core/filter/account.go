package filter

import (
	"context"

	"github.com/iam047801/tonidx/abi"
	"github.com/iam047801/tonidx/internal/addr"
	"github.com/iam047801/tonidx/internal/core"
)

type AccountsReq struct {
	Addresses   []*addr.Address // `form:"addresses"`
	LatestState bool            `form:"latest"`

	// contract data filter
	WithData      bool
	ContractTypes []abi.ContractName `form:"interface"`
	OwnerAddress  *addr.Address      // `form:"owner_address"`
	MinterAddress *addr.Address      // `form:"minter_address"`

	Order string `form:"order"` // ASC, DESC

	AfterTxLT *uint64 `form:"after"`
	Limit     int     `form:"limit"`
}

type AccountsRes struct {
	Total int                  `json:"total"`
	Rows  []*core.AccountState `json:"results"`
}

type AccountRepository interface {
	FilterAccounts(context.Context, *AccountsReq) (*AccountsRes, error)
}
