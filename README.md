# tonidx

The project fetches data from the TON blockchain and puts it in PostgreSQL and ClickHouse databases. 

## Overview

Before you start, take a look at the [official docs](https://ton.org/docs/learn/overviews/ton-blockchain).

Consider an arbitrary contract.
It has a state that is updated with any transaction on the contract's account.
This state contains the contract data.
The contract data can be complex,
but developers usually provide [get-methods](https://ton.org/docs/develop/func/functions#specifiers) in the contract.
You can retrieve data by executing these methods and possibly passing them arguments.
By parsing the contract code, you can check any contract for an arbitrary get-method (identified by function name).

TON has some standard tokens, such as
[TEP-62](https://github.com/ton-blockchain/TEPs/blob/master/text/0062-nft-standard.md),
[TEP-74](https://github.com/ton-blockchain/TEPs/blob/master/text/0074-jettons-standard.md).
Standard contracts have predefined get-method names and various types of acceptable incoming messages,
each with a different payload schema.
Standards also specify [tags](https://ton.org/docs/learn/overviews/tl-b-language#constructors) (or operation ids)
as the first 32 bits of the parsed message payload cell.
Therefore, you can attempt to match accounts found in the network to the standards by checking the presence of the get-methods and
matching found messages to these accounts by parsing the first 32 bits of the message payload.

For example, look at NFT standard tokens, which can be found [here](https://github.com/ton-blockchain/token-contract).
NFT item contract has one `get_nft_data` get method and two incoming [messages](https://github.com/ton-blockchain/token-contract/blob/main/nft/op-codes.fc)
(`transfer` with an operation id = `0x5fcc3d14`, `get_static_data` with an operation id = `0x2fcb26a2`).
Transfer payload has the following [schema](https://github.com/xssnick/tonutils-go/blob/master/ton/nft/item.go#L14).
If an arbitrary contract has a `get_nft_data` method, we can parse the operation id of messages sent to and from this contract.
If the operation id matches a known id, such as `0x5fcc3d14`, we attempt to parse the message data using the known schema
(new owner of NFT in the given example).

Known contract interfaces are initialized in [abi/known.go](/abi/known.go).

Go to [msg_schema.json](/docs/msg_schema.json) for an example of a message payload JSON schema.

Go to [API.md](/docs/API.md) to see working query examples.

### Project structure

| Folder            | Description                                                                      | 
|-------------------|----------------------------------------------------------------------------------|
| `abi`             | tlb cell parsing defined by json schema, known contract messages and get-methods |
|                   |                                                                                  |
| `api/http`        | JSON API documentation                                                           |
| `docs`            | database schemas used in this project, API query examples                        |
| `config`          | custom postgresql configuration                                                  |
|                   |                                                                                  |
| `core`            | contains project domain                                                          |
| `core/rndm`       | generation of random domain structures                                           |
| `core/filter`     | filters description                                                              |
| `core/aggregate`  | aggregation metrics description                                                  |
| `core/repository` | database repositories implementing filters and aggregation                       |
|                   |                                                                                  |
| `app`             | contains all services interfaces and theirs configs                              |
| `app/parser`      | service to parse contract data and message payloads to known contracts           | 
| `app/indexer`     | a service to scan blocks and save data from `parser` to a database               |
|                   |                                                                                  |
| `cmd`             | command line application and env parsers                                         |

## Starting it up

### Cloning repository

```shell
git clone https://github.com/tonindexer/anton tonidx
cd tonidx
docker-compose build
```

### Running tests

```shell
# run tests on abi package
go test -p 1 $(go list ./... | grep /abi) -covermode=count

# start databases up
docker-compose up -d postgres clickhouse
# run repositories tests
go test -p 1 $(go list ./... | grep /internal/core) -covermode=count
```

### Running linter

Firstly, install [`golangci-lint`](https://golangci-lint.run/usage/install/).

```shell
golangci-lint run 
```

### Configuration

Installation requires some environment variables.

```shell
cp .env.example .env
nano .env
```

| Name          | Description                       | Default  | Example                                                           |
|---------------|-----------------------------------|----------|-------------------------------------------------------------------|
| `DB_NAME`     | Database name                     |          | idx                                                               |
| `DB_USERNAME` | Database username                 |          | user                                                              |
| `DB_PASSWORD` | Database password                 |          | pass                                                              |
| `FROM_BLOCK`  | Master chain seq_no to start from | 22222022 | 23532000                                                          |
| `LITESERVERS` | Lite servers to connect to        |          | 135.181.177.59:53312 aF91CuUHuuOv9rm2W5+O/4h38M3sRm40DtSdRxQhmtQ= |
| `DEBUG_LOGS`  | Debug logs enabled                | false    | true                                                              |

### Running indexer and API

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
docker-compose exec indexer tonidx addOperation               \
    -name     [operation name]                                \
    -contract [contract interface name]                       \
    -opid     [operation id, example: 0x5fcc3d14]             \
    -schema   [message body schema]
```
