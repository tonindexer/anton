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
	GetBlocks(ctx context.Context, filter *core.BlockFilter, offset, limit int) ([]*core.Block, error)

	GetAccountStates(ctx context.Context, filter *core.AccountStateFilter, offset, limit int) ([]*core.AccountState, error)

	GetTransactions(ctx context.Context, filter *core.TransactionFilter, offset, limit int) ([]*core.Transaction, error)
	GetMessages(ctx context.Context, filter *core.MessageFilter, offset, limit int) ([]*core.Message, error)
}
