package app

import (
	"github.com/tonindexer/anton/internal/core"
	"github.com/tonindexer/anton/internal/core/repository"
)

type RescanConfig struct {
	ContractRepo core.ContractRepository
	BlockRepo    repository.Block
	AccountRepo  repository.Account
	MessageRepo  repository.Message

	Parser ParserService

	Workers int

	SelectLimit int
}

type RescanService interface {
	Start() error
	Stop()
}
