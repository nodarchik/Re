package cache

import (
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

// Cache interface for pack calculation results
type Cache interface {
	Get(key string) (map[int]int, int, bool)
	Set(key string, packs map[int]int, total int, ttl time.Duration)
	Clear()
	Stats() CacheStats
}

// CacheStats tracks cache performance
type CacheStats struct {
	Hits     int64
	Misses   int64
	HitRatio float64
	Size     int
}

// MemoryCache implements in-memory LRU cache with O(1) operations
type MemoryCache struct {
	items   map[string]*cacheItem
	head    *lruNode // Most recently used
	tail    *lruNode // Least recently used
	maxSize int
	mu      sync.RWMutex
	hits    int64
	misses  int64
}

type cacheItem struct {
	packs      map[int]int
	total      int
	expiration time.Time
	node       *lruNode // Reference to LRU node for O(1) access
}

type lruNode struct {
	key  string
	prev *lruNode
	next *lruNode
}

// NewMemoryCache creates a new in-memory cache with O(1) LRU
func NewMemoryCache(maxSize int) *MemoryCache {
	return &MemoryCache{
		items:   make(map[string]*cacheItem, maxSize),
		maxSize: maxSize,
	}
}

// Get retrieves a cached result with optimized locking
func (c *MemoryCache) Get(key string) (map[int]int, int, bool) {
	// Fast path: read with RLock for concurrency
	c.mu.RLock()
	item, exists := c.items[key]
	if !exists || time.Now().After(item.expiration) {
		c.mu.RUnlock()
		atomic.AddInt64(&c.misses, 1)
		return nil, 0, false
	}

	packs, total := item.packs, item.total
	c.mu.RUnlock()

	atomic.AddInt64(&c.hits, 1)

	// Move to front (most recently used) with write lock
	c.mu.Lock()
	c.moveToFront(item.node)
	c.mu.Unlock()

	return packs, total, true
}

// Set stores a result in cache with O(1) LRU update
func (c *MemoryCache) Set(key string, packs map[int]int, total int, ttl time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()

	now := time.Now()

	// Check if key already exists
	if item, exists := c.items[key]; exists {
		// Update existing item
		item.packs = packs
		item.total = total
		item.expiration = now.Add(ttl)
		c.moveToFront(item.node)
		return
	}

	// If at max size, evict least recently used item
	if len(c.items) >= c.maxSize {
		c.evictLRU()
	}

	// Create new node and add to front
	node := &lruNode{key: key}
	c.items[key] = &cacheItem{
		packs:      packs,
		total:      total,
		expiration: now.Add(ttl),
		node:       node,
	}
	c.addToFront(node)
}

// evictLRU removes the least recently used item in O(1)
func (c *MemoryCache) evictLRU() {
	if c.tail == nil {
		return
	}

	// Remove tail (least recently used)
	key := c.tail.key
	c.removeNode(c.tail)
	delete(c.items, key)
}

// addToFront adds a node to the front (most recently used)
func (c *MemoryCache) addToFront(node *lruNode) {
	node.next = c.head
	node.prev = nil

	if c.head != nil {
		c.head.prev = node
	}
	c.head = node

	if c.tail == nil {
		c.tail = node
	}
}

// removeNode removes a node from the linked list
func (c *MemoryCache) removeNode(node *lruNode) {
	if node.prev != nil {
		node.prev.next = node.next
	} else {
		c.head = node.next
	}

	if node.next != nil {
		node.next.prev = node.prev
	} else {
		c.tail = node.prev
	}
}

// moveToFront moves a node to the front (most recently used)
func (c *MemoryCache) moveToFront(node *lruNode) {
	if node == c.head {
		return // Already at front
	}

	c.removeNode(node)
	c.addToFront(node)
}

// Clear removes all cached items
func (c *MemoryCache) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.items = make(map[string]*cacheItem)
	c.head = nil
	c.tail = nil
	atomic.StoreInt64(&c.hits, 0)
	atomic.StoreInt64(&c.misses, 0)
}

// Stats returns cache statistics with atomic reads
func (c *MemoryCache) Stats() CacheStats {
	c.mu.RLock()
	size := len(c.items)
	c.mu.RUnlock()

	hits := atomic.LoadInt64(&c.hits)
	misses := atomic.LoadInt64(&c.misses)
	total := hits + misses
	hitRatio := 0.0
	if total > 0 {
		hitRatio = float64(hits) / float64(total)
	}

	return CacheStats{
		Hits:     hits,
		Misses:   misses,
		HitRatio: hitRatio,
		Size:     size,
	}
}

// GenerateCacheKey creates a cache key from amount and pack sizes
// Optimized: Uses string builder instead of JSON for 10-20x performance
func GenerateCacheKey(amount int, packSizes []int) string {
	var b strings.Builder
	b.Grow(32 + len(packSizes)*6) // Pre-allocate capacity
	b.WriteString("calc:")
	b.WriteString(strconv.Itoa(amount))
	b.WriteByte(':')
	for i, size := range packSizes {
		if i > 0 {
			b.WriteByte(',')
		}
		b.WriteString(strconv.Itoa(size))
	}
	return b.String()
}

// NoOpCache is a cache that does nothing (for disabling cache)
type NoOpCache struct{}

func (c *NoOpCache) Get(key string) (map[int]int, int, bool) {
	return nil, 0, false
}

func (c *NoOpCache) Set(key string, packs map[int]int, total int, ttl time.Duration) {
}

func (c *NoOpCache) Clear() {
}

func (c *NoOpCache) Stats() CacheStats {
	return CacheStats{}
}

// RedisCache implements Redis-backed cache (placeholder for Redis integration)
// To implement: use github.com/go-redis/redis/v9
/*
type RedisCache struct {
	client *redis.Client
	ttl    time.Duration
}

func NewRedisCache(addr string, ttl time.Duration) (*RedisCache, error) {
	client := redis.NewClient(&redis.Options{
		Addr: addr,
	})

	ctx := context.Background()
	if err := client.Ping(ctx).Err(); err != nil {
		return nil, err
	}

	return &RedisCache{
		client: client,
		ttl:    ttl,
	}, nil
}

func (c *RedisCache) Get(key string) (map[int]int, int, bool) {
	ctx := context.Background()
	val, err := c.client.Get(ctx, key).Result()
	if err != nil {
		return nil, 0, false
	}

	var result struct {
		Packs map[int]int `json:"packs"`
		Total int         `json:"total"`
	}

	if err := json.Unmarshal([]byte(val), &result); err != nil {
		return nil, 0, false
	}

	return result.Packs, result.Total, true
}

func (c *RedisCache) Set(key string, packs map[int]int, total int, ttl time.Duration) {
	ctx := context.Background()
	data := struct {
		Packs map[int]int `json:"packs"`
		Total int         `json:"total"`
	}{
		Packs: packs,
		Total: total,
	}

	val, _ := json.Marshal(data)
	c.client.Set(ctx, key, val, ttl)
}

func (c *RedisCache) Clear() {
	ctx := context.Background()
	c.client.FlushDB(ctx)
}
*/
