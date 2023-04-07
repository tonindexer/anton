package rndm

import (
	"github.com/tonindexer/anton/internal/core"
)

var (
	seqNo uint32 = 100000
)

func BlockID(workchain int32) *core.BlockID {
	return &core.BlockID{
		Workchain: workchain,
		Shard:     -9223372036854775808,
		SeqNo:     seqNo,
	}
}

func Block(workchain int32) *core.Block {
	seqNo++

	id := BlockID(workchain)

	return &core.Block{
		Workchain: id.Workchain,
		Shard:     id.Shard,
		SeqNo:     id.SeqNo,
		FileHash:  Bytes(32),
		RootHash:  Bytes(32),
	}
}

func Blocks(workchain int32, n int) (ret []*core.Block) {
	for i := 0; i < n; i++ {
		ret = append(ret, Block(workchain))
	}
	return ret
}

func MasterBlock() *core.Block {
	return Block(-1)
}
