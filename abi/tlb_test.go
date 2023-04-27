package abi_test

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/tonindexer/anton/abi"
)

func makeOperationDesc(t *testing.T, x any) string {
	d, err := abi.NewOperationDesc(x)
	require.Nil(t, err)

	n, err := d.New()
	require.Nil(t, err)

	nd, err := abi.NewOperationDesc(n)
	nd.Name = d.Name
	require.Nil(t, err)
	require.Equal(t, d, nd)

	j, err := json.Marshal(nd)
	require.Nil(t, err)

	return string(j)
}
