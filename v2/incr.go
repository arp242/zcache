package zcache

import "errors"

// Increment an item of type int by n. Returns an error if the item's value is
// not an int, or if it was not found. If there is no error, the new value is returned.
func (c *cache) IncrementInt(k string, n int) (int, error) {
	c.mu.Lock()
	v, found := c.items[k]
	if !found || v.Expired() {
		c.mu.Unlock()
		return 0, errors.New("zcache.Increment: item" + k + " not found")
	}
	rv, ok := v.Object.(int)
	if !ok {
		c.mu.Unlock()
		return 0, errors.New("the value for " + k + " is not an int")
	}
	nv := rv + n
	v.Object = nv
	c.items[k] = v
	c.mu.Unlock()
	return nv, nil
}
