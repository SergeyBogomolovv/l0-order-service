package cache

import (
	"container/list"
	"context"
	"sync"
	"time"
)

const janitorInterval = 2 * time.Minute

type entry struct {
	key        string
	value      []byte
	expiration time.Time
}

type LRUCache struct {
	capacity int
	mu       sync.Mutex
	ll       *list.List
	cache    map[string]*list.Element
	ttl      time.Duration
}

func NewLRUCache(capacity int, ttl time.Duration) *LRUCache {
	return &LRUCache{
		capacity: capacity,
		ll:       list.New(),
		cache:    make(map[string]*list.Element),
		ttl:      ttl,
	}
}

func (c *LRUCache) Get(key string) ([]byte, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if ele, ok := c.cache[key]; ok {
		ent := ele.Value.(*entry)
		if time.Now().After(ent.expiration) {
			c.removeElement(ele)
			return nil, false
		}
		c.ll.MoveToFront(ele)
		return ent.value, true
	}
	return nil, false
}

func (c *LRUCache) Set(key string, value []byte) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if ele, ok := c.cache[key]; ok {
		c.ll.MoveToFront(ele)
		ent := ele.Value.(*entry)
		ent.value = value
		ent.expiration = time.Now().Add(c.ttl)
		return
	}

	ent := &entry{key: key, value: value, expiration: time.Now().Add(c.ttl)}
	ele := c.ll.PushFront(ent)
	c.cache[key] = ele

	if c.ll.Len() > c.capacity {
		c.removeOldest()
	}
}

func (c *LRUCache) removeOldest() {
	ele := c.ll.Back()
	if ele != nil {
		c.removeElement(ele)
	}
}

func (c *LRUCache) removeElement(e *list.Element) {
	c.ll.Remove(e)
	ent := e.Value.(*entry)
	delete(c.cache, ent.key)
}

func (c *LRUCache) Size() int {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.ll.Len()
}

func (c *LRUCache) StartJanitor(ctx context.Context) {
	go func() {
		ticker := time.NewTicker(janitorInterval)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				c.cleanup()
			case <-ctx.Done():
				return
			}
		}
	}()
}

func (c *LRUCache) cleanup() {
	c.mu.Lock()
	defer c.mu.Unlock()
	for e := c.ll.Back(); e != nil; {
		prev := e.Prev()
		ent := e.Value.(*entry)
		if time.Now().After(ent.expiration) {
			c.removeElement(e)
		}
		e = prev
	}
}
