package rndm

import (
	"encoding/json"
	"math/rand"
	"time"

	"github.com/tonindexer/anton/addr"
	"github.com/tonindexer/anton/internal/core"
)

var (
	// operationNames        = []string{"nft_item_transfer", "nft_collection_item_mint"}
	msgLT uint64 = 1000
	msgTS        = time.Now().UTC()
)

// func OperationName() string {
// 	return operationNames[int(rand.Uint32())%len(operationNames)]
// }

func MessageFromTo(from, to *addr.Address) *core.Message {
	msgLT++
	msgTS = msgTS.Add(time.Minute)

	src, dst := Block(0), Block(0)

	return &core.Message{
		Type:            core.Internal,
		Hash:            Bytes(32),
		SrcAddress:      *from,
		SrcWorkchain:    src.Workchain,
		SrcShard:        src.Shard,
		SrcBlockSeqNo:   src.SeqNo,
		SrcTxLT:         msgLT,
		DstAddress:      *to,
		DstWorkchain:    dst.Workchain,
		DstShard:        dst.Shard,
		DstBlockSeqNo:   dst.SeqNo,
		DstTxLT:         msgLT,
		Amount:          BigInt(),
		IHRFee:          BigInt(),
		FwdFee:          BigInt(),
		Body:            Bytes(256),
		BodyHash:        Bytes(32),
		OperationID:     rand.Uint32(),
		TransferComment: String(8),
		DataJSON:        json.RawMessage(`{}`),
		StateInitCode:   Bytes(64),
		StateInitData:   Bytes(64),
		CreatedAt:       msgTS,
		CreatedLT:       msgLT,
	}
}

func MessageTo(to *addr.Address) *core.Message {
	return MessageFromTo(Address(), to)
}

func MessageFrom(from *addr.Address) *core.Message {
	return MessageFromTo(from, Address())
}

func Message() *core.Message {
	return MessageFromTo(Address(), Address())
}

func MessagesFrom(from *addr.Address, n int) (ret []*core.Message) {
	for i := 0; i < n; i++ {
		ret = append(ret, MessageFrom(from))
	}
	return
}

func MessagesTo(to *addr.Address, n int) (ret []*core.Message) {
	for i := 0; i < n; i++ {
		ret = append(ret, MessageTo(to))
	}
	return
}

func Messages(n int) (ret []*core.Message) {
	for i := 0; i < n; i++ {
		ret = append(ret, Message())
	}
	return
}
