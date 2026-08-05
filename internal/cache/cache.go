package cache

import (
	"strings"
	"sync"
	"time"

	lru "github.com/hashicorp/golang-lru/v2"

	"github.com/vendemas/imagekit/internal/metrics"
)

type entry struct {
	data        []byte
	contentType string
	expiresAt   time.Time
	notFound    bool
}

type LRUCache struct {
	mu     sync.RWMutex
	cache  *lru.Cache[string, *entry]
	ttl    time.Duration
}

func NewLRUCache(maxSizeMB, ttlSec int) *LRUCache {
	maxKeys := 10000
	if maxSizeMB > 0 {
		maxKeys = maxSizeMB * 20
	}

	c, err := lru.New[string, *entry](maxKeys)
	if err != nil {
		panic("failed to create lru cache: " + err.Error())
	}

	return &LRUCache{
		cache: c,
		ttl:   time.Duration(ttlSec) * time.Second,
	}
}

func (c *LRUCache) Get(key string) ([]byte, string, bool) {
	c.mu.RLock()
	e, ok := c.cache.Get(key)
	c.mu.RUnlock()

	if !ok || e.notFound {
		metrics.CacheMisses.Inc()
		return nil, "", false
	}

	if time.Now().After(e.expiresAt) {
		c.mu.Lock()
		c.cache.Remove(key)
		c.mu.Unlock()
		metrics.CacheEvictions.Inc()
		metrics.CacheMisses.Inc()
		return nil, "", false
	}

	metrics.CacheHits.Inc()
	return e.data, e.contentType, true
}

func (c *LRUCache) Set(key string, data []byte, contentType string) {
	if len(data) == 0 {
		return
	}

	e := &entry{
		data:        data,
		contentType: contentType,
		expiresAt:   time.Now().Add(c.ttl),
	}

	c.mu.Lock()
	c.cache.Add(key, e)
	size := c.cache.Len()
	c.mu.Unlock()

	metrics.CacheSize.Set(float64(size))
}

func (c *LRUCache) SetNotFound(key string, ttl time.Duration) {
	e := &entry{
		notFound:  true,
		expiresAt: time.Now().Add(ttl),
	}

	c.mu.Lock()
	c.cache.Add(key, e)
	size := c.cache.Len()
	c.mu.Unlock()

	metrics.CacheSize.Set(float64(size))
}

func (c *LRUCache) GetNotFound(key string) bool {
	c.mu.RLock()
	e, ok := c.cache.Get(key)
	c.mu.RUnlock()

	if !ok || !e.notFound {
		return false
	}

	if time.Now().After(e.expiresAt) {
		c.mu.Lock()
		c.cache.Remove(key)
		c.mu.Unlock()
		metrics.CacheEvictions.Inc()
		return false
	}

	return true
}

func (c *LRUCache) InvalidatePrefix(prefix string) {
	c.mu.Lock()
	keys := c.cache.Keys()
	for _, k := range keys {
		if strings.HasPrefix(k, prefix) {
			c.cache.Remove(k)
			metrics.CacheEvictions.Inc()
		}
	}
	newSize := c.cache.Len()
	c.mu.Unlock()
	metrics.CacheSize.Set(float64(newSize))
}

func (c *LRUCache) Len() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.cache.Len()
}

func (c *LRUCache) Purge() {
	c.mu.Lock()
	c.cache.Purge()
	c.mu.Unlock()
	metrics.CacheSize.Set(0)
}
