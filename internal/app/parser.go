package app

import (
	"context"

	"github.com/xssnick/tonutils-go/tlb"
	"github.com/xssnick/tonutils-go/ton"

	"github.com/iam047801/tonidx/abi"
	"github.com/iam047801/tonidx/internal/core"
	"github.com/iam047801/tonidx/internal/core/repository"
)

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

	DetermineInterfaces(ctx context.Context, acc *core.AccountState) ([]abi.ContractName, error)
	ParseAccountData(ctx context.Context, b *tlb.BlockInfo, acc *core.AccountState, types []abi.ContractName) (*core.AccountData, error)

	ParseMessagePayload(ctx context.Context, src, dst *core.AccountData, message *core.Message) (*core.MessagePayload, error)
}
