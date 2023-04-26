package abi_test

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/tonindexer/anton/abi"
)

func makeOperationDesc(t *testing.T, x any) string {
	d, err := abi.NewOperationDesc(x)
	assert.Nil(t, err)

	n, err := d.New()
	assert.Nil(t, err)

	nd, err := abi.NewOperationDesc(n)
	nd.Name = d.Name
	assert.Nil(t, err)
	assert.Equal(t, d, nd)

	j, err := json.Marshal(nd)
	assert.Nil(t, err)

	return string(j)
}
