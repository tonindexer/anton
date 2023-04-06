package filter

import (
	"context"

	"github.com/tonindexer/anton/internal/addr"
	"github.com/tonindexer/anton/internal/core"
)

type TransactionsReq struct {
	Hash      []byte // `form:"hash"`
	InMsgHash []byte // `form:"in_msg_hash"`

	Addresses []*addr.Address //

	Workchain *int32 `form:"workchain"`

	BlockID *core.BlockID

	WithAccountState    bool
	WithAccountData     bool
	WithMessages        bool
	WithMessagePayloads bool

	ExcludeColumn []string // TODO: support relations

	Order string `form:"order"` // ASC, DESC

	AfterTxLT *uint64 `form:"after"`
	Limit     int     `form:"limit"`
}

type TransactionsRes struct {
	Total int                 `json:"total"`
	Rows  []*core.Transaction `json:"results"`
}

type TransactionRepository interface {
	FilterTransactions(context.Context, *TransactionsReq) (*TransactionsRes, error)
}
