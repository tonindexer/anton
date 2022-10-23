package app

import (
	"time"

	"github.com/uptrace/go-clickhouse/ch"
)

type IndexerConfig struct {
	DB               *ch.DB
	Parser           ParserService
	FromBlock        uint32
	FetchBlockPeriod time.Duration
}

type IndexerService interface {
	Start() error
	Stop()
}
