package parser

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"math/big"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/xssnick/tonutils-go/tvm/cell"

	"github.com/tonindexer/anton/abi"
	"github.com/tonindexer/anton/addr"
	"github.com/tonindexer/anton/internal/core"
)

func mustFromB64(t *testing.T, b64 string) []byte {
	s, err := base64.StdEncoding.DecodeString(b64)
	require.NoError(t, err)
	return s
}

func mustFromBOC(t *testing.T, b64 string) *cell.Cell {
	s, err := base64.StdEncoding.DecodeString(b64)
	require.NoError(t, err)
	c, err := cell.FromBOC(s)
	require.NoError(t, err)
	return c
}

func TestService_checkPrevGetMethodExecution(t *testing.T) {
	var testCases = []struct {
		contract     abi.ContractName
		descJson     string
		executedJson string
		args         []any
		result       bool
	}{
		{
			contract: "nft_collection",
			descJson: `
{
	"name": "get_nft_content",
	"arguments": [
		{
			"name": "index",
			"stack_type": "int"
		},
		{
			"name": "individual_content",
			"stack_type": "cell"
		}
	],
	"return_values": [
		{
			"name": "full_content",
			"stack_type": "cell",
			"format": "content"
		}
	]
}`,
			executedJson: `
{
	"name": "get_nft_content",
	"arguments": [
		{
			"name": "index",
			"stack_type": "int"
		},
		{
			"name": "individual_content",
			"stack_type": "cell"
		}
	],
	"receives": [
		1.11966e+5,
		"te6cckEBAQEAMwAAYgFodHRwczovL25mdC5mcmFnbWVudC5jb20vbnVtYmVyLzg4ODA4MTk1NzU0Lmpzb26DXfO0"
	],
	"return_values": [
		{
			"name": "full_content",
			"stack_type": "cell",
			"format": "content"
		}
	],
	"returns": [
		{
			"URI": "https://nft.fragment.com/number/88808195754.json"
		}
	]
}`,
			args:   []any{big.NewInt(111966), mustFromBOC(t, "te6cckEBAQEAMwAAYgFodHRwczovL25mdC5mcmFnbWVudC5jb20vbnVtYmVyLzg4ODA4MTk1NzU0Lmpzb26DXfO0")},
			result: true,
		},
		{
			contract: "nft_collection",
			descJson: `
{
	"name": "get_nft_address_by_index",
	"arguments": [
		{
			"name": "index",
			"stack_type": "int",
			"format": "bytes"
		}
	],
	"return_values": [
		{
			"name": "address",
			"stack_type": "slice",
			"format": "addr"
		}
	]
}`,
			executedJson: `
{
	"name": "get_nft_address_by_index",
	"arguments": [
		{
			"name": "index",
			"stack_type": "int"
		}
	],
	"receives": [
		10
	],
	"return_values": [
		{
			"name": "address",
			"stack_type": "slice",
			"format": "addr"
		}
	],
	"returns": [
		"EQDHVwNhkIvqS3tJf0ScpM2kGd0Yi0PgGf_lZ1Vh0m7AyWD3"
	]
}`,
			args:   []any{big.NewInt(10)},
			result: false,
		},
		{
			contract: "nft_collection",
			descJson: `
{
	"name": "get_nft_address_by_index",
	"arguments": [
		{
			"name": "index",
			"stack_type": "int"
		}
	],
	"return_values": [
		{
			"name": "address",
			"stack_type": "slice",
			"format": "addr"
		}
	]
}`,
			executedJson: `
{
	"name": "get_nft_address_by_index",
	"arguments": [
		{
			"name": "index",
			"stack_type": "int"
		}
	],
	"receives": [
		10
	],
	"return_values": [
		{
			"name": "address",
			"stack_type": "slice",
			"format": "addr"
		}
	],
	"returns": [
		"EQDHVwNhkIvqS3tJf0ScpM2kGd0Yi0PgGf_lZ1Vh0m7AyWD3"
	]
}`,
			args:   []any{big.NewInt(10)},
			result: true,
		},
		{
			contract: "nft_collection",
			descJson: `
{
	"name": "get_nft_address_by_index",
	"arguments": [
		{
			"name": "index",
			"stack_type": "int",
			"format": "bytes"
		}
	],
	"return_values": [
		{
			"name": "address",
			"stack_type": "slice",
			"format": "addr"
		}
	]
}`,
			executedJson: `
{
	"name": "get_nft_address_by_index",
	"arguments": [
		{
			"name": "index",
			"stack_type": "int",
			"format": "bytes"
		}
	],
	"receives": [
		"08tzIyyFysK97F6dtnqpSZkuqqKit6y2gPvelDuXoPQ="
	],
	"return_values": [
		{
			"name": "address",
			"stack_type": "slice",
			"format": "addr"
		}
	],
	"returns": [
		"EQDHVwNhkIvqS3tJf0ScpM2kGd0Yi0PgGf_lZ1Vh0m7AyWD3"
	]
}`,
			args:   []any{mustFromB64(t, "08tzIyyFysK97F6dtnqpSZkuqqKit6y2gPvelDuXoPQ=")},
			result: false,
		},
		{
			contract: "jetton_minter",
			descJson: `
{
	"name": "get_wallet_address",
	"arguments": [
		{
			"name": "owner_address",
			"stack_type": "slice",
			"format": "addr"
		}
	],
	"return_values": [
		{
			"name": "wallet_address",
			"stack_type": "slice",
			"format": "addr"
		}
	]
}`,
			executedJson: `
{
	"name": "get_wallet_address",
	"arguments": [
		{
			"name": "owner_address",
			"stack_type": "slice",
			"format": "addr"
		}
	],
	"receives": [
		"EQDwKGXmxr9kV9bxUoi3o4eU9o6onrKw3g2sd57XAZFWV_kw"
	],
	"return_values": [
		{
			"name": "wallet_address",
			"stack_type": "slice",
			"format": "addr"
		}
	],
	"returns": [
		"EQAdeuTxkNGycqRS-MdwRGVtnEjiS1p7quWaA36Q2XlnOa4Q"
	]
}`,
			args:   []any{addr.MustFromBase64("EQDwKGXmxr9kV9bxUoi3o4eU9o6onrKw3g2sd57XAZFWV_kw").MustToTonutils()},
			result: true,
		},
	}

	for it := range testCases {
		ts := &testCases[it]

		var (
			desc abi.GetMethodDesc
			exec abi.GetMethodExecution
		)

		err := json.Unmarshal([]byte(ts.descJson), &desc)
		require.Nil(t, err)

		err = json.Unmarshal([]byte(ts.executedJson), &exec)
		require.Nil(t, err)

		id, res := (*Service)(nil).checkPrevGetMethodExecution(
			ts.contract,
			&desc,
			&core.AccountState{
				ExecutedGetMethods: map[abi.ContractName][]abi.GetMethodExecution{
					ts.contract: {exec},
				},
			},
			ts.args,
		)
		require.Equal(t, ts.result, res, fmt.Sprintf("test number %d", it))
		if res {
			require.Equal(t, 0, id)
		}
	}
}
