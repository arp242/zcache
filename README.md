zcache is an in-memory key:value store/cache with time-based evictions.

It is suitable for applications running on a single machine. Its major advantage
is that it's essentially a thread-safe map with expiration times. Any object can
be stored, for a given duration or forever, and the cache can be safely used by
multiple goroutines.

Although zcache isn't meant to be used as a persistent datastore, the contents
can be saved to and loaded from a file (using `c.Items()` to retrieve the items
map to serialize, and `NewFrom()` to create a cache from a deserialized one) to
recover from downtime quickly.

The canonical import path is `zgo.at/zcache`, and reference docs are at
https://godocs.io/zgo.at/zcache

This is a fork of https://github.com/patrickmn/go-cache – which no longer seems
actively maintained. There are two versions of zcache:

- v1 is intended to be 100% compatible with co-cache and a drop-in replacement
  with various enhancements.
- v2 makes various incompatible changes to the API: various functions calls are
  improved. This uses generics and requires Go 1.18.

This README documents v2; see README.v1.md for the v1 README. Both versions are
maintained. See the "changes" section below for a list of changes.

Usage
-----
Some examples from `example_test.go`:

```go
func ExampleSimple() {
	// Create a cache with a default expiration time of 5 minutes, and which
	// purges expired items every 10 minutes.
	//
	// This creates a cache with string keys and values, with Go 1.18 type
	// parameters.
	c := zcache.New[string, string](5*time.Minute, 10*time.Minute)

	// Set the value of the key "foo" to "bar", with the default expiration.
	c.Set("foo", "bar")

	// Set the value of the key "baz" to "never", with no expiration time. The
	// item won't be removed until it's removed with c.Delete("baz").
	c.SetWithExpire("baz", "never", zcache.NoExpiration)

	// Get the value associated with the key "foo" from the cache; due to the
	// use of type parameters this is a string, and no type assertions are
	// needed.
	foo, ok := c.Get("foo")
	if ok {
		fmt.Println(foo)
	}

	// Output: bar
}

func ExampleStruct() {
	type MyStruct struct{ Value string }

	// Create a new cache that stores a specific struct.
	c := zcache.New[string, *MyStruct](zcache.NoExpiration, zcache.NoExpiration)
	c.Set("cache", &MyStruct{Value: "value"})

	v, _ := c.Get("cache")
	fmt.Printf("%#v\n", v)

	// Output: &zcache_test.MyStruct{Value:"value"}
}

func ExampleAny() {
	// Create a new cache that stores any value, behaving similar to zcache v1
	// or go-cache.
	c := zcache.New[string, any](zcache.NoExpiration, zcache.NoExpiration)

	c.Set("a", "value 1")
	c.Set("b", 42)

	a, _ := c.Get("a")
	b, _ := c.Get("b")

	// This needs type assertions.
	p := func(a string, b int) { fmt.Println(a, b) }
	p(a.(string), b.(int))

	// Output: value 1 42
}

func ExampleProxy() {
	type Site struct {
		ID       int
		Hostname string
	}

	site := &Site{
		ID:       42,
		Hostname: "example.com",
	}

	// Create a new site which caches by site ID (int), and a "proxy" which
	// caches by the hostname (string).
	c := zcache.New[int, *Site](zcache.NoExpiration, zcache.NoExpiration)
	p := zcache.NewProxy[string, int, *Site](c)

	p.Set(42, "example.com", site)

	siteByID, ok := c.Get(42)
	fmt.Printf("%v %v\n", ok, siteByID)

	siteByHost, ok := p.Get("example.com")
	fmt.Printf("%v %v\n", ok, siteByHost)

	// They're both the same object/pointer.
	fmt.Printf("%v\n", siteByID == siteByHost)

	// Output:
	// true &{42 example.com}
	// true &{42 example.com}
	// true
}
```

Changes
-------
### Incompatible changes in v2
- Use type parameters instead of `map[string]interface{}`; you can get the same
  as before with `zcache.New[string, any](..)`, but if you know you will only
  store `MyStruct` you can use `zcache.New[string, *MyStruct](..)` for
  additional type safety.

- Remove `Save()`, `SaveFile()`, `Load()`, `LoadFile()`; you can still persist
  stuff to disk by using `Items()` and `NewFrom()`. These methods were already
  deprecated.

- Rename `Set()` to `SetWithExpire()`, and rename `SetDefault()` to `Set()`.
  Most of the time you want to use the default expiry time, so make that the
  easier path.

- The `Increment*` and `Decrement*` functions have been removed; you can replace
  them with `Modify()`:

      cache := New[string, int](DefaultExpiration, 0)
      cache.Set("one", 1)
      cache.Modify("one", func(v int) int { return v + 1 })

  The performance of this is roughly the same as the old Increment, and this is
  a more generic method that can also be used for other things like appending to
  a slice.

- Rename `Flush()` to `Reset()`; I think that more clearly conveys what it's
  intended for as `Flush()` is typically used to flush a buffer or the like.

### Compatible changes from go-cache
All these changes are in both v1 and v2:

- Add `Keys()` to list all keys.
- Add `Touch()` to update the expiry on an item.
- Add `GetStale()` to get items even after they've expired.
- Add `Pop()` to get an item and delete it.
- Add `Modify()` to atomically modify existing cache entries (e.g. lists, maps).
- Add `DeleteAll()` to remove all items from the cache with onEvicted call.
- Add `DeleteFunc()` to remove specific items from the cache atomically.
- Add `Rename()` to rename keys, retaining the value and expiry.
- Add `Proxy` type, to access cache items under a different key.
- Various small internal and documentation improvements.

See [issue-list.markdown](/issue-list.markdown) for a complete run-down of the
PRs/issues for go-cache and what was and wasn't included.

FAQ
---

### How can I limit the size of the cache? Is there an option for this?
Not really; zcache is intended as a thread-safe map with time-based eviction.
This keeps it nice and simple. Adding something like a LRU eviction mechanism
not only makes the code more complex, it also makes the library worse for cases
where you just want a map since it requires additional memory and makes some
operations more expensive (unless a new API is added which make the API worse
for those use cases).

So unless I or someone else comes up with a way to do this which doesn't detract
anything from the simple map use case, I'd rather not add it. Perhaps wrapping
`zcache.Cache` and overriding some methods could work, but I haven't looked at
it.

tl;dr: this isn't designed to solve every caching use case. That's a feature.
