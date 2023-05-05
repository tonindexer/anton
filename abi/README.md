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
   "body": [
      {                             // fields definitions
         "name": "query_id",        // field name
         "tlb_type": "## 64",       // field TL-B type
         "format": "uint64"         // describes how we should parse the field
      }, 
      {
         "name": "auction_config",
         "tlb_type": "^",
         "format": "struct",
         "struct_fields": [         // fields of inner structure
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

### Types mapping

While parsing TL-B cells by fields description, we are trying to parse data according to TL-B type and map it into some Golang type or structure.
Each TL-B type used in schemas has value equal to the structure tags in [tonutils-go](https://github.com/xssnick/tonutils-go/blob/4d0157009913e35d450c36e28018cd0686502439/tlb/loader.go#L24).
If it is not possible to parse the field using `tlb.LoadFromCell`, 
you can define your custom type with `LoadFromCell` method in `abi` package (for example, `TelemintText`) and register it in `tlb_types.go`.

Accepted TL-B types in `tlb_type`:
1. `## N` - integer with N bits; by default maps to `uintX` or `big.Int`
2. `^` - data is stored in the referenced cell; by default maps to `cell.Cell` or to custom struct, if `struct_fields` is defined
3. `.` - inner struct; by default maps to `cell.Cell` or to custom struct, if `struct_fields` is defined
4. [TODO] `[^]dict N [-> array [^]]` - dictionary with key size `N`, transformation is not supported yet
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
5. `bigInt` - integer with more than 64 bits, maps into `big.Int` wrapper
6. `cell` - TL-B cell, maps into [`cell.Cell`](https://github.com/xssnick/tonutils-go/blob/4d0157009913e35d450c36e28018cd0686502439/tvm/cell/cell.go#L11)
7. `dict` - TL-B dictionary (hashmap), maps into [`cell.Dictionary`](https://github.com/xssnick/tonutils-go/blob/4d0157009913e35d450c36e28018cd0686502439/tvm/cell/dict.go)
8. `magic` - TL-B constructor prefix, must not be used
9. `coins` - varInt 16, maps into `big.Int` wrapper
10. `addr` - TON address, maps into [`address.Address`](https://github.com/xssnick/tonutils-go/blob/4d0157009913e35d450c36e28018cd0686502439/address/addr.go#L21) wrapper
11. [TODO] `content_cell` - token data as in [TEP-64](https://github.com/ton-blockchain/TEPs/blob/master/text/0064-token-data-standard.md); [implementation](https://github.com/xssnick/tonutils-go/blob/b839942a7b7bc431cc610f2ca3d9ff0e03079586/ton/nft/content.go#L10)
12. `string` - [string snake](https://github.com/xssnick/tonutils-go/blob/4d0157009913e35d450c36e28018cd0686502439/tvm/cell/builder.go#L317) is stored in the cell
13. `telemintText` - variable length string with [this](https://github.com/TelegramMessenger/telemint/blob/main/telemint.tlb#L25) TL-B constructor

### Shared TL-B constructors

You can define some cell schema in contract interface `definitions` field and use it later in messages or contract data schemas.

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

### Get-methods

Each get-method consists of name (which is then used to get `method_id`), arguments and return values.

```json5
{
   "interface_name": "jetton_minter",
   "get_methods": [
      {
         "name": "get_wallet_address",          // get-method name
         "arguments": [
            {
               "name": "owner_address",         // argument name
               "stack_type": "slice",
               "format": "addr"
            }
         ],
         "return_values": [
            {
               "name": "jetton_wallet_address", // return value name
               "stack_type": "slice",           // type we load
               "format": "addr"                 // type we parse into
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
 
Accepted types to map from or into in `format` field:

1. `addr` - MsgAddress slice type
2. `bool` - map int to boolean
3. `uint8`, `uint16`, `uint32`, `uint64` - map int to an unsigned integer
4. `bigInt` - map integer bigger than 64 bits
5. `string` - load string snake from cell
6. `content` - load [TEP-64](https://github.com/ton-blockchain/TEPs/blob/master/text/0064-token-data-standard.md) standard token data into [`nft.ContentAny`](https://github.com/xssnick/tonutils-go/blob/b839942a7b7bc431cc610f2ca3d9ff0e03079586/ton/nft/content.go#L10)

## Known contracts

1. TEP-62 NFT Standard: [interfaces](/abi/known/tep62_nft.json), [description](https://github.com/ton-blockchain/TEPs/blob/master/text/0062-nft-standard.md), [contract code](https://github.com/ton-blockchain/token-contract/tree/main/nft)
2. TEP-74 Fungible tokens (Jettons) standard: [interfaces](/abi/known/tep74_jetton.json), [description](https://github.com/ton-blockchain/TEPs/blob/master/text/0074-jettons-standard.md), [contract code](https://github.com/ton-blockchain/token-contract/tree/main/ft)
3. TEP-81 DNS contracts: [interface](/abi/known/tep81_dns.json), [description](https://github.com/ton-blockchain/TEPs/blob/master/text/0081-dns-standard.md)
4. TEP-85 NFT SBT tokens: [interfaces](/abi/known/tep85_nft_sbt.json), [description](https://github.com/ton-blockchain/TEPs/blob/master/text/0085-sbt-standard.md)
5. Telemint contracts: [interfaces](/abi/known/telemint.json), [contract code](https://github.com/TelegramMessenger/telemint) 
6. Getgems contracts: [contract code](https://github.com/getgems-io/nft-contracts/blob/main/packages/contracts/sources)
7. Wallets: [interfaces](/abi/known/wallets.json), [tonweb](https://github.com/toncenter/tonweb/tree/0a5effd36a3f342f4aacabab728b1f9747085ad1/src/contract/wallet)
8. [STON.fi](https://ston.fi) DEX: [architecture](https://docs.ston.fi/docs/developer-section/architecture), [contract code](https://github.com/ston-fi/dex-core)
9. [Megaton.fi](https://megaton.fi) DEX: [architecture](https://docs.megaton.fi/developers/contract)
10. [Tonpay](https://thetonpay.app): [go-sdk](https://github.com/TheTonpay/tonpay-go-sdk), [js-sdk](https://github.com/TheTonpay/tonpay-js-sdk) 
