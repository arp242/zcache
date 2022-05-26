// Package zcache is an in-memory key:value store/cache with time-based evictions.
//
// It is suitable for applications running on a single machine. Its major
// advantage is that it's essentially a thread-safe map with expiration times.
// Any object can be stored, for a given duration or forever, and the cache can
// be safely used by multiple goroutines.
//
// Although zcache isn't meant to be used as a persistent datastore, the
// contents can be saved to and loaded from a file (using `c.Items()` to
// retrieve the items map to serialize, and `NewFrom()` to create a cache from a
// deserialized one) to recover from downtime quickly.
package zcache

import (
	"fmt"
	"runtime"
	"sync"
	"time"
)

const (
	// NoExpiration indicates a cache item never expires.
	NoExpiration time.Duration = -1

	// DefaultExpiration indicates to use the cache default expiration time.
	// Equivalent to passing in the same expiration duration as was given to
	// New() or NewFrom() when the cache was created (e.g. 5 minutes.)
	DefaultExpiration time.Duration = 0
)

type (
	// Cache is a thread-safe in-memory key/value store.
	Cache[K comparable, V any] struct {
		*cache[K, V] // If this is confusing, see the comment at newCacheWithJanitor()
	}

	cache[K comparable, V any] struct {
		defaultExpiration time.Duration
		items             map[K]Item[V]
		mu                sync.RWMutex
		onEvicted         func(K, V)
		janitor           *janitor[K, V]
	}

	// Item stored in the cache; it holds the value and the expiration time as
	// timestamp.
	Item[V any] struct {
		Object     V
		Expiration int64
	}
)

// New creates a new cache with a given expiration duration and cleanup
// interval.
//
// If the expiration duration is less than 1 (or NoExpiration) the items in the
// cache never expire (by default) and must be deleted manually.
//
// If the cleanup interval is less than 1 expired items are not deleted from the
// cache before calling c.DeleteExpired().
func New[K comparable, V any](defaultExpiration, cleanupInterval time.Duration) *Cache[K, V] {
	return newCacheWithJanitor(defaultExpiration, cleanupInterval, make(map[K]Item[V]))
}

// NewFrom creates a new cache like New() and populates the cache with the given
// items.
//
// The passed map will serve as the underlying map for the cache. This is useful
// for starting from a deserialized cache (serialized using e.g. gob.Encode() on
// c.Items()), or passing in e.g. make(map[string]Item, 500) to improve startup
// performance when the cache is expected to reach a certain minimum size.
//
// The map is *not* copied and only the cache's methods synchronize access to this
// map, so it is not recommended to keep any references to the map around after
// creating a cache. If need be, the map can be accessed at a later point using
// c.Items() (which creates a copy of the map).
//
// Note regarding serialization: When using e.g. gob, make sure to
// gob.Register() the individual types stored in the cache before encoding a map
// retrieved with c.Items() and to register those same types before decoding a
// blob containing an items map.
func NewFrom[K comparable, V any](defaultExpiration, cleanupInterval time.Duration, items map[K]Item[V]) *Cache[K, V] {
	return newCacheWithJanitor(defaultExpiration, cleanupInterval, items)
}

func newCache[K comparable, V any](de time.Duration, m map[K]Item[V]) *cache[K, V] {
	if de == 0 {
		de = -1
	}
	c := &cache[K, V]{
		defaultExpiration: de,
		items:             m,
	}
	return c
}

func newCacheWithJanitor[K comparable, V any](de time.Duration, ci time.Duration, m map[K]Item[V]) *Cache[K, V] {
	c := newCache(de, m)
	// This trick ensures that the janitor goroutine (which is running
	// DeleteExpired on c forever) does not keep the returned C object from
	// being garbage collected. When it is garbage collected, the finalizer
	// stops the janitor goroutine, after which c can be collected.
	C := &Cache[K, V]{c}
	if ci > 0 {
		runJanitor(c, ci)
		runtime.SetFinalizer(C, stopJanitor[K, V])
	}
	return C
}

// Set a cache item, replacing any existing item.
func (c *cache[K, V]) Set(k K, v V) { c.SetWithExpire(k, v, DefaultExpiration) }

