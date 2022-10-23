# tonidx

Project indexes data from TON blockchain: blocks, transactions, messages and contract data. 

## Overview

Known contract interfaces are initialized in `core/db/contract.go` and inserted to a database table `contract_interfaces`. 
Contract interfaces can be identified by some get methods. You can check any contract whether it has arbitrary get method 
using function `(*tlb.Account).HasGetMethod` of [tonutils-go](https://github.com/xssnick/tonutils-go) library. 

Contracts have acceptable messages. Each acceptable message can be represented as contract operation with 
operation id and message body schema. Usually operation id is the first 32 bits of message body [slice](https://ton.org/docs/#/func/types?id=atomic-types).

For example, let's look at FT and NFT standard tokens which can be found [here](https://github.com/ton-blockchain/token-contract/).
NFT item contract has one `get_nft_data` get method and two operations such as `transfer` 
with operation id = `0x5fcc3d14` and the following [schema](https://github.com/xssnick/tonutils-go/blob/0cf1be2f79276255f15e85a7274aba2d7f8fc52e/ton/nft/item.go#L14). 
So if arbitrary contract has these two get methods, we can try to determine operation id of
messages to this contract by parsing message body cell and getting the first 32 bits. If these 32 bits are equal to 
a known operation id, we can try to parse other message data with a schema (new owner of NFT item in this example).

### Project structure

| Folder            | Description                                                                         |
|-------------------|-------------------------------------------------------------------------------------|
| `core`            | contains project domain and all common interfaces                                   |
| `core/db`         | creates of db tables and inserts known ton contract interfaces                      |
| `core/repository` | database repositories                                                               |
|                   |                                                                                     |
| `app`             | contains all services interfaces and theirs configs                                 |
| `app/parser`      | a service to get readable transactions, messages, accounts and other TON structures |
| `app/indexer`     | a service to scan blocks and save data from `parser` to a database                  |
|                   |                                                                                     |
| `cmd`             | command line application and env parsers                                            |

### Reading docs
```shell
go install golang.org/x/tools/cmd/godoc
godoc -http=localhost:6060
```

## Installation

[//]: # (### docker)
[//]: # (```shell)
[//]: # (git clone https://github.com/iam047801/tonidx)
[//]: # (cd tonidx)
[//]: # (docker build -t indexer:latest .)
[//]: # (```)

### docker-compose
```shell
git clone https://github.com/iam047801/tonidx
cd tonidx
docker-compose build
```

## Configuration

### docker-compose
Docker compose installation requires some environment variables.
```shell
# Create .env file
cp .env.example .env
# Configure env
nano .env
```

### env

| Name         | Description                       | Default  | Example  |
|--------------|-----------------------------------|----------|----------|
| `DB_NAME`    | Database name                     |          | idx      |
| `FROM_BLOCK` | Master chain seq_no to start from | 22222022 | 23532000 |

## Starting

[//]: # (### docker)
[//]: # (```shell)
[//]: # (docker run -d -n indexer-service --env-file .env indexer:latest)
[//]: # (```)

### docker-compose
```shell
docker-compose up -d
docker-compose logs -f # reading logs
```

## Using

### Inserting contract interface

```shell
docker-compose exec indexer tonidx addInterface               \ 
    -name       [unique contract name]                        \
    -address    [optional contract address]                   \
    -code       [optional contract code]                      \
    -getmethods [optional get methods separated with commas]
```

### Inserting contract operation

```shell
docker-compose exec indexer tonidx addOperation   \ 
    -name     [operation name]                    \
    -contract [contract interface name]           \
    -opid     [operation id, example: 0x5fcc3d14] \
    -schema   [message body schema]
```

### Message body schema example
```json
[
    {
        "Name": "OperationID",
        "Type": "tlbMagic",
        "Tag": "tlb:\"#5fcc3d14\""
    },
    {
        "Name": "QueryID",
        "Type": "uint64",
        "Tag": "tlb:\"## 64\""
    },
    {
        "Name": "NewOwner",
        "Type": "address",
        "Tag": "tlb:\"addr\""
    },
    {
        "Name": "ResponseDestination",
        "Type": "address",
        "Tag": "tlb:\"addr\""
    }
]
```

### Connecting to the clickhouse

```shell
docker-compose exec clickhouse clickhouse-client
```

### Query examples (do not work well)

Get all nft sales with price more than 5 TON:

```clickhouse
SELECT hex(tx_hash), src_addr, dst_addr, amount / 1e9 AS ton
FROM messages
WHERE dst_addr IN (
    SELECT any(address)
    FROM accounts
    WHERE has(types, 'nft_sale')
    GROUP BY address
) AND ton > 5
```

Get amount of unique wallets with given versions:

```clickhouse
SELECT count() FROM (
    SELECT any(address)
    FROM accounts
    WHERE hasAny(types, ['wallet_V3R1', 'wallet_V3R2', 'wallet_V4R1', 'wallet_V4R2'])
    GROUP BY address
)
```

Count rich addresses with balance more than 100k TON:

```clickhouse
SELECT count() FROM (
    SELECT any(address), max(balance)
    FROM accounts
    WHERE hasAny(types, ['wallet_V3R1', 'wallet_V3R2', 'wallet_V4R1', 'wallet_V4R2'])
        AND balance > 100 * 1e3 * 1e9
    GROUP BY address
)
```

Select TOP 25 rich accounts, theirs types and balance:

```clickhouse
SELECT any(address), any(types), max(ton) FROM (
    SELECT address, types, balance / 1e9 AS ton
    FROM accounts
    WHERE last_tx_lt > 31180269000003
    ORDER BY balance DESC
    LIMIT 100
) GROUP BY address ORDER BY max(ton) DESC LIMIT 25
```
