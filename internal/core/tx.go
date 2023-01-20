package core

import (
	"context"
	"reflect"

	"github.com/uptrace/go-clickhouse/ch"
	"github.com/xssnick/tonutils-go/tlb"
)

type Transaction struct {
	ch.CHModel `ch:"transactions,partition:account_addr,round(created_at,-5)"`

	Hash        []byte `ch:",pk"`
	AccountAddr string `ch:",pk"`

	PrevTxHash []byte //
	PrevTxLT   uint64 //

	InMsgHash    []byte   `ch:",pk"`
	OutMsgHashes [][]byte //

	TotalFees uint64 // `ch:"type:UInt256"`

	StateUpdate []byte //
	Description []byte //

	OrigStatus AccountStatus `ch:",lc"`
	EndStatus  AccountStatus `ch:",lc"`

	CreatedLT uint64 `ch:",pk"`
	CreatedAT uint64 `ch:",pk"`
}

type MessageType string

const (
	Internal    = MessageType(tlb.MsgTypeInternal)
	ExternalIn  = MessageType(tlb.MsgTypeExternalIn)
	ExternalOut = MessageType(tlb.MsgTypeExternalOut)
)

type Message struct {
	ch.CHModel `ch:"messages,partition:type,tx_account_addr,round(amount,-9),round(created_at,-5)"`

	Type MessageType `ch:",lc"` // TODO: enum

	TxHash        []byte `ch:",pk"`
	TxAccountAddr string `ch:",pk"` // TODO: not needed, as we have incoming flag
	SourceTxHash  []byte `ch:",pk"` // match in_msg with out_msg by body_hash

	Incoming bool   `ch:",pk"`
	SrcAddr  string `ch:",pk"`
	DstAddr  string `ch:",pk"`

	Bounce  bool //
	Bounced bool //

	Amount uint64 // TODO: uint256

	IHRDisabled bool   //
	IHRFee      uint64 // TODO: uint256
	FwdFee      uint64 // TODO: uint256

	Body            []byte //
	BodyHash        []byte `ch:",pk"`
	OperationID     uint32 //
	TransferComment string //

	StateInitCode []byte //
	StateInitData []byte //

	CreatedLT uint64 `ch:",pk"`
	CreatedAt uint64 `ch:",pk"`
}

type ContractOperation struct {
	ch.CHModel `ch:"contract_operations"`

	Name         string                //
	ContractName ContractType          `ch:",pk"`
	OperationID  uint32                `ch:",pk"`
	Schema       string                //
	StructSchema []reflect.StructField `ch:"-"`
}

type MessagePayload struct {
	ch.CHModel `ch:"message_payloads,partition:contract_name,operation_id,operation_name"`

	ContractName  ContractType `ch:",lc"`
	TxHash        []byte       `ch:",pk"`
	PayloadHash   []byte       `ch:",pk"`
	DstAddr       string       `ch:",lc"`
	OperationID   uint32       //
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
