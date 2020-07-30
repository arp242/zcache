All PRs and issues from go-cache (excluding some fluff) and why they have or
haven't been included:

    INCLUDED   This was included, although perhaps not exactly as mentioned.
    DECLINED   Not adding this for now.
    TODO       TODO :-)


- INCLUCDED
  https://github.com/patrickmn/go-cache/pull/20
  https://github.com/patrickmn/go-cache/pull/66
  https://github.com/patrickmn/go-cache/pull/96
  https://github.com/patrickmn/go-cache/pull/126
  https://github.com/patrickmn/go-cache/issues/65

  Added as `Touch()`

- INCLUDED
  https://github.com/patrickmn/go-cache/pull/78
  https://github.com/patrickmn/go-cache/pull/81
  https://github.com/patrickmn/go-cache/pull/100

  All of them are essentially the same issue: do something with all keys. Added
  a `Keys()` method to return an (unsorted) list of keys.

- TODO

  TODO




- TODO
  https://github.com/patrickmn/go-cache/issues/118

  Add as Modify()?

  cache.Modify("key", func(x interface{}) interface{} {
        x = x.(map[string]string)
        x["foo"] = "qwe"
        return x
       // Do stuff while it's locked.
  })

- DECLINED
  https://github.com/patrickmn/go-cache/issues/49

  You can start your own goroutine if you want; this can be potential
  tricky/dangerous as you really don't want to spawn thousands of goroutines at
  the same time, and it may be surprising for some.

- DECLINED
  https://github.com/patrickmn/go-cache/issues/108

  The performance difference is not that large compared to a for loop (about
  970ns/op vs 1450 ns/op for 50 items, and it adds an alloc), it's not clear how
  to make a consistent API for this (how do you return found? what if there are
  duplicate keys?), and overall I don't really think it's worth it.

- DECLINED
  https://github.com/patrickmn/go-cache/issues/5
  https://github.com/patrickmn/go-cache/pull/17

  See FAQ; maybe we can add this as a wrapper and new `zcache.LRUCache` or some
  such. Max size is even harder, since getting the size of an object is
  non-trivial.

- TODO
  https://github.com/patrickmn/go-cache/issues/104

- DECLINED
  https://github.com/patrickmn/go-cache/pull/27

  Seems to solve a specific use case, but makes stuff quite a bit more complex
  and the performance regresses for some use cases.

- TODO
  https://github.com/patrickmn/go-cache/pull/42

  TODO: could possible include this.

- TODO
  https://github.com/patrickmn/go-cache/pull/47
  https://github.com/patrickmn/go-cache/pull/53
  https://github.com/patrickmn/go-cache/pull/63
  https://github.com/patrickmn/go-cache/issues/107

  Get expired cache items; could be useful, but not entirely sold on the
  API/name of either.
  TODO: look at this.

- TODO
  https://github.com/patrickmn/go-cache/pull/55

  Some of this looks useful.

  One way to do this would be to add a list of options in an (incompatible) API,
  would also solve integrate some of the other this:

      c.Get("key", zcache.Pop, z.cacheIncludeExpired) // Or as bitmask?

- DECLINED
  https://github.com/patrickmn/go-cache/issues/57
  https://github.com/patrickmn/go-cache/pull/58

  Unclear use case; although passing the Item instead of value to OnEvicted()
  wouldn't be a bad idea (but incompatible).

- DECLINED
  https://github.com/patrickmn/go-cache/pull/62

  This makes the entire increment/decrement stuff even worse; need to rethink
  that entire API. An option to set it if it doesn't exist would be better.

- DECLINED
  https://github.com/patrickmn/go-cache/pull/72

  Unclear if this is a good idea, because performance may either increase or
  regress. Won't include.

- DECLINED
  https://github.com/patrickmn/go-cache/pull/75/files

  Not a good idea IMO, makes Get() performance unpredictable, and can be solved
  by just running the janitor more often. Would also complicate the "get even if
  expired" functionality.

- TODO
  https://github.com/patrickmn/go-cache/pull/77

  Remove item and return value; don't like the function name but could include
  this. TODO

- DECLINED
  https://github.com/patrickmn/go-cache/pull/92
  https://github.com/patrickmn/go-cache/issues/116

  You don't really need this; you can define your own interfaces already.
  Mocking out a in-memory cache with a "fake" implementation also seems like a
  weird thing to do. Worst part is: this will lock down the API. Can't add new
  functions without breaking it.
  Not adding it.

- DECLINED
  https://github.com/patrickmn/go-cache/pull/94

  This PR makes things worse for everyone who doesn't use Prometheus (i.e. most
  people). Clearly this is not a good idea. You can still add it as a wrapper if
  you want.

- TODO
  https://github.com/patrickmn/go-cache/pull/97

  This is probably useful; think a bit about the API.

- TODO
  https://github.com/patrickmn/go-cache/pull/106
  https://github.com/patrickmn/go-cache/pull/113
  https://github.com/patrickmn/go-cache/pull/117

  These all address the same problem: populate data on a cache Get(); think
  about the best API.

- DECLINED https://github.com/patrickmn/go-cache/pull/122

  This is a breaking change, since Flush() now works different. You can also
  already do this by getting all the items and deleting one-by-one (or getting
  all the items, Flush(), and calling onEvict()).

- DECLINED
  https://github.com/patrickmn/go-cache/issues/48

  I'm not so sure this is actually a bug, as you're overwriting values.
