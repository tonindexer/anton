package abi_test

import (
	"encoding/base64"
	"testing"

	"github.com/stretchr/testify/require"
)

func mustBase64(t *testing.T, str string) []byte {
	ret, err := base64.StdEncoding.DecodeString(str)
	require.Nil(t, err)
	return ret
}
