# TON contract interface

## Overview

You can define messages schema going to and from contract, contract get-methods and account data in just one JSON schema.
Those schemas are used in Anton, it looks for contracts in the network, which fit interfaces it knows.

### Contract interface

Anton mainly determines contracts by the presence of get-methods in the contract code.
But if it is impossible to define your contracts by only get-methods (as in Telemint NFT collection contracts), 
you should define contract addresses in the network or a contract code BoC.

```json5
{
   "interface_name": "",  // name of the contract
   "addresses": [],       // optional known addresses
   "code_boc": "",        // optional contract code
   "definitions": {},     // map from definition name to cell schema
   "in_messages": [],     // possible incoming messages schema
   "out_messages": [],    // possible outgoing messages schema
   "get_methods": [],     // get-method names, return values and arguments
   "contract_data": []    // contract state data cell schema
}
```

### Message schema

Each message schema has operation name, operation code and field definitions. 
Each field definition has name, TL-B type and Go type, which will be used to map this field into a Golang structure.
Also, it is possible to define similarly described embedded structures in the field.

```json5
{
   "op_name": "nft_start_auction",  // operation name
   "op_code": "0x5fcc3d14",         // TL-B constructor prefix code (operation code)
   "body": [
      {                             // fields definitions
         "name": "query_id",        // field name
         "tlb_type": "## 64",       // describes how we should parse the field
         "format": "uint64"         // [optional] describes in what golang type should we map the given field
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
         // TODO: add omitempty flag (example, it's needed for the forward payload)
      }
   ]
}
```

### Contract data schema

Contract data schema has the structure same as each message body schema, as it has no operation id and name.

### Types mapping

While parsing TL-B cells by fields description, we are trying to parse data according to TL-B type and map it into some Golang type or structure.
Each TL-B type used in schemas has value equal to the structure tags in [tonutils-go](https://github.com/xssnick/tonutils-go/blob/4d0157009913e35d450c36e28018cd0686502439/tlb/loader.go#L24).
If it is not possible to parse the field using `tlb.LoadFromCell`, 
you can define your custom type with `LoadFromCell` method in `abi` package (example, `TelemintText`) and register it in `tlb_types.go`.

Accepted TL-B types in `tlb_type`:
1. `## N` - integer with N bits; by default maps to `uintX` or `big.Int`
2. `^` - data is stored in the referenced cell; by default maps to `cell.Cell` or to custom struct, if `struct_fields` is defined
3. `.` - inner struct; by default maps to `cell.Cell` or to custom struct, if `struct_fields` is defined
4. [TODO] `[^]dict N [-> array [^]]` - dictionary with key size `N`, transformation `->` can be applied to convert dict to array. 
   Example: `dict 256 -> array ^` will give you array of deserialized refs of values
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
11. `content_cell` - token data as in [TEP-64](https://github.com/ton-blockchain/TEPs/blob/master/text/0064-token-data-standard.md); [implementation](https://github.com/xssnick/tonutils-go/blob/b839942a7b7bc431cc610f2ca3d9ff0e03079586/ton/nft/content.go#L10)
12. `string_cell` - [string snake](https://github.com/xssnick/tonutils-go/blob/4d0157009913e35d450c36e28018cd0686502439/tvm/cell/builder.go#L317) is stored in the cell
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
               "func_type": "slice",            // type we are trying to load
               "format": "addr"               // type we map into
            }
         ],
         "return_values": [
            {
               "name": "jetton_wallet_address", // return value name
               "func_type": "slice",            // type we load
               "format": "addr"               // type we parse into
            }
         ]
      },
      {
         "name": "get_jetton_data",
         "return_values": [
            {
               "name": "total_supply",
               "func_type": "int",
               "format": "bigInt"
            },
            {
               "name": "mintable",
               "func_type": "int",
               "format": "bool"
            },
            {
               "name": "admin_address",
               "func_type": "slice",
               "format": "addr"
            }
         ]
      }
   ]
}
```

Accepted func argument types:

1. `int` - integer; by default maps from `big.Int`
2. `cell` - map from BoC
3. `slice` - cell slice

Accepted func return values types:

1. `int` - integer; by default maps into `big.Int`
2. `cell` - map to BoC
3. `slice` - load slice
4. `tuple` - TODO :(
 
Accepted types to map from or into in `format` field:

1. `addr` - MsgAddress slice type
2. `bool` - map int to boolean
3. `uint8`, `uint16`, `uint32`, `uint64` - map int to an unsigned integer
4. `bigInt` - map integer bigger than 64 bits
5. `string` - load string snake from cell

## Known contracts

1. TEP-62 NFT Standard: [interfaces](/abi/known/tep62_nft.json), [description](https://github.com/ton-blockchain/TEPs/blob/master/text/0062-nft-standard.md), [contract code](https://github.com/ton-blockchain/token-contract/tree/main/nft)
2. TEP-74 Fungible tokens (Jettons) standard: [interfaces](/abi/known/tep74_jetton.json), [description](https://github.com/ton-blockchain/TEPs/blob/master/text/0074-jettons-standard.md), [contract code](https://github.com/ton-blockchain/token-contract/tree/main/ft)
3. TEP-81 DNS contracts: [description](https://github.com/ton-blockchain/TEPs/blob/master/text/0081-dns-standard.md)
4. TEP-85 NFT SBT tokens: [description](https://github.com/ton-blockchain/TEPs/blob/master/text/0085-sbt-standard.md)
5. Getgems contracts: [contract code](https://github.com/getgems-io/nft-contracts/blob/main/packages/contracts/sources)
6. Wallets: [tonweb](https://github.com/toncenter/tonweb/tree/0a5effd36a3f342f4aacabab728b1f9747085ad1/src/contract/wallet)
