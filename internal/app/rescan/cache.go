package rescan

import (
	"sort"

	"github.com/rs/zerolog/log"

	"github.com/tonindexer/anton/addr"
	"github.com/tonindexer/anton/internal/core"
	"github.com/tonindexer/anton/lru"
)

const maxMinterStates int = 64

type minterState struct {
	acc      *core.AccountState
	nextTxLT uint64
}

type minterStatesCache struct {
	lru *lru.Cache[uint64, *minterState]
}

func newMinterStatesCache() *minterStatesCache {
	return &minterStatesCache{
		lru: lru.New[uint64, *minterState](maxMinterStates),
	}
}

type mintersCache struct {
	lru *lru.Cache[addr.Address, *minterStatesCache]
}

func newMinterStateCache(capacity int) *mintersCache {
	return &mintersCache{
		lru: lru.New[addr.Address, *minterStatesCache](capacity),
	}
}

func (c *mintersCache) put(k addr.Address, state *core.AccountState, nextTxLT uint64) {
	states, ok := c.lru.Get(k)
	if !ok {
		states = newMinterStatesCache()
		states.lru.Put(state.LastTxLT, &minterState{acc: state, nextTxLT: nextTxLT})
		c.lru.Put(k, states)
		return
	}

	states.lru.Put(state.LastTxLT, &minterState{acc: state, nextTxLT: nextTxLT})
}

func (c *mintersCache) get(k addr.Address, itemTxLT uint64) (state *core.AccountState, ok bool) {
	states, ok := c.lru.Get(k)
	if !ok {
		return nil, false
	}

	lts := states.lru.Keys()
	sort.Slice(lts, func(i, j int) bool { return lts[i] > lts[j] })

	for _, lt := range lts {
		if lt > itemTxLT {
			continue
		}

		minter, ok := states.lru.Get(lt)
		if !ok {
			log.Error().Str("addr", k.Base64()).Uint64("last_tx_lt", lt).Msg("cannot get minter state from cache")
			return nil, false
		}

		if minter.nextTxLT != 0 && itemTxLT > minter.nextTxLT {
			continue
		}

		return minter.acc, true
	}

	return nil, false
}
