package history

import (
	"context"

	"github.com/iam047801/tonidx/internal/addr"
)

type MessageMetric string

var (
	MessageCount          MessageMetric = "message_count"
	MessageStateInitCount MessageMetric = "state_init_count"
	MessageAmountSum      MessageMetric = "message_amount_sum"
)

type MessagesReq struct {
	Metric MessageMetric `form:"metric"`

	SrcAddresses []*addr.Address // `form:"src_address"`
	DstAddresses []*addr.Address // `form:"dst_address"`

	SrcContracts []string `form:"src_contract"`
	DstContracts []string `form:"dst_contract"`

	OperationNames []string `form:"operation_name"`

	MinterAddress *addr.Address // `form:"minter_address"`

	ReqParams
}

type MessagesRes struct {
	CountRes  `json:"count_results,omitempty"`
	BigIntRes `json:"sum_results,omitempty"`
}

type MessageRepository interface {
	AggregateMessagesHistory(ctx context.Context, req *MessagesReq) (*MessagesRes, error)
}