// Touch replaces the expiry of a key with the default expiration and returns
// the current value, if any.
//
// The boolean return value indicates if this item was set.
func (c *cache[K, V]) Touch(k K) (V, bool) { return c.TouchWithExpire(k, DefaultExpiration) }

// Add an item to the cache only if it doesn't exist yet or if it has expired.
//
// It will return an error if the cache key already exists.
func (c *cache[K, V]) Add(k K, v V) error { return c.AddWithExpire(k, v, DefaultExpiration) }

// Replace sets a new value for the key only if it already exists and isn't
// expired.
//
// It will return an error if the cache key doesn't exist.
func (c *cache[K, V]) Replace(k K, v V) error { return c.ReplaceWithExpire(k, v, DefaultExpiration) }

// SetWithExpire sets a cache item, replacing any existing item.
//
// If the duration is 0 (DefaultExpiration), the cache's default expiration time
// is used. If it is -1 (NoExpiration), the item never expires.
func (c *cache[K, V]) SetWithExpire(k K, v V, d time.Duration) {
	// "Inlining" of set
	var e int64
	if d == DefaultExpiration {
		d = c.defaultExpiration
	}
	if d > 0 {
		e = time.Now().Add(d).UnixNano()
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	c.items[k] = Item[V]{
		Object:     v,
		Expiration: e,
	}
}

// TouchWithExpire replaces the expiry of a key and returns the current value, if any.
//
// The boolean return value indicates if this item was set. If the duration is 0
// (DefaultExpiration), the cache's default expiration time is used. If it is -1
// (NoExpiration), the item never expires.
func (c *cache[K, V]) TouchWithExpire(k K, d time.Duration) (V, bool) {
	if d == DefaultExpiration {
		d = c.defaultExpiration
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	item, ok := c.items[k]
	if !ok {
		return c.zero(), false
	}

	item.Expiration = time.Now().Add(d).UnixNano()
	c.items[k] = item
	return item.Object, true
}

// AddWithExpire adds an item to the cache only if it doesn't exist yet, or if
// it has expired.
//
// It will return an error if the cache key already exists. If the duration is 0
// (DefaultExpiration), the cache's default expiration time is used. If it is -1
// (NoExpiration), the item never expires.
func (c *cache[K, V]) AddWithExpire(k K, v V, d time.Duration) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	_, ok := c.get(k)
	if ok {
		return fmt.Errorf("zcache.Add: item %v already exists", k)
	}
	c.set(k, v, d)
	return nil
}

// ReplaceWithExpire sets a new value for the key only if it already exists and isn't
// expired.
//
// It will return an error if the cache key doesn't exist. If the duration is 0
// (DefaultExpiration), the cache's default expiration time is used. If it is -1
// (NoExpiration), the item never expires.
func (c *cache[K, V]) ReplaceWithExpire(k K, v V, d time.Duration) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	_, ok := c.get(k)
	if !ok {
		return fmt.Errorf("zcache.Replace: item %v doesn't exist", k)
	}
	c.set(k, v, d)
	return nil
}

// Get an item from the cache.
//
// Returns the item or the zero value and a bool indicating whether the key is
// set.
func (c *cache[K, V]) Get(k K) (V, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	// "Inlining" of get and Expired
	item, ok := c.items[k]
	if !ok {
		return c.zero(), false
	}
	if item.Expiration > 0 && time.Now().UnixNano() > item.Expiration {
		return c.zero(), false
	}
	return item.Object, true
}

// GetStale gets an item from the cache without checking if it's expired.
//
// Returns the item or the zero value and a bool indicating whether the key was
// expired and a bool indicating whether the key was set.
func (c *cache[K, V]) GetStale(k K) (v V, expired bool, ok bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	// "Inlining" of get and Expired
	item, ok := c.items[k]
	if !ok {
		return c.zero(), false, false
	}
	return item.Object,
		item.Expiration > 0 && time.Now().UnixNano() > item.Expiration,
		true
}

