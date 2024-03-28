# Anton

This project is an open-source tool that extracts and organizes data from the TON blockchain, 
efficiently storing it in PostgreSQL and ClickHouse databases. 

## Overview

Before you start, take a look at the [official docs](https://ton.org/docs/learn/overviews/ton-blockchain).

Consider an arbitrary contract.
It has a state that is updated with any transaction on the contract's account.
Each state has the contract code and data.
The contract data can be complex, but developers typically provide [get-methods](https://ton.org/docs/develop/func/functions#specifiers) in the contract, which can be executed to retrieve the necessary data.
The TON has standard contracts (such as [TEP-62](https://github.com/ton-blockchain/TEPs/blob/master/text/0062-nft-standard.md), [TEP-74](https://github.com/ton-blockchain/TEPs/blob/master/text/0074-jettons-standard.md)), and they have predefined get-method names. 
Therefore, you can attempt to match accounts found in the network to these standards by checking the presence of the get-methods.
Contract standards also specify [TL-B constructor tags](https://docs.ton.org/develop/data-formats/tl-b-language#constructors) (or operation ids) for each acceptable message to contract, defined as the first 32 bits of the parsed message payload cell.
So you if you know standard of a given contract, you can determine the type of message to it (for example, NFT item transfer) by parsing the first 32 bits of message body. 

Anton allows you to define the contract interface in just one JSON schema. 
Format of every schema is described in detail in [abi/README.md](abi/README.md). 
Every schema comprises contract get-methods, as well as incoming and outgoing message schemas for the contract.
Once contract interfaces are defined and stored in the database, Anton begins scanning new blocks on the network.
The tool stores every account state, transaction, and message in the database.
For get-methods without arguments in the contract interface, Anton emulates these methods and saves the returned values to the database. 
When a message is sent to a known contract interface, Anton attempts to match the message to a known schema by comparing the parsed operation ID. 
If the message is successfully parsed using the identified schema, Anton also stores the parsed data.

To explore contract interfaces known to this project, visit the [abi/known](/abi/known) directory. 
This will provide you with an understanding of the various contract interfaces already supported and serve as examples for adding your own.

Currently, Anton offers a REST API for retrieving filtered and aggregated data from the databases. To see example queries, refer to the [API.md](/docs/API.md) file.

To explore how Anton stores data, visit the [migrations' directory](/migrations).

### Project structure

| Folder       | Description                                       | 
|--------------|---------------------------------------------------|
| `abi`        | get-methods and tlb cell parsing                  |
| `abi/known`  | contract interfaces known to this project         |
| `api/http`   | JSON API Swagger documentation                    |
| `docs`       | only API query examples for now                   |
| `config`     | custom postgresql configuration                   |
| `migrations` | database migrations                               |
| `internal`   | database repositories and services implementation |

### Internal directory structure

| Folder            | Description                                                                      |
|-------------------|----------------------------------------------------------------------------------|
| `core`            | contains project domain                                                          |
| `core/rndm`       | generates random domain structures                                               |
| `core/filter`     | describes filters                                                                |
| `core/aggregate`  | describes aggregation metrics                                                    |
| `core/repository` | implements database repositories with filters and aggregation                    |
| `app`             | contains all services interfaces and their configs                               |
| `app/parser`      | service determines contract interfaces, parse contract data and message payloads | 
| `app/fetcher`     | service concurrently fetches data from blockchain                                | 
| `app/indexer`     | service scans blocks and save parsed data to databases                           |
| `app/rescan`      | service parses data by updated contract description                              |
| `app/query`       | service aggregates database repositories                                         |
| `api/http`        | implements the REST API                                                          |

## Starting it up

### Cloning repository

```shell
git clone https://github.com/tonindexer/anton
cd anton
```

### Running tests

Run tests on abi package:

```shell
go test -p 1 $(go list ./... | grep /abi) -covermode=count
```

Run repositories tests:

```shell
# start databases up
docker compose -f docker-compose.yml -f docker-compose.dev.yml up -d postgres clickhouse

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

| Name                  | Description                        | Default | Example                                                            |
|-----------------------|------------------------------------|---------|--------------------------------------------------------------------|
| `DB_NAME`             | Database name                      |         | idx                                                                |
| `DB_USERNAME`         | Database username                  |         | user                                                               |
| `DB_PASSWORD`         | Database password                  |         | pass                                                               |
| `DB_CH_URL`           | Clickhouse URL to connect to       |         | clickhouse://clickhouse:9000/db_name?sslmode=disable               |
| `DB_PG_URL`           | PostgreSQL URL to connect to       |         | postgres://username:password@postgres:5432/db_name?sslmode=disable |
| `FROM_BLOCK`          | Master chain seq_no to start from  | 1       | 23532000                                                           |
| `WORKERS`             | Number of indexer workers          | 4       | 8                                                                  |
| `RESCAN_WORKERS`      | Number of rescan workers           | 4       | 8                                                                  |
| `RESCAN_SELECT_LIMIT` | Number of rows to fetch for rescan | 3000    | 1000                                                               |
| `LITESERVERS`         | Lite servers to connect to         |         | 135.181.177.59:53312 aF91CuUHuuOv9rm2W5+O/4h38M3sRm40DtSdRxQhmtQ=  |
| `DEBUG_LOGS`          | Debug logs enabled                 | false   | true                                                               |

### Building

```shell
# building it locally
go build -o anton .

# build local docker container via docker cli
docker build -t anton:latest .
# or via compose
docker compose -f docker-compose.yml -f docker-compose.dev.yml build

# pull public images
docker compose pull
```

### Running

We have several options for compose run via [override files](https://docs.docker.com/compose/extends/#multiple-compose-files):
* base (docker-compose.yml) - allows to run services with near default configuration;
* dev (docker-compose.dev.yml) - allows to rebuild Anton image locally and exposes databases ports;
* prod (docker-compose.prod.yml) - allows to configure and backup databases, requires at least 128GB RAM.

You can combine it by your own. Also, there are optional [profiles](https://docs.docker.com/compose/profiles/):
* migrate - runs optional migrations service.

Take a look at the following run examples:
```shell
# run base compose
docker compose up -d

# run dev compose (build docker image locally)
docker compose -f docker-compose.yml -f docker-compose.dev.yml up -d

# run prod compose
# WARNING: requires at least 128GB RAM
docker compose -f docker-compose.yml -f docker-compose.prod.yml up -d
```

To run Anton, you need at least one defined contract interface.
There are some known interfaces in the [abi/known](/abi/known) directory.
You can add them through this command:
```shell
docker compose exec rescan sh -c "anton contract addInterfaces /var/anton/known/*.json"
```

### Database schema migration

```shell
# run migrations service on running compose
docker compose run migrations
```

### Reading logs

```shell
docker compose logs -f
```

### Taking a backup

```shell
# starting up databases and API service
docker compose                      \
    -f docker-compose.yml           \
    -f docker-compose.prod.yml      \
        up -d postgres clickhouse web

# stop indexer
docker compose stop indexer

# create backup directories
mkdir backups backups/pg backups/ch

# backing up postgres
docker compose exec postgres pg_dump -U user db_name | gzip > backups/pg/1.pg.backup.gz

# backing up clickhouse (available only with docker-compose.prod.yml)
## connect to the clickhouse
docker compose exec clickhouse clickhouse-client
## execute backup command
# :) BACKUP DATABASE default TO File('/backups/1/');

# execute migrations through API service
docker compose exec web anton migrate up

# start up indexer
docker compose                      \
    -f docker-compose.yml           \
    -f docker-compose.prod.yml      \
        up -d indexer
```

## Using

### Showing archive nodes from global config

```shell
docker run tonindexer/anton archive [--testnet]
```

### Inserting contract interface

To add interfaces, you need to provide Anton with a contract description. 
It will select any interfaces not already present in the database, 
insert them, and initiate rescan tasks for messages and account states.

```shell
# add from stdin
cat abi/known/tep81_dns.json | docker compose exec -T web anton contract addInterfaces --stdin
# add from file
docker compose exec web anton contract addInterfaces "/var/anton/known/tep81_dns.json"
```

### Deleting contract interface

To delete an interface, provide a contract description along with the specific contract name you wish to remove. 
Anton will then delete the contract interface and its associated operations from the database 
and initiate rescan tasks to remove all parsed data related to this interface from messages and account states.  

```shell
docker compose exec rescan sh -c "anton contract deleteInterface -c nft_item /var/anton/known/*.json"
```

### Updating contract interface

To update a contract interface, you need to provide both the contract description 
and the specific name of the contract you're updating. 
Anton will then compare the provided contract interface description against the existing interface in the database. 
If there are any differences, Anton initiates rescan tasks to reparse data and fix these changes. 
This process may involve adding, deleting, or updating get-methods and contract operations.

```shell
docker compose exec rescan sh -c "anton contract updateInterface -c telemint_nft_item /var/anton/known/telemint.json"
```

### Adding address label

```shell
docker compose exec web anton label "EQDj5AA8mQvM5wJEQsFFFof79y3ZsuX6wowktWQFhz_Anton" "anton.tools"

# known tonscan labels
docker compose exec web anton label --tonscan
```
