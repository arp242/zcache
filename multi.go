package zcache

import (
	"sync"
	"time"
)

type (
	// A Keyset is a set of keys for a Cache. All operations are run on the keys
	// in the set.
	Keyset[K comparable, V any] struct {
		mu    sync.RWMutex
		cache *cache[K, V]
		keys  []K
	}
	multiRet[V any] struct {
		V  V
		Ok bool
	}
	staleRet[V any] struct {
		V       V
		expired bool
		Ok      bool
	}
	expireRet[V any] struct {
		V  V
		T  time.Time
		Ok bool
	}
)

// Keyset returns a new set of keys.
func (c *cache[K, V]) Keyset(k ...K) *Keyset[K, V] {
	return &Keyset[K, V]{cache: c, keys: k}
}

// Glob returns all keys matching the pattern.
func (c *cache[K, V]) Glob(patt string) *Keyset[K, V] {
	return &Keyset[K, V]{cache: c} // TODO: implement.
}

// Find keys with a function callback.
//
// The item will be included if the callback's first return argument is true.
// The loop will stop if the second return argument is true.
func (c *cache[K, V]) Find(filter func(key K, item Item[V]) (del, stop bool)) *Keyset[K, V] {
	return &Keyset[K, V]{cache: c} // TODO: implement.
}

// Keyset methods.

// Keys returns all keys in this keyset.
func (m *Keyset[K, V]) Keys() []K {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.keys
}

// Append new keys to this keyset.
func (m *Keyset[K, V]) Append(k ...K) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.keys = append(m.keys, k...)
}

// Reset this keyset to zero keys.
func (m *Keyset[K, V]) Reset() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.keys = make([]K, 0, 16)
}

// Cache methods.

func (m *Keyset[K, V]) Get() []multiRet[V] {
	var (
		keys = m.Keys()
		ret  = make([]multiRet[V], 0, len(keys))
	)

	m.cache.mu.RLock()
	defer m.cache.mu.RUnlock()
	for _, kk := range keys {
		item, ok := m.cache.items[kk]
		if !ok {
			ret = append(ret, multiRet[V]{})
			continue
		}
		if item.Expiration > 0 && time.Now().UnixNano() > item.Expiration {
			ret = append(ret, multiRet[V]{})
			continue
		}
		ret = append(ret, multiRet[V]{Ok: true, V: item.Object})
	}
	return ret
}
func (m *Keyset[K, V]) GetStale() []staleRet[V]                            { return nil }
func (m *Keyset[K, V]) GetWithExpire() []expireRet[V]                      { return nil }
func (m *Keyset[K, V]) Touch() []multiRet[V]                               { return nil }
func (m *Keyset[K, V]) TouchWithExpire(k K, d time.Duration) []multiRet[V] { return nil }
func (m *Keyset[K, V]) Delete()                                            {}
func (m *Keyset[K, V]) Pop() []multiRet[V]                                 { return nil }

// Setting and modifying values.
//
//   Keyset("key1", "key2").Set("val 1", "val 2")
//   Keyset("key1", "key2").Add("val 1", "val 2")
//
// Number of values for all of these must match number of keys.
//
// Not a huge fan of this API though... All other things being equal passing a
// struct slice with the key and value is better, IMHO.

func (m *Keyset[K, V]) Set(v ...V) {
	keys := m.Keys()
	if len(v) != len(keys) {
		// TODO: error?
		// return fmt.Errorf("zcache.Keyset.Set: Keyset has %d keys, but %d values given", len(v), len(keys))
	}

	m.cache.mu.RLock()
	defer m.cache.mu.RUnlock()
	for i, k := range keys {
		m.cache.set(k, v[i], m.cache.defaultExpiration)
	}
}
func (m *Keyset[K, V]) SetWithExpire(d time.Duration, v ...V)           {}
func (m *Keyset[K, V]) Add(v ...V) error                                { return nil }
func (m *Keyset[K, V]) AddWithExpire(d time.Duration, v ...V) error     { return nil }
func (m *Keyset[K, V]) Rename(dst ...K) bool                            { return false }
func (m *Keyset[K, V]) Replace(v ...V) error                            { return nil }
func (m *Keyset[K, V]) ReplaceWithExpire(d time.Duration, v ...V) error { return nil }
func (m *Keyset[K, V]) Modify(f func(K, V) V) []multiRet[V]             { return nil }
