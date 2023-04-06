package abi

import (
	"encoding/base64"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/xssnick/tonutils-go/tlb"
	"github.com/xssnick/tonutils-go/tvm/cell"
)

func mustBase64(t *testing.T, str string) []byte {
	ret, err := base64.StdEncoding.DecodeString(str)
	assert.Nil(t, err)
	return ret
}

func testMarshalSchema(t *testing.T, structT any, expected string) {
	raw, err := MarshalSchema(structT)
	assert.Nil(t, err)
	assert.Equal(t, expected, string(raw))

	got, err := UnmarshalSchema(raw)
	assert.Nil(t, err)

	gotRaw, err := MarshalSchema(got)
	assert.Nil(t, err)
	assert.Equal(t, raw, gotRaw)
}

func testUnmarshalSchema(t *testing.T, boc, schema, expect []byte) {
	payloadCell, err := cell.FromBOC(boc)
	assert.Nil(t, err)
	payloadSlice := payloadCell.BeginParse()

	s, err := UnmarshalSchema(schema)
	assert.Nil(t, err)

	schemaGot, err := MarshalSchema(s)
	assert.Nil(t, err)
	assert.Equal(t, schema, schemaGot)

	err = tlb.LoadFromCell(s, payloadSlice)
	assert.Nil(t, err)

	raw, err := json.Marshal(s)
	assert.Nil(t, err)
	assert.Equal(t, expect, raw)
}
