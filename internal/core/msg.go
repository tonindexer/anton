package core

import (
	"context"
	"encoding/json"
	"time"

	"github.com/uptrace/bun"
	"github.com/uptrace/bun/extra/bunbig"
	"github.com/uptrace/go-clickhouse/ch"
	"github.com/xssnick/tonutils-go/tlb"

	"github.com/tonindexer/anton/abi"
	"github.com/tonindexer/anton/addr"
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

	// TODO: migrate src/dst blocks to nullable fields
	// TODO: null addresses in clickhouse

	SrcAddress    addr.Address  `ch:"type:String" bun:"type:bytea,nullzero" json:"src_address,omitempty"`
	SrcTxLT       uint64        `bun:",nullzero" json:"src_tx_lt,omitempty"`
	SrcTxHash     []byte        `ch:"-" bun:"-" json:"src_tx_hash,omitempty"`
	SrcWorkchain  int32         `bun:"type:integer,notnull" json:"src_workchain"`
	SrcShard      int64         `bun:"type:bigint,notnull" json:"src_shard"`
	SrcBlockSeqNo uint32        `bun:"type:integer,notnull" json:"src_block_seq_no"`
	SrcState      *AccountState `ch:"-" bun:"rel:has-one,join:src_address=address,join:src_workchain=workchain,join:src_shard=shard,join:src_block_seq_no=block_seq_no" json:"src_state,omitempty"`

	DstAddress    addr.Address  `ch:"type:String" bun:"type:bytea,nullzero" json:"dst_address,omitempty"`
	DstTxLT       uint64        `bun:",nullzero" json:"dst_tx_lt,omitempty"`
	DstTxHash     []byte        `ch:"-" bun:"-" json:"dst_tx_hash,omitempty"`
	DstWorkchain  int32         `bun:"type:integer,notnull" json:"dst_workchain"`
	DstShard      int64         `bun:"type:bigint,notnull" json:"dst_shard"`
	DstBlockSeqNo uint32        `bun:"type:integer,notnull" json:"dst_block_seq_no"`
	DstState      *AccountState `ch:"-" bun:"rel:has-one,join:dst_address=address,join:dst_workchain=workchain,join:dst_shard=shard,join:dst_block_seq_no=block_seq_no" json:"dst_state,omitempty"`

	Bounce  bool `bun:",notnull" json:"bounce"`
	Bounced bool `bun:",notnull" json:"bounced"`

	Amount *bunbig.Int `ch:"type:UInt256" bun:"type:numeric" json:"amount,omitempty"`

	IHRDisabled bool        `bun:",notnull" json:"ihr_disabled"`
	IHRFee      *bunbig.Int `ch:"type:UInt256" bun:"type:numeric" json:"ihr_fee"`
	FwdFee      *bunbig.Int `ch:"type:UInt256" bun:"type:numeric" json:"fwd_fee"`

	Body            []byte `bun:"type:bytea" json:"body"`
	BodyHash        []byte `bun:"type:bytea" json:"body_hash"`
	OperationID     uint32 `json:"operation_id"`
	TransferComment string `json:"transfer_comment,omitempty"`

	StateInitCode []byte `bun:"type:bytea" json:"state_init_code,omitempty"`
	StateInitData []byte `bun:"type:bytea" json:"state_init_data,omitempty"`

	SrcContract abi.ContractName `ch:",lc" bun:",nullzero" json:"src_contract,omitempty"`
	DstContract abi.ContractName `ch:",lc" bun:",nullzero" json:"dst_contract,omitempty"`

	OperationName string          `ch:",lc" bun:",nullzero" json:"operation_name,omitempty"`
	DataJSON      json.RawMessage `ch:"type:String" bun:"type:jsonb" json:"data,omitempty"`
	Error         string          `json:"error,omitempty"`

	CreatedAt time.Time `bun:"type:timestamp without time zone,notnull" json:"created_at"`
	CreatedLT uint64    `bun:",notnull" json:"created_lt"`
}

type MessageRepository interface {
	AddMessages(ctx context.Context, tx bun.Tx, messages []*Message) error
	UpdateMessages(ctx context.Context, messages []*Message) error

	GetMessages(ctx context.Context, hash [][]byte) ([]*Message, error)

	// MatchMessagesByOperationDesc returns hashes of suitable messages for the given contract operation.
	MatchMessagesByOperationDesc(ctx context.Context,
		contractName abi.ContractName,
		msgType MessageType,
		outgoing bool,
		operationId uint32,
		afterAddress *addr.Address,
		afterTxLT uint64,
		limit int) ([][]byte, error)
}
