package cache

import "errors"

// Increment an item of type float32 or float64 by n. Returns an error if the
// item's value is not floating point, if it was not found, or if it is not
// possible to increment it by n.
// To retrieve the incremented value, use one of the specialized methods,
// e.g. IncrementFloat64.
func (c *cache) IncrementFloat(k string, n float64) error {
	c.mu.Lock()
	v, found := c.items[k]
	if !found || v.Expired() {
		c.mu.Unlock()
		return errors.New("zcache.IncrementFloat: item " + k + " not found")
	}
	switch v.Object.(type) {
	case float32:
		v.Object = v.Object.(float32) + float32(n)
	case float64:
		v.Object = v.Object.(float64) + n
	default:
		c.mu.Unlock()
		return errors.New("zcache.IncrementFloat: the value for " + k + " does not have type float32 or float64")
	}
	c.items[k] = v
	c.mu.Unlock()
	return nil
}

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

// Increment an item of type int8 by n. Returns an error if the item's value is
// not an int8, or if it was not found. If there is no error, the new value is returned.
func (c *cache) IncrementInt8(k string, n int8) (int8, error) {
	c.mu.Lock()
	v, found := c.items[k]
	if !found || v.Expired() {
		c.mu.Unlock()
		return 0, errors.New("zcache.Increment: item" + k + " not found")
	}
	rv, ok := v.Object.(int8)
	if !ok {
		c.mu.Unlock()
		return 0, errors.New("the value for " + k + " is not an int8")
	}
	nv := rv + n
	v.Object = nv
	c.items[k] = v
	c.mu.Unlock()
	return nv, nil
}

// Increment an item of type int16 by n. Returns an error if the item's value is
// not an int16, or if it was not found. If there is no error, the new value is returned.
func (c *cache) IncrementInt16(k string, n int16) (int16, error) {
	c.mu.Lock()
	v, found := c.items[k]
	if !found || v.Expired() {
		c.mu.Unlock()
		return 0, errors.New("zcache.Increment: item" + k + " not found")
	}
	rv, ok := v.Object.(int16)
	if !ok {
		c.mu.Unlock()
		return 0, errors.New("the value for " + k + " is not an int16")
	}
	nv := rv + n
	v.Object = nv
	c.items[k] = v
	c.mu.Unlock()
	return nv, nil
}

// Increment an item of type int32 by n. Returns an error if the item's value is
// not an int32, or if it was not found. If there is no error, the new value is returned.
func (c *cache) IncrementInt32(k string, n int32) (int32, error) {
	c.mu.Lock()
	v, found := c.items[k]
	if !found || v.Expired() {
		c.mu.Unlock()
		return 0, errors.New("zcache.Increment: item" + k + " not found")
	}
	rv, ok := v.Object.(int32)
	if !ok {
		c.mu.Unlock()
		return 0, errors.New("the value for " + k + " is not an int32")
	}
	nv := rv + n
	v.Object = nv
	c.items[k] = v
	c.mu.Unlock()
	return nv, nil
}

// Increment an item of type int64 by n. Returns an error if the item's value is
// not an int64, or if it was not found. If there is no error, the new value is returned.
func (c *cache) IncrementInt64(k string, n int64) (int64, error) {
	c.mu.Lock()
	v, found := c.items[k]
	if !found || v.Expired() {
		c.mu.Unlock()
		return 0, errors.New("zcache.Increment: item" + k + " not found")
	}
	rv, ok := v.Object.(int64)
	if !ok {
		c.mu.Unlock()
		return 0, errors.New("the value for " + k + " is not an int64")
	}
	nv := rv + n
	v.Object = nv
	c.items[k] = v
	c.mu.Unlock()
	return nv, nil
}

