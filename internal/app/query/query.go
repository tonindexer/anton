package query

import (
	"context"

	"github.com/pkg/errors"
	"github.com/xssnick/tonutils-go/ton"

	"github.com/tonindexer/anton/abi"
	"github.com/tonindexer/anton/addr"
	"github.com/tonindexer/anton/internal/app"
	"github.com/tonindexer/anton/internal/app/fetcher"
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

func (s *Service) GetDefinitions(ctx context.Context) (map[abi.TLBType]abi.TLBFieldsDesc, error) {
	return s.contractRepo.GetDefinitions(ctx)
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

func (s *Service) fetchSkippedAccounts(ctx context.Context, req *filter.AccountsReq, res *filter.AccountsRes) error {
	if !req.LatestState {
		return nil // historical states are not available for skipped accounts
	}

	all := make(map[addr.Address]bool)
	for _, a := range req.Addresses {
		all[*a] = true
	}

	found := make(map[addr.Address]bool)
	for _, r := range res.Rows {
		found[r.Address] = true
	}

	var skipped []addr.Address
	for a := range all {
		if found[a] {
			continue
		}
		if core.SkipAddress(a) {
			// fetch heavy skipped account states
			skipped = append(skipped, a)
			continue
		}
	}
	if len(skipped) == 0 {
		return nil
	}

	b, err := s.blockRepo.GetLastMasterBlock(ctx)
	if err != nil {
		return errors.Wrap(err, "get last master block")
	}
	master := &ton.BlockIDExt{
		Workchain: b.Workchain,
		Shard:     b.Shard,
		SeqNo:     b.SeqNo,
		RootHash:  b.RootHash,
		FileHash:  b.FileHash,
	}

	for _, a := range skipped {
		tu, err := a.ToTonutils()
		if err != nil {
			return errors.Wrap(err, "parse address")
		}

		acc, err := s.API.GetAccount(ctx, master, tu)
		if err != nil {
			return errors.Wrapf(err, "get %s account", a.Base64())
		}

		parsed := fetcher.MapAccount(master, acc)

		parsed.Label, err = s.accountRepo.GetAddressLabel(ctx, a)
		if err != nil && !errors.Is(err, core.ErrNotFound) {
			return errors.Wrap(err, "get address label")
		}

		res.Total += 1
		res.Rows = append(res.Rows, parsed)
	}

	// TODO: sort and limit inserted accounts

	return nil
}

func (s *Service) addGetMethodDescription(ctx context.Context, rows []*core.AccountState) error {
	for _, r := range rows {
		for name, methods := range r.ExecutedGetMethods {
			for it := range methods {
				d, err := s.contractRepo.GetMethodDescription(ctx, name, methods[it].Name)
				if err != nil {
					return errors.Wrapf(err, "cannot get %s get-method description of contract %s", methods[it].Name, name)
				}
				methods[it].Arguments = d.Arguments
				methods[it].ReturnValues = d.ReturnValues
			}
		}
	}
	return nil
}

func (s *Service) FilterAccounts(ctx context.Context, req *filter.AccountsReq) (*filter.AccountsRes, error) {
	res, err := s.accountRepo.FilterAccounts(ctx, req)
	if err != nil {
		return nil, err
	}
	if err := s.fetchSkippedAccounts(ctx, req, res); err != nil {
		return nil, err
	}
	if err := s.addGetMethodDescription(ctx, res.Rows); err != nil {
		return nil, err
	}
	return res, nil
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
