package rescan

import (
	"container/list"
	"sync"

	"github.com/tonindexer/anton/abi"
	"github.com/tonindexer/anton/addr"
)

type interfacesCacheItem struct {
	data   map[uint64][]abi.ContractName
	keyPtr *list.Element
}

// accountInterfacesCache implements LRU cache for account interfaces.
// For a given address it stores account interface updates.
// Used only in messages rescan.
type interfacesCache struct {
	queue    *list.List
	items    map[addr.Address]*interfacesCacheItem
	capacity int
	sync.RWMutex
}

func newInterfacesCache(capacity int) *interfacesCache {
	return &interfacesCache{
		queue:    list.New(),
		items:    map[addr.Address]*interfacesCacheItem{},
		capacity: capacity,
	}
}

func (c *interfacesCache) removeItem() {
	back := c.queue.Back()
	c.queue.Remove(back)
	delete(c.items, back.Value.(addr.Address)) //nolint:forcetypeassert // no need
}

func (c *interfacesCache) updateItem(item *interfacesCacheItem, k addr.Address, v map[uint64][]abi.ContractName) {
	item.data = v
	c.items[k] = item
	c.queue.MoveToFront(item.keyPtr)
}

func (c *interfacesCache) put(k addr.Address, v map[uint64][]abi.ContractName) {
	c.Lock()
	defer c.Unlock()

	if item, ok := c.items[k]; !ok {
		if c.capacity == len(c.items) {
			c.removeItem()
		}
		c.items[k] = &interfacesCacheItem{
			data:   v,
			keyPtr: c.queue.PushFront(k),
		}
	} else {
		c.updateItem(item, k, v) // actually it's not used
	}
}

func (c *interfacesCache) get(key addr.Address) (map[uint64][]abi.ContractName, bool) {
	c.RLock()
	defer c.RUnlock()

	if item, ok := c.items[key]; ok {
		c.queue.MoveToFront(item.keyPtr)
		return item.data, true
	}

	return nil, false
}