// Increment an item of type uint by n. Returns an error if the item's value is
// not an uint, or if it was not found. If there is no error, the new value is returned.
func (c *cache) IncrementUint(k string, n uint) (uint, error) {
	c.mu.Lock()
	v, found := c.items[k]
	if !found || v.Expired() {
		c.mu.Unlock()
		return 0, errors.New("zcache.Increment: item" + k + " not found")
	}
	rv, ok := v.Object.(uint)
	if !ok {
		c.mu.Unlock()
		return 0, errors.New("the value for " + k + " is not an uint")
	}
	nv := rv + n
	v.Object = nv
	c.items[k] = v
	c.mu.Unlock()
	return nv, nil
}

// Increment an item of type uintptr by n. Returns an error if the item's value is
// not an uintptr, or if it was not found. If there is no error, the new value is returned.
func (c *cache) IncrementUintptr(k string, n uintptr) (uintptr, error) {
	c.mu.Lock()
	v, found := c.items[k]
	if !found || v.Expired() {
		c.mu.Unlock()
		return 0, errors.New("zcache.Increment: item" + k + " not found")
	}
	rv, ok := v.Object.(uintptr)
	if !ok {
		c.mu.Unlock()
		return 0, errors.New("the value for " + k + " is not an uintptr")
	}
	nv := rv + n
	v.Object = nv
	c.items[k] = v
	c.mu.Unlock()
	return nv, nil
}

// Increment an item of type uint8 by n. Returns an error if the item's value is
// not an uint8, or if it was not found. If there is no error, the new value is returned.
func (c *cache) IncrementUint8(k string, n uint8) (uint8, error) {
	c.mu.Lock()
	v, found := c.items[k]
	if !found || v.Expired() {
		c.mu.Unlock()
		return 0, errors.New("zcache.Increment: item" + k + " not found")
	}
	rv, ok := v.Object.(uint8)
	if !ok {
		c.mu.Unlock()
		return 0, errors.New("the value for " + k + " is not an uint8")
	}
	nv := rv + n
	v.Object = nv
	c.items[k] = v
	c.mu.Unlock()
	return nv, nil
}

// Increment an item of type uint16 by n. Returns an error if the item's value is
// not an uint16, or if it was not found. If there is no error, the new value is returned.
func (c *cache) IncrementUint16(k string, n uint16) (uint16, error) {
	c.mu.Lock()
	v, found := c.items[k]
	if !found || v.Expired() {
		c.mu.Unlock()
		return 0, errors.New("zcache.Increment: item" + k + " not found")
	}
	rv, ok := v.Object.(uint16)
	if !ok {
		c.mu.Unlock()
		return 0, errors.New("the value for " + k + " is not an uint16")
	}
	nv := rv + n
	v.Object = nv
	c.items[k] = v
	c.mu.Unlock()
	return nv, nil
}

// Increment an item of type uint32 by n. Returns an error if the item's value is
// not an uint32, or if it was not found. If there is no error, the new value is returned.
func (c *cache) IncrementUint32(k string, n uint32) (uint32, error) {
	c.mu.Lock()
	v, found := c.items[k]
	if !found || v.Expired() {
		c.mu.Unlock()
		return 0, errors.New("zcache.Increment: item" + k + " not found")
	}
	rv, ok := v.Object.(uint32)
	if !ok {
		c.mu.Unlock()
		return 0, errors.New("the value for " + k + " is not an uint32")
	}
	nv := rv + n
	v.Object = nv
	c.items[k] = v
	c.mu.Unlock()
	return nv, nil
}

// Increment an item of type uint64 by n. Returns an error if the item's value is
// not an uint64, or if it was not found. If there is no error, the new value is returned.
func (c *cache) IncrementUint64(k string, n uint64) (uint64, error) {
	c.mu.Lock()
	v, found := c.items[k]
	if !found || v.Expired() {
		c.mu.Unlock()
		return 0, errors.New("zcache.Increment: item" + k + " not found")
	}
	rv, ok := v.Object.(uint64)
	if !ok {
		c.mu.Unlock()
		return 0, errors.New("the value for " + k + " is not an uint64")
	}
	nv := rv + n
	v.Object = nv
	c.items[k] = v
	c.mu.Unlock()
	return nv, nil
}

