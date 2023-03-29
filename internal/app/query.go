package app

import (
	"context"

	"github.com/tonindexer/anton/internal/core"
	"github.com/tonindexer/anton/internal/core/aggregate"
	"github.com/tonindexer/anton/internal/core/aggregate/history"
	"github.com/tonindexer/anton/internal/core/filter"
	"github.com/tonindexer/anton/internal/core/repository"
)

type QueryConfig struct {
	DB *repository.DB
}

type QueryService interface {
	GetStatistics(ctx context.Context) (*aggregate.Statistics, error)

	GetInterfaces(ctx context.Context) ([]*core.ContractInterface, error)
	GetOperations(ctx context.Context) ([]*core.ContractOperation, error)

	filter.BlockRepository
	filter.AccountRepository
	filter.TransactionRepository
	filter.MessageRepository

	aggregate.AccountRepository
	aggregate.MessageRepository

	history.AccountRepository
	history.TransactionRepository
	history.MessageRepository
}
