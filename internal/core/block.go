package core

import (
	"context"

	"github.com/uptrace/go-clickhouse/ch"
)

type BlockID struct {
	Workchain int32
	Shard     int64
	SeqNo     uint32
}

type BlockInfo struct {
	ch.CHModel `ch:"block_info,partition:workchain,shard"`

	Workchain     int32  `ch:",pk"`
	Shard         int64  `ch:",pk"`
	SeqNo         uint32 `ch:",pk"`
	FileHash      []byte `ch:",pk"`
	RootHash      []byte `ch:",pk"`
	MasterBlockID *BlockID
	ShardBlockIDs []*BlockID
}

// TODO: block data

type BlockRepository interface {
	GetLastMasterBlockInfo(ctx context.Context) (*BlockInfo, error)

	AddBlocksInfo(ctx context.Context, info []*BlockInfo) error
}