// Increment an item of type float32 by n. Returns an error if the item's value is
// not an float32, or if it was not found. If there is no error, the new value is returned.
func (c *cache) IncrementFloat32(k string, n float32) (float32, error) {
	c.mu.Lock()
	v, found := c.items[k]
	if !found || v.Expired() {
		c.mu.Unlock()
		return 0, errors.New("zcache.Increment: item" + k + " not found")
	}
	rv, ok := v.Object.(float32)
	if !ok {
		c.mu.Unlock()
		return 0, errors.New("the value for " + k + " is not an float32")
	}
	nv := rv + n
	v.Object = nv
	c.items[k] = v
	c.mu.Unlock()
	return nv, nil
}

// Increment an item of type float64 by n. Returns an error if the item's value is
// not an float64, or if it was not found. If there is no error, the new value is returned.
func (c *cache) IncrementFloat64(k string, n float64) (float64, error) {
	c.mu.Lock()
	v, found := c.items[k]
	if !found || v.Expired() {
		c.mu.Unlock()
		return 0, errors.New("zcache.Increment: item" + k + " not found")
	}
	rv, ok := v.Object.(float64)
	if !ok {
		c.mu.Unlock()
		return 0, errors.New("the value for " + k + " is not an float64")
	}
	nv := rv + n
	v.Object = nv
	c.items[k] = v
	c.mu.Unlock()
	return nv, nil
}

// Decrement an item of type float32 or float64 by n. Returns an error if the
// item's value is not floating point, if it was not found, or if it is not
// possible to decrement it by n.
// To retrieve the decremented value, use one of the specialized methods,
// e.g. DecrementFloat64.
func (c *cache) DecrementFloat(k string, n float64) error {
	c.mu.Lock()
	v, found := c.items[k]
	if !found || v.Expired() {
		c.mu.Unlock()
		return errors.New("zcache.DecrementFloat: item " + k + " not found")
	}
	switch v.Object.(type) {
	case float32:
		v.Object = v.Object.(float32) - float32(n)
	case float64:
		v.Object = v.Object.(float64) - n
	default:
		c.mu.Unlock()
		return errors.New("zcache.DecrementFloat: the value for " + k + " does not have type float32 or float64")
	}
	c.items[k] = v
	c.mu.Unlock()
	return nil
}

// Decrement an item of type int by n. Returns an error if the item's value is
// not an int, or if it was not found. If there is no error, the new value is returned.
func (c *cache) DecrementInt(k string, n int) (int, error) {
	c.mu.Lock()
	v, found := c.items[k]
	if !found || v.Expired() {
		c.mu.Unlock()
		return 0, errors.New("zcache.Decrement: item" + k + " not found")
	}
	rv, ok := v.Object.(int)
	if !ok {
		c.mu.Unlock()
		return 0, errors.New("the value for " + k + " is not an int")
	}
	nv := rv - n
	v.Object = nv
	c.items[k] = v
	c.mu.Unlock()
	return nv, nil
}

// Decrement an item of type int8 by n. Returns an error if the item's value is
// not an int8, or if it was not found. If there is no error, the new value is returned.
func (c *cache) DecrementInt8(k string, n int8) (int8, error) {
	c.mu.Lock()
	v, found := c.items[k]
	if !found || v.Expired() {
		c.mu.Unlock()
		return 0, errors.New("zcache.Decrement: item" + k + " not found")
	}
	rv, ok := v.Object.(int8)
	if !ok {
		c.mu.Unlock()
		return 0, errors.New("the value for " + k + " is not an int8")
	}
	nv := rv - n
	v.Object = nv
	c.items[k] = v
	c.mu.Unlock()
	return nv, nil
}

