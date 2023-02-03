package core

import (
	"context"

	"github.com/uptrace/bun"
	"github.com/uptrace/go-clickhouse/ch"
)

type BlockID struct {
	Workchain int32
	Shard     int64
	SeqNo     uint32
}

type Block struct {
	ch.CHModel    `ch:"block_info,partition:workchain,shard"`
	bun.BaseModel `bun:"table:block_info"`

	Workchain int32  `ch:",pk" bun:",pk,notnull"`
	Shard     int64  `ch:",pk" bun:",pk,notnull"`
	SeqNo     uint32 `ch:",pk" bun:",pk,notnull"`
	FileHash  []byte `ch:",pk" bun:",unique,notnull"`
	RootHash  []byte `ch:",pk" bun:",unique,notnull"`

	MasterBlockID *BlockID   `bun:"master_block_id,type:text"`
	ShardBlockIDs []*BlockID `bun:"shard_block_ids,type:text[]"`

	Transactions []*Transaction `ch:"-" bun:"rel:has-many,join:workchain=block_workchain,join:shard=block_shard,join:seq_no=block_seq_no"`

	// TODO: block info data
}

type BlockFilter struct {
	ID        *BlockID
	Workchain *int32
	FileHash  []byte
}

type BlockRepository interface {
	AddBlocks(ctx context.Context, info []*Block) error
	GetLastMasterBlock(ctx context.Context) (*Block, error)
	GetBlocks(ctx context.Context, filter *BlockFilter, offset, limit int) ([]*Block, error)
	GetBlocksTransactions(ctx context.Context, filter *BlockFilter, offset, limit int) ([]*Block, error)
}
