package parser

import (
	"encoding/hex"
	"testing"

	"github.com/xssnick/tonutils-go/address"

	"github.com/iam047801/tonidx/abi"
)

type txTestCase struct {
	addr       *address.Address
	txHash     []byte
	lt         uint64
	comment    string
	commentOut string
	opId       uint32
	opIdOut    uint32
	// contract   abi.ContractName
	// error      error
}

func mustHexDecode(h string) []byte {
	ret, err := hex.DecodeString(h)
	if err != nil {
		panic(err)
	}
	return ret
}

func TestService_ParseOperationID(t *testing.T) {
	cases := []*txTestCase{
		{
			addr:       address.MustParseAddr("EQDd3NPNrWCvTA1pOJ9WetUdDCY_pJaNZVq0JMaara-TIp90"),
			txHash:     mustHexDecode("2c4e497a6bdcddfb72d92874fdcbbfc77e023fd9dec685aa70b54ae973d7c3b5"),
			lt:         25410982000001,
			opId:       0xeac4f808,
			opIdOut:    0,
			commentOut: "0246562d-15c7-490a-8104-eed384bdc4db",
		}, {
			addr:    address.MustParseAddr("EQBF1wmCWU2Lb_jBZalOy0mqa5MIDAzUYeav_Z0sI3CM8Okr"),
			txHash:  mustHexDecode("5fd05e9cfe02c09d2e248db424805a767719cd65b73c099463a35c0e252fb4f5"),
			opId:    1,
			opIdOut: 1,
			lt:      31199023000003,
		},
	}

	for _, c := range cases {
		tx := getTransactionOnce(t, c.addr, c.lt, c.txHash)
		if tx.IO.In != nil {
			payload := tx.IO.In.Msg.Payload().ToBOC()

			opID, comment, err := abi.ParseOperationID(payload)
			if err != nil {
				t.Fatal(err)
			}
			if opID != c.opId {
				t.Fatalf("expected: %d, got: %d", c.opId, opID)
			}
			if comment != c.comment {
				t.Fatalf("expected: %s, got: %s", c.comment, comment)
			}
		}
		for _, out := range tx.IO.Out {
			payload := out.Msg.Payload().ToBOC()
			opID, comment, err := abi.ParseOperationID(payload)
			if err != nil {
				t.Fatal(err)
			}
			if opID != c.opIdOut {
				t.Fatalf("expected: %d, got: %d", c.opIdOut, opID)
			}
			if comment != c.commentOut {
				t.Fatalf("expected: %s, got: %s", c.commentOut, comment)
			}
		}
	}
}
