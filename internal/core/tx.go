package core

import (
	"context"

	"github.com/uptrace/bun"
	"github.com/uptrace/go-clickhouse/ch"
	"github.com/xssnick/tonutils-go/tlb"

	"github.com/iam047801/tonidx/abi"
)

type Transaction struct {
	ch.CHModel    `ch:"transactions,partition:block_workchain,block_shard,round(block_seq_no,-5),toYYYYMMDD(toDateTime(created_at))" json:"-"`
	bun.BaseModel `bun:"table:transactions" json:"-"`

	Hash    []byte        `ch:",pk" bun:"type:bytea,pk,notnull" json:"hash"`
	Address string        `ch:",pk" json:"address"`
	Account *AccountState `ch:"-" bun:"rel:has-one,join:address=address,join:created_lt=last_tx_lt" json:"account"`

	BlockWorkchain int32  `bun:",notnull" json:"block_workchain"`
	BlockShard     int64  `bun:",notnull" json:"block_shard"`
	BlockSeqNo     uint32 `bun:",notnull" json:"block_seq_no"`

	PrevTxHash []byte `bun:"type:bytea" json:"prev_tx_hash,omitempty"`
	PrevTxLT   uint64 `json:"prev_tx_lt,omitempty"`

	InMsgHash []byte     `json:"in_msg_hash"`
	InMsg     *Message   `ch:"-" bun:"rel:belongs-to,join:in_msg_hash=hash" json:"in_msg"` // `ch:"-" bun:"rel:belongs-to,join:in_msg_hash=hash"`
	OutMsg    []*Message `ch:"-" bun:"rel:has-many,join:address=src_address,join:created_lt=source_tx_lt" json:"out_msg,omitempty"`

	TotalFees uint64 `json:"total_fees"` // `ch:"type:UInt256"`

	StateUpdate []byte `bun:"type:bytea" json:"state_update"`
	Description []byte `bun:"type:bytea" json:"description"` // TODO: parse it (exit code, etc)

	OrigStatus AccountStatus `ch:",lc" bun:"type:account_status,notnull" json:"orig_status"`
	EndStatus  AccountStatus `ch:",lc" bun:"type:account_status,notnull" json:"end_status"`

	CreatedAt uint64 `bun:",notnull" json:"created_at"`
	CreatedLT uint64 `bun:",notnull" json:"created_lt"`
}

type MessageType string

const (
	Internal    = MessageType(tlb.MsgTypeInternal)
	ExternalIn  = MessageType(tlb.MsgTypeExternalIn)
	ExternalOut = MessageType(tlb.MsgTypeExternalOut)
)

type Message struct {
	ch.CHModel    `ch:"messages,partition:type,incoming,toYYYYMMDD(toDateTime(created_at))" json:"-"`
	bun.BaseModel `bun:"table:messages" json:"-"`

	Type MessageType `ch:",lc" bun:"type:message_type,notnull" json:"msg_type"` // TODO: ch enum

	Hash []byte `ch:",pk" bun:"type:bytea,pk,notnull" json:"hash"`

	SrcAddress string `json:"src_address"`
	DstAddress string `json:"dst_address"`
	// TODO: addr contract types

	// SourceTx initiates outgoing message
	// or contract can accept external incoming message in SourceTx
	SourceTxHash []byte       `bun:"type:bytea" json:"source_tx_hash,omitempty"`
	SourceTxLT   uint64       `json:"source_tx_lt,omitempty"`
	Source       *Transaction `ch:"-" bun:"-" json:"source,omitempty"`

	Bounce  bool `json:"bounce"`
	Bounced bool `json:"bounced"`

	Amount uint64 `json:"amount,omitempty"` // TODO: uint256

	IHRDisabled bool   `json:"ihr_disabled"`
	IHRFee      uint64 `json:"ihr_fee"` // TODO: uint256
	FwdFee      uint64 `json:"fwd_fee"` // TODO: uint256

	Body            []byte          `bun:"type:bytea" json:"body"`
	BodyHash        []byte          `bun:"type:bytea" json:"body_hash"`
	OperationID     uint32          `json:"operation_id,omitempty"`
	TransferComment string          `json:"transfer_comment,omitempty"`
	Payload         *MessagePayload `ch:"-" bun:"rel:belongs-to,join:hash=hash" json:"payload,omitempty"`

	StateInitCode []byte `bun:"type:bytea" json:"state_init_code,omitempty"`
	StateInitData []byte `bun:"type:bytea" json:"state_init_data,omitempty"`

	CreatedAt uint64 `bun:",notnull" json:"created_at"`
	CreatedLT uint64 `bun:",notnull" json:"created_lt"`
}

type MessagePayload struct {
	ch.CHModel    `ch:"message_payloads,partition:src_contract,partition:dst_contract,partition:toYYYYMMDD(toDateTime(created_at))" json:"-"`
	bun.BaseModel `bun:"table:message_payloads" json:"-"`

	Type MessageType `ch:",lc" bun:"type:message_type,notnull" json:"msg_type"`
	Hash []byte      `ch:",pk" bun:"type:bytea,pk,notnull" json:"hash"`

	SrcAddress  string           `json:"src_address,omitempty"`
	SrcContract abi.ContractName `ch:",lc" json:"src_contract,omitempty"`
	DstAddress  string           `json:"dst_address,omitempty"`
	DstContract abi.ContractName `ch:",lc" json:"dst_contract,omitempty"`

	BodyHash      []byte `bun:"type:bytea,notnull" json:"body_hash"`
	OperationID   uint32 `bun:",notnull" json:"operation_id"`
	OperationName string `ch:",lc" bun:",notnull" json:"operation_name"`
	DataJSON      string `json:"data_json"`

	CreatedAt uint64 `bun:",notnull" json:"created_at"`
	CreatedLT uint64 `bun:",notnull" json:"created_lt"`
}

type TransactionFilter struct {
	Hash []byte `form:"hash"`

	Address string `form:"address"`

	BlockID *BlockID

	WithAccountState    bool `form:"with_accounts"`
	WithAccountData     bool
	WithMessages        bool
	WithMessagePayloads bool
}

type MessageFilter struct {
	DBTx *bun.Tx

	Hash       []byte `form:"hash"`
	SrcAddress string `form:"src_address"`
	DstAddress string `form:"dst_address"`

	WithPayload    bool
	SrcContract    string   `form:"src_contract"`
	DstContract    string   `form:"dst_contract"`
	OperationNames []string `form:"operation_names"`
}

type TxRepository interface {
	AddTransactions(ctx context.Context, tx bun.Tx, transactions []*Transaction) error
	AddMessages(ctx context.Context, tx bun.Tx, messages []*Message) error
	AddMessagePayloads(ctx context.Context, tx bun.Tx, payloads []*MessagePayload) error
	GetTransactions(ctx context.Context, filter *TransactionFilter, offset, limit int) ([]*Transaction, error)
	GetMessages(ctx context.Context, filter *MessageFilter, offset, limit int) ([]*Message, error)
}
