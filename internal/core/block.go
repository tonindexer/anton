package core

import (
	"context"

	"github.com/uptrace/bun"
	"github.com/uptrace/go-clickhouse/ch"
)

type BlockID struct {
	Workchain int32  `json:"workchain"`
	Shard     int64  `json:"shard"`
	SeqNo     uint32 `json:"seq_no"`
}

type Block struct {
	ch.CHModel    `ch:"block_info,partition:workchain" json:"-"`
	bun.BaseModel `bun:"table:block_info" json:"-"`

	Workchain int32  `ch:",pk" bun:",pk,notnull" json:"workchain"`
	Shard     int64  `ch:",pk" bun:",pk,notnull" json:"shard"`
	SeqNo     uint32 `ch:",pk" bun:",pk,notnull" json:"seq_no"`

	FileHash []byte `ch:",pk" bun:"type:bytea,unique,notnull" json:"file_hash"`
	RootHash []byte `ch:",pk" bun:"type:bytea,unique,notnull" json:"root_hash"`

	MasterID *BlockID `ch:"-" bun:"embed:master_" json:"master,omitempty"`
	Shards   []*Block `ch:"-" bun:"rel:has-many,join:workchain=master_workchain,join:shard=master_shard,join:seq_no=master_seq_no" json:"shards,omitempty"`

	Transactions []*Transaction `ch:"-" bun:"rel:has-many,join:workchain=block_workchain,join:shard=block_shard,join:seq_no=block_seq_no" json:"transactions"`

	// TODO: block info data
}

type BlockFilter struct {
	Workchain *int32  `form:"workchain"`
	Shard     *int64  `form:"shard"`
	SeqNo     *uint32 `form:"seq_no"`
	FileHash  []byte  `form:"file_hash"`

	WithShards                     bool // TODO: array of relations as strings
	WithTransactionAccountState    bool
	WithTransactionAccountData     bool
	WithTransactions               bool `form:"with_transactions"`
	WithTransactionMessages        bool
	WithTransactionMessagePayloads bool

	Order string `form:"order"` // ASC, DESC

	AfterSeqNo *uint32 `form:"after"`
	Limit      int     `form:"limit"`
}

type BlockFiltered struct {
	Total int      `json:"total"`
	Rows  []*Block `json:"results"`
}

type BlockRepository interface {
	AddBlocks(ctx context.Context, tx bun.Tx, info []*Block) error
	GetLastMasterBlock(ctx context.Context) (*Block, error)
	GetBlocks(ctx context.Context, filter *BlockFilter) (*BlockFiltered, error)
}
