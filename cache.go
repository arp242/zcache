//go:generate sh -c "go run gen.go > incr.go"

package zcache

import (
	"encoding/gob"
	"errors"
	"io"
	"os"
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

// Item stored in the cache; it holds the value and the expiration time as
// timestamp.
type Item struct {
	Object     interface{}
	Expiration int64
}

// Expired reports if this item has expired.
func (item Item) Expired() bool {
	if item.Expiration == 0 {
		return false
	}
	return time.Now().UnixNano() > item.Expiration
}

type Cache struct {
	*cache // If this is confusing, see the comment at newCacheWithJanitor()
}

type cache struct {
	defaultExpiration time.Duration
	items             map[string]Item
	mu                sync.RWMutex
	onEvicted         func(string, interface{})
	janitor           *janitor
}

// Set a cache item, replacing any existing item.
//
// If the duration is 0 (DefaultExpiration), the cache's default expiration time
// is used. If it is -1 (NoExpiration), the item never expires.
func (c *cache) Set(k string, v interface{}, d time.Duration) {
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
	c.items[k] = Item{
		Object:     v,
		Expiration: e,
	}
}

// SetDefault calls Set() with the default expiration for this cache.
func (c *cache) SetDefault(k string, v interface{}) {
	c.Set(k, v, DefaultExpiration)
}

// Touch replaces the expiry of a key and returns the current value, if any.
func (c *cache) Touch(k string, d time.Duration) (interface{}, bool) {
	if d == DefaultExpiration {
		d = c.defaultExpiration
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	item, ok := c.items[k]
	if !ok {
		return nil, false
	}

	item.Expiration = time.Now().Add(d).UnixNano()
	c.items[k] = item
	return item.Object, true
}

func (c *cache) set(k string, v interface{}, d time.Duration) {
	var e int64
	if d == DefaultExpiration {
		d = c.defaultExpiration
	}
	if d > 0 {
		e = time.Now().Add(d).UnixNano()
	}
	c.items[k] = Item{
		Object:     v,
		Expiration: e,
	}
}

// Add an item to the cache only if it doesn't exist yet, or if it has expired.
//
// It will return an error if the cache key exists.
func (c *cache) Add(k string, v interface{}, d time.Duration) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	_, ok := c.get(k)
	if ok {
		return errors.New("zcache.Add: item " + k + "already exists")
	}
	c.set(k, v, d)
	return nil
}

// Replace sets a new value for the key only if it already exists and isn't
// expired.
//
// It will return an error if the cache key doesn't exist.
func (c *cache) Replace(k string, v interface{}, d time.Duration) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	_, ok := c.get(k)
	if !ok {
		return errors.New("zcache.Replace: item " + k + " doesn't exist")
	}
	c.set(k, v, d)
	return nil
}

// Get an item from the cache.
//
// Returns the item or nil and a bool indicating whether the key is set.
func (c *cache) Get(k string) (interface{}, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	// "Inlining" of get and Expired
	item, ok := c.items[k]
	if !ok {
		return nil, false
	}
	if item.Expiration > 0 && time.Now().UnixNano() > item.Expiration {
		return nil, false
	}
	return item.Object, true
}

var getOrSetOnce once

// GetOrSet attempts to get a cache key, and will populate it with the return
// value of f if it's not yet set.
//
// This call will block while f() is running, as will all subsequent calls to
// GetOrSet() for the same key (but not other functions, such as Get()). The
// cache is not locked while f() is running.
func (c *cache) GetOrSet(k string, f func() (interface{}, time.Duration)) interface{} {
	c.mu.RLock()
	item, ok := c.items[k]
	c.mu.RUnlock()

	if !ok || item.Expiration > 0 && time.Now().UnixNano() > item.Expiration {
		ran := getOrSetOnce.Do(k, func() {
			// TODO: what if a Set() call modifies the key in-between the first
			// GetOrSet() setting it, and the 2nd GetOrSet() getting the value?
			v, d := f()
			c.mu.Lock()
			c.set(k, v, d)
			c.mu.Unlock()
			item.Object = v
		})
		if ran {
			// TODO: because this is checked after the lock in Do(), resetting
			// the key here will cause it to run again. Should think of a better
			// way to do this in once.
			// Note this breaks "go test -count=2".
			go func() {
				time.Sleep(5 * time.Millisecond)
				getOrSetOnce.Forget(k)
			}()
		} else {
			c.mu.RLock()
			item, _ = c.items[k]
			c.mu.RUnlock()
		}
	}

	return item.Object
}

