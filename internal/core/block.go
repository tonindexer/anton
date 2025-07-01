package core

import (
	"context"
	"time"

	"github.com/uptrace/bun"
	"github.com/uptrace/go-clickhouse/ch"
	"github.com/xssnick/tonutils-go/ton"
)

type BlockID struct {
	Workchain int32  `json:"workchain"`
	Shard     int64  `json:"shard"`
	SeqNo     uint32 `json:"seq_no"`
}

func GetBlockID(b *ton.BlockIDExt) BlockID {
	return BlockID{
		Workchain: b.Workchain,
		Shard:     b.Shard,
		SeqNo:     b.SeqNo,
	}
}

type Block struct {
	ch.CHModel    `ch:"block_info" json:"-"`
	bun.BaseModel `bun:"table:block_info" json:"-"`

	Workchain int32  `ch:",pk" bun:"type:integer,pk,notnull" json:"workchain"`
	Shard     int64  `ch:",pk" bun:"type:bigint,pk,notnull" json:"shard"`
	SeqNo     uint32 `ch:",pk" bun:"type:integer,pk,notnull" json:"seq_no"`

	FileHash []byte `bun:"type:bytea,unique,notnull" json:"file_hash"`
	RootHash []byte `bun:"type:bytea,unique,notnull" json:"root_hash"`

	MasterID *BlockID `ch:"-" bun:"embed:master_" json:"master,omitempty"`
	Shards   []*Block `ch:"-" bun:"rel:has-many,join:workchain=master_workchain,join:shard=master_shard,join:seq_no=master_seq_no" json:"shards,omitempty"`

	TransactionsCount int             `ch:"-" bun:"transactions_count,scanonly" json:"transactions_count"`
	Transactions      []*Transaction  `ch:"-" bun:"rel:has-many,join:workchain=workchain,join:shard=shard,join:seq_no=block_seq_no" json:"transactions,omitempty"`
	Accounts          []*AccountState `ch:"-" bun:"rel:has-many,join:workchain=workchain,join:shard=shard,join:seq_no=block_seq_no" json:"accounts,omitempty"`

	// TODO: block info data

	ScannedAt time.Time `bun:"type:timestamp without time zone,notnull" json:"scanned_at"`
}

func (b *Block) ID() BlockID {
	return BlockID{
		Workchain: b.Workchain,
		Shard:     b.Shard,
		SeqNo:     b.SeqNo,
	}
}

type BlockRepository interface {
	AddBlocks(ctx context.Context, tx bun.Tx, info []*Block) error
	GetLastMasterBlock(ctx context.Context) (*Block, error)
	CountMasterBlocks(ctx context.Context) (int, error)
	GetMissedMasterBlocks(ctx context.Context) ([]uint32, error)
}
