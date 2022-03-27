package zcache

import (
	"sync"
)

// Proxy a cache, allowing access to the same cache entries with different keys.
//
// This is useful if you want to keep a cache which may be accessed by different
// keys in various different code paths. For example, a "site" may be accessed
// by ID or by CNAME.
//
// Proxies cache entries don't have an expiry and are never automatically
// deleted, the logic being that the same "proxy → key" mapping should always be
// valid. The items in the underlying cache can still be expired or deleted, and
// you can still manually call Delete() or Flush().
type Proxy[K comparable, V any] struct {
	cache *Cache[K, V]
	mu    sync.RWMutex
	m     map[K]K
}

// NewProxy creates a new proxied cache.
func NewProxy[K comparable, V any](c *Cache[K, V]) *Proxy[K, V] {
	return &Proxy[K, V]{cache: c, m: make(map[K]K)}
}

// Proxy items from "proxyKey" to "mainKey".
func (p *Proxy[K, V]) Proxy(mainKey, proxyKey K) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.m[proxyKey] = mainKey
}

// Delete stops proxying "proxyKey" to "mainKey".
//
// This only removes the proxy link, not the entry from the main cache.
func (p *Proxy[K, V]) Delete(proxyKey K) {
	p.mu.Lock()
	defer p.mu.Unlock()
	delete(p.m, proxyKey)
}

// Flush removes all proxied keys.
func (p *Proxy[K, V]) Flush() {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.m = make(map[K]K)
}

// Key gets the main key for this proxied entry, if it exist.
func (p *Proxy[K, V]) Key(proxyKey K) (K, bool) {
	p.mu.RLock()
	defer p.mu.RUnlock()
	mainKey, ok := p.m[proxyKey]
	return mainKey, ok
}

// Cache gets the associated cache.
func (p *Proxy[K, V]) Cache() *Cache[K, V] {
	return p.cache
}

// Set a new item in the main cache with the key mainKey, and proxy to that with
// proxyKey.
//
// This behaves like zcache.Cache.Set() otherwise.
func (p *Proxy[K, V]) Set(mainKey, proxyKey K, v V) {
	p.mu.Lock()
	p.m[proxyKey] = mainKey
	p.mu.Unlock()
	p.cache.Set(mainKey, v)
}

// Get a proxied cache item with zcache.Cache.Get()
func (p *Proxy[K, V]) Get(proxyKey K) (V, bool) {
	p.mu.RLock()
	mainKey, ok := p.m[proxyKey]
	if !ok {
		p.mu.RUnlock()
		return p.cache.zero(), false
	}
	p.mu.RUnlock()

	return p.cache.Get(mainKey)
}

// Items gets all items in this proxy, as proxyKey → mainKey
func (p *Proxy[K, V]) Items() map[K]K {
	p.mu.RLock()
	defer p.mu.RUnlock()

	m := make(map[K]K, len(p.m))
	for k, v := range p.m {
		m[k] = v
	}
	return m
}
