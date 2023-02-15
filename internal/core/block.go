package core

import (
	"context"

	"github.com/uptrace/bun"
	"github.com/uptrace/go-clickhouse/ch"
)

type BlockID struct {
	Workchain int32  `ch:",pk" bun:",pk,notnull" json:"w"`
	Shard     int64  `ch:",pk" bun:",pk,notnull" json:"s"`
	SeqNo     uint32 `ch:",pk" bun:",pk,notnull" json:"n"`
}

type Block struct {
	ch.CHModel    `ch:"block_info,partition:workchain,shard"`
	bun.BaseModel `bun:"table:block_info"`

	BlockID

	FileHash []byte `ch:",pk" bun:"type:bytea,unique,notnull"` // TODO: []byte here, go-bun bug on has-many
	RootHash []byte `ch:",pk" bun:"type:bytea,unique,notnull"`

	MasterID BlockID  `ch:"-" bun:"embed:master_"`
	Shards   []*Block `ch:"-" bun:"rel:has-many,join:workchain=master_workchain,join:shard=master_shard,join:seq_no=master_seq_no"`

	Transactions []*Transaction `ch:"-" bun:"rel:has-many,join:workchain=block_workchain,join:shard=block_shard,join:seq_no=block_seq_no"`

	// TODO: block info data
}

type BlockFilter struct {
	ID        *BlockID
	Workchain *int32
	FileHash  []byte

	WithShards       bool
	WithTransactions bool
}

type BlockRepository interface {
	AddBlocks(ctx context.Context, tx bun.Tx, info []*Block) error
	GetLastMasterBlock(ctx context.Context) (*Block, error)
	GetBlocks(ctx context.Context, filter *BlockFilter, offset, limit int) ([]*Block, error)
}