// GetStale gets an item from the cache without checking if it's expired.
//
// Returns the item or nil, a bool indicating that the item is expired, and a
// bool indicating whether the key was found.
func (c *cache) GetStale(k string) (v interface{}, expired bool, ok bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	// "Inlining" of get and Expired
	item, ok := c.items[k]
	if !ok {
		return nil, false, false
	}
	return item.Object,
		item.Expiration > 0 && time.Now().UnixNano() > item.Expiration,
		true
}

// GetWithExpiration returns an item and its expiration time from the cache.
// It returns the item or nil, the expiration time if one is set (if the item
// never expires a zero value for time.Time is returned), and a bool indicating
// whether the key was found.
func (c *cache) GetWithExpiration(k string) (interface{}, time.Time, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	// "Inlining" of get and Expired
	item, ok := c.items[k]
	if !ok {
		return nil, time.Time{}, false
	}

	if item.Expiration > 0 {
		if time.Now().UnixNano() > item.Expiration {
			return nil, time.Time{}, false
		}

		// Return the item and the expiration time
		return item.Object, time.Unix(0, item.Expiration), true
	}

	// If expiration <= 0 (i.e. no expiration time set) then return the item
	// and a zeroed time.Time
	return item.Object, time.Time{}, true
}

func (c *cache) get(k string) (interface{}, bool) {
	item, ok := c.items[k]
	if !ok {
		return nil, false
	}
	// "Inlining" of Expired
	if item.Expiration > 0 && time.Now().UnixNano() > item.Expiration {
		return nil, false
	}
	return item.Object, true
}

// Modify the value of an existing key; this can be used for appending to a list
// or setting map keys:
//
//   zcache.Modify("key", func(v interface{}) interface{} {
//         vv = v.(map[string]string)
//         vv["k"] = "v"
//         return vv
//   })
//
// This is not run for keys that are not set yet; the boolean return indicates
// if the key was set and if the function was applied.
func (c *cache) Modify(k string, f func(interface{}) interface{}) bool {
	c.mu.Lock()
	defer c.mu.Unlock()

	// "Inlining" of get and Expired
	item, ok := c.items[k]
	if !ok {
		return false
	}
	if item.Expiration > 0 && time.Now().UnixNano() > item.Expiration {
		return false
	}

	item.Object = f(item.Object)
	c.items[k] = item
	return true
}

// Increment an item of type int, int8, int16, int32, int64, uintptr, uint,
// uint8, uint32, or uint64, float32 or float64 by n. Returns an error if the
// item's value is not an integer, if it was not found, or if it is not
// possible to increment it by n. To retrieve the incremented value, use one
// of the specialized methods, e.g. IncrementInt64.
func (c *cache) Increment(k string, n int64) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	v, ok := c.items[k]
	if !ok || v.Expired() {
		return errors.New("zcache.Increment: item " + k + " not found")
	}
	switch v.Object.(type) {
	case int:
		v.Object = v.Object.(int) + int(n)
	case int8:
		v.Object = v.Object.(int8) + int8(n)
	case int16:
		v.Object = v.Object.(int16) + int16(n)
	case int32:
		v.Object = v.Object.(int32) + int32(n)
	case int64:
		v.Object = v.Object.(int64) + n
	case uint:
		v.Object = v.Object.(uint) + uint(n)
	case uintptr:
		v.Object = v.Object.(uintptr) + uintptr(n)
	case uint8:
		v.Object = v.Object.(uint8) + uint8(n)
	case uint16:
		v.Object = v.Object.(uint16) + uint16(n)
	case uint32:
		v.Object = v.Object.(uint32) + uint32(n)
	case uint64:
		v.Object = v.Object.(uint64) + uint64(n)
	case float32:
		v.Object = v.Object.(float32) + float32(n)
	case float64:
		v.Object = v.Object.(float64) + float64(n)
	default:
		return errors.New("zcache.Incremeny: the value for " + k + " is not an integer")
	}
	c.items[k] = v
	return nil
}

// Decrement an item of type int, int8, int16, int32, int64, uintptr, uint,
// uint8, uint32, or uint64, float32 or float64 by n. Returns an error if the
// item's value is not an integer, if it was not found, or if it is not
// possible to decrement it by n. To retrieve the decremented value, use one
// of the specialized methods, e.g. DecrementInt64.
func (c *cache) Decrement(k string, n int64) error {
	// TODO: Implement Increment and Decrement more cleanly.
	// (Cannot do Increment(k, n*-1) for uints.)
	c.mu.Lock()
	defer c.mu.Unlock()

	v, ok := c.items[k]
	if !ok || v.Expired() {
		return errors.New("zcache.Decrement: item not found")
	}
	switch v.Object.(type) {
	case int:
		v.Object = v.Object.(int) - int(n)
	case int8:
		v.Object = v.Object.(int8) - int8(n)
	case int16:
		v.Object = v.Object.(int16) - int16(n)
	case int32:
		v.Object = v.Object.(int32) - int32(n)
	case int64:
		v.Object = v.Object.(int64) - n
	case uint:
		v.Object = v.Object.(uint) - uint(n)
	case uintptr:
		v.Object = v.Object.(uintptr) - uintptr(n)
	case uint8:
		v.Object = v.Object.(uint8) - uint8(n)
	case uint16:
		v.Object = v.Object.(uint16) - uint16(n)
	case uint32:
		v.Object = v.Object.(uint32) - uint32(n)
	case uint64:
		v.Object = v.Object.(uint64) - uint64(n)
	case float32:
		v.Object = v.Object.(float32) - float32(n)
	case float64:
		v.Object = v.Object.(float64) - float64(n)
	default:
		return errors.New("zcache.Decrement: the value for " + k + " is not an integer")
	}
	c.items[k] = v
	return nil
}

