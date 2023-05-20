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

	// TODO: add constraints on tx lt
	SrcAddress addr.Address  `ch:"type:String" bun:"type:bytea,nullzero" json:"src_address,omitempty"`
	SrcTxLT    uint64        `json:"src_tx_lt,omitempty"`
	SrcState   *AccountState `ch:"-" bun:"rel:has-one,join:address=address,join:src_tx_lt=last_tx_lt" json:"src_state"`
	DstAddress addr.Address  `ch:"type:String" bun:"type:bytea,nullzero" json:"dst_address,omitempty"`
	DstTxLT    uint64        `json:"dst_tx_lt,omitempty"`
	DstState   *AccountState `ch:"-" bun:"rel:has-one,join:address=address,join:dst_tx_lt=last_tx_lt" json:"dst_state"`

	Bounce  bool `bun:",notnull" json:"bounce"`
	Bounced bool `bun:",notnull" json:"bounced"`

	Amount *bunbig.Int `ch:"type:UInt256" bun:"type:numeric" json:"amount,omitempty"`

	IHRDisabled bool        `bun:",notnull" json:"ihr_disabled"`
	IHRFee      *bunbig.Int `ch:"type:UInt256" bun:"type:numeric" json:"ihr_fee"`
	FwdFee      *bunbig.Int `ch:"type:UInt256" bun:"type:numeric" json:"fwd_fee"`

	Body            []byte `bun:"type:bytea" json:"body"`
	BodyHash        []byte `bun:"type:bytea" json:"body_hash"`
	OperationID     uint32 `json:"operation_id,omitempty"`
	TransferComment string `json:"transfer_comment,omitempty"`

	StateInitCode []byte `bun:"type:bytea" json:"state_init_code,omitempty"`
	StateInitData []byte `bun:"type:bytea" json:"state_init_data,omitempty"`

	SrcContract abi.ContractName `ch:",lc" json:"src_contract,omitempty"`
	DstContract abi.ContractName `ch:",lc" json:"dst_contract,omitempty"`

	// can be used to show all jetton or nft item transfers
	MinterAddress *addr.Address `ch:"type:String" bun:"type:bytea" json:"minter_address,omitempty"`

	OperationName string          `ch:",lc" bun:",notnull" json:"operation_name"`
	DataJSON      json.RawMessage `ch:"ch:type:JSON" bun:"type:jsonb" json:"data"`
	Error         string          `json:"error,omitempty"`

	CreatedAt time.Time `bun:"type:timestamp without time zone,notnull" json:"created_at"`
	CreatedLT uint64    `bun:",notnull" json:"created_lt"`
}

type MessageRepository interface {
	AddMessages(ctx context.Context, tx bun.Tx, messages []*Message) error
}