// GetWithExpire returns an item and its expiration time from the cache.
//
// It returns the item or the zero value, the expiration time if one is set (if
// the item never expires a zero value for time.Time is returned), and a bool
// indicating whether the key was set.
func (c *cache[K, V]) GetWithExpire(k K) (V, time.Time, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	// "Inlining" of get and Expired
	item, ok := c.items[k]
	if !ok {
		return c.zero(), time.Time{}, false
	}

	if item.Expiration > 0 {
		if time.Now().UnixNano() > item.Expiration {
			return c.zero(), time.Time{}, false
		}

		// Return the item and the expiration time
		return item.Object, time.Unix(0, item.Expiration), true
	}

	// If expiration <= 0 (i.e. no expiration time set) then return the item
	// and a zeroed time.Time
	return item.Object, time.Time{}, true
}

// Modify the value of an existing key.
//
// This is thread-safe; for example to increment a number:
//
//   cache.Modify("one", func(v int) int { return v + 1 })
//
// Or setting a map key:
//
//   cache.Modify("key", func(v map[string]string) map[string]string {
//         v["k"] = "v"
//         return v
//   })
//
// This is thread-safe and can be safely run by multiple goroutines modifying
// the same key. If you would use Get() + Set() then two goroutines may Get()
// the same value and the modification of one of them will be lost.
//
// This is not run for keys that are not set yet; the boolean return indicates
// if the key was set and if the function was applied.
func (c *cache[K, V]) Modify(k K, f func(V) V) (V, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()

	// "Inlining" of get and Expired
	item, ok := c.items[k]
	if !ok {
		return c.zero(), false
	}
	if item.Expiration > 0 && time.Now().UnixNano() > item.Expiration {
		return c.zero(), false
	}

	item.Object = f(item.Object)
	c.items[k] = item
	return item.Object, true
}

// Delete an item from the cache. Does nothing if the key is not in the cache.
func (c *cache[K, V]) Delete(k K) {
	c.mu.Lock()
	v, evicted := c.delete(k)
	c.mu.Unlock()
	if evicted {
		c.onEvicted(k, v)
	}
}

// Rename a key; the value and expiry will be left untouched; onEvicted will not
// be called.
//
// Existing keys will be overwritten; returns false is the src key doesn't
// exist.
func (c *cache[K, V]) Rename(src, dst K) bool {
	c.mu.Lock()
	defer c.mu.Unlock()

	// "Inlining" of get and Expired
	item, ok := c.items[src]
	if !ok {
		return false
	}
	if item.Expiration > 0 && time.Now().UnixNano() > item.Expiration {
		return false
	}

	delete(c.items, src)
	c.items[dst] = item
	return true
}

// Pop gets an item from the cache and deletes it.
//
// The bool return indicates if the item was set.
func (c *cache[K, V]) Pop(k K) (V, bool) {
	c.mu.Lock()

	// "Inlining" of get and Expired
	item, ok := c.items[k]
	if !ok {
		c.mu.Unlock()
		return c.zero(), false
	}
	if item.Expiration > 0 && time.Now().UnixNano() > item.Expiration {
		c.mu.Unlock()
		return c.zero(), false
	}

	v, evicted := c.delete(k)
	c.mu.Unlock()
	if evicted {
		c.onEvicted(k, v)
	}

	return item.Object, true
}

// DeleteExpired deletes all expired items from the cache.
func (c *cache[K, V]) DeleteExpired() {
	var evictedItems []keyAndValue[K, V]
	now := time.Now().UnixNano()
	c.mu.Lock()

	for k, v := range c.items {
		// "Inlining" of expired
		if v.Expiration > 0 && now > v.Expiration {
			ov, evicted := c.delete(k)
			if evicted {
				evictedItems = append(evictedItems, keyAndValue[K, V]{k, ov})
			}
		}
	}
	c.mu.Unlock()
	for _, v := range evictedItems {
		c.onEvicted(v.key, v.value)
	}
}

// OnEvicted sets an function to call when an item is evicted from the cache.
//
// The function is run with the key and value. This is also run when a cache
// item is is deleted manually, but *not* when it is overwritten.
//
// Can be set to nil to disable it (the default).
func (c *cache[K, V]) OnEvicted(f func(K, V)) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.onEvicted = f
}

