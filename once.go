// From: https://github.com/zgoat/zstd/blob/master/zsync/once.go

package zcache

import (
	"sync"
)

// once is an object that will perform exactly one action per key.
//
// This is mix between sync.Once and /x/sync/singleflight; like once, a function
// is only run once, and like singleflight it allows grouping per-key and
// returns if the function is already run.
//
// This implementation is a bit slower than the stdlib one; the benchmark
// regresses ~1.6ns/op to ~52ns/op on my system.
type once struct {
	m         sync.Mutex
	forgotten bool
	done      map[string]struct{}
}

// Do calls the function f for the given key if and only if Do is being called
// for the first time for this key. In other words, given:
//
// 	var once Once
//
// If once.Do("x", f) is called multiple times, only the first call will invoke
// f, even if f has a different value in each invocation. A new key or instance
// of Once is required for each function to execute.
//
// The return value tells you if f is run; it's true on the first caller, and
// false on all subsequent calls.
//
// Since f is niladic, it may be necessary to use a function literal to capture
// the arguments to a function to be invoked by Do:
//
// 	config.once.Do(func() { config.init(filename) })
//
// Because no call to Do returns until the one call to f returns, if f causes Do
// to be called, it will deadlock.
//
// If f panics, Do considers it to have returned; future calls of Do return
// without calling f.
func (o *once) Do(key string, f func()) bool {
	o.m.Lock()
	defer o.m.Unlock()

	if o.done == nil {
		o.done = make(map[string]struct{})
	}
	_, ok := o.done[key]
	if ok {
		return false
	}

	defer func() { o.done[key] = struct{}{} }()
	f()
	return true
}

// Forget about a key.
func (o *once) Forget(key string) {
	o.m.Lock()
	delete(o.done, key)
	o.m.Unlock()
}
