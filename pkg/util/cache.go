package util

import "sync"

type ConsistentCache struct {
	cache sync.Map
}

var tableMap ConsistentCache = ConsistentCache{cache: sync.Map{}}

func GetTableName(key string) string {
	if value, ok := tableMap.cache.Load(key); ok {
		return value.(string)
	}
	return ""

}

func SetTableName(key, value string) {
	tableMap.cache.Store(key, value)
}
