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
	GetInterfaces(context.Context) ([]*core.ContractInterface, error)
	GetOperations(ctx context.Context) ([]*core.ContractOperation, error)

	GetLastMasterBlock(ctx context.Context) (*core.Block, error)
	GetBlocks(ctx context.Context, filter *core.BlockFilter) ([]*core.Block, error)

	GetAccountStates(ctx context.Context, filter *core.AccountStateFilter) (*core.AccountStateFilterResults, error)

	GetTransactions(ctx context.Context, filter *core.TransactionFilter) ([]*core.Transaction, error)
	GetMessages(ctx context.Context, filter *core.MessageFilter) ([]*core.Message, error)
}
