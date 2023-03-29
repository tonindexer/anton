package app

import (
	"time"

	"github.com/tonindexer/anton/internal/core/repository"
)

type IndexerConfig struct {
	DB               *repository.DB
	Parser           ParserService
	FromBlock        uint32
	FetchBlockPeriod time.Duration
}

type IndexerService interface {
	Start() error
	Stop()
}
