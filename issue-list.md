All PRs and issues from go-cache (excluding some fluff) and why they have or
haven't been included.

Included
--------

This was included, although perhaps not exactly as mentioned, but the use case
or issue that was reported should be resolved.


- [add Expire ExpireAt command by zebozhuang](https://github.com/patrickmn/go-cache/pull/20)<br>
  [Update item with same expiry by pranjal5215](https://github.com/patrickmn/go-cache/pull/42)<br>
  [Add UpdateExpiration method. by dossy](https://github.com/patrickmn/go-cache/pull/66)<br>
  [add GetWithExpirationUpdate by sbabiv](https://github.com/patrickmn/go-cache/pull/96)<br>
  [GetWithExpirationUpdate - atomic implementation by paddlesteamer](https://github.com/patrickmn/go-cache/pull/126)<br>
  [best way to extend time on an not expired object?](https://github.com/patrickmn/go-cache/issues/65)<br>
  [Adding some utility functions to cache: by EVODelavega](https://github.com/patrickmn/go-cache/pull/55)

  Added as `Touch()`; PR 55 also added `Pop()`, which seems like a useful thing
  to add as well.

- [Add GetPossiblyExpired() by Freeaqingme](https://github.com/patrickmn/go-cache/pull/47)<br>
  [Add method to retrieve CacheItem rather than the value of the item by rahulraheja](https://github.com/patrickmn/go-cache/pull/53)<br>
  [Add functionality allowing to get expired value (stale) from cache by oskarwojciski](https://github.com/patrickmn/go-cache/pull/63)<br>
  [Get Expired Item](https://github.com/patrickmn/go-cache/issues/107)

  Get expired cache items; added as `GetStale()`.

  I didn't use `GetItem()` or `GetCacheItem()` as I felt that it should be clear
  from the name you're getting potentially expired items.

  Potentually, a `GetStaleWithExpiration()` could be added too; but I'm not sure
  how valuable that is.

- [Add Iterate by youjianglong](https://github.com/patrickmn/go-cache/pull/78)<br>
  [Add method for getting all cache keys by alex-ant](https://github.com/patrickmn/go-cache/pull/81)<br>
  [Add method for delete by regex rule by vmpartner](https://github.com/patrickmn/go-cache/pull/100)

  All of them are essentially the same issue: do something with all keys. Added
  a `Keys()` method to return an (unsorted) list of keys.

- [Add Map function (Read/Replace) in single lock](https://github.com/patrickmn/go-cache/issues/118)<br>
  [added atomic list-append operation by sgeisbacher](https://github.com/patrickmn/go-cache/pull/97)

  Both of these issues are essentially the same: the ability to atomically
  modify existing values. Instead of adding a []string-specific implementation a
  generic Modify() seems better to me, so add that.

- [Add remove method, if key exists, delete and return elements by yinbaoqiang](https://github.com/patrickmn/go-cache/pull/77)<br>

  Added as Pop()


Not included
------------

Issues and PRs that were *not* included with a short explanation why. You can
open an issue if you feel I made a mistake and we can look at it again :-)

- [Add OnMissing callback by adregner](https://github.com/patrickmn/go-cache/pull/106)<br>
  [Add Memoize by clitetailor](https://github.com/patrickmn/go-cache/pull/113)<br>
  [GetOrSet method to handle case for atomic get and set if not exists by technicianted](https://github.com/patrickmn/go-cache/pull/117)

  These all address the same problem: populate data on a cache Get() miss.

  The problem with a `GetOrSet(set func())`-type method is that the map will be
  locked while the `set` callback is running. This could be fixed by unlocking
  the map, but then it's no longer atomic and you need to be very careful to not
  spawn several `GetOrSet()`s (basically, it doesn't necessarily make things
  more convenient). Since a cache is useful for getting expensive-to-get data
  this seems like it could be a realistic problem.

  This is also the problem with an `OnMiss()` callback: you run the risk of
  spawning a bucketload of OnMiss() callbacks. I also don't especially care much
  for the UX of such a callback, since it's kind of a "action at a distance"
  thing.
  
  This could be solved with [`zsync.Once`]([zstd/once.go at master](https://github.com/zgoat/zstd/blob/master/zsync/once.go#L6)) though,
  then only subsequent GetOrSet calls will block. The downside is that is that
  keys may still be modified with Set() and other functions while this is
  running. I'm not sure if that's a big enough of an issue.

  I'm not entirely sure what the value of a simple `GetOrSet(k string,
  valueIfNotSet interface{})` is. If you already have the value, then why do you
  need this? You can just set it (or indeed, if you already have the value then
  why do you need a cache at all?)

  For now, I decided to not add it.

- [what if onEvicted func  very slow](https://github.com/patrickmn/go-cache/issues/49)<br>

  You can start your own goroutine if you want; this can be potential
  tricky/dangerous as you really don't want to spawn thousands of goroutines at
  the same time, and it may be surprising for some.

- [Feature request: add multiple get method.](https://github.com/patrickmn/go-cache/issues/108)<br>

  The performance difference is not that large compared to a for loop (about
  970ns/op vs 1450 ns/op for 50 items, and it adds an alloc), it's not clear how
  to make a consistent API for this (how do you return found? what if there are
  duplicate keys?), and overall I don't really think it's worth it.

- [Feature request: max size and/or max objects](https://github.com/patrickmn/go-cache/issues/5)<br>
  [An Unobtrusive LRU for the best time cache I've used for go by cognusion](https://github.com/patrickmn/go-cache/pull/17)

  See FAQ; maybe we can add this as a wrapper and new `zcache.LRUCache` or some
  such. Max size is even harder, since getting the size of an object is
  non-trivial.

- [Added BST for efficient deletion by beppeben](https://github.com/patrickmn/go-cache/pull/27)<br>

  Seems to solve a specific use case, but makes stuff quite a bit more complex
  and the performance regresses for some use cases.

- [expose a flag to indicate if it was expired or removed in OnEvicted()](https://github.com/patrickmn/go-cache/issues/57)<br>
  [add isExpired bool to OnEvicted callback signature by Ashtonian](https://github.com/patrickmn/go-cache/pull/58)

  Unclear use case; although passing the Item instead of value to OnEvicted()
  wouldn't be a bad idea (but incompatible).

- [Add function which increase int64 or set in cache if not exists yet by oskarwojciski](https://github.com/patrickmn/go-cache/pull/62)<br>

  This makes the entire increment/decrement stuff even worse; need to rethink
  that entire API. An option to set it if it doesn't exist would be better.

- [Changing RWMutexMap to sync.Map by vidmed](https://github.com/patrickmn/go-cache/pull/72)<br>

  Unclear if this is a good idea, because performance may either increase or
  regress. Won't include.

- [Delete from the cache on Get if the item expired (to trigger onEvicted) by fdurand](https://github.com/patrickmn/go-cache/pull/75/files)<br>

  Not a good idea IMO, makes Get() performance unpredictable, and can be solved
  by just running the janitor more often. Would also complicate the "get even if
  expired" functionality.

- [Add a Noop cache implementation by sylr](https://github.com/patrickmn/go-cache/pull/92)<br>
  [Request: Add formal interface for go-cache](https://github.com/patrickmn/go-cache/issues/116)

  You don't really need this; you can define your own interfaces already.
  Mocking out a in-memory cache with a "fake" implementation also seems like a
  weird thing to do. Worst part is: this will lock down the API. Can't add new
  functions without breaking it.
  Not adding it.

- [Add prometheus metrics by sylr](https://github.com/patrickmn/go-cache/pull/94)<br>

  This PR makes things worse for everyone who doesn't use Prometheus (i.e. most
  people). Clearly this is not a good idea. You can still add it as a wrapper if
  you want.

- [Flush calls onEvicted by pavelbazika](https://github.com/patrickmn/go-cache/pull/122)<br>

  This is a breaking change, since Flush() now works different. You can also
  already do this by getting all the items and deleting one-by-one (or getting
  all the items, Flush(), and calling onEvict()).

- [The OnEvicted function is not called if a value is re-set after expiration but before deletion](https://github.com/patrickmn/go-cache/issues/48)<br>

  I'm not so sure this is actually a bug, as you're overwriting values.

- [Allow querying of expiration, cleanup durations](https://github.com/patrickmn/go-cache/issues/104)<br>

  You can already get a list of items with Items(); so not sure what the use
  case is here? Not clear enough to do anything with as it stands.

- [Implemented a faster version of the Size() function](https://github.com/patrickmn/go-cache/pull/129)<br>

  Because this only counts primitives (and not maps, structs, slices) it's very
  limited. This kind of stuff is out-of-scope for v1 anyway.
