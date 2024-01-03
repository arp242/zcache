package zcache_test

import (
	"fmt"
	"slices"
	"testing"
	"time"

	"zgo.at/zcache/v2"
)

func ExampleCache_Keyset() {
	c := zcache.New[string, string](5*time.Minute, 10*time.Minute)
	c.Set("key1", "one")
	c.Set("key2", "two")

	keys := c.Keyset("key1", "key2", "not set")
	for _, k := range keys.Get() {
		fmt.Println(k)
	}
}

func TestKeyset(t *testing.T) {
	c := zcache.New[string, string](5*time.Minute, 10*time.Minute)
	var evicted [][2]string
	c.OnEvicted(func(k, v string) {
		evicted = append(evicted, [2]string{k, v})
	})
	c.Set("key1", "one")
	c.Set("key2", "two")
	c.SetWithExpire("key3", "two", 1)
	keys := c.Keyset("key1", "key2", "not set")

	{
		have := fmt.Sprintf("%v", keys.Get())
		want := `[{one true} {two true} { false}]`
		if have != want {
			t.Errorf("\nhave: %s\nwant: %s", have, want)
		}
	}
	{
		have := fmt.Sprintf("%v", c.Keyset("key1", "key2", "key3").Get())
		want := `[{one true} {two true} { false}]`
		if have != want {
			t.Errorf("\nhave: %s\nwant: %s", have, want)
		}
	}

	{
		c.Keyset("key2", "not set").Delete()
		have := fmt.Sprintf("%v", keys.Get())
		want := `[{one true} { false} { false}]`
		if have != want {
			t.Errorf("\nhave: %s\nwant: %s", have, want)
		}
		have = fmt.Sprintf("%v", evicted)
		want = `[[key2 two]]`
		if have != want {
			t.Errorf("\nhave: %s\nwant: %s", have, want)
		}
	}

	{
		c.Keyset("new1", "new2").Set("x", "y")
		keys.Append("new1", "new2")
		have := fmt.Sprintf("%v", keys.Get())
		want := `[{one true} { false} { false} {x true} {y true}]`
		if have != want {
			t.Errorf("\nhave: %s\nwant: %s", have, want)
		}
	}

	{
		keys.Reset()
		have := fmt.Sprintf("%v", keys.Get())
		want := `[]`
		if have != want {
			t.Errorf("\nhave: %s\nwant: %s", have, want)
		}
	}

	{
		keys := c.Find(func(k string, v zcache.Item[string]) (bool, bool) {
			return false, false
		})
		have := fmt.Sprintf("%v", keys.Keys())
		want := `[]`
		if have != want {
			t.Errorf("\nhave: %s\nwant: %s", have, want)
		}
	}
	{
		keys := c.Find(func(k string, v zcache.Item[string]) (bool, bool) {
			return true, false
		})
		kk := keys.Keys()
		slices.Sort(kk)
		have := fmt.Sprintf("%v", kk)
		want := `[key1 key3 new1 new2]`
		if have != want {
			t.Errorf("\nhave: %s\nwant: %s", have, want)
		}
	}
	{
		keys := c.Find(func(k string, v zcache.Item[string]) (bool, bool) {
			return true, true
		})
		if l := len(keys.Keys()); l != 1 {
			t.Errorf("len not 1: %d", l)
		}
	}
}

// func TestMultiGet(t *testing.T) {
// 	tc := New[int, string](time.Second, 0)
//
// 	tc.MultiSet([]int{0, 1, 3}, []string{"y0", "y1", "y3"})
// 	tc.SetWithExpire(5, "y5", time.Millisecond*10)
//
// 	values, existsArr := tc.MultiGet(0, 1, 2, 3, 4, 5)
// 	time.Sleep(time.Millisecond * 10)
//
// 	for i, exists := range existsArr {
// 		if exists {
// 			if i == 2 || i == 4 {
// 				t.Errorf("Item exists, but shouldn't %v", i)
// 			}
// 			value := values[i]
// 			if value != "y"+strconv.Itoa(i) {
// 				t.Errorf("Wrong Item found %v: %v", i, value)
// 			}
// 		} else {
// 			if !(i == 2 || i == 4) {
// 				t.Errorf("Item doesn't exists, but should %v", i)
// 			}
// 		}
// 	}
// }
