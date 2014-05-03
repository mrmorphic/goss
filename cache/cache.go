package cache

// A simple cache mechanism for testing performance. This is intended to be used for global state
// across requests, such as SiteConfig. The caller is responsible for re-adding values if the cache
// is missed.
// This is a very naive caching mechanism. It is only suitable for small caches at this point,
// as the expiry locks the set while checking

// Ideas to try here include:
//  * add to cache with a function that can be called in a goroutine to refresh the value on
//    expiry. set-and-forget. Duration still specified.
//  * add to cache with a policy expiry function. The cache will poll the policy expiry functions.

import (
	"sync"
	"time"
)

type CacheEntry struct {
	value  interface{}
	expiry time.Time
}

var mutex sync.Mutex

var cache map[string]*CacheEntry

func Store(key string, value interface{}, lifetime time.Duration) {
	entry := &CacheEntry{value: value, expiry: time.Now().Add(lifetime)}
	mutex.Lock()
	cache[key] = entry
	mutex.Unlock()
}

func Get(key string) interface{} {
	mutex.Lock()
	entry := cache[key]
	mutex.Unlock()
	if entry == nil {
		return nil
	}
	return entry.value
}

// start up a 1 second ping to expiry cache entries past their expiry.
func init() {
	cache = map[string]*CacheEntry{}

	ticker := time.NewTicker(time.Second)
	quit := make(chan struct{})
	go func() {
		for {
			select {
			case <-ticker.C:
				n := time.Now().UnixNano()
				mutex.Lock()
				for k, v := range cache {
					e := v.expiry.UnixNano()
					if e <= n {
						delete(cache, k)
					}
				}
				mutex.Unlock()
			case <-quit:
				ticker.Stop()
				return
			}
		}
	}()
}
