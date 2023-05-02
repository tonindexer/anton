package app

import (
	"context"

	"github.com/pkg/errors"
	"github.com/xssnick/tonutils-go/ton"

	"github.com/tonindexer/anton/internal/core"
	"github.com/tonindexer/anton/internal/core/repository"
)

var ErrImpossibleParsing = errors.New("parsing is impossible")

type ServerAddr struct {
	IPPort    string
	PubKeyB64 string
}

type ParserConfig struct {
	DB      *repository.DB
	Servers []*ServerAddr
}

type ParserService interface {
	API() *ton.APIClient

	ParseAccountData(ctx context.Context, acc *core.AccountState) (*core.AccountData, error)
	ParseMessagePayload(ctx context.Context, src, dst *core.AccountData, message *core.Message) (*core.MessagePayload, error)
}
