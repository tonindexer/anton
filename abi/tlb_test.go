package abi

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"reflect"
	"testing"

	"github.com/xssnick/tonutils-go/tlb"
	"github.com/xssnick/tonutils-go/tvm/cell"
)

func mustBase64(t *testing.T, str string) []byte {
	ret, err := base64.StdEncoding.DecodeString(str)
	if err != nil {
		t.Fatal(err)
	}
	return ret
}

func testMarshalSchema(t *testing.T, structT any, expected string) {
	raw, err := MarshalSchema(structT)
	if err != nil {
		t.Fatal(err)
	}
	if string(raw) != expected {
		t.Fatalf("(%s)\nexpected: %s\n     got: %s", reflect.TypeOf(structT), expected, string(raw))
	}

	got, err := UnmarshalSchema(raw)
	if err != nil {
		t.Fatal(err)
	}

	gotRaw, err := MarshalSchema(got)
	if err != nil {
		t.Fatal(err)
	}

	if !bytes.Equal(raw, gotRaw) {
		t.Fatalf("(%s)\nexpected: %s\n     got: %s", reflect.TypeOf(structT), raw, gotRaw)
	}
}

func testUnmarshalSchema(t *testing.T, boc, schema, expect []byte) {
	payloadCell, err := cell.FromBOC(boc)
	if err != nil {
		t.Fatal(err)
	}
	payloadSlice := payloadCell.BeginParse()

	s, err := UnmarshalSchema(schema)
	if err != nil {
		t.Fatal(err)
	}

	schemaGot, err := MarshalSchema(s)
	if err != nil {
		t.Fatal(err)
	}

	if !bytes.Equal(schemaGot, schema) {
		t.Fatalf("unmarshalled and marshalled schema is different\nexpected: %s\ngot: %s", schema, schemaGot)
	}

	if err = tlb.LoadFromCell(s, payloadSlice); err != nil {
		t.Fatal(err)
	}

	raw, err := json.Marshal(s)
	if err != nil {
		t.Fatal(err)
	}

	if !bytes.Equal(raw, expect) {
		t.Fatalf("expected: \t%s\ngot: \t\t%s", expect, raw)
	}
}
