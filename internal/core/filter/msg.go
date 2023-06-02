package filter

import (
	"context"

	"github.com/uptrace/bun"

	"github.com/tonindexer/anton/addr"
	"github.com/tonindexer/anton/internal/core"
)

type MessagesReq struct {
	DBTx *bun.Tx

	Hash         []byte          // `form:"hash"`
	SrcAddresses []*addr.Address // `form:"src_address"`
	DstAddresses []*addr.Address // `form:"dst_address"`
	OperationID  *uint32

	SrcContracts   []string `form:"src_contract"`
	DstContracts   []string `form:"dst_contract"`
	OperationNames []string `form:"operation_name"`

	Order string `form:"order"` // ASC, DESC

	AfterTxLT *uint64 `form:"after"`
	Limit     int     `form:"limit"`
}

type MessagesRes struct {
	Total int             `json:"total"`
	Rows  []*core.Message `json:"results"`
}

type MessageRepository interface {
	FilterMessages(context.Context, *MessagesReq) (*MessagesRes, error)
}
