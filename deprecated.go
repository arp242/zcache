package zcache

// DeleteFunc deletes and returns cache items matched by the filter function.
//
// The item will be deleted if the callback's first return argument is true. The
// loop will stop if the second return argument is true.
//
// OnEvicted is called for deleted items.
//
// Deprecated: Keyset can be used to operate on multiple values. For example
// mycache.Find(func(...) { ... }).Delete()
func (c *cache[K, V]) DeleteFunc(filter func(key K, item Item[V]) (del, stop bool)) map[K]Item[V] {
	keys := c.Find(filter)
	m := map[K]Item[V]{}
	for i, item := range keys.Get() {
		// Note this isn't the same as before – we don't have access to
		// Expiration. That's probably okay; it was a mistake this was exposed
		// in the first place.
		m[keys.Index(i)] = Item[V]{Object: item.V}
	}
	keys.Delete()
	return m
}
