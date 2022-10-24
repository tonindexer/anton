package core

import (
	"context"

	"github.com/uptrace/go-clickhouse/ch"
	"github.com/xssnick/tonutils-go/tlb"
)

type Transaction struct {
	ch.CHModel `ch:"transactions,partition:account_addr"`

	AccountAddr    string `ch:",pk"`
	AccountBalance uint64 // TODO: uint256
	Hash           []byte `ch:",pk"`
	LT             uint64 `ch:",pk"`
	PrevTxHash     []byte //
	PrevTxLT       uint64 //
	OutMsgCount    uint16 //
	TotalFees      uint64 // `ch:"type:UInt256"`
}

type MessageType string

const (
	Internal    = MessageType(tlb.MsgTypeInternal)
	ExternalIn  = MessageType(tlb.MsgTypeExternalIn)
	ExternalOut = MessageType(tlb.MsgTypeExternalOut)
)

type Message struct {
	ch.CHModel `ch:"messages,partition:type,src_addr,dst_addr,round(amount,-9)"`

	TxHash          []byte      `ch:",pk"`
	Hash            []byte      `ch:",pk"`
	Type            MessageType `ch:",lc"` // TODO: enum
	Incoming        bool        //
	SrcAddr         string      `ch:",lc"`
	DstAddr         string      `ch:",lc"`
	Bounce          bool        //
	Bounced         bool        //
	Amount          uint64      // TODO: uint256
	IHRDisabled     bool        //
	IHRFee          uint64      // TODO: uint256
	FwdFee          uint64      // TODO: uint256
	CreatedLT       uint64      `ch:",pk"`
	CreatedAt       uint32      `ch:",pk"`
	StateInitCode   []byte      //
	StateInitData   []byte      //
	Body            []byte      //
	BodyHash        []byte      //
	OperationID     uint32      //
	TransferComment string      //
	SourceTxHash    []byte      //
}

type MessagePayload struct {
	ch.CHModel `ch:"message_payloads,partition:contract_name,operation_id,operation_name"`

	ContractName  ContractType `ch:",lc"`
	TxHash        []byte       `ch:",pk"`
	MsgHash       []byte       `ch:",pk"`
	OperationID   uint32       `ch:",lc"`
	OperationName string       `ch:",lc"`
	DataJSON      string       //
}

type TxRepository interface {
	AddTransactions(ctx context.Context, tx []*Transaction) error
	GetTransactionByHash(ctx context.Context, txHash []byte) (*Transaction, error)
	AddMessages(ctx context.Context, m []*Message) error
	GetMessageByHash(ctx context.Context, msgHash []byte) (*Message, error)
	AddMessagePayloads(ctx context.Context, payloads []*MessagePayload) error
}
