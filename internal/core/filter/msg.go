package filter

import (
	"context"

	"github.com/uptrace/bun"

	"github.com/iam047801/tonidx/internal/addr"
	"github.com/iam047801/tonidx/internal/core"
)

type MessagesReq struct {
	DBTx *bun.Tx

	Hash         []byte          // `form:"hash"`
	SrcAddresses []*addr.Address // `form:"src_address"`
	DstAddresses []*addr.Address // `form:"dst_address"`

	WithPayload    bool
	SrcContracts   []string      `form:"src_contract"`
	DstContracts   []string      `form:"dst_contract"`
	OperationNames []string      `form:"operation_name"`
	MinterAddress  *addr.Address // `form:"minter_address"`

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
