package abi_test

import (
	"encoding/base64"
	"encoding/json"
	"math/big"
	"testing"

	"github.com/xssnick/tonutils-go/address"
	"github.com/xssnick/tonutils-go/tlb"
	"github.com/xssnick/tonutils-go/tvm/cell"

	"github.com/stretchr/testify/require"

	"github.com/tonindexer/anton/abi"
)

type Payload struct {
	SmallInt  uint32   `tlb:"## 32"`
	BigInt    *big.Int `tlb:"## 128"`
	RefStruct struct {
		Addr *address.Address `tlb:"addr"`
	} `tlb:"^"`
	EmbedStruct struct {
		Bits []byte `tlb:"bits 32"`
	} `tlb:"^"`
	MaybeCell  *cell.Cell `tlb:"maybe ^"`
	EitherCell *cell.Cell `tlb:"either ^ ."`
}

type Operation struct {
	_       tlb.Magic `tlb:"#00000001"`
	Payload Payload   `tlb:"^"`
}

var testPayloadShortSchema = `[{"name":"small_int","tlb_type":"## 32"},{"name":"big_int","tlb_type":"## 128"},{"name":"ref_struct","tlb_type":"^","struct_fields":[{"name":"addr","tlb_type":"addr"}]},{"name":"embed_struct","tlb_type":"^","struct_fields":[{"name":"bits","tlb_type":"bits 32"}]},{"name":"maybe_cell","tlb_type":"maybe ^"},{"name":"either_cell","tlb_type":"either ^ ."}]`
var testPayloadFullSchema = `[{"name":"small_int","tlb_type":"## 32","format":"uint32"},{"name":"big_int","tlb_type":"## 128","format":"bigInt"},{"name":"ref_struct","tlb_type":"^","format":"struct","struct_fields":[{"name":"addr","tlb_type":"addr","format":"addr"}]},{"name":"embed_struct","tlb_type":"^","format":"struct","struct_fields":[{"name":"bits","tlb_type":"bits 32","format":"bytes"}]},{"name":"maybe_cell","tlb_type":"maybe ^","format":"cell"},{"name":"either_cell","tlb_type":"either ^ .","format":"cell"}]`

func TestNewTLBDesc(t *testing.T) {
	var d1, d2 abi.TLBFieldsDesc

	// test json unmarshal
	err := json.Unmarshal([]byte(testPayloadShortSchema), &d1)
	require.Nil(t, err)

	x, err := d1.New()
	require.Nil(t, err)

	// test structure description
	d2, err = abi.NewTLBDesc(x)
	require.Nil(t, err)

	_, err = d2.New()
	require.Nil(t, err)

	j, err := json.Marshal(d2)
	require.Nil(t, err)
	require.Equal(t, testPayloadFullSchema, string(j))
}

func TestNewOperationDesc(t *testing.T) {
	d1, err := abi.NewOperationDesc(&Operation{})
	require.Nil(t, err)

	x, err := d1.New()
	require.Nil(t, err)

	d2, err := abi.NewOperationDesc(x)
	require.Nil(t, err)

	_, err = d2.New()
	require.Nil(t, err)

	d2.Name = d1.Name
	j1, err := json.Marshal(d1)
	require.Nil(t, err)
	j2, err := json.Marshal(d2)
	require.Nil(t, err)

	require.Equal(t, string(j1), string(j2))
}

func TestTLBFieldsDesc_LoadFromCell(t *testing.T) {
	var (
		p Payload
		d abi.TLBFieldsDesc
	)

	p.SmallInt = 42
	p.BigInt, _ = new(big.Int).SetString("8000000000000000000000000", 10)
	p.RefStruct.Addr = address.MustParseAddr("EQDj5AA8mQvM5wJEQsFFFof79y3ZsuX6wowktWQFhz_Anton")
	p.EmbedStruct.Bits = []byte("asdf")
	p.MaybeCell = nil
	p.EitherCell = cell.BeginCell().MustStoreStringSnake("either").EndCell()

	err := json.Unmarshal([]byte(testPayloadShortSchema), &d)
	require.Nil(t, err)

	x, err := d.New()
	require.Nil(t, err)

	c, err := tlb.ToCell(&p)
	require.Nil(t, err)

	err = tlb.LoadFromCell(x, c.BeginParse())
	require.Nil(t, err)

	j, err := json.Marshal(&x)
	require.Nil(t, err)

	exp := `{"small_int":42,"big_int":8000000000000000000000000,"ref_struct":{"addr":"EQDj5AA8mQvM5wJEQsFFFof79y3ZsuX6wowktWQFhz_Anton"},"embed_struct":{"bits":"YXNkZg=="},"maybe_cell":null,"either_cell":"te6cckEBAQEACAAADGVpdGhlcskJ1lc="}`
	require.Equal(t, exp, string(j))
}

