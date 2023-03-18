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

type Transaction struct {
	ch.CHModel    `ch:"transactions,partition:block_workchain,block_shard,round(block_seq_no,-5),toYYYYMMDD(toDateTime(created_at))" json:"-"`
	bun.BaseModel `bun:"table:transactions" json:"-"`

	Hash    []byte        `ch:",pk" bun:"type:bytea,pk,notnull" json:"hash"`
	Address addr.Address  `ch:"type:String,pk" bun:"type:bytea,notnull" json:"address"`
	Account *AccountState `ch:"-" bun:"rel:has-one,join:address=address,join:created_lt=last_tx_lt" json:"account"`

	BlockWorkchain int32  `bun:",notnull" json:"block_workchain"`
	BlockShard     int64  `bun:",notnull" json:"block_shard"`
	BlockSeqNo     uint32 `bun:",notnull" json:"block_seq_no"`

	PrevTxHash []byte `bun:"type:bytea" json:"prev_tx_hash,omitempty"`
	PrevTxLT   uint64 `json:"prev_tx_lt,omitempty"`

	InMsgHash []byte      `json:"in_msg_hash"`
	InMsg     *Message    `ch:"-" bun:"rel:belongs-to,join:in_msg_hash=hash" json:"in_msg"`
	InAmount  *bunbig.Int `ch:"type:UInt256" bun:"type:numeric,notnull" json:"in_amount"`

	OutMsg      []*Message  `ch:"-" bun:"rel:has-many,join:address=src_address,join:created_lt=source_tx_lt" json:"out_msg,omitempty"`
	OutMsgCount uint16      `bun:",notnull" json:"out_msg_count"`
	OutAmount   *bunbig.Int `ch:"type:UInt256" bun:"type:numeric,notnull" json:"out_amount"`

	TotalFees *bunbig.Int `ch:"type:UInt256" bun:"type:numeric" json:"total_fees"`

	StateUpdate []byte `bun:"type:bytea" json:"state_update,omitempty"`
	Description []byte `bun:"type:bytea" json:"description,omitempty"`

	OrigStatus AccountStatus `ch:",lc" bun:"type:account_status,notnull" json:"orig_status"`
	EndStatus  AccountStatus `ch:",lc" bun:"type:account_status,notnull" json:"end_status"`

	CreatedAt time.Time `bun:"type:timestamp,notnull" json:"created_at"`
	CreatedLT uint64    `bun:",notnull" json:"created_lt"`
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

	CreatedAt time.Time `bun:"type:timestamp,notnull" json:"created_at"`
	CreatedLT uint64    `bun:",notnull" json:"created_lt"`

	Known bool `ch:"-" bun:"-" json:"-"`
}

type MessagePayload struct {
	ch.CHModel    `ch:"message_payloads,partition:src_contract,partition:dst_contract,partition:toYYYYMMDD(toDateTime(created_at))" json:"-"`
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
	DataJSON      json.RawMessage `ch:"type:String" bun:"type:jsonb" json:"data"`

	// TODO: save fields from parsed data to payloads table

	CreatedAt time.Time `bun:"type:timestamp,notnull" json:"created_at"`
	CreatedLT uint64    `bun:",notnull" json:"created_lt"`

	Error string `json:"error,omitempty"`
}

type TransactionFilter struct {
	Hash []byte // `form:"hash"`

	Addresses []*addr.Address //

	Workchain *int32 `form:"workchain"`

	BlockID *BlockID

	WithAccountState    bool
	WithAccountData     bool
	WithMessages        bool
	WithMessagePayloads bool

	Order string `form:"order"` // ASC, DESC

	AfterTxLT *uint64 `form:"after"`
	Limit     int     `form:"limit"`
}

type MessageFilter struct {
	DBTx *bun.Tx

	Hash         []byte          // `form:"hash"`
	SrcAddresses []*addr.Address // `form:"src_address"`
	DstAddresses []*addr.Address // `form:"dst_address"`

	WithPayload    bool
	SrcContracts   []string `form:"src_contract"`
	DstContracts   []string `form:"dst_contract"`
	OperationNames []string `form:"operation_name"`

	Order string `form:"order"` // ASC, DESC

	AfterTxLT *uint64 `form:"after"`
	Limit     int     `form:"limit"`
}

type TxRepository interface {
	AddTransactions(ctx context.Context, tx bun.Tx, transactions []*Transaction) error
	AddMessages(ctx context.Context, tx bun.Tx, messages []*Message) error
	AddMessagePayloads(ctx context.Context, tx bun.Tx, payloads []*MessagePayload) error
	GetTransactions(ctx context.Context, filter *TransactionFilter) ([]*Transaction, error)
	GetMessages(ctx context.Context, filter *MessageFilter) ([]*Message, error)
}
