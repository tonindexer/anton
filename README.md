# tonidx

Project fetches data from TON blockchain and put it in PostgreSQL and ClickHouse databases. 

## Overview

Before you start, take a look at [official docs](https://ton.org/docs/learn/overviews/ton-blockchain).

Consider an arbitrary contract.
It has its own state that is updated with any transaction that occurs on the contract's account. 
This state contains the contract data.
The contract data can be in a complex format, 
but developers usually provide [get-methods](https://ton.org/docs/develop/func/functions#specifiers) in the contract. 
By executing these methods and possibly passing them arguments, you can retrieve data.
You can check any contract for the presence of an arbitrary get-method (identified by function name) by parsing the contract code.

TON has some standard tokens, such as
[TEP-62](https://github.com/ton-blockchain/TEPs/blob/master/text/0062-nft-standard.md),
[TEP-74](https://github.com/ton-blockchain/TEPs/blob/master/text/0074-jettons-standard.md).
Standard contracts have predefined get-method names and various types of acceptable incoming messages, 
each with a different payload schema.
Standards also specify [tags](https://ton.org/docs/learn/overviews/tl-b-language#constructors) (or operation ids) 
as the first 32 bits of the parsed message payload cell.
Therefore, you can attempt to match accounts found in the network to the standards by checking for the presence of the get-methods and 
matching found messages to these accounts by parsing the first 32 bits of the message payload.

For example, let's look at NFT standard tokens which can be found [here](https://github.com/ton-blockchain/token-contract).
NFT item contract has one `get_nft_data` get method and two incoming [messages](https://github.com/ton-blockchain/token-contract/blob/main/nft/op-codes.fc) 
(`transfer` with an operation id = `0x5fcc3d14`, `get_static_data` with an operation id = `0x2fcb26a2`). 
Transfer payload has the following [schema](https://github.com/xssnick/tonutils-go/blob/master/ton/nft/item.go#L14).
If an arbitrary contract has a `get_nft_data` method, we can parse the operation id of messages sent to and from this contract. 
If the operation id matches a known id, such as `0x5fcc3d14`, we attempt to parse the message data using the known schema 
(new owner of NFT in the given example).

Known contract interfaces are initialized in [abi/known.go](/abi/known.go).

Go to [MODELS.md](/MODELS.md) to get more detailed description of models used in this project and contracts known to this project.

Go to [data_tree.json](/data_tree.json) to see an example of real parsed masterchain block.

Go to [msg_schema.json](/msg_schema.json) to see an example of message payload schema defined by json.

### Project structure

| Folder            | Description                                                                      |
|-------------------|----------------------------------------------------------------------------------|
| `abi`             | tlb cell parsing defined by json schema, known contract messages and get-methods |
|                   |                                                                                  |
| `core`            | contains project domain and all common interfaces                                |
| `core/repository` | database repositories                                                            |
|                   |                                                                                  |
| `app`             | contains all services interfaces and theirs configs                              |
| `app/parser`      | service to parse contract data and message payloads to known contracts           |
| `app/indexer`     | a service to scan blocks and save data from `parser` to a database               |
|                   |                                                                                  |
| `cmd`             | command line application and env parsers                                         |

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
| `DEBUG_LOGS`  | Debug logs enabled                | false    | true                                                              |

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