func TestTLBFieldsDesc_LoadFromCell_DictToMap(t *testing.T) {
	d := []byte(`[
  {
    "name": "order_tag",
    "tlb_type": "$0010",
    "format": "tag"
  },
  {
    "name": "expiration",
    "tlb_type": "## 32"
  },
  {
    "name": "direction",
    "tlb_type": "## 1"
  },
  {
    "name": "amount",
    "tlb_type": ".",
    "format": "coins"
  },
  {
    "name": "leverage",
    "tlb_type": "## 64"
  },
  {
    "name": "limit_price",
    "tlb_type": ".",
    "format": "coins"
  },
  {
    "name": "stop_price",
    "tlb_type": ".",
    "format": "coins"
  },
  {
    "name": "stop_trigger_price",
    "tlb_type": ".",
    "format": "coins"
  },
  {
    "name": "take_trigger_price",
    "tlb_type": ".",
    "format": "coins"
  }
]`)

	var descD abi.TLBFieldsDesc

	err := json.Unmarshal(d, &descD)
	require.Nil(t, err)

	err = abi.RegisterDefinitions(map[abi.TLBType]abi.TLBFieldsDesc{
		"take_order": descD,
	})
	if err != nil {
		require.Nil(t, err)
	}

	j := []byte(`[
  {
    "name": "dict_uint_3",
    "tlb_type": "dict inline 3 -> ^",
    "format": "take_order"
  }
]`)

	var desc abi.TLBFieldsDesc

	err = json.Unmarshal(j, &desc)
	require.Nil(t, err)

	body, err := base64.StdEncoding.DecodeString(`te6cckEBBQEAUwACAdQDAQEBIAIAQSZS6uXai6Q7dAAAAAAAWWgvACEeGjAAIU3JOAIO5rKAQAEBIAQAQSZS5ufKi6Q7dAAAAAAAWWgvACEeGjAAIU3JOAIO5rKAQPxznzQ=`)
	require.Nil(t, err)

	c, err := cell.FromBOC(body)
	require.Nil(t, err)

	got, err := desc.FromCell(c)
	require.Nil(t, err)

	j, err = json.Marshal(got)
	require.Nil(t, err)

	require.Equal(t,
		`{"dict_uint_3":{"0":{"order_tag":{},"expiration":1697541756,"direction":1,"amount":"100000000000","leverage":3000000000,"limit_price":"600000000","stop_price":"0","stop_trigger_price":"700000000","take_trigger_price":"500000000"},"1":{"order_tag":{},"expiration":1697558109,"direction":1,"amount":"100000000000","leverage":3000000000,"limit_price":"600000000","stop_price":"0","stop_trigger_price":"700000000","take_trigger_price":"500000000"}}}`,
		string(j))
}