// Delete an item from the cache. Does nothing if the key is not in the cache.
func (c *cache) Delete(k string) {
	c.mu.Lock()
	v, evicted := c.delete(k)
	c.mu.Unlock()
	if evicted {
		c.onEvicted(k, v)
	}
}

// Pop gets an item from the cache and deletes it.
func (c *cache) Pop(k string) (interface{}, bool) {
	c.mu.Lock()

	// "Inlining" of get and Expired
	item, ok := c.items[k]
	if !ok {
		c.mu.Unlock()
		return nil, false
	}
	if item.Expiration > 0 && time.Now().UnixNano() > item.Expiration {
		c.mu.Unlock()
		return nil, false
	}

	v, evicted := c.delete(k)
	c.mu.Unlock()
	if evicted {
		c.onEvicted(k, v)
	}

	return item.Object, true
}

func (c *cache) delete(k string) (interface{}, bool) {
	if c.onEvicted != nil {
		if v, ok := c.items[k]; ok {
			delete(c.items, k)
			return v.Object, true
		}
	}
	delete(c.items, k)
	return nil, false
}

type keyAndValue struct {
	key   string
	value interface{}
}

// DeleteExpired deletes all expired items from the cache.
func (c *cache) DeleteExpired() {
	var evictedItems []keyAndValue
	now := time.Now().UnixNano()
	c.mu.Lock()

	for k, v := range c.items {
		// "Inlining" of expired
		if v.Expiration > 0 && now > v.Expiration {
			ov, evicted := c.delete(k)
			if evicted {
				evictedItems = append(evictedItems, keyAndValue{k, ov})
			}
		}
	}
	c.mu.Unlock()
	for _, v := range evictedItems {
		c.onEvicted(v.key, v.value)
	}
}

// OnEvicted sets an (optional) function that is called with the key and value
// when an item is evicted from the cache. (Including when it is deleted
// manually, but not when it is overwritten.) Set to nil to disable.
func (c *cache) OnEvicted(f func(string, interface{})) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.onEvicted = f
}

// Save the cache's items (using Gob) to an io.Writer.
//
// NOTE: This method is deprecated in favor of c.Items() and NewFrom() (see the
// documentation for NewFrom().)
func (c *cache) Save(w io.Writer) (err error) {
	enc := gob.NewEncoder(w)
	defer func() {
		if rec := recover(); rec != nil {
			err = errors.New("zcache.Save: error registering item types with Gob library")
		}
	}()
	c.mu.RLock()
	defer c.mu.RUnlock()
	for _, v := range c.items {
		gob.Register(v.Object)
	}
	err = enc.Encode(&c.items)
	return
}

// SaveFile writes the cache's items to the given filename, creating the file if
// it doesn't exist, and overwriting it if it does.
//
// NOTE: This method is deprecated in favor of c.Items() and NewFrom() (see the
// documentation for NewFrom().)
func (c *cache) SaveFile(fname string) error {
	fp, err := os.Create(fname)
	if err != nil {
		return err
	}
	err = c.Save(fp)
	if err != nil {
		fp.Close()
		return err
	}
	return fp.Close()
}

// Load (Gob-serialized) cache items from an io.Reader, excluding any items with
// keys that already exist (and haven't expired) in the current cache.
//
// NOTE: This method is deprecated in favor of c.Items() and NewFrom() (see the
// documentation for NewFrom().)
func (c *cache) Load(r io.Reader) error {
	dec := gob.NewDecoder(r)
	items := map[string]Item{}
	err := dec.Decode(&items)
	if err == nil {
		c.mu.Lock()
		defer c.mu.Unlock()
		for k, v := range items {
			ov, ok := c.items[k]
			if !ok || ov.Expired() {
				c.items[k] = v
			}
		}
	}
	return err
}

