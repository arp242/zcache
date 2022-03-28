package zcache

import (
	"sync"
)

// Proxy a cache, allowing access to the same cache entries with different keys.
//
// This is useful if you want to keep a cache which may be accessed by different
// keys in various different code paths. For example, a "site" may be accessed
// by ID or by CNAME. Proxy keys can have a different type than cache keys.
//
// Proxy keys  don't have an expiry and are never automatically deleted, the
// logic being that the same "proxy → key" mapping should always be valid. The
// items in the underlying cache can still be expired or deleted, and you can
// still manually call Delete() or Reset().
type Proxy[ProxyK, MainK comparable, V any] struct {
	cache *Cache[MainK, V]
	mu    sync.RWMutex
	m     map[ProxyK]MainK
}

// NewProxy creates a new proxied cache.
func NewProxy[ProxyK, MainK comparable, V any](c *Cache[MainK, V]) *Proxy[ProxyK, MainK, V] {
	return &Proxy[ProxyK, MainK, V]{cache: c, m: make(map[ProxyK]MainK)}
}

// Proxy items from "proxyKey" to "mainKey".
func (p *Proxy[ProxyK, MainK, V]) Proxy(mainKey MainK, proxyKey ProxyK) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.m[proxyKey] = mainKey
}

// Delete stops proxying "proxyKey" to "mainKey".
//
// This only removes the proxy link, not the entry from the main cache.
func (p *Proxy[ProxyK, MainK, V]) Delete(proxyKey ProxyK) {
	p.mu.Lock()
	defer p.mu.Unlock()
	delete(p.m, proxyKey)
}

// Reset removes all proxied keys (but not the underlying cache).
func (p *Proxy[ProxyK, MainK, V]) Reset() {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.m = make(map[ProxyK]MainK)
}

// Key gets the main key for this proxied entry, if it exist.
//
// The boolean value indicates if this proxy key is set.
func (p *Proxy[ProxyK, MainK, V]) Key(proxyKey ProxyK) (MainK, bool) {
	p.mu.RLock()
	defer p.mu.RUnlock()
	mainKey, ok := p.m[proxyKey]
	return mainKey, ok
}

// Cache gets the associated cache.
func (p *Proxy[ProxyK, MainK, V]) Cache() *Cache[MainK, V] {
	return p.cache
}

// Set a new item in the main cache with the key mainKey, and proxy to that with
// proxyKey.
//
// This behaves like zcache.Cache.Set() otherwise.
func (p *Proxy[ProxyK, MainK, V]) Set(mainKey MainK, proxyKey ProxyK, v V) {
	p.mu.Lock()
	p.m[proxyKey] = mainKey
	p.mu.Unlock()
	p.cache.Set(mainKey, v)
}

// Get a proxied cache item with zcache.Cache.Get()
func (p *Proxy[ProxyK, MainK, V]) Get(proxyKey ProxyK) (V, bool) {
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
func (p *Proxy[ProxyK, MainK, V]) Items() map[ProxyK]MainK {
	p.mu.RLock()
	defer p.mu.RUnlock()

	m := make(map[ProxyK]MainK, len(p.m))
	for k, v := range p.m {
		m[k] = v
	}
	return m
}
