package aggregate

import (
	"context"

	"github.com/uptrace/bun/extra/bunbig"

	"github.com/tonindexer/anton/addr"
)

type MessagesReq struct {
	Address *addr.Address

	OrderBy string `form:"order_by"` // amount / count
	Limit   int    `form:"limit"`
}

type MessagesRes struct {
	RecvCount  int         `json:"received_count"`
	RecvAmount *bunbig.Int `json:"received_ton_amount"`

	SentCount  int         `json:"sent_count"`
	SentAmount *bunbig.Int `json:"sent_ton_amount"`

	RecvByAddress []struct {
		Sender *addr.Address `ch:"type:String" json:"sender"`
		Amount *bunbig.Int   `json:"amount"`
		Count  int           `json:"count"`
	} `json:"received_from_address"`

	SentByAddress []struct {
		Receiver *addr.Address `ch:"type:String" json:"receiver"`
		Amount   *bunbig.Int   `json:"amount"`
		Count    int           `json:"count"`
	} `json:"sent_to_address"`
}

type MessageRepository interface {
	AggregateMessages(ctx context.Context, req *MessagesReq) (*MessagesRes, error)
}
