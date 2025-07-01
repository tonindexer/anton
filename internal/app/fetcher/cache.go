package fetcher

import (
	"encoding/hex"
	"sync"
	"time"

	"github.com/xssnick/tonutils-go/ton"
)

var cacheInvalidation = 10 * time.Minute

type blocksCache struct {
	masterBlocks map[uint32]*ton.BlockIDExt
	shardsInfo   map[uint32][]*ton.BlockIDExt
	lastCleared  time.Time
	sync.Mutex
}

func newBlocksCache() *blocksCache {
	return &blocksCache{
		masterBlocks: map[uint32]*ton.BlockIDExt{},
		shardsInfo:   map[uint32][]*ton.BlockIDExt{},
		lastCleared:  time.Now(),
	}
}

func (c *blocksCache) clearCaches() {
	if time.Since(c.lastCleared) < cacheInvalidation {
		return
	}
	c.masterBlocks = map[uint32]*ton.BlockIDExt{}
	c.shardsInfo = map[uint32][]*ton.BlockIDExt{}
	c.lastCleared = time.Now()
}

func (c *blocksCache) getMaster(seqNo uint32) (*ton.BlockIDExt, bool) {
	c.Lock()
	defer c.Unlock()

	m, ok := c.masterBlocks[seqNo]
	return m, ok
}

func (c *blocksCache) setMaster(master *ton.BlockIDExt) {
	c.Lock()
	defer c.Unlock()

	c.masterBlocks[master.SeqNo] = master
	c.clearCaches()
}

func (c *blocksCache) getShards(master *ton.BlockIDExt) ([]*ton.BlockIDExt, bool) {
	c.Lock()
	defer c.Unlock()

	m, ok := c.shardsInfo[master.SeqNo]
	return m, ok
}

func (c *blocksCache) setShards(master *ton.BlockIDExt, shards []*ton.BlockIDExt) {
	c.Lock()
	defer c.Unlock()

	c.shardsInfo[master.SeqNo] = shards
	c.clearCaches()
}

type librariesCache struct {
	libs map[string]*libDescription
	sync.Mutex
}

func newLibrariesCache() *librariesCache {
	return &librariesCache{
		libs: map[string]*libDescription{},
	}
}

func (c *librariesCache) get(hash []byte) *libDescription {
	c.Lock()
	defer c.Unlock()

	l, ok := c.libs[hex.EncodeToString(hash)]
	if ok {
		return l
	}

	return nil
}

func (c *librariesCache) set(hash []byte, desc *libDescription) {
	c.Lock()
	defer c.Unlock()

	h := hex.EncodeToString(hash)

	_, ok := c.libs[h]
	if ok {
		return
	}

	c.libs[h] = desc
}
