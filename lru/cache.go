package lru

import (
	"container/list"
	"sync"
)

type Item[V any] struct {
	data   V
	keyPtr *list.Element
}

type Cache[K comparable, V any] struct {
	queue    *list.List
	items    map[K]*Item[V]
	capacity int
	mx       sync.Mutex
}

func New[K comparable, V any](capacity int) *Cache[K, V] {
	return &Cache[K, V]{
		queue:    list.New(),
		items:    map[K]*Item[V]{},
		capacity: capacity,
	}
}

func (c *Cache[K, V]) removeItem() {
	back := c.queue.Back()
	c.queue.Remove(back)
	delete(c.items, back.Value.(K)) //nolint:forcetypeassert // no need
}

func (c *Cache[K, V]) updateItem(item *Item[V], k K, v V) {
	item.data = v
	c.items[k] = item
	c.queue.MoveToFront(item.keyPtr)
}

func (c *Cache[K, V]) Put(k K, v V) {
	c.mx.Lock()
	defer c.mx.Unlock()

	if item, ok := c.items[k]; !ok {
		if c.capacity == len(c.items) {
			c.removeItem()
		}
		c.items[k] = &Item[V]{
			data:   v,
			keyPtr: c.queue.PushFront(k),
		}
	} else {
		c.updateItem(item, k, v) // actually it's not used
	}
}

func (c *Cache[K, V]) Get(key K) (v V, ok bool) { //nolint:ireturn // returns generic interface (V) of type param any
	c.mx.Lock()
	defer c.mx.Unlock()

	if item, ok := c.items[key]; ok {
		c.queue.MoveToFront(item.keyPtr)
		return item.data, true
	}

	return v, false
}

func (c *Cache[K, V]) Keys() (keys []K) {
	c.mx.Lock()
	defer c.mx.Unlock()

	for k := range c.items {
		keys = append(keys, k)
	}
	return keys
}