// Decrement an item of type int16 by n. Returns an error if the item's value is
// not an int16, or if it was not found. If there is no error, the new value is returned.
func (c *cache) DecrementInt16(k string, n int16) (int16, error) {
	c.mu.Lock()
	v, found := c.items[k]
	if !found || v.Expired() {
		c.mu.Unlock()
		return 0, errors.New("zcache.Decrement: item" + k + " not found")
	}
	rv, ok := v.Object.(int16)
	if !ok {
		c.mu.Unlock()
		return 0, errors.New("the value for " + k + " is not an int16")
	}
	nv := rv - n
	v.Object = nv
	c.items[k] = v
	c.mu.Unlock()
	return nv, nil
}

// Decrement an item of type int32 by n. Returns an error if the item's value is
// not an int32, or if it was not found. If there is no error, the new value is returned.
func (c *cache) DecrementInt32(k string, n int32) (int32, error) {
	c.mu.Lock()
	v, found := c.items[k]
	if !found || v.Expired() {
		c.mu.Unlock()
		return 0, errors.New("zcache.Decrement: item" + k + " not found")
	}
	rv, ok := v.Object.(int32)
	if !ok {
		c.mu.Unlock()
		return 0, errors.New("the value for " + k + " is not an int32")
	}
	nv := rv - n
	v.Object = nv
	c.items[k] = v
	c.mu.Unlock()
	return nv, nil
}

// Decrement an item of type int64 by n. Returns an error if the item's value is
// not an int64, or if it was not found. If there is no error, the new value is returned.
func (c *cache) DecrementInt64(k string, n int64) (int64, error) {
	c.mu.Lock()
	v, found := c.items[k]
	if !found || v.Expired() {
		c.mu.Unlock()
		return 0, errors.New("zcache.Decrement: item" + k + " not found")
	}
	rv, ok := v.Object.(int64)
	if !ok {
		c.mu.Unlock()
		return 0, errors.New("the value for " + k + " is not an int64")
	}
	nv := rv - n
	v.Object = nv
	c.items[k] = v
	c.mu.Unlock()
	return nv, nil
}

// Decrement an item of type uint by n. Returns an error if the item's value is
// not an uint, or if it was not found. If there is no error, the new value is returned.
func (c *cache) DecrementUint(k string, n uint) (uint, error) {
	c.mu.Lock()
	v, found := c.items[k]
	if !found || v.Expired() {
		c.mu.Unlock()
		return 0, errors.New("zcache.Decrement: item" + k + " not found")
	}
	rv, ok := v.Object.(uint)
	if !ok {
		c.mu.Unlock()
		return 0, errors.New("the value for " + k + " is not an uint")
	}
	nv := rv - n
	v.Object = nv
	c.items[k] = v
	c.mu.Unlock()
	return nv, nil
}

// Decrement an item of type uintptr by n. Returns an error if the item's value is
// not an uintptr, or if it was not found. If there is no error, the new value is returned.
func (c *cache) DecrementUintptr(k string, n uintptr) (uintptr, error) {
	c.mu.Lock()
	v, found := c.items[k]
	if !found || v.Expired() {
		c.mu.Unlock()
		return 0, errors.New("zcache.Decrement: item" + k + " not found")
	}
	rv, ok := v.Object.(uintptr)
	if !ok {
		c.mu.Unlock()
		return 0, errors.New("the value for " + k + " is not an uintptr")
	}
	nv := rv - n
	v.Object = nv
	c.items[k] = v
	c.mu.Unlock()
	return nv, nil
}

// Decrement an item of type uint8 by n. Returns an error if the item's value is
// not an uint8, or if it was not found. If there is no error, the new value is returned.
func (c *cache) DecrementUint8(k string, n uint8) (uint8, error) {
	c.mu.Lock()
	v, found := c.items[k]
	if !found || v.Expired() {
		c.mu.Unlock()
		return 0, errors.New("zcache.Decrement: item" + k + " not found")
	}
	rv, ok := v.Object.(uint8)
	if !ok {
		c.mu.Unlock()
		return 0, errors.New("the value for " + k + " is not an uint8")
	}
	nv := rv - n
	v.Object = nv
	c.items[k] = v
	c.mu.Unlock()
	return nv, nil
}

