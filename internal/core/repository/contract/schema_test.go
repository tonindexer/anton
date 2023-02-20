package contract

import (
	"reflect"
	"testing"

	"github.com/xssnick/tonutils-go/address"
	"github.com/xssnick/tonutils-go/tlb"

	"github.com/iam047801/tonidx/internal/core"
)

func TestMarshal(t *testing.T) {
	op := &core.ContractOperation{
		Name:         "nft_item_transfer",
		ContractName: core.NFTItem,
		OperationID:  0x5fcc3d14,
		StructSchema: []reflect.StructField{
			{Name: "OperationID", Type: reflect.TypeOf(tlb.Magic{}), Tag: `tlb:"#5fcc3d14"`},
			{Name: "QueryID", Type: reflect.TypeOf(uint64(0)), Tag: `tlb:"## 64"`},
			{Name: "NewOwner", Type: reflect.TypeOf((*address.Address)(nil)), Tag: `tlb:"addr"`},
			{Name: "ResponseDestination", Type: reflect.TypeOf((*address.Address)(nil)), Tag: `tlb:"addr"`},
			//	CustomPayload       *cell.Cell       `tlb:"maybe ^"`
			//	ForwardAmount       tlb.Coins        `tlb:"."`
			//	ForwardPayload      *cell.Cell       `tlb:"either . ^"`
		},
	}

	json, err := marshalStructSchema(op.StructSchema)
	if err != nil {
		t.Fatal(err)
	}

	t.Logf("%s", json)
}
