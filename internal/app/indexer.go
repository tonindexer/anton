package app

import (
	"github.com/xssnick/tonutils-go/ton"

	"github.com/tonindexer/anton/internal/core/repository"
)

type IndexerConfig struct {
	DB *repository.DB

	API ton.APIClientWrapped

	Fetcher FetcherService
	Parser  ParserService

	FromBlock uint32
	Workers   int
}

type IndexerService interface {
	Start() error
	Stop()
}
