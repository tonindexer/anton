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
   "interface": "",    // name of the contract
   "addresses": [],    // optional known addresses
   "code_boc": "",     // optional contract code
   "im_messages": [],  // possible incoming messages schema
   "out_messages": [], // possible outgoing messages schema
   "get_methods": [],  // get-method names, return values and arguments
   "contract_data": [] // contract state data cell schema
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
         "go_type": "uint64"        // describes in what golang type should we map the given field
      }, 
      {
         "name": "auction_config",
         "tlb_type": "^",
         "go_type": "struct",
         "struct_fields": [         // fields of inner structure
            {
               "name": "beneficiary_address",
               "go_type": "addr",
               "tlb_type": "addr"
            }
         ]
      }
   ]
}
```

### Contract data schema

Contract data schema has the structure same as each message body schema, as it has no operation id and name.

### Types mapping

While parsing TL-B cells by fields description, we are trying to parse data according to TL-B type and map it into some Golang type or structure.
Each TL-B type used in schemas has value equal to the structure tags in [tonutils-go](https://github.com/xssnick/tonutils-go/blob/master/tlb/loader.go#L24).
If it is not possible to parse the field using `tlb.LoadFromCell`, 
you can define your custom type with `LoadFromCell` method in `abi` package (example, `TelemintText`) and register it in `tlb.go`.

Accepted TL-B types:
1. `## N` - integer with N bits
2. `^` - data is stored in the referenced cell
3. `.` - inner struct
4. `[^]dict N [-> array [^]]` - dictionary with key size `N`, transformation `->` can be applied to convert dict to array. 
   Example: `dict 256 -> array ^` will give you array of deserialized refs of values
5. `bits N` - bit slice N len to []byte
6. `bool` - 1 bit boolean
7. `addr` - ton address
8. `maybe` - reads 1 bit, and loads rest if its 1, can be used in combination with others only
9. `either X Y` - reads 1 bit, if its 0 - loads X, if 1 - loads Y

Accepted Go types:
1. `struct` - embed structure, maps into structure described by `struct_fields`
2. `bytes` - byte slice, maps into `[]byte`
3. `bool` - boolean
4. `uint8`, `uint16`, `uint32`, `uint64` - unsigned integers
5. `bigInt` - integer with more than 64 bits, maps into `big.Int` wrapper
6. `cell` - TL-B cell, maps into [`cell.Cell`](https://github.com/xssnick/tonutils-go/blob/master/tvm/cell/cell.go#L19)
7. `magic` - TL-B constructor prefix, must not be used
8. `coins` - varInt 16, maps into `big.Int` wrapper
9. `addr` - TON address, maps into [`address.Address`](https://github.com/xssnick/tonutils-go/blob/master/address/addr.go#L21) wrapper
10. `telemintText` - variable length string with [this](https://github.com/TelegramMessenger/telemint/blob/main/telemint.tlb#L25) TL-B constructor
