package app

import (
	"context"

	"github.com/uptrace/go-clickhouse/ch"
	"github.com/xssnick/tonutils-go/address"
	"github.com/xssnick/tonutils-go/tlb"
	"github.com/xssnick/tonutils-go/ton"

	"github.com/iam047801/tonidx/internal/core"
)

type ServerAddr struct {
	IPPort    string
	PubKeyB64 string
}

type ParserConfig struct {
	DB      *ch.DB
	Servers []*ServerAddr
}

type ParserService interface {
	API() *ton.APIClient

	GetBlockTransactions(ctx context.Context, b *tlb.BlockInfo) ([]*tlb.Transaction, error)
	ParseBlockTransactions(ctx context.Context, b *tlb.BlockInfo, blockTx []*tlb.Transaction) ([]*core.Transaction, error)
	ParseBlockMessages(ctx context.Context, b *tlb.BlockInfo, blockTx []*tlb.Transaction) ([]*core.Message, error)

	ContractInterfaces(ctx context.Context, acc *tlb.Account) ([]core.ContractType, error)
	ParseAccount(ctx context.Context, master *tlb.BlockInfo, addr *address.Address) (*core.Account, error)
	ParseAccountData(ctx context.Context, master *tlb.BlockInfo, acc *core.Account) (*core.AccountData, error)

	ParseMessagePayload(ctx context.Context, src, dst *core.Account, message *core.Message) (*core.MessagePayload, error)
}
