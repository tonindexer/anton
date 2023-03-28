package query

import (
	"context"

	"github.com/iam047801/tonidx/internal/app"
	"github.com/iam047801/tonidx/internal/core"
	"github.com/iam047801/tonidx/internal/core/aggregate"
	"github.com/iam047801/tonidx/internal/core/filter"
	"github.com/iam047801/tonidx/internal/core/repository"
	"github.com/iam047801/tonidx/internal/core/repository/account"
	"github.com/iam047801/tonidx/internal/core/repository/block"
	"github.com/iam047801/tonidx/internal/core/repository/contract"
	"github.com/iam047801/tonidx/internal/core/repository/msg"
	"github.com/iam047801/tonidx/internal/core/repository/tx"
)

var _ app.QueryService = (*Service)(nil)

type Service struct {
	cfg *app.QueryConfig

	contractRepo core.ContractRepository
	blockRepo    repository.Block
	txRepo       repository.Transaction
	msgRepo      repository.Message
	accountRepo  repository.Account
}

func NewService(_ context.Context, cfg *app.QueryConfig) (*Service, error) {
	var s = new(Service)

	s.cfg = cfg
	ch, pg := s.cfg.DB.CH, s.cfg.DB.PG
	s.txRepo = tx.NewRepository(ch, pg)
	s.msgRepo = msg.NewRepository(ch, pg)
	s.blockRepo = block.NewRepository(ch, pg)
	s.accountRepo = account.NewRepository(ch, pg)
	s.contractRepo = contract.NewRepository(pg)

	return s, nil
}

func (s *Service) GetStatistics(ctx context.Context) (*aggregate.Statistics, error) {
	return aggregate.GetStatistics(ctx, s.cfg.DB.CH, s.cfg.DB.PG)
}

func (s *Service) GetInterfaces(ctx context.Context) ([]*core.ContractInterface, error) {
	return s.contractRepo.GetInterfaces(ctx)
}

func (s *Service) GetOperations(ctx context.Context) ([]*core.ContractOperation, error) {
	return s.contractRepo.GetOperations(ctx)
}

func (s *Service) FilterBlocks(ctx context.Context, req *filter.BlocksReq) (*filter.BlocksRes, error) {
	return s.blockRepo.FilterBlocks(ctx, req)
}

func (s *Service) FilterAccounts(ctx context.Context, req *filter.AccountsReq) (*filter.AccountsRes, error) {
	return s.accountRepo.FilterAccounts(ctx, req)
}

func (s *Service) AggregateAccounts(ctx context.Context, req *aggregate.AccountsReq) (*aggregate.AccountsRes, error) {
	return s.accountRepo.AggregateAccounts(ctx, req)
}

func (s *Service) FilterTransactions(ctx context.Context, req *filter.TransactionsReq) (*filter.TransactionsRes, error) {
	return s.txRepo.FilterTransactions(ctx, req)
}

func (s *Service) FilterMessages(ctx context.Context, req *filter.MessagesReq) (*filter.MessagesRes, error) {
	return s.msgRepo.FilterMessages(ctx, req)
}

func (s *Service) AggregateMessages(ctx context.Context, req *aggregate.MessagesReq) (*aggregate.MessagesRes, error) {
	return s.msgRepo.AggregateMessages(ctx, req)
}
