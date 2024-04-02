package rescan

import (
	"container/list"
	"sync"

	"github.com/tonindexer/anton/abi"
	"github.com/tonindexer/anton/addr"
	"github.com/tonindexer/anton/internal/core"
)

type interfacesCacheItem struct {
	data   map[uint64][]abi.ContractName
	keyPtr *list.Element
}

// interfacesCache implements LRU cache for account interfaces.
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

type minterStateCacheItem struct {
	state    *core.AccountState
	nextTxLT uint64
	keyPtr   *list.Element
}

// minterStateCache implements LRU cache for minter account states,
// which are used for rescanning of nft items and jetton wallets.
type minterStateCache struct {
	list     *list.List
	items    map[addr.Address]*minterStateCacheItem
	capacity int
	sync.RWMutex
}

func newMinterStateCache(capacity int) *minterStateCache {
	return &minterStateCache{
		list:     list.New(),
		items:    map[addr.Address]*minterStateCacheItem{},
		capacity: capacity,
	}
}

func (c *minterStateCache) removeItem() {
	back := c.list.Back()
	c.list.Remove(back)
	delete(c.items, back.Value.(addr.Address)) //nolint:forcetypeassert // no need
}

func (c *minterStateCache) updateItem(item *minterStateCacheItem, k addr.Address, state *core.AccountState, nextTxLT uint64) {
	item.state = state
	item.nextTxLT = nextTxLT
	c.items[k] = item
	c.list.MoveToFront(item.keyPtr)
}

func (c *minterStateCache) put(k addr.Address, state *core.AccountState, nextTxLT uint64) {
	c.Lock()
	defer c.Unlock()

	if item, ok := c.items[k]; !ok {
		if c.capacity == len(c.items) {
			c.removeItem()
		}
		c.items[k] = &minterStateCacheItem{
			state:    state,
			nextTxLT: nextTxLT,
			keyPtr:   c.list.PushFront(k),
		}
	} else {
		c.updateItem(item, k, state, nextTxLT) // it's never used
	}
}

func (c *minterStateCache) get(key addr.Address, itemTxLT uint64) (state *core.AccountState, ok bool) {
	c.Lock()
	defer c.Unlock()

	minter, ok := c.items[key]
	if !ok {
		return nil, false
	}

	c.list.MoveToFront(minter.keyPtr)

	if itemTxLT < minter.state.LastTxLT || (minter.nextTxLT != 0 && itemTxLT > minter.nextTxLT) {
		// as we are processing item state,
		// which is later than the next minter state or earlier than the current minter state in cache,
		// we should remove the current minter state and get the new one from the database

		front := c.list.Front()
		c.list.Remove(front)
		delete(c.items, front.Value.(addr.Address)) //nolint:forcetypeassert // no need

		return nil, false
	}

	return minter.state, true
}
