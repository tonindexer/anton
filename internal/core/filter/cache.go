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
	lastCleanup      time.Time
}

func NewCache(ttl time.Duration) *Cache {
	return &Cache{
		msgCountCache:    make(map[string]CacheEntry),
		msgCountCacheTTL: ttl,
		lastCleanup:      time.Now(),
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

func (c *Cache) cleanupExpiredEntries() {
	if time.Since(c.lastCleanup) < time.Minute {
		return
	}
	now := time.Now()
	for k, entry := range c.msgCountCache {
		if now.Sub(entry.UpdatedAt) > c.msgCountCacheTTL {
			delete(c.msgCountCache, k)
		}
	}
	c.lastCleanup = now
}

func (c *Cache) Get(filterReq any) (count int, maxSeqNo uint64, err error) {
	k, err := getCacheKey(filterReq)
	if err != nil {
		return 0, 0, err
	}

	c.msgCountCacheMx.Lock()
	defer c.msgCountCacheMx.Unlock()

	c.cleanupExpiredEntries()

	entry, ok := c.msgCountCache[k]
	if !ok {
		return 0, 0, core.ErrNotFound
	}

	return entry.Count, entry.MaxSeqNo, nil
}
