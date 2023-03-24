package app

import (
	"context"

	"github.com/iam047801/tonidx/internal/core"
	"github.com/iam047801/tonidx/internal/core/aggregate"
	"github.com/iam047801/tonidx/internal/core/filter"
	"github.com/iam047801/tonidx/internal/core/repository"
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
}
