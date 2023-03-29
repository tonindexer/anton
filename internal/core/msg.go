package core

import (
	"context"
	"encoding/json"
	"time"

	"github.com/uptrace/bun"
	"github.com/uptrace/bun/extra/bunbig"
	"github.com/uptrace/go-clickhouse/ch"
	"github.com/xssnick/tonutils-go/tlb"

	"github.com/iam047801/tonidx/abi"
	"github.com/iam047801/tonidx/internal/addr"
)

type MessageType string

const (
	Internal    = MessageType(tlb.MsgTypeInternal)
	ExternalIn  = MessageType(tlb.MsgTypeExternalIn)
	ExternalOut = MessageType(tlb.MsgTypeExternalOut)
)

type Message struct {
	ch.CHModel    `ch:"messages,partition:toYYYYMM(created_at)" json:"-"`
	bun.BaseModel `bun:"table:messages" json:"-"`

	Type MessageType `ch:",lc" bun:"type:message_type,notnull" json:"type"` // TODO: ch enum

	Hash []byte `ch:",pk" bun:"type:bytea,pk,notnull"  json:"hash"`

	SrcAddress addr.Address `ch:"type:String" bun:"type:bytea,nullzero" json:"src_address,omitempty"`
	DstAddress addr.Address `ch:"type:String" bun:"type:bytea,nullzero" json:"dst_address,omitempty"`

	// SourceTx initiates outgoing message.
	// For external incoming messages SourceTx == nil.
	SourceTxHash []byte       `bun:"type:bytea" json:"source_tx_hash,omitempty"`
	SourceTxLT   uint64       `json:"source_tx_lt,omitempty"`
	Source       *Transaction `ch:"-" bun:"-" json:"source,omitempty"` // TODO: join it

	Bounce  bool `bun:",notnull" json:"bounce"`
	Bounced bool `bun:",notnull" json:"bounced"`

	Amount *bunbig.Int `ch:"type:UInt256" bun:"type:numeric" json:"amount,omitempty"`

	IHRDisabled bool        `bun:",notnull" json:"ihr_disabled"`
	IHRFee      *bunbig.Int `ch:"type:UInt256" bun:"type:numeric" json:"ihr_fee"`
	FwdFee      *bunbig.Int `ch:"type:UInt256" bun:"type:numeric" json:"fwd_fee"`

	Body            []byte          `bun:"type:bytea" json:"body"`
	BodyHash        []byte          `bun:"type:bytea" json:"body_hash"`
	OperationID     uint32          `json:"operation_id,omitempty"`
	TransferComment string          `json:"transfer_comment,omitempty"`
	Payload         *MessagePayload `ch:"-" bun:"rel:belongs-to,join:hash=hash" json:"payload,omitempty"`

	StateInitCode []byte `bun:"type:bytea" json:"state_init_code,omitempty"`
	StateInitData []byte `bun:"type:bytea" json:"state_init_data,omitempty"`

	CreatedAt time.Time `bun:"type:timestamp without time zone,notnull" json:"created_at"`
	CreatedLT uint64    `bun:",notnull" json:"created_lt"`

	Known bool `ch:"-" bun:"-" json:"-"`
}

type MessagePayload struct {
	ch.CHModel    `ch:"message_payloads,partition:toYYYYMM(created_at)" json:"-"`
	bun.BaseModel `bun:"table:message_payloads" json:"-"`

	Type MessageType `ch:",lc" bun:"type:message_type,notnull" json:"type"`
	Hash []byte      `ch:",pk" bun:"type:bytea,pk,notnull" json:"hash"`

	SrcAddress  addr.Address     `ch:"type:String" bun:"type:bytea,nullzero" json:"src_address,omitempty"`
	SrcContract abi.ContractName `ch:",lc" json:"src_contract,omitempty"`
	DstAddress  addr.Address     `ch:"type:String" bun:"type:bytea,nullzero" json:"dst_address,omitempty"`
	DstContract abi.ContractName `ch:",lc" json:"dst_contract,omitempty"`

	Amount *bunbig.Int `ch:"type:UInt256" bun:"type:numeric" json:"amount,omitempty"`

	BodyHash      []byte          `bun:"type:bytea,notnull" json:"body_hash"`
	OperationID   uint32          `bun:",notnull" json:"operation_id"`
	OperationName string          `ch:",lc" bun:",notnull" json:"operation_name"`
	DataJSON      json.RawMessage `ch:"-" bun:"type:jsonb" json:"data"` // TODO: https://github.com/uptrace/go-clickhouse/issues/22

	// TODO: save fields from parsed data to payloads table

	// can be used to show all jetton or nft item transfers
	MinterAddress *addr.Address `ch:"type:String" bun:"type:bytea" json:"minter_address,omitempty"`

	CreatedAt time.Time `bun:"type:timestamp without time zone,notnull" json:"created_at"`
	CreatedLT uint64    `bun:",notnull" json:"created_lt"`

	Error string `json:"error,omitempty"`
}

type MessageRepository interface {
	AddMessages(ctx context.Context, tx bun.Tx, messages []*Message) error
	AddMessagePayloads(ctx context.Context, tx bun.Tx, payloads []*MessagePayload) error
}
