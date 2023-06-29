package filter

import (
	"context"

	"github.com/tonindexer/anton/abi"
	"github.com/tonindexer/anton/addr"
	"github.com/tonindexer/anton/internal/core"
)

type LabelsReq struct {
	Name       string               `form:"name"`
	Categories []core.LabelCategory `form:"category"`
	Offset     int                  `form:"offset"`
	Limit      int                  `form:"limit"`
}

type LabelsRes struct {
	Total int                  `json:"total"`
	Rows  []*core.AddressLabel `json:"results"`
}

type AccountsReq struct {
	Addresses   []*addr.Address // `form:"addresses"`
	LatestState bool            `form:"latest"`

	// contract data filter
	ContractTypes []abi.ContractName `form:"interface"`
	OwnerAddress  *addr.Address      // `form:"owner_address"`
	MinterAddress *addr.Address      // `form:"minter_address"`

	ExcludeColumn []string // TODO: support relations

	Order string `form:"order"` // ASC, DESC

	AfterTxLT *uint64 `form:"after"`
	Limit     int     `form:"limit"`
}

type AccountsRes struct {
	Total int                  `json:"total"`
	Rows  []*core.AccountState `json:"results"`
}

type AccountRepository interface {
	FilterLabels(context.Context, *LabelsReq) (*LabelsRes, error)
	FilterAccounts(context.Context, *AccountsReq) (*AccountsRes, error)
}
