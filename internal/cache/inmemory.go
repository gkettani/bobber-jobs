package cache

import "sync"

type InMemoryCache struct {
	cache map[string]string
	mutex sync.RWMutex
}

func NewInMemoryCache() *InMemoryCache {
	return &InMemoryCache{
		cache: make(map[string]string),
	}
}

func (c *InMemoryCache) Get(key string) (string, bool) {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	value, exists := c.cache[key]
	return value, exists
}

func (c *InMemoryCache) Set(key, value string) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	c.cache[key] = value
}

func (c *InMemoryCache) Exists(key string) bool {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	_, exists := c.cache[key]
	return exists
}

func (c *InMemoryCache) Delete(key string) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	delete(c.cache, key)
}
