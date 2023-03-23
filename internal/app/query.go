package app

import (
	"context"

	"github.com/iam047801/tonidx/internal/core"
	"github.com/iam047801/tonidx/internal/core/repository"
)

type QueryConfig struct {
	DB *repository.DB
}

type QueryService interface {
	GetStatistics(ctx context.Context) (*repository.Statistics, error)

	GetInterfaces(ctx context.Context) ([]*core.ContractInterface, error)
	GetOperations(ctx context.Context) ([]*core.ContractOperation, error)

	GetLastMasterBlock(ctx context.Context) (*core.Block, error)
	GetBlocks(ctx context.Context, filter *core.BlockFilter) (*core.BlockFiltered, error)

	GetAccountStates(ctx context.Context, filter *core.AccountStateFilter) (*core.AccountStateFiltered, error)
	AggregateAccountStates(ctx context.Context, req *core.AccountStateAggregate) (*core.AccountStateAggregated, error)

	GetTransactions(ctx context.Context, filter *core.TransactionFilter) (*core.TransactionFiltered, error)
	GetMessages(ctx context.Context, filter *core.MessageFilter) (*core.MessageFiltered, error)
}
