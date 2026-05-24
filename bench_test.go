package zcache

import (
	"runtime"
	"strconv"
	"sync"
	"testing"
	"time"
)

func benchmarkGet(b *testing.B, exp time.Duration) {
	b.StopTimer()
	c := New[string, any](exp, 0)
	c.Set("foo", "bar")
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		c.Get("foo")
	}
}

func benchmarkGetConcurrent(b *testing.B, exp time.Duration) {
	b.StopTimer()
	c := New[string, any](exp, 0)
	c.Set("foo", "bar")
	wg := new(sync.WaitGroup)
	workers := runtime.NumCPU()
	each := b.N / workers
	wg.Add(workers)
	b.StartTimer()
	for range workers {
		go func() {
			for range each {
				c.Get("foo")
			}
			wg.Done()
		}()
	}
	wg.Wait()
}

func benchmarkSet(b *testing.B, exp time.Duration) {
	b.StopTimer()
	c := New[string, any](exp, 0)
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		c.Set("foo", "bar")
	}
}

func BenchmarkGetExpiring(b *testing.B)              { benchmarkGet(b, 5*time.Minute) }
func BenchmarkGetNotExpiring(b *testing.B)           { benchmarkGet(b, NoExpiration) }
func BenchmarkGetConcurrentExpiring(b *testing.B)    { benchmarkGetConcurrent(b, 5*time.Minute) }
func BenchmarkGetConcurrentNotExpiring(b *testing.B) { benchmarkGetConcurrent(b, NoExpiration) }
func BenchmarkSetExpiring(b *testing.B)              { benchmarkSet(b, 5*time.Minute) }
func BenchmarkSetNotExpiring(b *testing.B)           { benchmarkSet(b, NoExpiration) }

func BenchmarkRWMutexMapGet(b *testing.B) {
	b.StopTimer()
	m := map[string]string{
		"foo": "bar",
	}
	mu := sync.RWMutex{}
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		mu.RLock()
		_ = m["foo"]
		mu.RUnlock()
	}
}

func BenchmarkRWMutexInterfaceMapGetStruct(b *testing.B) {
	b.StopTimer()
	s := struct{ name string }{name: "foo"}
	m := map[any]string{
		s: "bar",
	}
	mu := sync.RWMutex{}
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		mu.RLock()
		_ = m[s]
		mu.RUnlock()
	}
}

func BenchmarkRWMutexInterfaceMapGetString(b *testing.B) {
	b.StopTimer()
	m := map[any]string{
		"foo": "bar",
	}
	mu := sync.RWMutex{}
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		mu.RLock()
		_ = m["foo"]
		mu.RUnlock()
	}
}

func BenchmarkRWMutexMapGetConcurrent(b *testing.B) {
	b.StopTimer()
	m := map[string]string{
		"foo": "bar",
	}
	mu := sync.RWMutex{}
	wg := new(sync.WaitGroup)
	workers := runtime.NumCPU()
	each := b.N / workers
	wg.Add(workers)
	b.StartTimer()
	for range workers {
		go func() {
			for range each {
				mu.RLock()
				_ = m["foo"]
				mu.RUnlock()
			}
			wg.Done()
		}()
	}
	wg.Wait()
}

func BenchmarkRWMutexMapSet(b *testing.B) {
	b.StopTimer()
	m := map[string]string{}
	mu := sync.RWMutex{}
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		mu.Lock()
		m["foo"] = "bar"
		mu.Unlock()
	}
}

func BenchmarkCacheSetDelete(b *testing.B) {
	b.StopTimer()
	c := New[string, any](DefaultExpiration, 0)
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		c.Set("foo", "bar")
		c.Delete("foo")
	}
}

func BenchmarkRWMutexMapSetDelete(b *testing.B) {
	b.StopTimer()
	m := map[string]string{}
	mu := sync.RWMutex{}
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		mu.Lock()
		m["foo"] = "bar"
		mu.Unlock()
		mu.Lock()
		delete(m, "foo")
		mu.Unlock()
	}
}

func BenchmarkCacheSetDeleteSingleLock(b *testing.B) {
	b.StopTimer()
	c := New[string, any](DefaultExpiration, 0)
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		c.mu.Lock()
		c.set("foo", "bar", DefaultExpiration)
		c.delete("foo")
		c.mu.Unlock()
	}
}

func BenchmarkRWMutexMapSetDeleteSingleLock(b *testing.B) {
	b.StopTimer()
	m := map[string]string{}
	mu := sync.RWMutex{}
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		mu.Lock()
		m["foo"] = "bar"
		delete(m, "foo")
		mu.Unlock()
	}
}

func BenchmarkDeleteExpiredLoop(b *testing.B) {
	b.StopTimer()
	c := New[string, any](5*time.Minute, 0)
	c.mu.Lock()
	for i := range 100000 {
		c.set(strconv.Itoa(i), "bar", DefaultExpiration)
	}
	c.mu.Unlock()
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		c.DeleteExpired()
	}
}
