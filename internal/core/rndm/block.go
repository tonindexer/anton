package rndm

import (
	"github.com/tonindexer/anton/internal/core"
)

var (
	seqNo uint32 = 100000
)

func Block(workchain int32) *core.Block {
	seqNo++

	return &core.Block{
		Workchain: workchain,
		Shard:     -9223372036854775808,
		SeqNo:     seqNo,
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
