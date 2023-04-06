package rndm

import (
	"encoding/json"
	"math/rand"
	"reflect"
	"time"

	"github.com/iancoleman/strcase"
	"github.com/uptrace/bun"
	"github.com/uptrace/go-clickhouse/ch"

	"github.com/tonindexer/anton/abi"
	"github.com/tonindexer/anton/internal/addr"
	"github.com/tonindexer/anton/internal/core"
)

var (
	operationNames []string
	msgLT          uint64 = 1000
	msgTS                 = time.Now().UTC()
)

func initOperationNames() {
	operations := []any{
		(*abi.NFTItemTransfer)(nil), (*abi.NFTCollectionItemMint)(nil),
	}
	for _, op := range operations {
		operationNames = append(operationNames, strcase.ToSnake(reflect.TypeOf(op).Elem().Name()))
	}
}

func OperationName() string {
	if operationNames == nil {
		initOperationNames()
	}
	return operationNames[int(rand.Uint32())%len(operationNames)]
}

func MessageFromTo(from, to *addr.Address) *core.Message {
	msgLT++
	msgTS = msgTS.Add(time.Minute)

	return &core.Message{
		Type:            core.Internal,
		Hash:            Bytes(32),
		SrcAddress:      *from,
		DstAddress:      *to,
		SourceTxHash:    Bytes(32),
		SourceTxLT:      msgLT,
		Amount:          BigInt(),
		IHRFee:          BigInt(),
		FwdFee:          BigInt(),
		Body:            Bytes(256),
		BodyHash:        Bytes(32),
		OperationID:     rand.Uint32(),
		TransferComment: String(8),
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

func MessageOperationToContract(msg *core.Message, op string, to abi.ContractName) *core.MessagePayload {
	dataJSON, err := json.Marshal(&abi.NFTItemTransfer{})
	if err != nil {
		panic(err)
	}

	return &core.MessagePayload{
		CHModel:       ch.CHModel{},
		BaseModel:     bun.BaseModel{},
		Type:          msg.Type,
		Hash:          msg.Hash,
		SrcAddress:    msg.SrcAddress,
		SrcContract:   ContractNames(&msg.SrcAddress)[0],
		DstAddress:    msg.DstAddress,
		DstContract:   to,
		Amount:        msg.Amount,
		BodyHash:      msg.BodyHash,
		OperationID:   msg.OperationID,
		OperationName: op,
		DataJSON:      dataJSON,
		MinterAddress: Address(),
		CreatedAt:     msg.CreatedAt,
		CreatedLT:     msg.CreatedLT,
		Error:         String(42),
	}
}

func MessageToContract(msg *core.Message, to abi.ContractName) *core.MessagePayload {
	return MessageOperationToContract(msg, OperationName(), to)
}

func MessagesToContract(msg []*core.Message, to abi.ContractName) (ret []*core.MessagePayload) {
	for _, m := range msg {
		ret = append(ret, MessageOperationToContract(m, OperationName(), to))
	}
	return
}

func MessagePayload(msg *core.Message) *core.MessagePayload {
	return MessageToContract(msg, ContractNames(&msg.DstAddress)[0])
}

func MessagePayloads(messages []*core.Message) (ret []*core.MessagePayload) {
	for _, msg := range messages {
		ret = append(ret, MessagePayload(msg))
	}
	return
}
