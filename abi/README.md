# TON contract interface

## Overview

You can define schema of contract get-methods and messages going to and from contract in just one JSON schema.

### Contract interface

Anton mainly determines contracts by the presence of get-methods in the contract code.
But if it is impossible to identify your contracts by only get-methods (as in Telemint NFT collection contracts), 
you should define contract addresses in the network or a contract code Bag of Cells.

```json5
{
  "interface_name": "",  // name of the contract
  "addresses": [],       // optional contract addresses
  "code_boc": "",        // optional contract code BoC
  "definitions": {},     // map definition name to cell schema
  "in_messages": [],     // possible incoming messages schema
  "out_messages": [],    // possible outgoing messages schema
  "get_methods": []      // get-method names, return values and arguments
}
```

### Message schema

Each message schema has operation name, operation code and field definitions. 
Each field definition has name, TL-B type and format, which shows how to parse cell.
Also, it is possible to define similarly described embedded structures in each field in `struct_fields`.

```json5
{
  "op_name": "nft_start_auction",  // operation name
  "op_code": "0x5fcc3d14",         // TL-B constructor prefix code (operation code)
  "type": "external_out",          // message type: internal, external_in, external_out
  "body": [
    {                              // fields definitions
      "name": "query_id",          // field name
      "tlb_type": "## 64",         // field TL-B type
      "format": "uint64"           // describes how we should parse the field
    }, 
    {
      "name": "auction_config",
      "tlb_type": "^",
      "format": "struct",
      "struct_fields": [           // fields of inner structure
        {
          "name": "beneficiary_address", 
          "tlb_type": "addr", 
          "format": "addr"
        }
      ]
    }
  ]
}
```

