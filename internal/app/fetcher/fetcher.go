package fetcher

import (
	"sync"

	"github.com/tonindexer/anton/internal/app"
	"github.com/tonindexer/anton/internal/core"
	"github.com/tonindexer/anton/lru"
)

var _ app.FetcherService = (*Service)(nil)

const statesCacheLen = 16384

type getAccountRes struct {
	acc *core.AccountState
	err error
}

type Service struct {
	*app.FetcherConfig

	masterWorkchain int32
	masterShard     uint64

	minterStatesCache        *lru.Cache[core.AccountStateID, *core.AccountState]
	minterStatesCacheLocks   *lru.Cache[core.AccountStateID, *sync.Once]
	minterStatesCacheLocksMx sync.Mutex

	accBlockStatesCache        *lru.Cache[core.AccountBlockStateID, getAccountRes]
	accBlockStatesCacheLocks   *lru.Cache[core.AccountBlockStateID, *sync.Once]
	accBlockStatesCacheLocksMx sync.Mutex

	blocks    *blocksCache
	libraries *librariesCache
}

func NewService(cfg *app.FetcherConfig) *Service {
	return &Service{
		FetcherConfig:            cfg,
		masterWorkchain:          -1,
		masterShard:              0x8000000000000000,
		minterStatesCache:        lru.New[core.AccountStateID, *core.AccountState](statesCacheLen),
		minterStatesCacheLocks:   lru.New[core.AccountStateID, *sync.Once](statesCacheLen),
		accBlockStatesCache:      lru.New[core.AccountBlockStateID, getAccountRes](statesCacheLen),
		accBlockStatesCacheLocks: lru.New[core.AccountBlockStateID, *sync.Once](statesCacheLen),
		blocks:                   newBlocksCache(),
		libraries:                newLibrariesCache(),
	}
}
