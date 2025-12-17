package cache

import (
	"container/list"
	"context"
	"strings"
	"sync"
	"time"
)

type localEntry struct {
	key       string
	value     []byte
	expiresAt time.Time
}

// LocalCache 简单的并发安全 LRU+TTL 本地缓存
type LocalCache struct {
	mu         sync.RWMutex
	items      map[string]*list.Element
	ll         *list.List
	capacity   int
	defaultTTL time.Duration
	codec      Codec
}

func NewLocalCache(capacity int, defaultTTL time.Duration, codec Codec) *LocalCache {
	if capacity <= 0 {
		capacity = 1024
	}
	if codec == nil {
		codec = JSONCodec{}
	}
	return &LocalCache{
		items:      make(map[string]*list.Element, capacity),
		ll:         list.New(),
		capacity:   capacity,
		defaultTTL: defaultTTL,
		codec:      codec,
	}
}

func (c *LocalCache) Get(ctx context.Context, key string, dst any) (bool, error) {
	_ = ctx
	c.mu.Lock()
	elem, ok := c.items[key]
	if !ok {
		c.mu.Unlock()
		return false, nil
	}
	ent := elem.Value.(localEntry)
	if ent.expiresAt.IsZero() || ent.expiresAt.After(time.Now()) {
		c.ll.MoveToFront(elem)
		c.mu.Unlock()
		return c.decode(ent.value, dst)
	}
	// expired
	c.ll.Remove(elem)
	delete(c.items, key)
	c.mu.Unlock()
	return false, nil
}

func (c *LocalCache) Set(ctx context.Context, key string, val any, ttl time.Duration) error {
	_ = ctx
	data, err := c.codec.Marshal(val)
	if err != nil {
		return err
	}
	if ttl <= 0 {
		ttl = c.defaultTTL
	}
	var expiresAt time.Time
	if ttl > 0 {
		expiresAt = time.Now().Add(ttl)
	}

	c.mu.Lock()
	defer c.mu.Unlock()
	if elem, ok := c.items[key]; ok {
		elem.Value = localEntry{key: key, value: data, expiresAt: expiresAt}
		c.ll.MoveToFront(elem)
		return nil
	}
	elem := c.ll.PushFront(localEntry{key: key, value: data, expiresAt: expiresAt})
	c.items[key] = elem
	if c.ll.Len() > c.capacity {
		c.evictOldest()
	}
	return nil
}

func (c *LocalCache) Delete(ctx context.Context, key string) error {
	_ = ctx
	c.mu.Lock()
	if elem, ok := c.items[key]; ok {
		c.ll.Remove(elem)
		delete(c.items, key)
	}
	c.mu.Unlock()
	return nil
}

func (c *LocalCache) DeleteByPrefix(ctx context.Context, prefix string) error {
	_ = ctx
	if prefix == "" {
		return nil
	}
	c.mu.Lock()
	for k, elem := range c.items {
		if strings.HasPrefix(k, prefix) {
			c.ll.Remove(elem)
			delete(c.items, k)
		}
	}
	c.mu.Unlock()
	return nil
}

func (c *LocalCache) evictOldest() {
	elem := c.ll.Back()
	if elem == nil {
		return
	}
	c.ll.Remove(elem)
	ent := elem.Value.(localEntry)
	delete(c.items, ent.key)
}

func (c *LocalCache) decode(data []byte, dst any) (bool, error) {
	if dst == nil {
		return true, nil
	}
	if err := c.codec.Unmarshal(data, dst); err != nil {
		return false, err
	}
	return true, nil
}
