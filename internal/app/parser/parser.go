package parser

import (
	"context"
	"fmt"

	"github.com/pkg/errors"
	"github.com/xssnick/tonutils-go/liteclient"
	"github.com/xssnick/tonutils-go/tl"
	"github.com/xssnick/tonutils-go/tlb"
	"github.com/xssnick/tonutils-go/ton"
	"github.com/xssnick/tonutils-go/tvm/cell"

	"github.com/tonindexer/anton/internal/app"
	"github.com/tonindexer/anton/internal/core"
	"github.com/tonindexer/anton/internal/core/repository/contract"
)

var _ app.ParserService = (*Service)(nil)

type Service struct {
	cfg *app.ParserConfig

	api *ton.APIClient

	bcConfig *cell.Cell // TODO: init it

	contractRepo core.ContractRepository
}

func getBlockchainConfig(ctx context.Context, api *ton.APIClient) (*cell.Cell, error) {
	var res tl.Serializable

	b, err := api.GetMasterchainInfo(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "get masterchain info")
	}

	err = api.Client().QueryLiteserver(ctx, ton.GetConfigAll{Mode: 0, BlockID: b}, &res)
	if err != nil {
		return nil, err
	}

	switch t := res.(type) {
	case ton.ConfigAll:
		var state tlb.ShardStateUnsplit

		c, err := cell.FromBOC(t.ConfigProof)
		if err != nil {
			return nil, err
		}

		ref, err := c.BeginParse().LoadRef()
		if err != nil {
			return nil, err
		}

		err = tlb.LoadFromCell(&state, ref)
		if err != nil {
			return nil, err
		}

		return state.McStateExtra.ConfigParams.Config.ToCell()

	case ton.LSError:
		return nil, t

	default:
		return nil, fmt.Errorf("got unknown response")
	}
}

func NewService(ctx context.Context, cfg *app.ParserConfig) (*Service, error) {
	var (
		s   = new(Service)
		err error
	)

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

	s.bcConfig, err = getBlockchainConfig(ctx, s.api)
	if err != nil {
		return nil, err
	}

	return s, nil
}

func (s *Service) API() *ton.APIClient {
	return s.api
}
