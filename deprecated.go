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