func TestTLBFieldsDesc_LoadFromCell_DefinitionsUnion(t *testing.T) {
	j := []byte(`{
  "interface": "jetton_vault",
  "definitions": {
    "native_asset": [
      {
        "name": "native_asset",
        "tlb_type": "$0000",
        "format": "tag"
      }
    ],
    "jetton_asset": [
      {
        "name": "jetton_asset",
        "tlb_type": "$0001",
        "format": "tag"
      },
      {
        "name": "workchain_id",
        "tlb_type": "## 8",
        "format": "int8"
      },
      {
        "name": "jetton_address",
        "tlb_type": "## 32",
        "format": "uint32"
      }
    ],
    "pool_params": [
      {
        "name": "is_stable",
        "tlb_type": "bool"
      },
      {
        "name": "asset0",
        "tlb_type": "[native_asset,jetton_asset]"
      },
      {
        "name": "asset1",
        "tlb_type": "[native_asset,jetton_asset]"
      }
    ],
    "deposit_liquidity": [
      {
        "name": "deposit_liquidity",
        "tlb_type": "#40e108d6",
        "format": "tag"
      },
      {
        "name": "pool_params",
        "tlb_type": ".",
        "format": "pool_params"
      },
      {
        "name": "min_lp_amount",
        "tlb_type": ".",
        "format": "coins"
      },
      {
        "name": "asset0_target_balance",
        "tlb_type": ".",
        "format": "coins"
      },
      {
        "name": "asset1_target_balance",
        "tlb_type": ".",
        "format": "coins"
      }
    ],
    "swap_step_params": [
      {
        "name": "swap_kind",
        "tlb_type": "## 1",
        "format": "uint8"
      },
      {
        "name": "limit",
        "tlb_type": ".",
        "format": "coins"
      },
      {
        "name": "next",
        "tlb_type": "maybe ^",
        "format": "cell"
      }
    ],
    "swap_step": [
      {
        "name": "pool_addr",
        "tlb_type": "addr"
      },
      {
        "name": "params",
        "tlb_type": ".",
        "format": "swap_step_params"
      }
    ],
    "swap_params": [
      {
        "name": "deadline",
        "tlb_type": "## 32",
        "format": "uint32"
      },
      {
        "name": "recipient_addr",
        "tlb_type": "addr",
        "format": "addr"
      },
      {
        "name": "referral_addr",
        "tlb_type": "addr",
        "format": "addr"
      }
    ],
    "swap": [
      {
        "name": "swap",
        "tlb_type": "#e3a0d482",
        "format": "tag"
      },
      {
        "name": "swap_step",
        "tlb_type": ".",
        "format": "swap_step"
      },
      {
        "name": "swap_params",
        "tlb_type": "^",
        "format": "swap_params"
      }
    ]
  },
  "in_messages": [
    {
      "op_name": "jetton_transfer_notification",
      "op_code": "0x7362d09c",
      "body": [
        {
          "name": "query_id",
          "tlb_type": "## 64",
          "format": "uint64"
        },
        {
          "name": "amount",
          "tlb_type": ".",
          "format": "coins"
        },
        {
          "name": "sender",
          "tlb_type": "addr",
          "format": "addr"
        },
        {
          "name": "forward_payload",
          "tlb_type": "either . ^",
          "format": "struct",
          "struct_fields": [
            {
              "name": "value",
              "tlb_type": "[deposit_liquidity,swap]"
            }
          ]
        }
      ]
    }
  ]
}`)

	var i abi.InterfaceDesc

	err := json.Unmarshal(j, &i)
	require.Nil(t, err)

	err = abi.RegisterDefinitions(i.Definitions)
	require.Nil(t, err)

	// body, err := base64.StdEncoding.DecodeString(`te6cckEBAwEAaAABaHNi0JwAAAJovkRKsWAQCc9A8DgB1SCJzWjksBzMruGpmYclZnRmWWc2C4h83mgaryg4Sd8BAU3joNSCgB8qWTCcZtGRrvci/dxNr39DgHo5VaQVpPrkLnQ6xUf9YEACAAkAAAAAAshothE=`)
	// require.Nil(t, err)
	body, err := base64.StdEncoding.DecodeString(`te6cckEBAgEAbQABanNi0JwyfTMSEZO+g3BHRfuilJTYAeemCBfQgjsDZ8gsCXc5s2WGdfHXgF8BeUI/Z5BgBeJ5AQBlQOEI1gCASDNL0r145tjeHJCluCTXAVnklS2GGhVjDwdcXsmCWviCHc1lADgjov3RSkppLrJXaA==`)
	require.Nil(t, err)

	c, err := cell.FromBOC(body)
	require.Nil(t, err)

	got, err := i.InMessages[0].FromCell(c)
	require.Nil(t, err)

	j, err = json.Marshal(got)
	require.Nil(t, err)

	require.Equal(t,
		// `{"query_id":2648892000945,"amount":"1102144868099","sender":"EQDqkETmtHJYDmZXcNTMw5KzOjMss5sFxD5vNA1XlBwk7_mo","forward_payload":{"value":{"swap":{},"swap_step":{"pool_addr":"EQD5UsmE4zaMjXe5F-7ibXv6HAPRyq0grSfXIXOh1io_6xmV","params":{"swap_kind":0,"limit":"0","next":null}},"swap_params":{"deadline":0,"recipient_addr":"NONE","referral_addr":"NONE"}}}}`,
		`{"query_id":3638120226682551939,"amount":"1253854400825677","sender":"EQDz0wQL6EEdgbPkFgS7nNmywzr468AvgLyhH7PIMALxPB6G","forward_payload":{"value":{"deposit_liquidity":{},"pool_params":{"is_stable":false,"asset_0":{"native_asset":{}},"asset_1":{"jetton_asset":{},"workchain_id":0,"jetton_address":2422642597}},"min_lp_amount":"49289848313582100","asset_0_target_balance":"135747634478277169790071850","asset_1_target_balance":"30291957672135140790470162860"}}}`,
		string(j))
}
