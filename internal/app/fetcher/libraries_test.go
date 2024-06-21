package fetcher

import (
	"context"
	"encoding/base64"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/xssnick/tonutils-go/tvm/cell"

	"github.com/tonindexer/anton/abi"
	"github.com/tonindexer/anton/addr"
)

func TestService_getAccountLibraries(t *testing.T) {
	s := newService(t)

	ctx := context.Background()

	m, err := s.API.GetMasterchainInfo(ctx)
	require.NoError(t, err)

	addresses := []string{
		"EQBQ1N2kXSnH1Kn-Ds9hAb975AH4ONhnQQKFxoVJQcjk-R3d",
		"EQA4p5Xw8jOXwZFq2bg50b68E8lQTSHjnorzb6eu7-Bi7Vf5",
		"EQAOtILmq55sXyBVzZcIbT8k_llRMX4cB4_CqHoOXGI5Pt_k",
		"EQDK-nuKv1Rvb9YeV84e707l0FaLujTsj-TMsrQ1jX1M71Id",
		"EQAtxlncJm6z8o9lEVDUzdhQ3xFlFnsQwoUInnlTFccbU1KN",
		"EQDynReiCeK8xlKRbYArpp4jyzZuF6-tYfhFM0O5ulOs5H0L",
	}

	for _, addrStr := range addresses {
		a := addr.MustFromBase64(addrStr)

		raw, err := s.API.GetAccount(ctx, m, a.MustToTonutils())
		require.NoError(t, err)

		_, err = s.getAccountLibraries(ctx, *a, raw)
		require.NoError(t, err)
	}
}

func TestService_getAccountLibraries_emulate(t *testing.T) {
	s := newService(t)

	ctx := context.Background()

	m, err := s.API.GetMasterchainInfo(ctx)
	require.NoError(t, err)

	addrStr := "0:38a795f0f23397c1916ad9b839d1bebc13c9504d21e39e8af36fa7aeefe062ed"

	a := addr.MustFromString(addrStr)

	raw, err := s.API.GetAccount(ctx, m, a.MustToTonutils())
	require.NoError(t, err)

	acc := MapAccount(m, raw)

	lib, err := s.getAccountLibraries(ctx, *a, raw)
	require.NoError(t, err)

	acc.Libraries = lib.ToBOC()

	codeBase64, dataBase64, librariesBase64 :=
		base64.StdEncoding.EncodeToString(acc.Code),
		base64.StdEncoding.EncodeToString(acc.Data),
		base64.StdEncoding.EncodeToString(acc.Libraries)

	e, err := abi.NewEmulatorBase64(acc.Address.MustToTonutils(), codeBase64, dataBase64, bcConfigBase64, librariesBase64)
	require.NoError(t, err)

	retValues := []abi.VmValueDesc{
		{
			Name:      "contract_data",
			StackType: abi.VmCell,
		},
	}

	retStack, err := e.RunGetMethod(ctx, "get_position_manager_contract_data", nil, retValues)
	require.NoError(t, err)
	require.Equal(t, 1, len(retStack))

	c, ok := retStack[0].Payload.(*cell.Cell)
	require.True(t, ok)
	t.Logf("%x", c.ToBOC())
}
