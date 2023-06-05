package app

import (
	"context"
	"fmt"
	"time"

	"github.com/rs/zerolog/log"
	"github.com/xssnick/tonutils-go/ton"

	"github.com/tonindexer/anton/internal/core"
)

type FetcherConfig struct {
	API *ton.APIClient

	Parser ParserService
}

func TimeTrack(start time.Time, fun string, args ...any) {
	elapsed := float64(time.Since(start)) / 1e9
	if elapsed < 0.1 {
		return
	}
	log.Debug().Str("func", fmt.Sprintf(fun, args...)).Float64("elapsed", elapsed).Msg("timer")
}

type FetcherService interface {
	UnseenBlocks(ctx context.Context, masterSeqNo uint32) (master *ton.BlockIDExt, shards []*ton.BlockIDExt, err error)
	BlockTransactions(ctx context.Context, b *ton.BlockIDExt) ([]*core.Transaction, error)
}
