package parser

import (
	"context"

	"github.com/pkg/errors"
	"github.com/xssnick/tonutils-go/liteclient"
	"github.com/xssnick/tonutils-go/ton"

	"github.com/tonindexer/anton/internal/app"
	"github.com/tonindexer/anton/internal/core"
	"github.com/tonindexer/anton/internal/core/repository/contract"
)

var _ app.ParserService = (*Service)(nil)

type Service struct {
	cfg *app.ParserConfig

	api *ton.APIClient

	contractRepo core.ContractRepository
}

func NewService(ctx context.Context, cfg *app.ParserConfig) (*Service, error) {
	var s = new(Service)

	s.cfg = cfg
	_, pg := s.cfg.DB.CH, s.cfg.DB.PG
	s.contractRepo = contract.NewRepository(pg)

	client := liteclient.NewConnectionPool()
	for _, n := range cfg.Servers {
		if err := client.AddConnection(ctx, n.IPPort, n.PubKeyB64); err != nil {
			return nil, errors.Wrapf(err, "cannot add connection (host = '%s', key = '%s')", n.IPPort, n.PubKeyB64)
		}
	}

	s.api = ton.NewAPIClient(client)

	return s, nil
}

func (s *Service) API() *ton.APIClient {
	return s.api
}
