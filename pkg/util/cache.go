package util

import "sync"

type ConsistentCache struct {
	cache sync.Map
}

func (c *ConsistentCache) Get(key string) string {
	if value, ok := c.cache.Load(key); ok {
		return value.(string)
	}
	return ""

}

func (c *ConsistentCache) Set(key, value string) {
	c.cache.Store(key, value)
}
