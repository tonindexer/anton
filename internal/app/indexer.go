package app

import (
	"time"

	"github.com/xssnick/tonutils-go/ton"

	"github.com/tonindexer/anton/internal/core/repository"
)

type IndexerConfig struct {
	DB *repository.DB

	API *ton.APIClient

	Fetcher FetcherService
	Parser  ParserService

	FromBlock        uint32
	FetchBlockPeriod time.Duration
}

type IndexerService interface {
	Start() error
	Stop()
}
