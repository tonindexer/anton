package contract

import (
	"sync"
	"time"

	"github.com/tonindexer/anton/abi"
	"github.com/tonindexer/anton/internal/core"
)

var cacheInvalidation = 10 * time.Second

type cache struct {
	interfaces  []*core.ContractInterface
	operations  []*core.ContractOperation
	lastCleared time.Time
	sync.RWMutex
}

func newCache() *cache {
	return &cache{lastCleared: time.Now()}
}

func (c *cache) clearCaches() {
	if time.Since(c.lastCleared) < cacheInvalidation {
		return
	}
	c.interfaces = nil
	c.operations = nil
	c.lastCleared = time.Now()
}

func (c *cache) setInterfaces(interfaces []*core.ContractInterface) {
	c.Lock()
	defer c.Unlock()
	c.interfaces = interfaces
}

func (c *cache) getInterfaces() []*core.ContractInterface {
	c.RLock()
	defer c.RUnlock()
	return c.interfaces
}

func (c *cache) setOperations(operations []*core.ContractOperation) {
	c.Lock()
	defer c.Unlock()
	c.operations = operations
}

func (c *cache) getOperations() []*core.ContractOperation {
	c.RLock()
	defer c.RUnlock()
	return c.operations
}

func (c *cache) getOperationByID(types []abi.ContractName, outgoing bool, id uint32) *core.ContractOperation {
	c.RLock()
	defer c.RUnlock()

	for _, op := range c.operations {
		if op.Outgoing != outgoing {
			continue
		}
		if op.OperationID != id {
			continue
		}
		for _, t := range types {
			if op.ContractName == t {
				return op
			}
		}
	}

	return nil
}
