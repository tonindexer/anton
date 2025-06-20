package filter

import (
	"encoding/json"
	"sync"
	"time"

	"github.com/tonindexer/anton/internal/core"
)

type CacheEntry struct {
	Count     int
	MaxSeqNo  uint64
	UpdatedAt time.Time
}

type Cache struct {
	msgCountCache    map[string]CacheEntry
	msgCountCacheMx  sync.Mutex
	msgCountCacheTTL time.Duration
}

func NewCache(ttl time.Duration) *Cache {
	return &Cache{
		msgCountCache:    make(map[string]CacheEntry),
		msgCountCacheTTL: ttl,
	}
}

func getCacheKey(req any) (string, error) {
	bytes, err := json.Marshal(req)
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}

func (c *Cache) Set(filterReq any, count int, maxSeqNo uint64) error {
	k, err := getCacheKey(filterReq)
	if err != nil {
		return err
	}

	c.msgCountCacheMx.Lock()
	defer c.msgCountCacheMx.Unlock()

	c.msgCountCache[k] = CacheEntry{
		Count:     count,
		MaxSeqNo:  maxSeqNo,
		UpdatedAt: time.Now(),
	}

	return nil
}

func (c *Cache) Get(filterReq any) (count int, maxSeqNo uint64, err error) {
	k, err := getCacheKey(filterReq)
	if err != nil {
		return 0, 0, err
	}

	c.msgCountCacheMx.Lock()
	defer c.msgCountCacheMx.Unlock()

	entry, ok := c.msgCountCache[k]
	if !ok {
		return 0, 0, core.ErrNotFound
	}
	if time.Since(entry.UpdatedAt) > c.msgCountCacheTTL {
		delete(c.msgCountCache, k)
		return 0, 0, core.ErrNotFound
	}

	return entry.Count, entry.MaxSeqNo, nil
}
