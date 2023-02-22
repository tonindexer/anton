package query

import (
	"context"

	"github.com/iam047801/tonidx/internal/app"
	"github.com/iam047801/tonidx/internal/core"
	"github.com/iam047801/tonidx/internal/core/repository/account"
	"github.com/iam047801/tonidx/internal/core/repository/block"
	"github.com/iam047801/tonidx/internal/core/repository/contract"
	"github.com/iam047801/tonidx/internal/core/repository/tx"
)

var _ app.QueryService = (*Service)(nil)

type Service struct {
	cfg *app.QueryConfig

	contractRepo core.ContractRepository
	blockRepo    core.BlockRepository
	txRepo       core.TxRepository
	accountRepo  core.AccountRepository
}

func NewService(_ context.Context, cfg *app.QueryConfig) (*Service, error) {
	var s = new(Service)

	s.cfg = cfg
	ch, pg := s.cfg.DB.CH, s.cfg.DB.PG
	s.txRepo = tx.NewRepository(ch, pg)
	s.blockRepo = block.NewRepository(ch, pg)
	s.accountRepo = account.NewRepository(ch, pg)
	s.contractRepo = contract.NewRepository(pg)

	return s, nil
}

func (s *Service) GetInterfaces(ctx context.Context) ([]*core.ContractInterface, error) {
	return s.contractRepo.GetInterfaces(ctx)
}

func (s *Service) GetOperations(ctx context.Context) ([]*core.ContractOperation, error) {
	return s.contractRepo.GetOperations(ctx)
}

func (s *Service) GetLastMasterBlock(ctx context.Context) (*core.Block, error) {
	return s.blockRepo.GetLastMasterBlock(ctx)
}

func (s *Service) GetBlocks(ctx context.Context, filter *core.BlockFilter, offset, limit int) ([]*core.Block, error) {
	return s.blockRepo.GetBlocks(ctx, filter, offset, limit)
}

func (s *Service) GetAccountStates(ctx context.Context, filter *core.AccountStateFilter, offset, limit int) ([]*core.AccountState, error) {
	return s.accountRepo.GetAccountStates(ctx, filter, offset, limit)
}

func (s *Service) GetTransactions(ctx context.Context, filter *core.TransactionFilter, offset, limit int) ([]*core.Transaction, error) {
	return s.txRepo.GetTransactions(ctx, filter, offset, limit)
}

func (s *Service) GetMessages(ctx context.Context, filter *core.MessageFilter, offset, limit int) ([]*core.Message, error) {
	return s.txRepo.GetMessages(ctx, filter, offset, limit)
}
