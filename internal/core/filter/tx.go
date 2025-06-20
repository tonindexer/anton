package filter

import (
	"context"

	"github.com/tonindexer/anton/addr"
	"github.com/tonindexer/anton/internal/core"
)

type TransactionsFilter struct {
	Hash      []byte // `form:"hash"`
	InMsgHash []byte // `form:"in_msg_hash"`

	Addresses []*addr.Address //

	Workchain *int32 `form:"workchain"`

	BlockID *core.BlockID
}

type TransactionsReq struct {
	TransactionsFilter

	WithAccountState bool
	WithMessages     bool

	ExcludeColumn []string // TODO: support relations

	Order string `form:"order"` // ASC, DESC

	CreatedLT *uint64 `form:"created_lt"`

	AfterTxLT *uint64 `form:"after"`
	Limit     int     `form:"limit"`
	Count     bool    `form:"count"`
}

type TransactionsRes struct {
	Total int                 `json:"total,omitempty"`
	Rows  []*core.Transaction `json:"results"`
}

type TransactionRepository interface {
	FilterTransactions(context.Context, *TransactionsReq) (*TransactionsRes, error)
}
