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
	tc := New(exp, 0)
	tc.Set("foo", "bar", DefaultExpiration)
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		tc.Get("foo")
	}
}

func benchmarkGetConcurrent(b *testing.B, exp time.Duration) {
	b.StopTimer()
	tc := New(exp, 0)
	tc.Set("foo", "bar", DefaultExpiration)
	wg := new(sync.WaitGroup)
	workers := runtime.NumCPU()
	each := b.N / workers
	wg.Add(workers)
	b.StartTimer()
	for i := 0; i < workers; i++ {
		go func() {
			for j := 0; j < each; j++ {
				tc.Get("foo")
			}
			wg.Done()
		}()
	}
	wg.Wait()
}

func benchmarkSet(b *testing.B, exp time.Duration) {
	b.StopTimer()
	tc := New(exp, 0)
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		tc.Set("foo", "bar", DefaultExpiration)
	}
}

func benchmarkGetManyConcurrent(b *testing.B, exp time.Duration) {
	// This is the same as BenchmarkCacheGetConcurrent, but its result
	// can be compared against BenchmarkShardedCacheGetManyConcurrent
	// in sharded_test.go.
	b.StopTimer()
	n := 10000
	tc := New(exp, 0)
	keys := make([]string, n)
	for i := 0; i < n; i++ {
		k := "foo" + strconv.Itoa(i)
		keys[i] = k
		tc.Set(k, "bar", DefaultExpiration)
	}
	each := b.N / n
	wg := new(sync.WaitGroup)
	wg.Add(n)
	for _, v := range keys {
		go func(k string) {
			for j := 0; j < each; j++ {
				tc.Get(k)
			}
			wg.Done()
		}(v)
	}
	b.StartTimer()
	wg.Wait()
}

func BenchmarkGetExpiring(b *testing.B)                  { benchmarkGet(b, 5*time.Minute) }
func BenchmarkGetNotExpiring(b *testing.B)               { benchmarkGet(b, NoExpiration) }
func BenchmarkGetConcurrentExpiring(b *testing.B)        { benchmarkGetConcurrent(b, 5*time.Minute) }
func BenchmarkGetConcurrentNotExpiring(b *testing.B)     { benchmarkGetConcurrent(b, NoExpiration) }
func BenchmarkSetExpiring(b *testing.B)                  { benchmarkSet(b, 5*time.Minute) }
func BenchmarkSetNotExpiring(b *testing.B)               { benchmarkSet(b, NoExpiration) }
func BenchmarkGetManyConcurrentExpiring(b *testing.B)    { benchmarkGetManyConcurrent(b, 5*time.Minute) }
func BenchmarkGetManyConcurrentNotExpiring(b *testing.B) { benchmarkGetManyConcurrent(b, NoExpiration) }

func BenchmarkRWMutexMapGet(b *testing.B) {
	b.StopTimer()
	m := map[string]string{
		"foo": "bar",
	}
	mu := sync.RWMutex{}
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		mu.RLock()
		_, _ = m["foo"]
		mu.RUnlock()
	}
}

func BenchmarkRWMutexInterfaceMapGetStruct(b *testing.B) {
	b.StopTimer()
	s := struct{ name string }{name: "foo"}
	m := map[interface{}]string{
		s: "bar",
	}
	mu := sync.RWMutex{}
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		mu.RLock()
		_, _ = m[s]
		mu.RUnlock()
	}
}

func BenchmarkRWMutexInterfaceMapGetString(b *testing.B) {
	b.StopTimer()
	m := map[interface{}]string{
		"foo": "bar",
	}
	mu := sync.RWMutex{}
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		mu.RLock()
		_, _ = m["foo"]
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
	for i := 0; i < workers; i++ {
		go func() {
			for j := 0; j < each; j++ {
				mu.RLock()
				_, _ = m["foo"]
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
	tc := New(DefaultExpiration, 0)
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		tc.Set("foo", "bar", DefaultExpiration)
		tc.Delete("foo")
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
	tc := New(DefaultExpiration, 0)
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		tc.mu.Lock()
		tc.set("foo", "bar", DefaultExpiration)
		tc.delete("foo")
		tc.mu.Unlock()
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

func BenchmarkIncrementInt(b *testing.B) {
	b.StopTimer()
	tc := New(DefaultExpiration, 0)
	tc.Set("foo", 0, DefaultExpiration)
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		tc.IncrementInt("foo", 1)
	}
}

func BenchmarkDeleteExpiredLoop(b *testing.B) {
	b.StopTimer()
	tc := New(5*time.Minute, 0)
	tc.mu.Lock()
	for i := 0; i < 100000; i++ {
		tc.set(strconv.Itoa(i), "bar", DefaultExpiration)
	}
	tc.mu.Unlock()
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		tc.DeleteExpired()
	}
}
