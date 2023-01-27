package core

import (
	"context"
	"reflect"

	"github.com/uptrace/go-clickhouse/ch"
	"github.com/xssnick/tonutils-go/tlb"
)

type Transaction struct {
	ch.CHModel `ch:"transactions,partition:block_workchain,block_shard,round(block_seq_no,-5),toYYYYMMDD(toDateTime(created_at))"`

	Address string `ch:",pk"`
	Hash    []byte `ch:",pk"`

	BlockWorkchain int32  //
	BlockShard     int64  //
	BlockSeqNo     uint32 //

	PrevTxHash []byte //
	PrevTxLT   uint64 //

	InMsgBodyHash    []byte   //
	OutMsgBodyHashes [][]byte //

	TotalFees uint64 // `ch:"type:UInt256"`

	StateUpdate []byte //
	Description []byte //

	OrigStatus AccountStatus `ch:",lc"`
	EndStatus  AccountStatus `ch:",lc"`

	CreatedAT uint64 //
	CreatedLT uint64 //
}

type MessageType string

const (
	Internal    = MessageType(tlb.MsgTypeInternal)
	ExternalIn  = MessageType(tlb.MsgTypeExternalIn)
	ExternalOut = MessageType(tlb.MsgTypeExternalOut)
)

type Message struct {
	ch.CHModel `ch:"messages,partition:type,incoming,toYYYYMMDD(toDateTime(created_at))"`

	Type MessageType `ch:",lc"` // TODO: enum

	Incoming     bool   `ch:",pk"`
	TxAddress    string `ch:",pk"`
	TxHash       []byte `ch:",pk"`
	SourceTxHash []byte //

	SrcAddress string //
	DstAddress string //
	// TODO: addr contract types

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

	CreatedAt uint64 //
	CreatedLT uint64 //
}

type ContractOperation struct {
	ch.CHModel `ch:"contract_operations"`

	Name         string                //
	ContractName ContractType          `ch:",pk"`
	Outgoing     bool                  // if operation is going from contract
	OperationID  uint32                `ch:",pk"`
	Schema       string                //
	StructSchema []reflect.StructField `ch:"-"`
}

type MessagePayload struct {
	ch.CHModel `ch:"message_payloads,partition:incoming,src_contract,dst_contract,toYYYYMMDD(toDateTime(created_at))"`

	// Type MessageType `ch:",lc"` // TODO: not only incoming messages

	Incoming    bool         `ch:",pk"`
	TxAddress   string       `ch:",pk"`
	TxHash      []byte       `ch:",pk"`
	SrcAddress  string       //
	SrcContract ContractType `ch:",lc"`
	DstAddress  string       //
	DstContract ContractType `ch:",lc"`

	BodyHash []byte `ch:",pk"`

	OperationID   uint32 //
	OperationName string `ch:",lc"`
	DataJSON      string //

	CreatedAt uint64 //
	CreatedLT uint64 //
}

type TxRepository interface {
	AddTransactions(ctx context.Context, tx []*Transaction) error
	GetTransactionByHash(ctx context.Context, txHash []byte) (*Transaction, error)
	AddMessages(ctx context.Context, m []*Message) error
	GetMessageByHash(ctx context.Context, msgHash []byte) (*Message, error)
	AddMessagePayloads(ctx context.Context, payloads []*MessagePayload) error
}
