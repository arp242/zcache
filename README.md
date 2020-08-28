zcache is an in-memory key:value store/cache similar to memcached. It is
suitable for applications running on a single machine. Its major advantage is
that it's essentially a thread-safe `map[string]interface{}` with expiration
times and doesn't need to serialize or transmit its contents over the network.

Any object can be stored, for a given duration or forever, and the cache can be
safely used by multiple goroutines.

Although zcache isn't meant to be used as a persistent datastore, the entire
cache can be saved to and loaded from a file (using `c.Items()` to retrieve the
items map to serialize, and `NewFrom()` to create a cache from a deserialized
one) to recover from downtime quickly. (See the docs for `NewFrom()` for
caveats.)

The canonical import path is `zgo.at/zcache`, and reference docs are at
https://pkg.go.dev/zgo.at/zcache

---

This is a fork of https://github.com/patrickmn/go-cache – which no longer seems
actively maintained. v1 is intended to be 100% compatible and a drop-in
replacement.

See [issue-list.markdown](/issue-list.markdown) for a complete run-down of the
PRs/issues for go-cache and what was and wasn't included; in short:

- Add `GetOrSet()` to set and return a value if a key doesn't exist yet.
- Add `Keys()` to list all keys
- Add `Touch()` to update the expiry on an item.
- Add `GetStale()` to get items even after they've expired.
- Add `Pop()` to get an item and delete it.
- Add `Modify()` to atomically modify existing cache entries (e.g. lists, maps).
- Add `DeleteAll()` to remove all items from the cache with onEvicted call.
- Add `DeleteFunc()` to remove specific items from the cache atomically.
- Various small internal and documentation improvements.

NOTE: there is no "v1" release yet, and the API or semantics may still change
based on feedback; in particular, the `GetOrSet()` method is rather tricky (see
comments in the issue-list.markdown file).


Usage
-----

```go
import (
    "fmt"
    "time"

    "zgo.at/zcache"
)

func main() {
    // Create a cache with a default expiration time of 5 minutes, and which
    // purges expired items every 10 minutes
    c := zcache.New(5*time.Minute, 10*time.Minute)

    // Set the value of the key "foo" to "bar", with the default expiration time
    c.Set("foo", "bar", zcache.DefaultExpiration)

    // Set the value of the key "baz" to 42, with no expiration time
    // (the item won't be removed until it is re-set, or removed using
    // c.Delete("baz")
    c.Set("baz", 42, zcache.NoExpiration)

    // Get the string associated with the key "foo" from the cache
    foo, ok := c.Get("foo")
    if ok {
        fmt.Println(foo)
    }

    // Since Go is statically typed, and cache values can be anything, type
    // assertion is needed when values are being passed to functions that don't
    // take arbitrary types, (i.e. interface{}). The simplest way to do this for
    // values which will only be used once--e.g. for passing to another
    // function--is:
    foo, ok := c.Get("foo")
    if ok {
        MyFunction(foo.(string))
    }

    // This gets tedious if the value is used several times in the same function.
    // You might do either of the following instead:
    if x, ok := c.Get("foo"); ok {
        foo := x.(string)
        // ...
    }
    // or
    var foo string
    if x, ok := c.Get("foo"); ok {
        foo = x.(string)
    }
    // ...
    // foo can then be passed around freely as a string

    // Want performance? Store pointers!
    c.Set("foo", &MyStruct, zcache.DefaultExpiration)
    if x, ok := c.Get("foo"); ok {
        foo := x.(*MyStruct)
        // ...
    }
}
```

FAQ
---

### How can I limit the size of the cache? Is there an option for this?

Not really; zcache is intended as a thread-safe `map[string]interface{}` with
time-based eviction. This keeps it nice and simple. Adding something like a LRU
eviction mechanism not only makes the code more complex, it also makes the
library worse for cases where you just want a `map[string]interface{}` since it
requires additional memory and makes some operations more expensive (unless a
new API is added which make the API worse for those use cases).

So unless I or someone else comes up with a way to do this which doesn't detract
anything from the simple `map[string]interface{}` use case, I'd rather not add
it. Perhaps wrapping `zcache.Cache` and overriding some methods could work, but
I haven't looked at it.

tl;dr: this isn't designed to solve every caching use case. That's a feature.
