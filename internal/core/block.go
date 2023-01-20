package core

import (
	"context"

	"github.com/uptrace/go-clickhouse/ch"
)

type MasterBlockInfo struct {
	ch.CHModel `ch:"master_block_info"`

	Workchain       int32  `ch:",pk"`
	Shard           int64  `ch:",pk"`
	SeqNo           uint32 `ch:",pk"`
	RootHash        []byte `ch:",pk"`
	FileHash        []byte `ch:",pk"`
	ShardFileHashes [][]byte
}

type ShardBlockInfo struct {
	ch.CHModel `ch:"shards_block_info,partition:workchain,shard"`

	Workchain      int32  `ch:",pk"`
	Shard          int64  `ch:",pk"`
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
