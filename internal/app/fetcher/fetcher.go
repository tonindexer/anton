package fetcher

import (
	"sync"

	"github.com/tonindexer/anton/internal/app"
	"github.com/tonindexer/anton/internal/core"
	"github.com/tonindexer/anton/lru"
)

var _ app.FetcherService = (*Service)(nil)

const minterStatesCacheLen = 16384

type Service struct {
	*app.FetcherConfig

	masterWorkchain int32
	masterShard     uint64

	minterStatesCache        *lru.Cache[core.AccountStateID, *core.AccountState]
	minterStatesCacheLocks   *lru.Cache[core.AccountStateID, *sync.Once]
	minterStatesCacheLocksMx sync.Mutex

	accounts  *accountCache
	blocks    *blocksCache
	libraries *librariesCache
}

func NewService(cfg *app.FetcherConfig) *Service {
	return &Service{
		FetcherConfig:          cfg,
		masterWorkchain:        -1,
		masterShard:            0x8000000000000000,
		minterStatesCache:      lru.New[core.AccountStateID, *core.AccountState](minterStatesCacheLen),
		minterStatesCacheLocks: lru.New[core.AccountStateID, *sync.Once](minterStatesCacheLen),
		accounts:               newAccountCache(),
		blocks:                 newBlocksCache(),
		libraries:              newLibrariesCache(),
	}
}
