package fetcher

import (
	"sync"
	"time"

	"github.com/xssnick/tonutils-go/ton"

	"github.com/tonindexer/anton/addr"
	"github.com/tonindexer/anton/internal/core"
)

var cacheInvalidation = time.Second

type blocksCache struct {
	masterBlocks map[uint32]*ton.BlockIDExt
	shardsInfo   map[uint32][]*ton.BlockIDExt
	lastUsed     map[uint32]time.Time
	sync.Mutex
}

func newBlocksCache() *blocksCache {
	return &blocksCache{
		masterBlocks: map[uint32]*ton.BlockIDExt{},
		shardsInfo:   map[uint32][]*ton.BlockIDExt{},
		lastUsed:     map[uint32]time.Time{},
	}
}

func (c *blocksCache) clearCaches() {
	for seq, ts := range c.lastUsed {
		if ts.Add(cacheInvalidation).Before(time.Now()) {
			delete(c.masterBlocks, seq)
			delete(c.shardsInfo, seq)
		}
	}
}

func (c *blocksCache) getMaster(seqNo uint32) (*ton.BlockIDExt, bool) {
	c.Lock()
	defer c.Unlock()

	m, ok := c.masterBlocks[seqNo]
	if ok {
		c.lastUsed[seqNo] = time.Now()
	}
	return m, ok
}

func (c *blocksCache) setMaster(master *ton.BlockIDExt) {
	c.Lock()
	defer c.Unlock()

	c.masterBlocks[master.SeqNo] = master
	c.lastUsed[master.SeqNo] = time.Now()
	c.clearCaches()
}

func (c *blocksCache) getShards(master *ton.BlockIDExt) ([]*ton.BlockIDExt, bool) {
	c.Lock()
	defer c.Unlock()

	m, ok := c.shardsInfo[master.SeqNo]
	if ok {
		c.lastUsed[master.SeqNo] = time.Now()
	}
	return m, ok
}

func (c *blocksCache) setShards(master *ton.BlockIDExt, shards []*ton.BlockIDExt) {
	c.Lock()
	defer c.Unlock()

	c.shardsInfo[master.SeqNo] = shards
	c.lastUsed[master.SeqNo] = time.Now()
	c.clearCaches()
}

type accountCache struct {
	m        map[core.BlockID]map[addr.Address]*core.AccountState
	lastUsed map[core.BlockID]time.Time
	sync.Mutex
}

func (c *accountCache) clearCaches() {
	for seq, ts := range c.lastUsed {
		if ts.Add(cacheInvalidation).Before(time.Now()) {
			delete(c.m, seq)
		}
	}
}

func newAccountCache() *accountCache {
	return &accountCache{
		m:        map[core.BlockID]map[addr.Address]*core.AccountState{},
		lastUsed: map[core.BlockID]time.Time{},
	}
}

func (c *accountCache) get(bExt *ton.BlockIDExt, a addr.Address) (*core.AccountState, bool) {
	c.Lock()
	defer c.Unlock()

	b := core.GetBlockID(bExt)

	m, ok := c.m[b]
	if !ok {
		return nil, false
	}

	acc, ok := m[a]
	if !ok {
		return nil, false
	}

	c.lastUsed[b] = time.Now()
	return acc, true
}

func (c *accountCache) set(bExt *ton.BlockIDExt, acc *core.AccountState) {
	c.Lock()
	defer c.Unlock()

	b := core.GetBlockID(bExt)

	if _, ok := c.m[b]; !ok {
		c.m[b] = map[addr.Address]*core.AccountState{}
	}

	c.m[b][acc.Address] = acc
	c.lastUsed[b] = time.Now()
	c.clearCaches()
}
