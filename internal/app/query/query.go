package query

import (
	"context"

	"github.com/pkg/errors"

	"github.com/tonindexer/anton/internal/app"
	"github.com/tonindexer/anton/internal/core"
	"github.com/tonindexer/anton/internal/core/aggregate"
	"github.com/tonindexer/anton/internal/core/aggregate/history"
	"github.com/tonindexer/anton/internal/core/filter"
	"github.com/tonindexer/anton/internal/core/repository"
	"github.com/tonindexer/anton/internal/core/repository/account"
	"github.com/tonindexer/anton/internal/core/repository/block"
	"github.com/tonindexer/anton/internal/core/repository/contract"
	"github.com/tonindexer/anton/internal/core/repository/msg"
	"github.com/tonindexer/anton/internal/core/repository/tx"
)

var _ app.QueryService = (*Service)(nil)

type Service struct {
	*app.QueryConfig

	contractRepo core.ContractRepository
	blockRepo    repository.Block
	txRepo       repository.Transaction
	msgRepo      repository.Message
	accountRepo  repository.Account
}

func NewService(_ context.Context, cfg *app.QueryConfig) (*Service, error) {
	var s = new(Service)

	s.QueryConfig = cfg
	ch, pg := s.DB.CH, s.DB.PG
	s.txRepo = tx.NewRepository(ch, pg)
	s.msgRepo = msg.NewRepository(ch, pg)
	s.blockRepo = block.NewRepository(ch, pg)
	s.accountRepo = account.NewRepository(ch, pg)
	s.contractRepo = contract.NewRepository(pg)

	return s, nil
}

func (s *Service) GetStatistics(ctx context.Context) (*aggregate.Statistics, error) {
	return aggregate.GetStatistics(ctx, s.DB.CH, s.DB.PG)
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

func (s *Service) GetLabelCategories(_ context.Context) ([]core.LabelCategory, error) {
	return []core.LabelCategory{core.Scam, core.CentralizedExchange}, nil
}

func (s *Service) FilterLabels(ctx context.Context, req *filter.LabelsReq) (*filter.LabelsRes, error) {
	return s.accountRepo.FilterLabels(ctx, req)
}

func (s *Service) FilterAccounts(ctx context.Context, req *filter.AccountsReq) (*filter.AccountsRes, error) {
	return s.accountRepo.FilterAccounts(ctx, req)
}

func (s *Service) AggregateAccounts(ctx context.Context, req *aggregate.AccountsReq) (*aggregate.AccountsRes, error) {
	return s.accountRepo.AggregateAccounts(ctx, req)
}

func (s *Service) AggregateAccountsHistory(ctx context.Context, req *history.AccountsReq) (*history.AccountsRes, error) {
	return s.accountRepo.AggregateAccountsHistory(ctx, req)
}

func (s *Service) FilterTransactions(ctx context.Context, req *filter.TransactionsReq) (*filter.TransactionsRes, error) {
	return s.txRepo.FilterTransactions(ctx, req)
}

func (s *Service) AggregateTransactionsHistory(ctx context.Context, req *history.TransactionsReq) (*history.TransactionsRes, error) {
	return s.txRepo.AggregateTransactionsHistory(ctx, req)
}

func (s *Service) FilterMessages(ctx context.Context, req *filter.MessagesReq) (*filter.MessagesRes, error) {
	if req.OperationID != nil && len(req.OperationNames) > 0 {
		return nil, errors.Wrap(core.ErrInvalidArg, "filter is available either on operation name or operation id")
	}
	return s.msgRepo.FilterMessages(ctx, req)
}

func (s *Service) AggregateMessages(ctx context.Context, req *aggregate.MessagesReq) (*aggregate.MessagesRes, error) {
	return s.msgRepo.AggregateMessages(ctx, req)
}

func (s *Service) AggregateMessagesHistory(ctx context.Context, req *history.MessagesReq) (*history.MessagesRes, error) {
	return s.msgRepo.AggregateMessagesHistory(ctx, req)
}
