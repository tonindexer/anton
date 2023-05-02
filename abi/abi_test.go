package abi_test

import (
	"encoding/base64"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/tonindexer/anton/abi"
)

func mustBase64(t *testing.T, str string) []byte {
	ret, err := base64.StdEncoding.DecodeString(str)
	require.Nil(t, err)
	return ret
}

func TestInterfaceDesc_UnmarshalJSON(t *testing.T) {
	var d abi.InterfaceDesc

	j := []byte(`
{
    "interface_name": "telemint_nft_collection",
    "addresses": [
      "EQAOQdwdw8kGftJCSFgOErM1mBjYPe4DBPq8-AhF6vr9si5N"
    ]
}`)
	err := json.Unmarshal(j, &d)
	require.Nil(t, err)
	require.Equal(t, 1, len(d.Addresses))
	require.Equal(t, "EQAOQdwdw8kGftJCSFgOErM1mBjYPe4DBPq8-AhF6vr9si5N", d.Addresses[0].Base64())
}
