package filter

import (
	"context"

	"github.com/tonindexer/anton/internal/core"
)

type BlocksFilter struct {
	Workchain *int32  `form:"workchain"`
	Shard     *int64  `form:"shard"`
	SeqNo     *uint32 `form:"seq_no"`
	FileHash  []byte  `form:"file_hash"`
}

type BlocksReq struct {
	BlocksFilter

	WithShards                  bool // TODO: array of relations as strings
	WithAccountStates           bool
	WithTransactionAccountState bool
	WithTransactions            bool `form:"with_transactions"`
	WithTransactionMessages     bool

	ExcludeColumn []string // TODO: support relations

	Order string `form:"order"` // ASC, DESC

	AfterSeqNo *uint32 `form:"after"`
	Limit      int     `form:"limit"`
	Count      bool    `form:"count"`
}

type BlocksRes struct {
	Total int           `json:"total,omitempty"`
	Rows  []*core.Block `json:"results"`
}

type BlockRepository interface {
	FilterBlocks(context.Context, *BlocksReq) (*BlocksRes, error)
}
