package fetcher

import (
	"github.com/tonindexer/anton/internal/app"
)

var _ app.FetcherService = (*Service)(nil)

type Service struct {
	*app.FetcherConfig

	masterWorkchain int32
	masterShard     int64

	accounts *accountCache
	blocks   *blocksCache
}

func NewService(cfg *app.FetcherConfig) *Service {
	return &Service{
		FetcherConfig:   cfg,
		masterWorkchain: -1,
		masterShard:     0x8000000000000000,
		accounts:        newAccountCache(),
		blocks:          newBlocksCache(),
	}
}
