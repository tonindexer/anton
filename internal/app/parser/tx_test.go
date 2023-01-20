package parser

import (
	"encoding/hex"
	"testing"

	"github.com/pkg/errors"
	"github.com/xssnick/tonutils-go/address"
	"github.com/xssnick/tonutils-go/tlb"

	"github.com/iam047801/tonidx/internal/core"
)

type txTestCase struct {
	addr     *address.Address
	txHash   []byte
	lt       uint64
	opId     uint32
	contract core.ContractType
	error    error
}

func TestService_parseOperation(t *testing.T) {
	cases := []*txTestCase{
		{
			addr:   address.MustParseAddr("EQDd3NPNrWCvTA1pOJ9WetUdDCY_pJaNZVq0JMaara-TIp90"),
			txHash: core.MustHexDecode("2c4e497a6bdcddfb72d92874fdcbbfc77e023fd9dec685aa70b54ae973d7c3b5"),
			lt:     25410982000001,
		}, {
			addr:   address.MustParseAddr("EQBF1wmCWU2Lb_jBZalOy0mqa5MIDAzUYeav_Z0sI3CM8Okr"),
			txHash: core.MustHexDecode("5fd05e9cfe02c09d2e248db424805a767719cd65b73c099463a35c0e252fb4f5"),
			lt:     31199023000003,
		},
	}

	for _, c := range cases {
		tx := getTransactionOnce(t, c.addr, c.lt, c.txHash)
		if tx.IO.In != nil {
			payload := tx.IO.In.Msg.Payload().ToBOC()
			msg := &core.Message{Body: payload}
			if err := testService(t).parseOperation(msg); err != nil {
				t.Fatal(err)
			}
			t.Logf("in msg payload: op = %x, comment = %s, %x (%d)", msg.OperationID, msg.TransferComment, payload, len(payload))
		}
		for i, out := range tx.IO.Out {
			payload := out.Msg.Payload().ToBOC()
			msg := &core.Message{Body: payload}
			if err := testService(t).parseOperation(msg); err != nil {
				t.Fatal(err)
			}
			t.Logf("[%d] out msg payload: op = %x, comment = %s, %x (%d)", i, msg.OperationID, msg.TransferComment, payload, len(payload))
		}
	}
}

func TestService_ParseBlockMessages(t *testing.T) {
	addr := address.MustParseAddr("EQBd9yllUaY2RSxVME0OG73RzJ_OOwOfaqrqrEDceOSUuuan")
	txHash, _ := hex.DecodeString("ac865cf93192f72709c63bbb3d5d64c3c84fc82643db9f8ed72a18453311c1ac")
	lt := uint64(31177094000001)

	tx := getTransactionOnce(t, addr, lt, txHash)
	if tx.IO.In != nil {
		payload := tx.IO.In.Msg.Payload().ToBOC()
		t.Logf("in msg payload: %x (%d)", payload, len(payload))
	}
	for i, out := range tx.IO.Out {
		payload := out.Msg.Payload().ToBOC()
		t.Logf("[%d] out msg payload: %x (%d)", i, payload, len(payload))
	}

	messages, err := testService(t).ParseBlockMessages(ctx, getCurrentMaster(t), []*tlb.Transaction{tx})
	if err != nil {
		t.Fatal(err)
	}
	for _, msg := range messages {
		t.Logf("%+v", msg)
	}
}

func TestService_ParseMessagePayload(t *testing.T) {
	cases := []*txTestCase{
		{
			addr:     address.MustParseAddr("EQBF1wmCWU2Lb_jBZalOy0mqa5MIDAzUYeav_Z0sI3CM8Okr"),
			txHash:   core.MustHexDecode("5fd05e9cfe02c09d2e248db424805a767719cd65b73c099463a35c0e252fb4f5"),
			lt:       31199023000003,
			opId:     1,
			contract: core.NFTCollection,
			error:    core.ErrNotAvailable,
		}, {
			addr:     address.MustParseAddr("EQDdjoU_zWiyHdRW8U3NHguZt_dMvUCChJ4JHCYK0PSJD2FT"),
			txHash:   core.MustHexDecode("30688324e25f16da78e1ab1d82c384c4b75160cdfc57d620ba95c6d63ac47ea9"),
			lt:       31177444000003,
			opId:     0x5FCC3D14,
			contract: core.NFTItem,
		},
	}

	for _, c := range cases {
		tx := getTransactionOnce(t, c.addr, c.lt, c.txHash)
		payload := tx.IO.In.Msg.Payload().ToBOC()
		acc := &core.Account{Types: []string{string(c.contract)}}
		msg := &core.Message{Body: payload, OperationID: c.opId}

		parsed, err := testService(t).ParseMessagePayload(ctx, nil, acc, msg)
		if err != nil && !errors.Is(err, c.error) {
			t.Fatal(err)
		}
		t.Logf("in msg payload: %+v", parsed)
	}
}
