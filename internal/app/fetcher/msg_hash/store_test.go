package msg_hash

import (
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/xssnick/tonutils-go/tlb"
	"github.com/xssnick/tonutils-go/tvm/cell"
)

func TestMessageHash(t *testing.T) {
	boc, err := base64.StdEncoding.DecodeString("te6cckECBAEAAR0AA69oASMHseOMLtTOzNLikEih0deCMkOfpbEoKBf6paLY9GcjAD1oM7csliVgwVQWX2NJFMmuMwd7eKQC5qAsT7sqSWS/z6erQAYXhOIAAGrLlROtBNC3/X8bAQIDCEICuikYyJR+myWvmsG4gzV3VBc+WBL4B6PW5kKhRwlZU5UAhwCAG26EZAMiSZdaT9/nb07G7krKUuZrixwBQSSIrv7HC5zwAeNtMkLGaGxnMtFWA30ih6RtkEJcPo9X/7T7tSePPCP+AKkXjUUZAAAAAAAAAABLLQXgCAD/ZY5/Z2Ta6w8b6QBTfUEPDAQcImJeWM3qvgOeR8ZFMQAjz1KOv+3yBZDnGPDFi8WYoBUREP8nejEWbWd8wSf45cQF6MF87g==")
	require.NoError(t, err)

	msgCell, err := cell.FromBOC(boc)
	require.NoError(t, err)
	require.Equal(t, hex.EncodeToString(msgCell.Hash()), "b3cfae95c4b916758edc6c8ca9c8ae837aeb40bd9981b4d9ca094358a36a780e")

	fmt.Println(hex.EncodeToString(msgCell.Hash()))

	msg := new(tlb.InternalMessage)
	err = tlb.LoadFromCell(msg, msgCell.BeginParse())
	require.NoError(t, err)

	gotCell, err := toCell(tlb.Message{MsgType: tlb.MsgTypeInternal, Msg: msg})
	require.NoError(t, err)
	require.Equal(t, hex.EncodeToString(gotCell.Hash()), "669e35112e147e5f88768d0e7c86482b1ce292e2ad7c4fbb977c651ea059ec1b")
}
