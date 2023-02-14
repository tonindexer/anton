package core

import (
	"context"

	"github.com/uptrace/bun"
	"github.com/uptrace/go-clickhouse/ch"
	"github.com/xssnick/tonutils-go/tlb"
)

type Transaction struct {
	ch.CHModel    `ch:"transactions,partition:block_workchain,block_shard,round(block_seq_no,-5),toYYYYMMDD(toDateTime(created_at))"`
	bun.BaseModel `bun:"table:transactions"`

	Hash    []byte        `ch:",pk" bun:"type:bytea,pk,notnull"`
	Address string        `ch:",pk"`
	Account *AccountState `ch:"-" bun:"rel:has-one,join:address=address,join:hash=last_tx_hash"`

	BlockWorkchain int32  `bun:",notnull"`
	BlockShard     int64  `bun:",notnull"`
	BlockSeqNo     uint32 `bun:",notnull"`
	BlockFileHash  []byte `bun:"type:bytea,notnull"`

	PrevTxHash []byte `bun:"type:bytea"`
	PrevTxLT   uint64 //

	InMsgHash []byte     `bun:"type:bytea"`
	InMsg     *Message   `ch:"-" bun:"rel:belongs-to,join:in_msg_hash=hash"`
	OutMsg    []*Message `ch:"-" bun:"rel:has-many,join:hash=tx_hash"`

	TotalFees uint64 // `ch:"type:UInt256"`

	StateUpdate []byte `bun:"type:bytea"`
	Description []byte `bun:"type:bytea"` // TODO: parse it (exit code, etc)

	OrigStatus AccountStatus `ch:",lc" bun:"type:account_status,notnull"`
	EndStatus  AccountStatus `ch:",lc" bun:"type:account_status,notnull"`

	CreatedAt uint64 `bun:",notnull"`
	CreatedLT uint64 `bun:",notnull"`
}

type MessageType string

const (
	Internal    = MessageType(tlb.MsgTypeInternal)
	ExternalIn  = MessageType(tlb.MsgTypeExternalIn)
	ExternalOut = MessageType(tlb.MsgTypeExternalOut)
)

type Message struct {
	ch.CHModel    `ch:"messages,partition:type,incoming,toYYYYMMDD(toDateTime(created_at))"`
	bun.BaseModel `bun:"table:messages"`

	Type MessageType `ch:",lc" bun:"type:message_type,notnull"` // TODO: ch enum

	Hash         []byte       `bun:"type:bytea,pk,notnull"`
	SourceTxHash []byte       `bun:"type:bytea"`
	SourceTx     *Transaction `bun:"rel:has-one,join:source_tx_hash=hash"`

	Incoming  bool   `ch:",pk" bun:",pk,notnull"`
	TxHash    []byte `ch:",pk" bun:"type:bytea,notnull"`
	TxAddress string `ch:",pk" bun:",notnull"`

	SrcAddress string //
	DstAddress string //
	// TODO: addr contract types

	Bounce  bool //
	Bounced bool //

	Amount uint64 // TODO: uint256

	IHRDisabled bool   //
	IHRFee      uint64 // TODO: uint256
	FwdFee      uint64 // TODO: uint256

	Body            []byte          `bun:"type:bytea"`
	BodyHash        []byte          `bun:"type:bytea"`
	OperationID     uint32          //
	TransferComment string          //
	Payload         *MessagePayload `ch:"-" bun:"rel:belongs-to,join:hash=hash"`

	StateInitCode []byte `bun:"type:bytea"`
	StateInitData []byte `bun:"type:bytea"`

	CreatedAt uint64 `bun:",notnull"`
	CreatedLT uint64 `bun:",notnull"`
}

type MessagePayload struct {
	ch.CHModel    `ch:"message_payloads,partition:incoming,src_contract,dst_contract,toYYYYMMDD(toDateTime(created_at))"`
	bun.BaseModel `bun:"table:message_payloads"`

	Type MessageType `ch:",lc" bun:"type:message_type,notnull"`
	Hash []byte      `bun:"type:bytea,pk,notnull"`

	Incoming  bool   `ch:",pk" bun:",pk,notnull"`
	TxAddress string `ch:",pk" bun:",notnull"`
	TxHash    []byte `ch:",pk" bun:"type:bytea,notnull"`

	SrcAddress  string       //
	SrcContract ContractType `ch:",lc"`
	DstAddress  string       //
	DstContract ContractType `ch:",lc"`

	BodyHash      []byte `ch:",pk" bun:"type:bytea,notnull"`
	OperationID   uint32 `bun:",notnull"`
	OperationName string `ch:",lc" bun:",notnull"`
	DataJSON      string //

	CreatedAt uint64 `bun:",notnull"`
	CreatedLT uint64 `bun:",notnull"`
}

type TransactionFilter struct {
	Hash []byte

	Address string

	BlockID       *BlockID
	BlockFileHash []byte

	WithMessages bool
}

type MessageFilter struct {
	Hash       []byte
	SrcAddress string
	DstAddress string

	WithSource bool

	WithPayload   bool
	SrcContract   string
	DstContract   string
	OperationName string
}

type TxRepository interface {
	AddTransactions(ctx context.Context, tx bun.Tx, transactions []*Transaction) error
	AddMessages(ctx context.Context, tx bun.Tx, messages []*Message) error
	AddMessagePayloads(ctx context.Context, tx bun.Tx, payloads []*MessagePayload) error
	GetSourceMessageTxHash(ctx context.Context, from, to string, lt uint64) (ret []byte, err error)
	GetTransactions(ctx context.Context, filter *TransactionFilter, offset, limit int) ([]*Transaction, error)
	GetMessages(ctx context.Context, filter *MessageFilter, offset, limit int) ([]*Message, error)
}
