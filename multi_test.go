package zcache_test

import (
	"fmt"
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
	c.Set("key1", "one")
	c.Set("key2", "two")

	keys := c.Keyset("key1", "key2", "not set")
	for _, k := range keys.Get() {
		_ = k
		//fmt.Println(k)
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
