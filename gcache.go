package gcache

import (
	"context"
	"sync"
	"time"
)

type cacheItem[T any] struct {
	value     T
	createdAt time.Time
	ttl       time.Duration
}

type shard[T any] struct {
	sync.RWMutex
	items map[string]*cacheItem[T]
}

type InMemoryCache[T any] struct {
	shards    []*shard[T]
	numShards int
	evictTime time.Duration
	ctx       context.Context
	cancel    context.CancelFunc
}

func (c *InMemoryCache[T]) shardIndex(key string) int {
	hash := fnv32(key)
	return int(hash % uint32(c.numShards))
}

func (c *InMemoryCache[T]) Set(key string, value T, ttl time.Duration) {
	shardIndex := c.shardIndex(key)
	shard := c.shards[shardIndex]

	shard.Lock()
	defer shard.Unlock()

	if item, exists := shard.items[key]; exists {
		item.value = value
		item.createdAt = time.Now()
		item.ttl = ttl
	} else {
		shard.items[key] = &cacheItem[T]{
			value:     value,
			createdAt: time.Now(),
			ttl:       ttl,
		}
	}
}

func (c *InMemoryCache[T]) Get(key string) (x T, ok bool) {
	shardIndex := c.shardIndex(key)
	shard := c.shards[shardIndex]

	shard.RLock()
	defer shard.RUnlock()

	item, exists := shard.items[key]
	if !exists {
		return
	}

	return item.value, true
}

func (c *InMemoryCache[T]) Stop() {
	c.cancel()
}
func (c *InMemoryCache[T]) StartEviction() {
	for i := 0; i < c.numShards; i++ {
		go c.startShardEviction(i)
	}
}

func (c *InMemoryCache[T]) startShardEviction(shardIndex int) {
	ticker := time.NewTicker(c.evictTime)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			c.evictExpiredItems(shardIndex)
		case <-c.ctx.Done():
			return
		}
	}
}

func (c *InMemoryCache[T]) evictExpiredItems(shardIndex int) {
	shard := c.shards[shardIndex]

	shard.Lock()
	defer shard.Unlock()

	now := time.Now()
	for key, item := range shard.items {
		if now.Sub(item.createdAt) > item.ttl {
			delete(shard.items, key)
		}
	}
}

func NewInMemoryCache[T any](numShards int, evictTime time.Duration) *InMemoryCache[T] {

	ctx, cancel := context.WithCancel(context.Background())

	cache := &InMemoryCache[T]{
		numShards: numShards,
		evictTime: evictTime,
		ctx:       ctx,
		cancel:    cancel,
	}

	cache.shards = make([]*shard[T], numShards)
	for i := 0; i < numShards; i++ {
		cache.shards[i] = &shard[T]{
			items: make(map[string]*cacheItem[T]),
		}
	}

	cache.StartEviction()

	return cache
}

func fnv32(key string) uint32 {
	const (
		offset uint32 = 2166136261
		prime  uint32 = 16777619
	)

	hash := offset
	for i := 0; i < len(key); i++ {
		hash ^= uint32(key[i])
		hash *= prime
	}
	return hash
}
