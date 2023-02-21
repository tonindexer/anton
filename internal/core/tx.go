package core

import (
	"context"

	"github.com/uptrace/bun"
	"github.com/uptrace/go-clickhouse/ch"
	"github.com/xssnick/tonutils-go/tlb"

	"github.com/iam047801/tonidx/abi"
)

type Transaction struct {
	ch.CHModel    `ch:"transactions,partition:block_workchain,block_shard,round(block_seq_no,-5),toYYYYMMDD(toDateTime(created_at))"`
	bun.BaseModel `bun:"table:transactions"`

	Hash    []byte        `ch:",pk" bun:"type:bytea,pk,notnull"`
	Address string        `ch:",pk"`
	Account *AccountState `ch:"-" bun:"rel:has-one,join:address=address,join:created_lt=last_tx_lt"`

	BlockWorkchain int32  `bun:",notnull"`
	BlockShard     int64  `bun:",notnull"`
	BlockSeqNo     uint32 `bun:",notnull"`

	PrevTxHash []byte `bun:"type:bytea"`
	PrevTxLT   uint64 //

	InMsgHash []byte     //
	InMsg     *Message   `ch:"-" bun:"rel:belongs-to,join:in_msg_hash=hash"` // `ch:"-" bun:"rel:belongs-to,join:in_msg_hash=hash"`
	OutMsg    []*Message `ch:"-" bun:"rel:has-many,join:address=src_address,join:created_lt=source_tx_lt"`

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

	Hash []byte `ch:",pk" bun:"type:bytea,pk,notnull"`

	SrcAddress string //
	DstAddress string //
	// TODO: addr contract types

	// SourceTx initiates outgoing message
	// or contract can accept external incoming message in SourceTx
	SourceTxHash []byte       `bun:"type:bytea"`
	SourceTxLT   uint64       //
	Source       *Transaction `ch:"-" bun:"-"`

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
	ch.CHModel    `ch:"message_payloads,partition:src_contract,partition:dst_contract,partition:toYYYYMMDD(toDateTime(created_at))"`
	bun.BaseModel `bun:"table:message_payloads"`

	Type MessageType `ch:",lc" bun:"type:message_type,notnull"`
	Hash []byte      `ch:",pk" bun:"type:bytea,pk,notnull"`

	SrcAddress  string           //
	SrcContract abi.ContractName `ch:",lc"`
	DstAddress  string           //
	DstContract abi.ContractName `ch:",lc"`

	BodyHash      []byte `bun:"type:bytea,notnull"`
	OperationID   uint32 `bun:",notnull"`
	OperationName string `ch:",lc" bun:",notnull"`
	DataJSON      string //

	CreatedAt uint64 `bun:",notnull"`
	CreatedLT uint64 `bun:",notnull"`
}

type TransactionFilter struct {
	Hash []byte

	Address string

	BlockID *BlockID

	WithAccountState    bool
	WithAccountData     bool
	WithMessages        bool
	WithMessagePayloads bool
}

type MessageFilter struct {
	Hash       []byte
	SrcAddress string
	DstAddress string

	WithPayload   bool
	SrcContract   string
	DstContract   string
	OperationName string
}

type TxRepository interface {
	AddTransactions(ctx context.Context, tx bun.Tx, transactions []*Transaction) error
	AddMessages(ctx context.Context, tx bun.Tx, messages []*Message) error
	AddMessagePayloads(ctx context.Context, tx bun.Tx, payloads []*MessagePayload) error
	GetTransactions(ctx context.Context, filter *TransactionFilter, offset, limit int) ([]*Transaction, error)
	GetMessages(ctx context.Context, filter *MessageFilter, offset, limit int) ([]*Message, error)
}
