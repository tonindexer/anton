# JSON API query examples

This documentation provides the list of available endpoints in the TON indexer API. 
Each endpoint accepts GET method requests and parameters in the URL query. The responses are in JSON format.

You can find detailed description for all parameters in [generated swagger documentation](https://anton.tools/api/v0/swagger).

## GetStatistics

Returns statistics on blocks, transactions, messages, and accounts.

### Endpoint: `/statistics`

### Request

```shell
curl -X GET 'https://anton.tools/api/v0/statistics'
```

### Response

```json
{
  "first_masterchain_block": 25000000,
  "last_masterchain_block": 28401995, 
  "masterchain_block_count": 3401996,

  "address_count": 1126481,
  "parsed_address_count": 998585,

  "account_count": 12350136,
  "parsed_account_count": 10693463,

  "transaction_count": 32888654,
  "message_count": 19542237,
  "parsed_message_count": 4911229,

  "contract_interface_count": 25,
  "contract_operation_count": 26,

  "account_count_by_status": [
    {
      "status": "ACTIVE",
      "count": 1094632
    }, {
      "status": "UNINIT", 
      "count": 31835
    }, {
      "status": "FROZEN",
      "count": 14
    }
  ],

  "account_count_by_interfaces": [
    {
      "interfaces": [ "nft_item" ],
      "count": 208234
    }, {
      "interfaces": [ "nft_sale" ], 
      "count": 184199
    },{
      "interfaces": [ "telemint_nft_item", "nft_royalty", "nft_item" ],
      "count": 136584
    }, {
      "interfaces": [ "wallet_v4r2_interface", "wallet", "wallet_v4r2" ],
      "count": 114627
    }, {
      "interfaces": [ "jetton_wallet" ],
      "count": 104451
    }
  ],

  "message_count_by_operation": [
    {
      "operation": "jetton_transfer",
      "count": 770003
    }, {
      "operation": "jetton_internal_transfer",
      "count": 748801
    }, {
      "operation": "nft_item_ownership_assigned",
      "count": 593450
    }, {
      "operation": "jetton_transfer_notification",
      "count": 542072
    }, {
      "operation": "nft_item_transfer",
      "count": 440577
    }
  ]
}
```

## GetContractInterfaces

Returns known contract interfaces or known contract addresses.
Interfaces are determined by `get-methods`.
But some contracts cannot be discovered only by `get-methods`.
For example, [telemint](https://github.com/TelegramMessenger/telemint/tree/main/func) NFT collections has all get-methods
identical to standard NFT collections, but they have some specific messages not seen in other collections.
That's why we also match contracts found in the network to interfaces by code hash or addresses.

### Endpoint: `/contract/interfaces`

### Request

```shell
curl -X GET 'https://anton.tools/api/v0/contract/interfaces'
```

### Response

```json
{
  "total": 25,
  "results": [
    {
      "name": "nft_collection",
      "get_methods": [
        "get_collection_data",
        "get_nft_address_by_index",
        "get_nft_content"
      ],
      "get_method_hashes": [ 102491, 92067, 68445 ]
    },
    {
      "name": "jetton_wallet",
      "get_methods": [
        "get_wallet_data"
      ],
      "get_method_hashes": [ 97026 ]
    },
    {
      "name": "telemint_nft_collection",
      "addresses": [
        {
          "hex": "0:0e41dc1dc3c9067ed24248580e12b3359818d83dee0304fabcf80845eafafdb2",
          "base64": "EQAOQdwdw8kGftJCSFgOErM1mBjYPe4DBPq8-AhF6vr9si5N"
        }, {
          "hex": "0:80d78a35f955a14b679faa887ff4cd5bfc0f43b4a4eea2a7e6927f3701b273c2",
          "base64": "EQCA14o1-VWhS2efqoh_9M1b_A9DtKTuoqfmkn83AbJzwnPi"
        }
      ]
    },
    {
      "name": "wallet_v4r2",
      "code": "... etc ...",
      "code_hash": "... etc ..."
    }
  ]
}
```

## GetContractOperations

After we match contracts in the network to interfaces, we can parse messages going to and from them.
Schemas of messages returned here is identical to what is defined in [msg_schema.json](/docs/msg_schema.json). 

### Endpoint: `/contract/operations`

### Request

```shell
curl -X GET 'https://anton.tools/api/v0/contract/operations'
```

### Response

```json
{
  "total": 26,
  "results": [
    {
      "name": "jetton_transfer_notification",
      "contract_name": "jetton_wallet",
      "outgoing": true,
      "operation_id": 1935855772,
      "schema": [
        {
          "tag": "tlb:\"#7362d09c\"",
          "name": "Op",
          "type": "magic"
        },
        {
          "tag": "tlb:\"## 64\"",
          "name": "QueryID",
          "type": "uint64"
        },
        {
          "tag": "tlb:\".\"",
          "name": "Amount",
          "type": "coins"
        },
        {
          "tag": "tlb:\"addr\"",
          "name": "Sender",
          "type": "address"
        }
      ]
    },
    {
      "name": "jetton_transfer",
      "contract_name": "jetton_wallet",
      "outgoing": false,
      "operation_id": 260734629,
      "schema": [
        {
          "tag": "tlb:\"#0f8a7ea5\"",
          "name": "Op",
          "type": "magic"
        },
        {
          "tag": "tlb:\"## 64\"",
          "name": "QueryID",
          "type": "uint64"
        },
        {
          "tag": "tlb:\".\"",
          "name": "Amount",
          "type": "coins"
        },
        {
          "tag": "tlb:\"addr\"",
          "name": "Destination",
          "type": "address"
        },
        {
          "tag": "tlb:\"addr\"",
          "name": "ResponseDestination",
          "type": "address"
        }
      ]
    }
  ]
}
```

## GetAccounts

Returns filtered account states and their parsed data.
The filter can be set by addresses, interfaces and owner or minter addresses (for FT and NFT items).
If `latest=true` parameter is set, it returns only the latest known account state for every address.

### Endpoint: `/accounts`

### Request

```shell
curl -X GET 'https://anton.tools/api/v0/accounts?latest=true&interface=nft_collection&order=DESC&after=36418223000005&limit=1'
```

### Response

```json
{
  "total": 3760,
  "results": [
    {
      "address": {
        "hex": "0:80d78a35f955a14b679faa887ff4cd5bfc0f43b4a4eea2a7e6927f3701b273c2",
        "base64": "EQCA14o1-VWhS2efqoh_9M1b_A9DtKTuoqfmkn83AbJzwnPi"
      },
      "is_active": true,
      "status": "ACTIVE",
      "balance": 220108298203,
      "last_tx_lt": 36418077000003,
      "last_tx_hash": "DGVtWg4TWZKO8fzGtHIvKYp3gSKZugnnnMTQJn+G1z8=",
      "state_data": {
        "types": [ "dns_resolver", "nft_collection", "telemint_nft_dns", "telemint_nft_collection" ],
        "next_item_index": -1,
        "content_uri": "https://nft.fragment.com/usernames.json"
      },
      "code_hash": "TU65v6KyENiQgzpxK+JECB/drbB/Ax82HqXPfxoTBlE=",
      "data_hash": "BFliyR1a8gnsippQtiTiSUD9cIqvYo0/x/Z38rqeXCE=",
      "get_method_hashes": [ 7, 38, 39, 102491, 123660, 66763, 68445, 92067, 524287 ],
      "updated_at": "2023-03-29T02:28:35Z"
    }
  ]
}
```

## AggregateAccounts

Returns statistics on account states.
Currently, only minter address can be set.
With NFT minter address set, it returns number of items, number of owners, counts items owned by each owner, counts number of unique owners for each item.
With FT minter address set, it returns number of wallets, total supply and supply owned by each wallet owner.

### Endpoint: `/accounts/aggregated`

### Request

```shell
# get statistics on telegram usernames
curl -X GET 'https://anton.tools/api/v0/accounts/aggregated?minter_address=EQCA14o1-VWhS2efqoh_9M1b_A9DtKTuoqfmkn83AbJzwnPi&limit=3'
# get statistics on Megaton MEGA-WTON LP tokens
curl -X GET 'https://anton.tools/api/v0/accounts/aggregated?minter_address=EQCSbsYkouaBFzc-4UnVbhNlbqSAzTy9cdnJzEm116Hc5JQw&limit=3'
```

### Response for NFT collection

```json
{
  "items": 41062,
  "owners_count": 10196,
  "owned_items": [
    {
      "owner_address": null,
      "items_count": 1487
    },
    {
      "owner_address": {
        "hex": "0:e1949d7df75d9209b483e605deffca558b4dd2df9f78874e453f0fe23f9032dc",
        "base64": "EQDhlJ19912SCbSD5gXe_8pVi03S3594h05FPw_iP5Ay3Fnv"
      },
      "items_count": 566
    },
    {
      "owner_address": {
        "hex": "0:e2e0df769417a45b1cca1ee49255380fb413c7337cc336b7bd48c4e99095db66",
        "base64": "EQDi4N92lBekWxzKHuSSVTgPtBPHM3zDNre9SMTpkJXbZupR"
      },
      "items_count": 530
    }
  ],
  "unique_owners": [
    {
      "item_address": {
        "hex": "0:7e0f6271c583c7a8dc4444b728c7b3156a8a213ad64f0afe5cb8294eac23c8cc",
        "base64": "EQB-D2JxxYPHqNxERLcox7MVaoohOtZPCv5cuClOrCPIzFZb"
      },
      "owners_count": 19
    },
    {
      "item_address": {
        "hex": "0:40618cec50911b3fd6c7d9d86f348350fe589ba2a7c40391b5b6281bbaec2015",
        "base64": "EQBAYYzsUJEbP9bH2dhvNINQ_liboqfEA5G1tigbuuwgFQ4o"
      },
      "owners_count": 19
    },
    {
      "item_address": {
        "hex": "0:9bbfe27993adee1a673d22ec396d7bca2fac73e9a26cdbb02772abbfa0fc925b",
        "base64": "EQCbv-J5k63uGmc9Iuw5bXvKL6xz6aJs27Ancqu_oPySW7On"
      },
      "owners_count": 18
    }
  ]
}
```

### Response for FT minter

```json
{
  "wallets": 342,
  "total_supply": 110854457707651,
  "owned_balance": [
    {
      "wallet_address": {
        "hex": "0:95aaf5945f6502e37e334e0bce32f506b5a75beefe370a725ec6e880935cc364",
        "base64": "EQCVqvWUX2UC434zTgvOMvUGtadb7v43CnJexuiAk1zDZORq"
      },
      "owner_address": {
        "hex": "0:3488aedff63a7d6c166666601e787070de8b0204c3303b60465d1fd45e33cbff",
        "base64": "EQA0iK7f9jp9bBZmZmAeeHBw3osCBMMwO2BGXR_UXjPL_-T6"
      },
      "balance": 31048693850429
    },
    {
      "wallet_address": {
        "hex": "0:924603ba9c5ef80bf0b7601de78a540db04f583d13ba67181f47dffa20e746a2",
        "base64": "EQCSRgO6nF74C_C3YB3nilQNsE9YPRO6ZxgfR9_6IOdGokaI"
      },
      "owner_address": {
        "hex": "0:574a5f8046c7a5f49f4d796daf6d4b2db32e0102ba2c014a6e50fc461b57951d",
        "base64": "EQBXSl-ARsel9J9NeW2vbUstsy4BArosAUpuUPxGG1eVHbDu"
      },
      "balance": 17494042821858
    },
    {
      "wallet_address": {
        "hex": "0:430f795c88715f23acda8a2f85b26d9f144c0fcf5b66e401984d8c59631692e4",
        "base64": "EQBDD3lciHFfI6zaii-Fsm2fFEwPz1tm5AGYTYxZYxaS5LqE"
      },
      "owner_address": {
        "hex": "0:bcdc2501992f392f58c1ce1815d01c1d2bdbe589e766103fcc4acba9d534ef43",
        "base64": "EQC83CUBmS85L1jBzhgV0BwdK9vliedmED_MSsup1TTvQ1rI"
      },
      "balance": 5663947335295
    }
  ]
}
```

## AggregateAccountsHistory

Returns time series for a given metric.
The filter can be set for an interface or minter address.

### Endpoint: `/accounts/aggregated/history`

### Request

```shell
# number of active address on each day for first 10 days of 2023 year
curl -X GET 'https://anton.tools/api/v0/accounts/aggregated/history?metric=active_addresses&from=2023-01-01T00%3A00%3A00Z&to=2023-01-11T00%3A00%3A00Z&interval=24h'
# number of active wallets of Lavandos jetton for each day
curl -X GET 'https://anton.tools/api/v0/accounts/aggregated/history?metric=active_addresses&minter_address=EQBl3gg6AAdjgjO2ZoNU5Q5EzUIl8XMNZrix8Z5dJmkHUfxI&interval=24h'
# number of active NFT items for each day
curl -X GET 'https://anton.tools/api/v0/accounts/aggregated/history?metric=active_addresses&interface=nft_item&interval=24h'
```

### Response

```json
{
  "count_results": [
    {
      "Value": 79,
      "Timestamp": "2022-12-24T00:00:00Z"
    },
    {
      "Value": 280,
      "Timestamp": "2022-12-25T00:00:00Z"
    },
    {
      "Value": 818,
      "Timestamp": "2022-12-26T00:00:00Z"
    },
    {
      "Value": 1462,
      "Timestamp": "2022-12-27T00:00:00Z"
    },
    {
      "Value": 515,
      "Timestamp": "2022-12-28T00:00:00Z"
    }
  ]
}
```

## GetTransactions

Returns filtered transactions, account states, messages and parsed data for each transaction.
The filter can be set by transaction address, hash, incoming message hash and workchain.

### Endpoint: `/transactions`

### Request

```shell
curl -X GET 'https://anton.tools/api/v0/transactions?address=EQBl3gg6AAdjgjO2ZoNU5Q5EzUIl8XMNZrix8Z5dJmkHUfxI&workchain=0&order=DESC&limit=1'
```

### Response

```json
{
  "total": 50,
  "results": [
    {
      "address": {
        "hex": "0:65de083a0007638233b6668354e50e44cd4225f1730d66b8b1f19e5d26690751",
        "base64": "EQBl3gg6AAdjgjO2ZoNU5Q5EzUIl8XMNZrix8Z5dJmkHUfxI"
      },
      "hash": "a6CrW7Dyk65NNRNi69psMZv0lNGN1ShzeMZf0FX++0U=",
      "account": {
        "is_active": true,
        "status": "ACTIVE",
        "balance": 240616906,
        "last_tx_lt": 36001780000007,
        "last_tx_hash": "a6CrW7Dyk65NNRNi69psMZv0lNGN1ShzeMZf0FX++0U=",
        "state_data": {
          "types": [
            "jetton_minter"
          ],
          "content_name": "Lavandos",
          "content_description": "This is a universal token for use in all areas of the decentralized Internet in the TON blockchain, web3, Telegram bots, TON sites. Issue of 4.6 billion coins. Telegram channels: Englishversion: @lave_eng Русскоязычная версия: @lavetoken, @lavefoundation, versión en español: @lave_esp, www.lavetoken.com",
          "content_image": "https://i.ibb.co/Bj5KqK4/IMG-20221213-115545-207.png",
          "total_supply": 4600000000000000000,
          "mintable": true,
          "admin_addr": {
            "hex": "0:0000000000000000000000000000000000000000000000000000000000000000",
            "base64": "EQAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAM9c"
          }
        },
        "code_hash": "mg+Y3W+/Il7vgWXk5kQX7pMffuoABlNDnntdzcBkTNY=",
        "data_hash": "Hoys/NBmNLvFGvdK/0GlzPZV2ZUjXBaKyrtFx+q/uwk=",
        "get_method_hashes": [ 106029, 10, 103289 ],
        "updated_at": "2023-03-12T23:40:07Z"
      },
      "block_workchain": 0,
      "block_shard": -9223372036854776000,
      "block_seq_no": 33606182,
      "prev_tx_hash": "0JHX+bF9GP4QA+VPYDPBQwcc9D7i4uLw9Bext0tBUnc=",
      "prev_tx_lt": 35368059000003,
      "in_msg_hash": "J1plFWLNeyzVZnZUAC85sAJ+KYFb/gL21iVoyVzzxS8=",
      "in_msg": {
        "type": "INTERNAL",
        "hash": "J1plFWLNeyzVZnZUAC85sAJ+KYFb/gL21iVoyVzzxS8=",
        "src_address": {
          "hex": "0:067f10d4354d3d02fe1911e9a3018a8c371e1d710b00f304b6b46af5e8aa2e77",
          "base64": "EQAGfxDUNU09Av4ZEemjAYqMNx4dcQsA8wS2tGr16Koud43P"
        },
        "dst_address": {
          "hex": "0:65de083a0007638233b6668354e50e44cd4225f1730d66b8b1f19e5d26690751",
          "base64": "EQBl3gg6AAdjgjO2ZoNU5Q5EzUIl8XMNZrix8Z5dJmkHUfxI"
        },
        "source_tx_hash": "8/cBWTYy2E+UnIWwggmv7lJTegXt9q/Tr/rpCB6JRhA=",
        "source_tx_lt": 36001780000005,
        "bounce": false,
        "bounced": false,
        "amount": 4809163,
        "ihr_disabled": true,
        "ihr_fee": 0,
        "fwd_fee": 666672,
        "body": "te6cckEBAQEADgAAGNUydtsAAAAAAAAAAPfBmNw=",
        "body_hash": "la37uGspaEdb70INRgaUtfLiBGlqXsrLMlKwhwvngQY=",
        "operation_id": 3576854235,
        "created_at": "2023-03-12T23:40:07Z",
        "created_lt": 36001780000006
      },
      "in_amount": 4809163,
      "out_msg_count": 0,
      "out_amount": 0,
      "total_fees": 3332042,
      "state_update": "te6cckEBAQEAQwAAgnIt4Gk2psKI3LyTTJhHH5cPqoQEGzezoxTQbHRaElXfEagWHRG7gq6d1aaQJ1Z8EZ4Z8aq7zHvrhnDnbOMO8HIIHLFJnw==",
      "description": "te6cckEBAgEAYAABGQzE0HaI0lhy0GPyvgkBAJxBAshLJAAAAf/+AAAAQQAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAABXYNxb",
      "orig_status": "ACTIVE",
      "end_status": "ACTIVE",
      "created_at": "2023-03-12T23:40:07Z",
      "created_lt": 36001780000007
    }
  ]
}
```

## AggregateTransactionsHistory

Returns time series for a given metric.
The filter can be set by addresses or workchain.

### Endpoint: `/transactions/aggregated/history`

### Request

```shell
# count transactions on telegram usernames collection on each day
curl -X GET 'https://anton.tools/api/v0/transactions/aggregated/history?metric=transaction_count&address=EQCA14o1-VWhS2efqoh_9M1b_A9DtKTuoqfmkn83AbJzwnPi&interval=24h'
```

### Response

```json
{
  "count_results": [
    {
      "Value": 283,
      "Timestamp": "2022-11-19T00:00:00Z"
    },
    {
      "Value": 427,
      "Timestamp": "2022-11-20T00:00:00Z"
    },
    {
      "Value": 223,
      "Timestamp": "2022-11-21T00:00:00Z"
    }
  ]
}
```

## GetMessages

Returns filtered messages theirs parsed data.
The filter can be set by addresses, minter address, contract interfaces and operation names.
With minter address set, you can view all messages on all NFT items in a given NFT collection or all jetton wallets.

### Endpoint: `/messages`

### Request

```shell
# show all telegram username transfers
curl -X GET 'https://anton.tools/api/v0/messages?operation_name=nft_item_transfer&minter_address=EQCA14o1-VWhS2efqoh_9M1b_A9DtKTuoqfmkn83AbJzwnPi&order=DESC&limit=1'
# show all Lavandos jetton transfers
curl -X GET 'https://anton.tools/api/v0/messages?operation_name=jetton_transfer&minter_address=EQBl3gg6AAdjgjO2ZoNU5Q5EzUIl8XMNZrix8Z5dJmkHUfxI&order=DESC&limit=1'
```

### Response

```json
{
  "total": 4469,
  "results": [
    {
      "type": "INTERNAL",
      "hash": "bgghlVNjVIIxDeywW5BDHzyRARxm7Wy3ct41TzdMPFw=",
      "src_address": {
        "hex": "0:4503985e1bbcb8fe8f7fb4bde36b8124bf6e1dbb504f7d7d239002b5686370c6",
        "base64": "EQBFA5heG7y4_o9_tL3ja4Ekv24du1BPfX0jkAK1aGNwxuIg"
      },
      "dst_address": {
        "hex": "0:f0967a6cce456cf182b2b4c2ef5cf6674715a0c0023a50f8a431875db5e4a96a",
        "base64": "EQDwlnpszkVs8YKytMLvXPZnRxWgwAI6UPikMYddteSpau3R"
      },
      "source_tx_hash": "FmewZ5DjjZMYWSxF8EBwvibssjDL984AK4UVHH3r9Ms=",
      "source_tx_lt": 36415940000001,
      "bounce": true,
      "bounced": false,
      "amount": 66844722,
      "ihr_disabled": true,
      "ihr_fee": 0,
      "fwd_fee": 1173343,
      "body": "te6cckEBAQEAVQAApV/MPRQAAAAAAAAAAIAFGLh2uaY6rfzuALJbmpcZTWR/gaXOoXn9krmJMF/UXrABFA5heG7y4/o9/tL3ja4Ekv24du1BPfX0jkAK1aGNwxhzEtAIrWp/+A==",
      "body_hash": "ETsP1cVj0O7Wyq6echVZrSTOyqjNrzUpn8s404g2pns=",
      "operation_id": 1607220500,
      "payload": {
        "dst_contract": "nft_item",
        "operation_name": "nft_item_transfer",
        "data": {
          "NewOwner": "EQAoxcO1zTHVb-dwBZLc1LjKayP8DS51C8_slcxJgv6i9bSB",
          "ForwardAmount": "10000000",
          "ResponseDestination": "EQBFA5heG7y4_o9_tL3ja4Ekv24du1BPfX0jkAK1aGNwxuIg"
        },
        "minter_address": {
          "hex": "0:80d78a35f955a14b679faa887ff4cd5bfc0f43b4a4eea2a7e6927f3701b273c2",
          "base64": "EQCA14o1-VWhS2efqoh_9M1b_A9DtKTuoqfmkn83AbJzwnPi"
        }
      },
      "created_at": "2023-03-29T00:26:22Z",
      "created_lt": 36415940000003
    }
  ]
}
```

## AggregateMessages

Returns statistics on messages on a given address. 
It counts messages and sums amount received from and sent to each address.

### Endpoint: `/messages/aggregated`

### Request

```shell
curl -X GET 'https://anton.tools/api/v0/messages/aggregated?address=0%3A83dfd552e63729b472fcbcc8c45ebcc6691702558b68ec7527e1ba403a0f31a8&order_by=count&limit=3'
```

### Response

```json
{
  "received_count": 182,
  "received_ton_amount": 14175360552821,
  "sent_count": 54,
  "sent_ton_amount": 14215521645679012,
  "received_from_address": [
    {
      "sender": {
        "hex": "0:0000000000000000000000000000000000000000000000000000000000000000",
        "base64": "AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA"
      },
      "amount": 0,
      "count": 54
    },
    {
      "sender": {
        "hex": "0:2260dd494a8a96d917807b6f0dfec864e07281295c36b93ed5017412c5022dd8",
        "base64": "EQAiYN1JSoqW2ReAe28N_shk4HKBKVw2uT7VAXQSxQIt2DVO"
      },
      "amount": 61000000,
      "count": 7
    },
    {
      "sender": {
        "hex": "0:04f89eb5fe2f355dd30a99384409ea2e07687e704a53c44d5b559442fdd67921",
        "base64": "EQAE-J61_i81XdMKmThECeouB2h-cEpTxE1bVZRC_dZ5IbOA"
      },
      "amount": 60000,
      "count": 6
    }
  ],
  "sent_to_address": [
    {
      "receiver": {
        "hex": "0:7e81a5f3a00f3b08ae10ab02a543c66aef565bce303580a6a817e4b8005fb2c2",
        "base64": "EQB-gaXzoA87CK4QqwKlQ8Zq71ZbzjA1gKaoF-S4AF-ywjlJ"
      },
      "amount": 102600000000000,
      "count": 9
    },
    {
      "receiver": {
        "hex": "0:314afa005f740b53d1e0c149615730d2e0400593829a02f34f4e8a4f573f6856",
        "base64": "EQAxSvoAX3QLU9HgwUlhVzDS4EAFk4KaAvNPTopPVz9oVoFg"
      },
      "amount": 21100000000000,
      "count": 4
    },
    {
      "receiver": {
        "hex": "0:c3f1da8ecda8f8cd42bace224ea3f1b6971eaa7f54c492d4d190527b4f573f7c",
        "base64": "EQDD8dqOzaj4zUK6ziJOo_G2lx6qf1TEktTRkFJ7T1c_fPQb"
      },
      "amount": 159482000000000,
      "count": 3
    }
  ]
}
```

Returns time series for a given metric.
The filter is same as in message getter.

## AggregateMessagesHistory

### Endpoint: `/messages/aggregated/history`

### Request

```shell
# count telegram NFT username transfers by each day
curl -X GET 'https://anton.tools/api/v0/messages/aggregated/history?metric=message_count&operation_name=nft_item_transfer&minter_address=EQCA14o1-VWhS2efqoh_9M1b_A9DtKTuoqfmkn83AbJzwnPi&from=2023-03-01T00%3A00%3A00Z&to=2023-03-11T00%3A00%3A00Z&interval=24h'
```

### Response

```json
{
  "count_results": [
    {
      "Value": 57,
      "Timestamp": "2023-03-04T00:00:00Z"
    },
    {
      "Value": 79,
      "Timestamp": "2023-03-05T00:00:00Z"
    },
    {
      "Value": 95,
      "Timestamp": "2023-03-06T00:00:00Z"
    },
    {
      "Value": 109,
      "Timestamp": "2023-03-07T00:00:00Z"
    },
    {
      "Value": 76,
      "Timestamp": "2023-03-08T00:00:00Z"
    },
    {
      "Value": 119,
      "Timestamp": "2023-03-09T00:00:00Z"
    },
    {
      "Value": 123,
      "Timestamp": "2023-03-10T00:00:00Z"
    }
  ]
}
```

## GetBlocks

Returns filtered blocks. 
The filter can be set by workchain, shard or sequence number. 
It can also return all transactions with account states, messages and parsed payloads in resulted blocks if `with_transactions=true` parameter is set.

### Endpoint: `/blocks`

### Request

```shell
curl -X GET 'https://anton.tools/api/v0/blocks?workchain=-1&seq_no=27777772&with_transactions=true'
curl -X GET 'https://anton.tools/api/v0/blocks?workchain=-1&with_transactions=true&order=DESC&limit=3'
```

### Response

```json
{
  "total": 1,
  "results": [
    {
      "workchain": -1,
      "shard": -9223372036854775808,
      "seq_no": 27777772,
      "file_hash": "lUA19886qXi4I+2db0rZK5Trnk9LBvDThC/esTuXnko=",
      "root_hash": "cI9+U74z6CzMdlnmIC+hLFTga8NC/wt1bzw0czrlOG4=",
      "shards": [
        {
          "workchain": 0,
          "shard": -9223372036854775808,
          "seq_no": 33356549,
          "file_hash": "PZmFr6nL3BtaSuLEufCNDEIzrzY1XfErI9EHlyIgjvs=",
          "root_hash": "SQQCndNOXRKVjVliGrvwJSOuHGmUyqXXUZR36dlF6DA=",
          "master": {
            "workchain": -1,
            "shard": -9223372036854775808,
            "seq_no": 27777772
          },
          "transactions": [
            {
              "hash": "TR4zXnRuVclDrSEgHdq2KnPAJ/DyDZBuLafsheyr6bU=",
              "address": {
                "hex": "0:dfdd409cdb59a25868c8d16bfeeeaaf67ab37cd5b057e37117e8348238a1aeb5",
                "base64": "EQDf3UCc21miWGjI0Wv-7qr2erN81bBX43EX6DSCOKGutUtB"
              },
              "account": {
                "is_active": true,
                "status": "ACTIVE",
                "balance": 47470763,
                "last_tx_lt": 35745768000007,
                "last_tx_hash": "TR4zXnRuVclDrSEgHdq2KnPAJ/DyDZBuLafsheyr6bU=",
                "state_data": {
                  "types": [
                    "jetton_wallet"
                  ],
                  "owner_address": {
                    "hex": "0:d4fed4363f93d48238b828adc359700641455f1527c1cdbe81a40ad4d1cdb14b",
                    "base64": "EQDU_tQ2P5PUgji4KK3DWXAGQUVfFSfBzb6BpArU0c2xS9eg"
                  },
                  "minter_address": {
                    "hex": "0:4f0156ba7e3322831b271c5df7510ddabae7d5ae0d99f250594d8f51fa2b1f8c",
                    "base64": "EQBPAVa6fjMigxsnHF33UQ3auufVrg2Z8lBZTY9R-isfjIFr"
                  },
                  "updated_at": "2023-03-03T09:52:23Z"
                },
                "code_hash": "zScFSHsYSN1XMaS2UzYGxAejOtgpJSFpCK0sh8VDKUo=",
                "data_hash": "TkQR45Cp3hXfdi1y4RhqSiACtk3jbnO/g+44RBrZxds=",
                "get_method_hashes": [ 10, 11, 97026, 8, 9 ],
                "updated_at": "2023-03-03T09:52:23Z"
              },
              "prev_tx_hash": "Mx8zXcG8nh33Ow35W2OgWvqYIGP/qVQFS4ZBtukC9dU=",
              "prev_tx_lt": 35196543000003,
              "in_msg_hash": "lLrI1C3UKanmszG2QneolARNCtztRI9bPTBTRvS34C0=",
              "in_msg": {
                "type": "INTERNAL",
                "src_address": {
                  "hex": "0:4f0156ba7e3322831b271c5df7510ddabae7d5ae0d99f250594d8f51fa2b1f8c",
                  "base64": "EQBPAVa6fjMigxsnHF33UQ3auufVrg2Z8lBZTY9R-isfjIFr"
                },
                "dst_address": {
                  "hex": "0:dfdd409cdb59a25868c8d16bfeeeaaf67ab37cd5b057e37117e8348238a1aeb5",
                  "base64": "EQDf3UCc21miWGjI0Wv-7qr2erN81bBX43EX6DSCOKGutUtB"
                },
                "source_tx_hash": "q8tTEMWSD6fOB6NlZbLcr4Nu9OoiQZX2O+5l9NP3OV8=",
                "source_tx_lt": 35745768000005,
                "bounce": true,
                "bounced": false,
                "amount": 51000000,
                "ihr_disabled": true,
                "ihr_fee": 0,
                "fwd_fee": 6746052,
                "body": "te6cckEBAQEAOAAAaxeNRRmwM11hQ+832kO5rKACAGp/ahsfyepBHFwUVuGsuAMgoq+Kk+Dm30DSBWpo5tilmHoSAoyDqgM=",
                "body_hash": "q+ISQ0DYTqh6YAzdLwQ9T/Y+vo+hEinlFzsXA5xctyA=",
                "operation_id": 395134233,
                "state_init_code": "... etc ...",
                "state_init_data": "... etc ...",
                "created_at": "2023-03-03T09:52:23Z",
                "created_lt": 35745768000006
              },
              "in_amount": 51000000,
              "out_msg": [
                {
                  "type": "INTERNAL",
                  "hash": "2n2P9xWpxwocWawfLwjdFdpZGTJfuTNSRIHjakS4QoU=",
                  "src_address": {
                    "hex": "0:dfdd409cdb59a25868c8d16bfeeeaaf67ab37cd5b057e37117e8348238a1aeb5",
                    "base64": "EQDf3UCc21miWGjI0Wv-7qr2erN81bBX43EX6DSCOKGutUtB"
                  },
                  "dst_address": {
                    "hex": "0:d4fed4363f93d48238b828adc359700641455f1527c1cdbe81a40ad4d1cdb14b",
                    "base64": "EQDU_tQ2P5PUgji4KK3DWXAGQUVfFSfBzb6BpArU0c2xS9eg"
                  },
                  "source_tx_hash": "TR4zXnRuVclDrSEgHdq2KnPAJ/DyDZBuLafsheyr6bU=",
                  "source_tx_lt": 35745768000007,
                  "bounce": false,
                  "bounced": false,
                  "amount": 1000000,
                  "ihr_disabled": true,
                  "ihr_fee": 0,
                  "fwd_fee": 823340,
                  "body": "te6cckEBAQEAEwAAIXNi0JywM11hQ+832kO5rKABHgmLcg==",
                  "body_hash": "D1tVS8p2LrFI9mVJSpjUokTsqqI7rWWObRWmaS0p1VA=",
                  "operation_id": 1935855772,
                  "payload": {
                    "type": "INTERNAL",
                    "hash": "2n2P9xWpxwocWawfLwjdFdpZGTJfuTNSRIHjakS4QoU=",
                    "src_contract": "jetton_wallet",
                    "amount": 1000000,
                    "body_hash": "D1tVS8p2LrFI9mVJSpjUokTsqqI7rWWObRWmaS0p1VA=",
                    "operation_id": 1935855772,
                    "operation_name": "jetton_transfer_notification",
                    "data": {
                      "Amount": "1000000000",
                      "Sender": "NONE",
                      "QueryID": 12696594446820521946,
                      "ForwardPayload": {}
                    },
                    "minter_address": {
                      "hex": "0:4f0156ba7e3322831b271c5df7510ddabae7d5ae0d99f250594d8f51fa2b1f8c",
                      "base64": "EQBPAVa6fjMigxsnHF33UQ3auufVrg2Z8lBZTY9R-isfjIFr"
                    }
                  },
                  "created_at": "2023-03-03T09:52:23Z",
                  "created_lt": 35745768000008
                },
                {
                  "type": "INTERNAL",
                  "hash": "gUgAO6zkgwluZGOnN42BEUGOhhGJOPDLuKphoe0T4Q8=",
                  "src_address": {
                    "hex": "0:dfdd409cdb59a25868c8d16bfeeeaaf67ab37cd5b057e37117e8348238a1aeb5",
                    "base64": "EQDf3UCc21miWGjI0Wv-7qr2erN81bBX43EX6DSCOKGutUtB"
                  },
                  "dst_address": {
                    "hex": "0:d4fed4363f93d48238b828adc359700641455f1527c1cdbe81a40ad4d1cdb14b",
                    "base64": "EQDU_tQ2P5PUgji4KK3DWXAGQUVfFSfBzb6BpArU0c2xS9eg"
                  },
                  "source_tx_hash": "TR4zXnRuVclDrSEgHdq2KnPAJ/DyDZBuLafsheyr6bU=",
                  "source_tx_lt": 35745768000007,
                  "bounce": false,
                  "bounced": false,
                  "amount": 32253948,
                  "ihr_disabled": true,
                  "ihr_fee": 0,
                  "fwd_fee": 666672,
                  "body": "te6cckEBAQEADgAAGNUydtuwM11hQ+832qU6+DU=",
                  "body_hash": "LpCrlJ+tCC8r5f/kWrw70+A+2eJra4L28IOPsvBKERQ=",
                  "operation_id": 3576854235,
                  "created_at": "2023-03-03T09:52:23Z",
                  "created_lt": 35745768000009
                }
              ],
              "out_msg_count": 2,
              "out_amount": 33253948,
              "total_fees": 11073124,
              "state_update": "te6cckEBAQEAQwAAgnJPlEk7gfd/iZVomnxQhqvtfJrgnhUKP+JjaAkFcUtRf4zWEJArsPi1OdUditdkxtJ44czppybYedI8zeylY0pfzduuXg==",
              "description": "te6cckEBAwEAnAACGwTBuJAJAMKMsBhy1sERAgEAb8mRDTxMLXhwAAAAAAAEAAAAAAAEl7U3tVgSOUt7tDQW3YhjRWAsHORCfYlHHjFVaSuRghZA0DMMAJxE0qsc4AAAAAAAAAAAvAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAVRGyV",
              "orig_status": "ACTIVE",
              "end_status": "ACTIVE",
              "created_at": "2023-03-03T09:52:23Z",
              "created_lt": 35745768000007
            }
          ]
        }
      ]
    }
  ]
}
```


