package core

import (
	"context"
	"time"

	"github.com/uptrace/bun"
	"github.com/uptrace/bun/extra/bunbig"
	"github.com/uptrace/go-clickhouse/ch"

	"github.com/iam047801/tonidx/internal/addr"
)

type Transaction struct {
	ch.CHModel    `ch:"transactions,partition:toYYYYMM(created_at)" json:"-"`
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

	CreatedAt time.Time `bun:"type:timestamp without time zone,notnull" json:"created_at"`
	CreatedLT uint64    `bun:",notnull" json:"created_lt"`
}

type TransactionRepository interface {
	AddTransactions(ctx context.Context, tx bun.Tx, transactions []*Transaction) error
}
