package parser

import (
	"context"

	"github.com/pkg/errors"
	"github.com/uptrace/go-clickhouse/ch"
	"github.com/xssnick/tonutils-go/liteclient"
	"github.com/xssnick/tonutils-go/ton"

	"github.com/iam047801/tonidx/internal/app"
	"github.com/iam047801/tonidx/internal/core"
	"github.com/iam047801/tonidx/internal/core/repository/account"
	"github.com/iam047801/tonidx/internal/core/repository/tx"
)

var _ app.ParserService = (*Service)(nil)

type Service struct {
	cfg *app.ParserConfig

	api *ton.APIClient

	db          *ch.DB
	txRepo      core.TxRepository
	accountRepo core.AccountRepository
}

func NewService(ctx context.Context, cfg *app.ParserConfig) (*Service, error) {
	var s = new(Service)

	s.cfg = cfg
	s.db = cfg.DB
	s.txRepo = tx.NewRepository(s.db)
	s.accountRepo = account.NewRepository(s.db)

	nodes := mainnetArchive
	if cfg.Testnet {
		nodes = testnetArchive
	}

	client := liteclient.NewConnectionPool()
	for _, n := range nodes {
		if err := client.AddConnection(ctx, n.addr, n.key); err != nil {
			return nil, errors.Wrap(err, "cannot add connection")
		}
	}

	s.api = ton.NewAPIClient(client)

	return s, nil
}

func (s *Service) API() *ton.APIClient {
	return s.api
}