// Decrement an item of type uint16 by n. Returns an error if the item's value is
// not an uint16, or if it was not found. If there is no error, the new value is returned.
func (c *cache) DecrementUint16(k string, n uint16) (uint16, error) {
	c.mu.Lock()
	v, found := c.items[k]
	if !found || v.Expired() {
		c.mu.Unlock()
		return 0, errors.New("zcache.Decrement: item" + k + " not found")
	}
	rv, ok := v.Object.(uint16)
	if !ok {
		c.mu.Unlock()
		return 0, errors.New("the value for " + k + " is not an uint16")
	}
	nv := rv - n
	v.Object = nv
	c.items[k] = v
	c.mu.Unlock()
	return nv, nil
}

// Decrement an item of type uint32 by n. Returns an error if the item's value is
// not an uint32, or if it was not found. If there is no error, the new value is returned.
func (c *cache) DecrementUint32(k string, n uint32) (uint32, error) {
	c.mu.Lock()
	v, found := c.items[k]
	if !found || v.Expired() {
		c.mu.Unlock()
		return 0, errors.New("zcache.Decrement: item" + k + " not found")
	}
	rv, ok := v.Object.(uint32)
	if !ok {
		c.mu.Unlock()
		return 0, errors.New("the value for " + k + " is not an uint32")
	}
	nv := rv - n
	v.Object = nv
	c.items[k] = v
	c.mu.Unlock()
	return nv, nil
}

// Decrement an item of type uint64 by n. Returns an error if the item's value is
// not an uint64, or if it was not found. If there is no error, the new value is returned.
func (c *cache) DecrementUint64(k string, n uint64) (uint64, error) {
	c.mu.Lock()
	v, found := c.items[k]
	if !found || v.Expired() {
		c.mu.Unlock()
		return 0, errors.New("zcache.Decrement: item" + k + " not found")
	}
	rv, ok := v.Object.(uint64)
	if !ok {
		c.mu.Unlock()
		return 0, errors.New("the value for " + k + " is not an uint64")
	}
	nv := rv - n
	v.Object = nv
	c.items[k] = v
	c.mu.Unlock()
	return nv, nil
}

// Decrement an item of type float32 by n. Returns an error if the item's value is
// not an float32, or if it was not found. If there is no error, the new value is returned.
func (c *cache) DecrementFloat32(k string, n float32) (float32, error) {
	c.mu.Lock()
	v, found := c.items[k]
	if !found || v.Expired() {
		c.mu.Unlock()
		return 0, errors.New("zcache.Decrement: item" + k + " not found")
	}
	rv, ok := v.Object.(float32)
	if !ok {
		c.mu.Unlock()
		return 0, errors.New("the value for " + k + " is not an float32")
	}
	nv := rv - n
	v.Object = nv
	c.items[k] = v
	c.mu.Unlock()
	return nv, nil
}

// Decrement an item of type float64 by n. Returns an error if the item's value is
// not an float64, or if it was not found. If there is no error, the new value is returned.
func (c *cache) DecrementFloat64(k string, n float64) (float64, error) {
	c.mu.Lock()
	v, found := c.items[k]
	if !found || v.Expired() {
		c.mu.Unlock()
		return 0, errors.New("zcache.Decrement: item" + k + " not found")
	}
	rv, ok := v.Object.(float64)
	if !ok {
		c.mu.Unlock()
		return 0, errors.New("the value for " + k + " is not an float64")
	}
	nv := rv - n
	v.Object = nv
	c.items[k] = v
	c.mu.Unlock()
	return nv, nil
}
