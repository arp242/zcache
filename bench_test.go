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
	tc := New[string, any](exp, 0)
	tc.Set("foo", "bar")
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		tc.Get("foo")
	}
}

func benchmarkGetConcurrent(b *testing.B, exp time.Duration) {
	b.StopTimer()
	tc := New[string, any](exp, 0)
	tc.Set("foo", "bar")
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
	tc := New[string, any](exp, 0)
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		tc.Set("foo", "bar")
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
	m := map[interface{}]string{
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
	m := map[interface{}]string{
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
	for i := 0; i < workers; i++ {
		go func() {
			for j := 0; j < each; j++ {
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
	tc := New[string, any](DefaultExpiration, 0)
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		tc.Set("foo", "bar")
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
	tc := New[string, any](DefaultExpiration, 0)
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

func BenchmarkDeleteExpiredLoop(b *testing.B) {
	b.StopTimer()
	tc := New[string, any](5*time.Minute, 0)
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

func repeat(n int, s ...string) []string {
	r := make([]string, 0, len(s)*n)
	for i := 0; i < n; i++ {
		r = append(r, s...)
	}
	return r
}

func BenchmarkKS(b *testing.B) {
	b.Run("Get", func(b *testing.B) {
		tc := New[string, any](0, 0)
		tc.Set("foo", "bar")
		tc.Set("bar", "barzxc")
		tc.Set("asd", "barzxc")

		do := func(b *testing.B, keys []string) {
			b.Run("one", func(b *testing.B) {
				for i := 0; i < b.N; i++ {
					for _, k := range keys {
						tc.Get(k)
					}
				}
			})
			b.Run("mul", func(b *testing.B) {
				for i := 0; i < b.N; i++ {
					tc.Keyset(keys...).Get()
				}
			})
		}

		s := []string{"foo", "bar", "baz", "asd", "zcx", "qwe", "fdg", "zxcz", "asdsada", "qweqewqe"}
		b.Run("0", func(b *testing.B) { do(b, []string{}) })
		b.Run("10", func(b *testing.B) { do(b, repeat(1, s...)) })
		b.Run("50", func(b *testing.B) { do(b, repeat(5, s...)) })
		b.Run("100", func(b *testing.B) { do(b, repeat(10, s...)) })
		b.Run("500", func(b *testing.B) { do(b, repeat(50, s...)) })
		b.Run("1000", func(b *testing.B) { do(b, repeat(100, s...)) })
	})

	b.Run("Set", func(b *testing.B) {
		tc := New[string, string](0, 0)

		do := func(b *testing.B, keys []string, vals []string) {
			b.Run("one", func(b *testing.B) {
				for i := 0; i < b.N; i++ {
					for _, k := range keys {
						tc.Set(k, "test value")
					}
				}
			})
			b.Run("mul", func(b *testing.B) {
				for i := 0; i < b.N; i++ {
					tc.Keyset(keys...).Set(vals...)
				}
			})
		}

		s := []string{"foo", "bar", "baz", "asd", "zcx", "qwe", "fdg", "zxcz", "asdsada", "qweqewqe"}
		vals := repeat(10, "test value")
		b.Run("0", func(b *testing.B) { do(b, []string{}, []string{}) })
		b.Run("10", func(b *testing.B) { do(b, repeat(1, s...), repeat(1, vals...)) })
		b.Run("50", func(b *testing.B) { do(b, repeat(5, s...), repeat(5, vals...)) })
		b.Run("100", func(b *testing.B) { do(b, repeat(10, s...), repeat(10, vals...)) })
		b.Run("500", func(b *testing.B) { do(b, repeat(50, s...), repeat(50, vals...)) })
		b.Run("1000", func(b *testing.B) { do(b, repeat(100, s...), repeat(100, vals...)) })
	})
}
