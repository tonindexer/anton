module github.com/iam047801/tonidx

go 1.18

replace github.com/uptrace/go-clickhouse v0.3.0 => github.com/iam047801/go-clickhouse v0.0.0-20230227133911-77a45625ed0b // branch with go-clickhouse dirty fixes

require (
	github.com/allisson/go-env v0.3.0
	github.com/iancoleman/strcase v0.2.0
	github.com/pkg/errors v0.9.1
	github.com/rs/zerolog v1.29.0
	github.com/sigurn/crc16 v0.0.0-20211026045750-20ab5afb07e3
	github.com/uptrace/bun v1.1.12
	github.com/uptrace/bun/dialect/pgdialect v1.1.12
	github.com/uptrace/bun/driver/pgdriver v1.1.12
	github.com/uptrace/bun/extra/bunbig v1.1.12
	github.com/uptrace/go-clickhouse v0.3.0
	github.com/xssnick/tonutils-go v1.5.2
)

require (
	github.com/codemodus/kace v0.5.1 // indirect
	github.com/jinzhu/inflection v1.0.0 // indirect
	github.com/mattn/go-colorable v0.1.13 // indirect
	github.com/mattn/go-isatty v0.0.17 // indirect
	github.com/oasisprotocol/curve25519-voi v0.0.0-20220328075252-7dd334e3daae // indirect
	github.com/pierrec/lz4/v4 v4.1.17 // indirect
	github.com/tmthrgd/go-hex v0.0.0-20190904060850-447a3041c3bc // indirect
	github.com/vmihailenco/msgpack/v5 v5.3.5 // indirect
	github.com/vmihailenco/tagparser/v2 v2.0.0 // indirect
	go.opentelemetry.io/otel v1.13.0 // indirect
	go.opentelemetry.io/otel/trace v1.13.0 // indirect
	golang.org/x/crypto v0.6.0 // indirect
	golang.org/x/exp v0.0.0-20230213192124-5e25df0256eb // indirect
	golang.org/x/sys v0.5.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
	mellium.im/sasl v0.3.1 // indirect
)