// Items returns a copy of all unexpired items in the cache.
func (c *cache[K, V]) Items() map[K]Item[V] {
	c.mu.RLock()
	defer c.mu.RUnlock()

	m := make(map[K]Item[V], len(c.items))
	now := time.Now().UnixNano()
	for k, v := range c.items {
		// "Inlining" of Expired
		if v.Expiration > 0 && now > v.Expiration {
			continue
		}
		m[k] = v
	}
	return m
}

// Keys gets a list of all keys, in no particular order.
func (c *cache[K, V]) Keys() []K {
	c.mu.RLock()
	defer c.mu.RUnlock()

	keys := make([]K, 0, len(c.items))
	now := time.Now().UnixNano()
	for k, v := range c.items {
		// "Inlining" of Expired
		if v.Expiration > 0 && now > v.Expiration {
			continue
		}
		keys = append(keys, k)
	}
	return keys
}

// ItemCount returns the number of items in the cache.
//
// This may include items that have expired but have not yet been cleaned up.
func (c *cache[K, V]) ItemCount() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return len(c.items)
}

// Reset deletes all items from the cache without calling OnEvicted.
func (c *cache[K, V]) Reset() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.items = map[K]Item[V]{}
}

// DeleteAll deletes all items from the cache and returns them.
//
// This calls OnEvicted for returned items.
func (c *cache[K, V]) DeleteAll() map[K]Item[V] {
	c.mu.Lock()
	items := c.items
	c.items = map[K]Item[V]{}
	c.mu.Unlock()

	if c.onEvicted != nil {
		for k, v := range items {
			c.onEvicted(k, v.Object)
		}
	}

	return items
}

// DeleteFunc deletes and returns cache items matched by the filter function.
//
// The item will be deleted if the callback's first return argument is true. The
// loop will stop if the second return argument is true.
//
// OnEvicted is called for deleted items.
func (c *cache[K, V]) DeleteFunc(filter func(key K, item Item[V]) (del, stop bool)) map[K]Item[V] {
	c.mu.Lock()
	m := map[K]Item[V]{}
	for k, v := range c.items {
		del, stop := filter(k, v)
		if del {
			m[k] = Item[V]{
				Object:     v.Object,
				Expiration: v.Expiration,
			}
			c.delete(k)
		}
		if stop {
			break
		}
	}
	c.mu.Unlock()

	if c.onEvicted != nil {
		for k, v := range m {
			c.onEvicted(k, v.Object)
		}
	}

	return m
}

func (c *cache[K, V]) set(k K, v V, d time.Duration) {
	var e int64
	if d == DefaultExpiration {
		d = c.defaultExpiration
	}
	if d > 0 {
		e = time.Now().Add(d).UnixNano()
	}
	c.items[k] = Item[V]{
		Object:     v,
		Expiration: e,
	}
}

func (c *cache[K, V]) get(k K) (V, bool) {
	item, ok := c.items[k]
	if !ok {
		return c.zero(), false
	}
	// "Inlining" of Expired
	if item.Expiration > 0 && time.Now().UnixNano() > item.Expiration {
		return c.zero(), false
	}
	return item.Object, true
}

func (c *cache[K, V]) delete(k K) (V, bool) {
	if c.onEvicted != nil {
		if v, ok := c.items[k]; ok {
			delete(c.items, k)
			return v.Object, true
		}
	}
	delete(c.items, k)

	return c.zero(), false
}

func (c *cache[K, V]) zero() V {
	var zeroValue V
	return zeroValue
}

type keyAndValue[K comparable, V any] struct {
	key   K
	value V
}

type janitor[K comparable, V any] struct {
	Interval time.Duration
	stop     chan bool
}

func (j *janitor[K, V]) run(c *cache[K, V]) {
	ticker := time.NewTicker(j.Interval)
	for {
		select {
		case <-ticker.C:
			c.DeleteExpired()
		case <-j.stop:
			ticker.Stop()
			return
		}
	}
}

func stopJanitor[K comparable, V any](c *Cache[K, V]) {
	c.janitor.stop <- true
}

func runJanitor[K comparable, V any](c *cache[K, V], ci time.Duration) {
	j := &janitor[K, V]{
		Interval: ci,
		stop:     make(chan bool),
	}
	c.janitor = j
	go j.run(c)
}