// LoadFile reads cache items from the given filename, excluding any items with
// keys that already exist in the current cache.
//
// NOTE: This method is deprecated in favor of c.Items() and NewFrom() (see the
// documentation for NewFrom().)
func (c *cache) LoadFile(fname string) error {
	fp, err := os.Open(fname)
	if err != nil {
		return err
	}
	err = c.Load(fp)
	if err != nil {
		fp.Close()
		return err
	}
	return fp.Close()
}

// Items returns a copy of all unexpired items in the cache.
func (c *cache) Items() map[string]Item {
	c.mu.RLock()
	defer c.mu.RUnlock()

	m := make(map[string]Item, len(c.items))
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
func (c *cache) Keys() []string {
	c.mu.RLock()
	defer c.mu.RUnlock()

	keys := make([]string, 0, len(c.items))
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
// This may include items that have expired, but have not yet been cleaned up.
func (c *cache) ItemCount() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return len(c.items)
}

// Flush deletes all items from the cache without calling onEvicted.
//
// This is a way to reset the cache to its original state.
func (c *cache) Flush() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.items = map[string]Item{}
}

// DeleteAll deletes all items from the cache and returns them.
// Note that onEvicted is called on returned items.
func (c *cache) DeleteAll() map[string]Item {
	c.mu.Lock()
	items := c.items
	c.items = map[string]Item{}
	c.mu.Unlock()

	if c.onEvicted != nil {
		for k, v := range items {
			c.onEvicted(k, v.Object)
		}
	}

	return items
}

// Filter is the function definition of the DeleteFunc method parameter.
// See DeleteFunc for more information.
type Filter func(key string, item Item) (del bool, stop bool)

// DeleteFunc deletes and returns filtered items from the cache.
// If `del` is true for `fn` call for an item, the item is deleted from the cache and returned.
// And if `stop` is returned as true, the filter won't be applied to the rest of the items in the cache.
// Note that onEvicted is called on returned items.
func (c *cache) DeleteFunc(fn Filter) map[string]Item {
	c.mu.Lock()
	m := map[string]Item{}
	for k, v := range c.items {
		del, stop := fn(k, v)

		if del {
			m[k] = Item{
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

type janitor struct {
	Interval time.Duration
	stop     chan bool
}

func (j *janitor) run(c *cache) {
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

func stopJanitor(c *Cache) {
	c.janitor.stop <- true
}

func runJanitor(c *cache, ci time.Duration) {
	j := &janitor{
		Interval: ci,
		stop:     make(chan bool),
	}
	c.janitor = j
	go j.run(c)
}

func newCache(de time.Duration, m map[string]Item) *cache {
	if de == 0 {
		de = -1
	}
	c := &cache{
		defaultExpiration: de,
		items:             m,
	}
	return c
}

func newCacheWithJanitor(de time.Duration, ci time.Duration, m map[string]Item) *Cache {
	c := newCache(de, m)
	// This trick ensures that the janitor goroutine (which – if enabled – is
	// running DeleteExpired on c forever) does not keep the returned C object
	// from being garbage collected. When it is garbage collected, the finalizer
	// stops the janitor goroutine, after which c can be collected.
	C := &Cache{c}
	if ci > 0 {
		runJanitor(c, ci)
		runtime.SetFinalizer(C, stopJanitor)
	}
	return C
}

// New creates a new cache with a given default expiration duration and cleanup
// interval.
//
// If the expiration duration is less than one (or NoExpiration), the items in
// the cache never expire (by default), and must be deleted manually.
//
// If the cleanup interval is less than one, expired items are not deleted from
// the cache before calling c.DeleteExpired().
func New(defaultExpiration, cleanupInterval time.Duration) *Cache {
	return newCacheWithJanitor(defaultExpiration, cleanupInterval, make(map[string]Item))
}

// NewFrom creates a new cache like New() and populates the cache with the given
// items.
//
// The passed map will serve as the underlying map for the cache. This is useful
// for starting from a deserialized cache (serialized using e.g. gob.Encode() on
// c.Items()), or passing in e.g. make(map[string]Item, 500) to improve startup
// performance when the cache is expected to reach a certain minimum size.
//
// The map is not copied and only the cache's methods synchronize access to this
// map, so it is not recommended to keep any references to the map around after
// creating a cache. If need be, the map can be accessed at a later point using
// c.Items() (subject to the same caveat.)
//
// Note regarding serialization: When using e.g. gob, make sure to
// gob.Register() the individual types stored in the cache before encoding a map
// retrieved with c.Items(), and to register those same types before decoding a
// blob containing an items map.
func NewFrom(defaultExpiration, cleanupInterval time.Duration, items map[string]Item) *Cache {
	return newCacheWithJanitor(defaultExpiration, cleanupInterval, items)
}
