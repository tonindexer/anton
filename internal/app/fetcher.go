package app

import (
	"context"

	"github.com/xssnick/tonutils-go/ton"

	"github.com/tonindexer/anton/internal/core"
	"github.com/tonindexer/anton/internal/core/filter"
)

type FetcherConfig struct {
	API ton.APIClientWrapped

	AccountRepo filter.AccountRepository

	Parser ParserService
}

type FetcherService interface {
	LookupMaster(ctx context.Context, api ton.APIClientWrapped, seqNo uint32) (*ton.BlockIDExt, error)
	UnseenBlocks(ctx context.Context, masterSeqNo uint32) (master *ton.BlockIDExt, shards []*ton.BlockIDExt, err error)
	UnseenShards(ctx context.Context, master *ton.BlockIDExt) (shards []*ton.BlockIDExt, err error)
	BlockTransactions(ctx context.Context, master, b *ton.BlockIDExt) ([]*core.Transaction, error)
}
