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
type Proxy struct {
	cache *Cache
	mu    sync.RWMutex
	m     map[string]string
}

// NewProxy creates a new proxied cache.
func NewProxy(c *Cache) *Proxy {
	return &Proxy{cache: c, m: make(map[string]string)}
}

// Proxy items from "proxyKey" to "mainKey".
func (p *Proxy) Proxy(mainKey, proxyKey string) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.m[proxyKey] = mainKey
}

// Delete stops proxying "proxyKey" to "mainKey".
//
// This only removes the proxy link, not the entry from the main cache.
func (p *Proxy) Delete(proxyKey string) {
	p.mu.Lock()
	defer p.mu.Unlock()
	delete(p.m, proxyKey)
}

// Flush removes all proxied keys.
func (p *Proxy) Flush() {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.m = make(map[string]string)
}

// Key gets the main key for this proxied entry, if it exist.
func (p *Proxy) Key(proxyKey string) (string, bool) {
	p.mu.RLock()
	defer p.mu.RUnlock()
	mainKey, ok := p.m[proxyKey]
	return mainKey, ok
}

// Cache gets the associated cache.
func (p *Proxy) Cache() *Cache {
	return p.cache
}

// Set a new item in the main cache with the key mainKey, and proxy to that with
// proxyKey.
//
// This behaves like zcache.Cache.SetDefault() otherwise.
func (p *Proxy) Set(mainKey, proxyKey string, v interface{}) {
	p.mu.Lock()
	p.m[proxyKey] = mainKey
	p.mu.Unlock()
	p.cache.SetDefault(mainKey, v)
}

// Get a proxied cache item with zcache.Cache.Get()
func (p *Proxy) Get(proxyKey string) (interface{}, bool) {
	p.mu.RLock()
	mainKey, ok := p.m[proxyKey]
	if !ok {
		p.mu.RUnlock()
		return nil, false
	}
	p.mu.RUnlock()

	return p.cache.Get(mainKey)
}

// Items gets all items in this proxy, as proxyKey → mainKey
func (p *Proxy) Items() map[string]string {
	p.mu.RLock()
	defer p.mu.RUnlock()

	m := make(map[string]string, len(p.m))
	for k, v := range p.m {
		m[k] = v
	}
	return m
}
