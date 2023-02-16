# tonidx

Project fetches data from TON blockchain and put it in PostgreSQL and ClickHouse databases. 

## Overview

Contract interfaces can be identified by some get methods.
You can check any contract whether it has an arbitrary get method by parsing contract code.
Also contracts have acceptable incoming messages (internal or external).
Each acceptable message can be represented as a contract operation with operation id and message body schema. 
Usually operation id is the first 32 bits of message body. 
If message is a regular TON transfer, operation id is equal to zero.
Then contract also can have formalized outgoing messages, each with specific operation id.

For example, let's look at FT and NFT standard tokens which can be found [here](https://github.com/ton-blockchain/token-contract/).
NFT item contract has one `get_nft_data` get method and two incoming [operations](https://github.com/ton-blockchain/token-contract/blob/main/nft/op-codes.fc) 
(`transfer` with operation id = `0x5fcc3d14`, `get_static_data` with operation id = `0x2fcb26a2`). 
Transfer payload has the following [schema](https://github.com/xssnick/tonutils-go/blob/master/ton/nft/item.go#L14). 
So if arbitrary contract has these two get methods, we try to determine operation id of
messages to (or from) this contract by parsing message body cell and getting first 32 bits.
If these 32 bits are equal to a known operation id (suppose `0x5fcc3d14`), we try to parse other message data with the known schema (new owner of NFT item).

Known contract interfaces are initialized in `core/repository/known.go` and inserted to a database table `contract_interfaces`.

Go to [MODELS.md](/MODELS.md) to get more detailed description of models used in this project and contracts known to this project.

Go to [data_tree.json](data_tree.json) to see an example of a real parsed masterchain block.

### Project structure

| Folder            | Description                                                            |
|-------------------|------------------------------------------------------------------------|
| `core`            | contains project domain and all common interfaces                      |
| `core/repository` | database repositories                                                  |
|                   |                                                                        |
| `app`             | contains all services interfaces and theirs configs                    |
| `app/parser`      | service to parse contract data and message payloads to known contracts |
| `app/indexer`     | a service to scan blocks and save data from `parser` to a database     |
|                   |                                                                        |
| `cmd`             | command line application and env parsers                               |

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

| Name          | Description                       | Default  | Example                                                           |
|---------------|-----------------------------------|----------|-------------------------------------------------------------------|
| `DB_NAME`     | Database name                     |          | idx                                                               |
| `DB_USER`     | Database username                 |          | user                                                              |
| `DB_PASSWORD` | Database password                 |          | pass                                                              |
| `FROM_BLOCK`  | Master chain seq_no to start from | 22222022 | 23532000                                                          |
| `LITESERVERS` | Lite servers to connect to        |          | 135.181.177.59:53312 aF91CuUHuuOv9rm2W5+O/4h38M3sRm40DtSdRxQhmtQ= |

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