While parsing TL-B cells by fields description, we are trying to parse data according to TL-B type and map it into some Golang type or structure.
Each TL-B type used in schemas has value equal to the structure tags in [tonutils-go](https://github.com/xssnick/tonutils-go/blob/4d0157009913e35d450c36e28018cd0686502439/tlb/loader.go#L24).
If it is not possible to parse the field using `tlb.LoadFromCell`, 
you can define your custom type with `LoadFromCell` method in `abi` package (for example, `TelemintText`) and register it in `tlb_types.go`.

Accepted TL-B types in `tlb_type`:
1. `## N` - integer with N bits; by default maps to `uintX` or `big.Int`
2. `^` - data is stored in the referenced cell; by default maps to `cell.Cell` or to custom struct, if `struct_fields` is defined
3. `.` - inner struct; by default maps to `cell.Cell` or to custom struct, if `struct_fields` is defined
4. `[^]dict [inline] N [-> [^]]` - dictionary with key size `N`, transformation to `map` is done through `->`
5. `bits N` - bit slice N len; by default maps to `[]byte`
6. `bool` - 1 bit boolean; by default maps to `bool`
7. `addr` - ton address; by default maps to `addr.Address`
8. `maybe` - reads 1 bit, and loads rest if its 1, can be used in combination with others only; by default maps to `cell.Cell` or to custom struct, if `struct_fields` is defined
9. `either X Y` - reads 1 bit, if its 0 - loads X, if 1 - loads Y; by default maps to `cell.Cell` or to custom struct, if `struct_fields` is defined

Accepted types of `format`:
1. `struct` - embed structure, maps into structure described by `struct_fields`
2. `bytes` - byte slice, maps into `[]byte`
3. `bool` - boolean (can be used only on `tlb_type = bool`)
4. `uint8`, `uint16`, `uint32`, `uint64` - unsigned integers
5. `int8`, `int16`, `int32`, `int64` - unsigned integers
6. `bigInt` - integer with more than 64 bits, maps into `big.Int` wrapper
7. `cell` - TL-B cell, maps into [`cell.Cell`](https://github.com/xssnick/tonutils-go/blob/4d0157009913e35d450c36e28018cd0686502439/tvm/cell/cell.go#L11)
8. `dict` - TL-B dictionary (hashmap), maps into [`cell.Dictionary`](https://github.com/xssnick/tonutils-go/blob/4d0157009913e35d450c36e28018cd0686502439/tvm/cell/dict.go)
9. `tag` - TL-B constructor prefix
10. `coins` - varInt 16, maps into `big.Int` wrapper
11. `addr` - TON address, maps into [`address.Address`](https://github.com/xssnick/tonutils-go/blob/4d0157009913e35d450c36e28018cd0686502439/address/addr.go#L21) wrapper
12. [TODO] `content_cell` - token data as in [TEP-64](https://github.com/ton-blockchain/TEPs/blob/master/text/0064-token-data-standard.md); [implementation](https://github.com/xssnick/tonutils-go/blob/b839942a7b7bc431cc610f2ca3d9ff0e03079586/ton/nft/content.go#L10)
13. `string` - [string snake](https://github.com/xssnick/tonutils-go/blob/4d0157009913e35d450c36e28018cd0686502439/tvm/cell/builder.go#L317) is stored in the cell
14. `telemintText` - variable length string with [this](https://github.com/TelegramMessenger/telemint/blob/main/telemint.tlb#L25) TL-B constructor

### Get-methods

Each get-method consists of name (which is then used to get `method_id`), arguments and return values.

```json5
{
  "interface_name": "jetton_minter",
  "get_methods": [
    {
      "name": "get_wallet_address",         // get-method name
      "arguments": [
        {
          "name": "owner_address",          // argument name
          "stack_type": "slice",
          "format": "addr"
        }
      ],
      "return_values": [
        {
          "name": "jetton_wallet_address",  // return value name
          "stack_type": "slice",            // type we load
          "format": "addr"                  // type we parse into
        }
      ]
    },
    {
      "name": "get_jetton_data",
      "return_values": [
        {
          "name": "total_supply",
          "stack_type": "int",
          "format": "bigInt"
        },
        {
          "name": "mintable",
          "stack_type": "int",
          "format": "bool"
        },
        {
          "name": "admin_address",
          "stack_type": "slice",
          "format": "addr"
        }
      ]
    }
  ]
}
```

Accepted argument stack types:

1. `int` - integer; by default maps from `big.Int`
2. `cell` - map from BoC
3. `slice` - cell slice

Accepted return values stack types:

1. `int` - integer; by default maps into `big.Int`
2. `cell` - map to BoC
3. `slice` - load slice
4. [TODO] `tuple`
 
Accepted types to map from or parse into in `format` field:

1. `addr` - MsgAddress slice type
2. `bool` - map int to boolean
3. `uint8`, `uint16`, `uint32`, `uint64` - map int to an unsigned integer
4. `int8`, `int16`, `int32`, `int64` - map int to an signed integer
5. `bigInt` - map integer bigger than 64 bits
6. `string` - load string snake from cell
7. `bytes` - convert big int to bytes
8. `content` - load [TEP-64](https://github.com/ton-blockchain/TEPs/blob/master/text/0064-token-data-standard.md) standard token data into [`nft.ContentAny`](https://github.com/xssnick/tonutils-go/blob/b839942a7b7bc431cc610f2ca3d9ff0e03079586/ton/nft/content.go#L10)
9. `struct` - define struct_fields to parse cell

### Shared TL-B constructors

You can define some cell schema in `definitions` field of contract interface.

You can use those definitions in message schemas:

```json
{
  "interface_name": "telemint_nft_item",
  "addresses": [
    "EQAOQdwdw8kGftJCSFgOErM1mBjYPe4DBPq8-AhF6vr9si5N",
    "EQCA14o1-VWhS2efqoh_9M1b_A9DtKTuoqfmkn83AbJzwnPi"
  ],
  "definitions": {
    "auction_config": [
      {
        "name": "beneficiary_address",
        "tlb_type": "addr"
      }
    ]
  },
  "in_messages": [
    {
      "op_name": "teleitem_start_auction",
      "op_code": "0x487a8e81",
      "body": [
        {
          "name": "query_id", 
          "tlb_type": "## 64"
        },
        {
          "name": "auction_config",
          "tlb_type": "^",
          "format": "auction_config"
        }
      ]
    }
  ]
}
```

Or use them in get-method return values' schema:

```json5
{
  "interface_name": "amm",
  "definitions": {
    "amm_state": [
      {
        "name": "quote_asset_reserve",
        "tlb_type": ".",
        "format": "coins"
      },
      // ...
    ]
  }, 
  "get_methods": [
    {
      "name": "get_amm_data", 
      "return_values": [
        // ...
        {
          "name": "amm_state",
          "stack_type": "cell",
          "format": "amm_state"
        },
      ]
    }
  ]
}
```

### Union of TLB types

You can make some definitions with tags in the beginning of cell and use them later in unions. See the following example:

```json5
{
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
      // ...
    ],
    "pool_params": [
      // ...
      {
        "name": "asset0",
        "tlb_type": "[native_asset,jetton_asset]"
      },
      // ...
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
      // ...
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
      // ...
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
          "struct_fields": [{
            "name": "value", 
            "tlb_type": "[deposit_liquidity,swap]"
          }]
        }
      ]
    }
  ]
}
```

Here we define two structs in the interface: `deposit_liquidity` and `swap`.
Then our contract interface accepts incoming `jetton_transfer_notification`.
Inside forward payload there may be a cell, which corresponds to either `deposit_liquidity`, either `swap`.
If Anton finds a message with `jetton_transfer_notification` operation, he will try to determine the structure 
of forward payload by tag in the beginning of cell.  

After parsing `deposit_liquidity` transfer notification message body will look like this:

```json
{
  "query_id": 3638120226682551939,
  "amount": "1253854400825677",
  "sender": "EQDz0wQL6EEdgbPkFgS7nNmywzr468AvgLyhH7PIMALxPB6G",
  "forward_payload": {
    "value": {
      "deposit_liquidity": {},
      "pool_params": {
        "is_stable": false,
        "asset_0": {
          "native_asset": {}
        },
        "asset_1": {
          "jetton_asset": {},
          "workchain_id": 0,
          "jetton_address": 2422642597
        }
      },
      "min_lp_amount": "49289848313582100",
      "asset_0_target_balance": "135747634478277169790071850",
      "asset_1_target_balance": "30291957672135140790470162860"
    }
  }
}
```

### Dictionary transformation

You can define the format of the dictionary values, so Anton will be able to parse it into the golang `map`.

In the following example, we use defined `limit_order` as a dictionary value:
```json5
{
  // ...
  "definitions": {
    "limit_order": [
      {
        "name": "order_tag",
        "tlb_type": "$0010",
        "format": "tag"
      },
      {
        "name": "expiration",
        "tlb_type": "## 32"
      },
      // ...
    ]
  },
  // ...
  "in_message": {
    // ...
    "body": [
      {
        "name": "dict_3_bit_key",
        "tlb_type": "dict inline 3 -> ^",
        "format": "limit_order"
      }
    ]
  }
}
```

Or we can use defined `orders` union as a dictionary value, but for the union we're setting `tlb_type` field instead of `format`.
```json5
{
  // ...
  "definitions": {
    "take_order": [
      {
        "name": "take_order_tag",
        "tlb_type": "$0001",
        "format": "tag"
      },
      {
        "name": "expiration",
        "tlb_type": "## 32"
      },
      // ...
    ],
    "limit_order": [
      {
        "name": "order_tag",
        "tlb_type": "$0010",
        "format": "tag"
      },
      {
        "name": "expiration",
        "tlb_type": "## 32"
      },
      // ...
    ]
  },
  // ...
  "in_message": {
    // ...
    "body": [
      {
        "name": "dict_3_bit_key",
        "tlb_type": "dict inline 3 -> ^ [take_order,limit_order]"
      }
    ]
  }
}
```

## Known contracts

1. TEP-62 NFT Standard: [interfaces](/abi/known/tep62_nft.json), [description](https://github.com/ton-blockchain/TEPs/blob/master/text/0062-nft-standard.md), [contract code](https://github.com/ton-blockchain/token-contract/tree/main/nft)
2. TEP-74 Fungible tokens (Jettons) standard: [interfaces](/abi/known/tep74_jetton.json), [description](https://github.com/ton-blockchain/TEPs/blob/master/text/0074-jettons-standard.md), [contract code](https://github.com/ton-blockchain/token-contract/tree/main/ft)
3. TEP-81 DNS contracts: [interface](/abi/known/tep81_dns.json), [description](https://github.com/ton-blockchain/TEPs/blob/master/text/0081-dns-standard.md)
4. TEP-85 NFT SBT tokens: [interfaces](/abi/known/tep85_nft_sbt.json), [description](https://github.com/ton-blockchain/TEPs/blob/master/text/0085-sbt-standard.md)
5. Telemint contracts: [interfaces](/abi/known/telemint.json), [contract code](https://github.com/TelegramMessenger/telemint)
6. Getgems contracts: [interfaces](/abi/known/getgems.json), [contract code](https://github.com/getgems-io/nft-contracts/blob/main/packages/contracts/sources)
7. Wallets: [interfaces](/abi/known/wallets.json), [tonweb](https://github.com/toncenter/tonweb/tree/0a5effd36a3f342f4aacabab728b1f9747085ad1/src/contract/wallet)
8. [STON.fi](https://ston.fi) DEX: [architecture](https://docs.ston.fi/docs/developer-section/architecture), [contract code](https://github.com/ston-fi/dex-core)
9. [Megaton.fi](https://megaton.fi) DEX: [architecture](https://docs.megaton.fi/developers/contract)
10. [Tonpay](https://thetonpay.app): [go-sdk](https://github.com/TheTonpay/tonpay-go-sdk), [js-sdk](https://github.com/TheTonpay/tonpay-js-sdk)

## Converting Golang struct to JSON schema

You can convert Golang struct with described tlb tags to the JSON schema by using `abi.NewTLBDesc` and `abi.NewOperationDesc` functions.
See an example in [`tlb_test.go`](/abi/tlb_test.go) file.
