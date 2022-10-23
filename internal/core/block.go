package core

import (
	"context"

	"github.com/uptrace/go-clickhouse/ch"
)

type MasterBlockInfo struct {
	ch.CHModel `ch:"master_block_info"`

	Workchain int32
	Shard     int64
	SeqNo     uint32 `ch:",pk"`
	RootHash  []byte `ch:",pk"`
	FileHash  []byte `ch:",pk"`
}

type ShardBlockInfo struct {
	ch.CHModel `ch:"shards_block_info,partition:shard"`

	Workchain      int32
	Shard          int64
	SeqNo          uint32 `ch:",pk"`
	RootHash       []byte `ch:",pk"`
	FileHash       []byte `ch:",pk"`
	MasterFileHash []byte `ch:",pk"`
}

// TODO: block data

type BlockRepository interface {
	AddMasterBlockInfo(ctx context.Context, info *MasterBlockInfo) error
	GetLastMasterBlockInfo(ctx context.Context) (*MasterBlockInfo, error)
	AddShardBlocksInfo(ctx context.Context, info []*ShardBlockInfo) error
}
