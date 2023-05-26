package app

import (
	"context"
	"time"

	"github.com/rs/zerolog/log"
	"github.com/xssnick/tonutils-go/ton"

	"github.com/tonindexer/anton/internal/core"
)

type FetcherConfig struct {
	API *ton.APIClient

	Parser ParserService
}

func TimeTrack(start time.Time, name string) {
	elapsed := float64(time.Since(start)) / 1e9
	if elapsed < 0.05 {
		return
	}
	log.Debug().Str("func", name).Float64("elapsed", elapsed).Msg("")
}

type FetcherService interface {
	UnseenBlocks(ctx context.Context, masterSeqNo uint32) (master *ton.BlockIDExt, shards []*ton.BlockIDExt, err error)
	BlockTransactions(ctx context.Context, b *ton.BlockIDExt) ([]*core.Transaction, error)
}
