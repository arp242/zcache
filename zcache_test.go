package zcache

import (
	"bytes"
	"fmt"
	"reflect"
	"runtime"
	"sort"
	"sync"
	"testing"
	"time"
)

func wantKeys(t *testing.T, c *Cache[string, any], want []string, dontWant []string) {
	t.Helper()

	for _, k := range want {
		_, ok := c.Get(k)
		if !ok {
			t.Errorf("key not found: %q", k)
		}
	}

	for _, k := range dontWant {
		v, ok := c.Get(k)
		if ok {
			t.Errorf("key %q found with value %v", k, v)
		}
		if v != nil {
			t.Error("v is not nil:", v)
		}
	}
}

func TestCache(t *testing.T) {
	c := New[string, any](DefaultExpiration, 0)

	v1, ok := c.Get("a")
	if ok || v1 != nil {
		t.Error("Getting A found value that shouldn't exist:", v1)
	}
	v2, ok := c.Get("b")
	if ok || v2 != nil {
		t.Error("Getting B found value that shouldn't exist:", v2)
	}
	v3, ok := c.Get("c")
	if ok || v3 != nil {
		t.Error("Getting C found value that shouldn't exist:", v3)
	}

	c.Set("a", 1)
	c.Set("b", "b")
	c.Set("c", 3.5)

	v, ok := c.Get("a")
	if !ok {
		t.Error("a was not found while getting a2")
	}
	if v == nil {
		t.Error("v for a is nil")
	} else if a2 := v.(int); a2+2 != 3 {
		t.Error("a2 (which should be 1) plus 2 does not equal 3; value:", a2)
	}

	v, ok = c.Get("b")
	if !ok {
		t.Error("b was not found while getting b2")
	}
	if v == nil {
		t.Error("v for b is nil")
	} else if b2 := v.(string); b2+"B" != "bB" {
		t.Error("b2 (which should be b) plus B does not equal bB; value:", b2)
	}

	v, ok = c.Get("c")
	if !ok {
		t.Error("c was not found while getting c2")
	}
	if v == nil {
		t.Error("v for c is nil")
	} else if c2 := v.(float64); c2+1.2 != 4.7 {
		t.Error("c2 (which should be 3.5) plus 1.2 does not equal 4.7; value:", c2)
	}
}

func TestCacheTimes(t *testing.T) {
	var found bool

	c := New[string, int](50*time.Millisecond, 1*time.Millisecond)
	c.Set("a", 1)
	c.SetWithExpire("b", 2, NoExpiration)
	c.SetWithExpire("c", 3, 20*time.Millisecond)
	c.SetWithExpire("d", 4, 70*time.Millisecond)

	<-time.After(25 * time.Millisecond)
	_, found = c.Get("c")
	if found {
		t.Error("Found c when it should have been automatically deleted")
	}

	<-time.After(30 * time.Millisecond)
	_, found = c.Get("a")
	if found {
		t.Error("Found a when it should have been automatically deleted")
	}

	_, found = c.Get("b")
	if !found {
		t.Error("Did not find b even though it was set to never expire")
	}

	_, found = c.Get("d")
	if !found {
		t.Error("Did not find d even though it was set to expire later than the default")
	}

	<-time.After(20 * time.Millisecond)
	_, found = c.Get("d")
	if found {
		t.Error("Found d when it should have been automatically deleted (later than the default)")
	}
}

func TestNewFrom(t *testing.T) {
	m := map[string]Item[int]{
		"a": {
			Object:     1,
			Expiration: 0,
		},
		"b": {
			Object:     2,
			Expiration: 0,
		},
	}
	c := NewFrom[string, int](DefaultExpiration, 0, m)
	a, found := c.Get("a")
	if !found {
		t.Fatal("Did not find a")
	}
	if a != 1 {
		t.Fatal("a is not 1")
	}
	b, found := c.Get("b")
	if !found {
		t.Fatal("Did not find b")
	}
	if b != 2 {
		t.Fatal("b is not 2")
	}
}

func TestStorePointerToStruct(t *testing.T) {
	type TestStruct struct {
		Num      int
		Children []*TestStruct
	}

	c := New[string, any](DefaultExpiration, 0)
	c.Set("foo", &TestStruct{Num: 1})
	v, found := c.Get("foo")
	if !found {
		t.Fatal("*TestStruct was not found for foo")
	}
	foo := v.(*TestStruct)
	foo.Num++

	y, found := c.Get("foo")
	if !found {
		t.Fatal("*TestStruct was not found for foo (second time)")
	}
	bar := y.(*TestStruct)
	if bar.Num != 2 {
		t.Fatal("TestStruct.Num is not 2")
	}
}

func TestOnEvicted(t *testing.T) {
	c := New[string, int](DefaultExpiration, 0)
	c.Set("foo", 3)
	if c.onEvicted != nil {
		t.Fatal("c.onEvicted is not nil")
	}
	works := false
	c.OnEvicted(func(k string, v int) {
		if k == "foo" && v == 3 {
			works = true
		}
		c.Set("bar", 4)
	})
	c.Delete("foo")
	v, _ := c.Get("bar")
	if !works {
		t.Error("works bool not true")
	}
	if v != 4 {
		t.Error("bar was not 4")
	}
}

func TestTouch(t *testing.T) {
	c := New[string, string](DefaultExpiration, 0)

	c.SetWithExpire("a", "b", 5*time.Second)
	_, first, _ := c.GetWithExpire("a")
	v, ok := c.TouchWithExpire("a", 10*time.Second)
	if !ok {
		t.Fatal("!ok")
	}
	_, second, _ := c.GetWithExpire("a")
	if v != "b" {
		t.Error("wrong value")
	}
	if first.Equal(second) {
		t.Errorf("not updated\nfirst:  %s\nsecond: %s", first, second)
	}
}

func TestGetWithExpire(t *testing.T) {
	c := New[string, any](DefaultExpiration, 0)

	v1, expiration, ok := c.GetWithExpire("a")
	if ok || v1 != nil || !expiration.IsZero() {
		t.Error("Getting A found value that shouldn't exist:", v1)
	}

	v2, expiration, ok := c.GetWithExpire("b")
	if ok || v2 != nil || !expiration.IsZero() {
		t.Error("Getting B found value that shouldn't exist:", v2)
	}

	v3, expiration, ok := c.GetWithExpire("c")
	if ok || v3 != nil || !expiration.IsZero() {
		t.Error("Getting C found value that shouldn't exist:", v3)
	}

	c.Set("a", 1)
	c.Set("b", "b")
	c.Set("c", 3.5)
	c.SetWithExpire("d", 1, NoExpiration)
	c.SetWithExpire("e", 1, 50*time.Millisecond)

	v, expiration, ok := c.GetWithExpire("a")
	if !ok {
		t.Error("a was not found while getting a2")
	}
	if v == nil {
		t.Error("v for a is nil")
	} else if a2 := v.(int); a2+2 != 3 {
		t.Error("a2 (which should be 1) plus 2 does not equal 3; value:", a2)
	}
	if !expiration.IsZero() {
		t.Error("expiration for a is not a zeroed time")
	}

	v, expiration, ok = c.GetWithExpire("b")
	if !ok {
		t.Error("b was not found while getting b2")
	}
	if v == nil {
		t.Error("v for b is nil")
	} else if b2 := v.(string); b2+"B" != "bB" {
		t.Error("b2 (which should be b) plus B does not equal bB; value:", b2)
	}
	if !expiration.IsZero() {
		t.Error("expiration for b is not a zeroed time")
	}

	v, expiration, ok = c.GetWithExpire("c")
	if !ok {
		t.Error("c was not found while getting c2")
	}
	if v == nil {
		t.Error("v for c is nil")
	} else if c2 := v.(float64); c2+1.2 != 4.7 {
		t.Error("c2 (which should be 3.5) plus 1.2 does not equal 4.7; value:", c2)
	}
	if !expiration.IsZero() {
		t.Error("expiration for c is not a zeroed time")
	}

	v, expiration, ok = c.GetWithExpire("d")
	if !ok {
		t.Error("d was not found while getting d2")
	}
	if v == nil {
		t.Error("v for d is nil")
	} else if d2 := v.(int); d2+2 != 3 {
		t.Error("d (which should be 1) plus 2 does not equal 3; value:", d2)
	}
	if !expiration.IsZero() {
		t.Error("expiration for d is not a zeroed time")
	}

	v, expiration, ok = c.GetWithExpire("e")
	if !ok {
		t.Error("e was not found while getting e2")
	}
	if v == nil {
		t.Error("v for e is nil")
	} else if e2 := v.(int); e2+2 != 3 {
		t.Error("e (which should be 1) plus 2 does not equal 3; value:", e2)
	}
	if expiration.UnixNano() != c.items["e"].Expiration {
		t.Error("expiration for e is not the correct time")
	}
	if expiration.UnixNano() < time.Now().UnixNano() {
		t.Error("expiration for e is in the past")
	}
}

func TestGetStale(t *testing.T) {
	c := New[string, any](5*time.Millisecond, 0)

	c.Set("x", "y")

	v, exp, ok := c.GetStale("x")
	if !ok {
		t.Errorf("Did not get expired item: %v", v)
	}
	if exp {
		t.Error("exp set")
	}
	if v.(string) != "y" {
		t.Errorf("value wrong: %v", v)
	}

	time.Sleep(10 * time.Millisecond)

	v, ok = c.Get("x")
	if ok || v != nil {
		t.Fatalf("Get retrieved expired item: %v", v)
	}

	v, exp, ok = c.GetStale("x")
	if !ok {
		t.Errorf("Did not get expired item: %v", v)
	}
	if !exp {
		t.Error("exp not set")
	}
	if v.(string) != "y" {
		t.Errorf("value wrong: %v", v)
	}
}

func TestAdd(t *testing.T) {
	c := New[string, any](DefaultExpiration, 0)
	err := c.Add("foo", "bar")
	if err != nil {
		t.Error("Couldn't add foo even though it shouldn't exist")
	}
	err = c.Add("foo", "baz")
	if err == nil {
		t.Error("Successfully added another foo when it should have returned an error")
	}
}

func TestReplace(t *testing.T) {
	c := New[string, string](DefaultExpiration, 0)
	err := c.Replace("foo", "bar")
	if err == nil {
		t.Error("Replaced foo when it shouldn't exist")
	}
	c.Set("foo", "bar")
	err = c.Replace("foo", "bar")
	if err != nil {
		t.Error("Couldn't replace existing key foo")
	}
}

func TestDelete(t *testing.T) {
	c := New[string, any](DefaultExpiration, 0)

	c.Set("foo", "bar")
	c.Delete("foo")
	wantKeys(t, c, []string{}, []string{"foo"})
}

type onEvictTest struct {
	sync.Mutex
	items []struct {
		k string
		v interface{}
	}
}

func (o *onEvictTest) add(k string, v interface{}) {
	if k == "race" {
		return
	}
	o.Lock()
	o.items = append(o.items, struct {
		k string
		v interface{}
	}{k, v})
	o.Unlock()
}

func TestPop(t *testing.T) {
	c := New[string, any](DefaultExpiration, 0)

	var onEvict onEvictTest
	c.OnEvicted(onEvict.add)

	c.Set("foo", "val")

	v, ok := c.Pop("foo")
	wantKeys(t, c, []string{}, []string{"foo"})
	if !ok {
		t.Error("ok is false")
	}
	if v.(string) != "val" {
		t.Errorf("wrong value: %v", v)
	}

	v, ok = c.Pop("nonexistent")
	if ok {
		t.Error("ok is true")
	}
	if v != nil {
		t.Errorf("v is not nil")
	}

	if fmt.Sprintf("%v", onEvict.items) != `[{foo val}]` {
		t.Errorf("onEvicted: %v", onEvict.items)
	}
}

func TestModify(t *testing.T) {
	c := New[string, []string](DefaultExpiration, 0)

	c.Set("k", []string{"x"})
	v, ok := c.Modify("k", func(v []string) []string {
		return append(v, "y")
	})
	if !ok {
		t.Error("ok is false")
	}
	if fmt.Sprintf("%v", v) != `[x y]` {
		t.Errorf("value wrong: %v", v)
	}

	_, ok = c.Modify("doesntexist", func(v []string) []string {
		t.Error("should not be called")
		return nil
	})
	if ok {
		t.Error("ok is true")
	}

	v, ok = c.Modify("k", func(v []string) []string { return nil })
	if !ok {
		t.Error("ok not set")
	}
	if v != nil {
		t.Error("v not nil")
	}

	c = New[string, []string](1, 0)
	c.Set("expired", []string{"x"})
	time.Sleep(time.Nanosecond)
	_, ok = c.Modify("expired", func(v []string) []string {
		t.Error("should not be called")
		return nil
	})
	if ok {
		t.Error("ok is true")
	}
}

func TestModifyIncrement(t *testing.T) {
	c := New[string, int](DefaultExpiration, 0)
	c.Set("one", 1)

	have, _ := c.Modify("one", func(v int) int { return v + 2 })
	if have != 3 {
		t.Fatal()
	}

	have, _ = c.Modify("one", func(v int) int { return v - 1 })
	if have != 2 {
		t.Fatal()
	}
}

func TestModifySet(t *testing.T) {
	c := New[string, []string](DefaultExpiration, 0)

	c.Set("k", []string{"x"})
	v, ok := c.ModifySet("k", func(v []string, ok bool) []string {
		if !ok {
			t.Error("ok is false")
		}
		return append(v, "y")
	})
	if fmt.Sprintf("%v", v) != `[x y]` {
		t.Errorf("value wrong: %v", v)
	}
	if !ok {
		t.Error("ok false")
	}

	v, ok = c.ModifySet("doesntexist", func(v []string, ok bool) []string {
		if ok {
			t.Error("ok is true")
		}
		return nil
	})
	if v != nil {
		t.Errorf("v not nil: %v", v)
	}
	if ok {
		t.Error("ok true")
	}

	v, ok = c.ModifySet("doesntexist", func(v []string, ok bool) []string {
		if !ok {
			t.Error("ok is false")
		}
		return nil
	})
	if v != nil {
		t.Errorf("v not nil: %v", v)
	}
	if !ok {
		t.Error("ok false")
	}
	v, ok = c.Get("doesntexist")
	if !ok {
		t.Error("ok is false")
	}
	if v != nil {
		t.Errorf("v not nil: %v", v)
	}
	v, ok = c.ModifySet("doesntexist", func(v []string, ok bool) []string {
		return nil
	})
	if !ok {
		t.Error("ok false")
	}
	if v != nil {
		t.Errorf("v not nil: %v", v)
	}
	v, ok = c.Get("doesntexist")
	if !ok {
		t.Error("ok is false")
	}
	if v != nil {
		t.Errorf("v not nil: %v", v)
	}

	c = New[string, []string](1, 0)
	c.Set("expired", []string{"x"})
	time.Sleep(time.Nanosecond)
	v, ok = c.ModifySet("expired", func(v []string, ok bool) []string {
		if ok {
			t.Error("ok is true")
		}
		return []string{"a", "b"}
	})
	if fmt.Sprintf("%v", v) != `[a b]` {
		t.Errorf("value wrong: %v", v)
	}
	if ok {
		t.Error("ok true")
	}
}

func TestModifySetIncrement(t *testing.T) {
	c := New[string, int](DefaultExpiration, 0)
	c.Set("one", 1)

	have, ok := c.ModifySet("one", func(v int, ok bool) int { return v + 2 })
	if have != 3 {
		t.Fatal()
	}
	if !ok {
		t.Fatal()
	}

	have, ok = c.ModifySet("one", func(v int, ok bool) int { return v - 1 })
	if have != 2 {
		t.Fatal()
	}
	if !ok {
		t.Fatal()
	}
}

func TestItems(t *testing.T) {
	c := New[string, any](DefaultExpiration, 1*time.Millisecond)
	c.Set("foo", "1")
	c.Set("bar", "2")
	c.Set("baz", "3")
	c.SetWithExpire("exp", "4", 1)
	time.Sleep(10 * time.Millisecond)
	if n := c.ItemCount(); n != 3 {
		t.Errorf("Item count is not 3 but %d", n)
	}

	keys := c.Keys()
	sort.Strings(keys)
	if fmt.Sprintf("%v", keys) != "[bar baz foo]" {
		t.Errorf("%v", keys)
	}

	want := map[string]Item[any]{
		"foo": {Object: "1"},
		"bar": {Object: "2"},
		"baz": {Object: "3"},
	}
	if !reflect.DeepEqual(c.Items(), want) {
		t.Errorf("%v", c.Items())
	}
}

func TestReset(t *testing.T) {
	c := New[string, any](DefaultExpiration, 0)
	c.Set("foo", "bar")
	c.Set("baz", "yes")
	c.Reset()
	v, found := c.Get("foo")
	if found {
		t.Error("foo was found, but it should have been deleted")
	}
	if v != nil {
		t.Error("v is not nil:", v)
	}
	v, found = c.Get("baz")
	if found {
		t.Error("baz was found, but it should have been deleted")
	}
	if v != nil {
		t.Error("v is not nil:", v)
	}
}

func TestDeleteAll(t *testing.T) {
	c := New[string, any](DefaultExpiration, 0)
	c.Set("foo", 3)
	if c.onEvicted != nil {
		t.Fatal("c.onEvicted is not nil")
	}
	works := false
	c.OnEvicted(func(k string, v interface{}) {
		if k == "foo" && v.(int) == 3 {
			works = true
		}
	})
	c.DeleteAll()
	if !works {
		t.Error("works bool not true")
	}
}

func TestDeleteFunc(t *testing.T) {
	c := New[string, any](NoExpiration, 0)
	c.Set("foo", 3)
	c.Set("bar", 4)

	works := false
	c.OnEvicted(func(k string, v interface{}) {
		if k == "foo" && v.(int) == 3 {
			works = true
		}
	})

	c.DeleteFunc(func(k string, v Item[any]) (bool, bool) {
		return k == "foo" && v.Object.(int) == 3, false
	})

	if !works {
		t.Error("onEvicted isn't called for 'foo'")
	}

	_, found := c.Get("bar")
	if !found {
		t.Error("bar shouldn't be removed from the cache")
	}

	c.Set("boo", 5)

	count := c.ItemCount()

	// Only one item should be deleted here
	c.DeleteFunc(func(k string, v Item[any]) (bool, bool) {
		return true, true
	})

	if c.ItemCount() != count-1 {
		t.Errorf("unexpected number of items in the cache. item count expected %d, found %d", count-1, c.ItemCount())
	}
}

// Make sure the janitor is stopped after GC frees up.
func TestFinal(t *testing.T) {
	has := func() bool {
		s := make([]byte, 8192)
		runtime.Stack(s, true)
		return bytes.Contains(s, []byte("zgo.at/zcache/v2.(*janitor[...]).run"))
	}

	c := New[string, any](10*time.Millisecond, 10*time.Millisecond)
	c.Set("asd", "zxc")

	if !has() {
		t.Fatal("no janitor goroutine before GC")
	}
	runtime.GC()
	time.Sleep(10 * time.Millisecond)
	if has() {
		t.Fatal("still have janitor goroutine after GC")
	}
}

func TestRename(t *testing.T) {
	c := New[string, int](NoExpiration, 0)
	c.Set("foo", 3)
	c.SetWithExpire("bar", 4, time.Nanosecond)
	time.Sleep(time.Nanosecond)

	if c.Rename("nonex", "asd") {
		t.Error("nonex reported as existing")
	}
	if c.Rename("bar", "expired") {
		t.Error("bar reported as existing (should be expired)")
	}
	if v, exp, ok := c.GetStale("bar"); !ok || v != 4 {
		t.Errorf(`GetStale("Bar"): v=%v; exp=%v; ok=%v`, v, exp, ok)
	}

	if !c.Rename("foo", "RENAME") {
		t.Error(`rename "foo" to "RENAME" failed`)
	}

	if v, ok := c.Get("RENAME"); !ok || v != 3 {
		t.Errorf(`Get("RENAME"): v=%v; ok=%v`, v, ok)
	}

	if v, ok := c.Get("foo"); ok {
		t.Errorf(`Get("foo"): v=%v; ok=%v`, v, ok)
	}
}

func TestGetOrAdd(t *testing.T) {
	c := New[string, int](NoExpiration, 0)
	v := c.GetOrAdd("a", 1)
	if v != 1 {
		t.Errorf("GetOrAdd did not return the added value: v=%v", v)
	}
	if v2 := c.GetOrAdd("a", 2); v2 != v {
		t.Errorf("GetOrAdd did not return the existing value: v=%v; v2=%v", v, v2)
	}
	if _, exp, _ := c.GetWithExpire("a"); !exp.Equal(time.Time{}) {
		t.Errorf("GetOrAdd did not add the default no expiration: exp=%v", exp)
	}
	if v2 := c.GetOrAddWithExpire("b", 2, time.Second); v2 != 2 {
		t.Errorf("GetOrAddWithExpire did not return the added value: v2=%v", v2)
	}
	_, expiration, _ := c.GetWithExpire("b")
	time.Sleep(time.Nanosecond)
	if expiration.After(time.Now().Add(time.Second)) {
		t.Errorf("GetOrAddWithExpire did not add the non-default expiration: expiration:%v", expiration)
	}
}

func TestTouchOrAdd(t *testing.T) {
	c := New[string, int](NoExpiration, 0)
	v := c.TouchOrAdd("a", 1)
	if v != 1 {
		t.Errorf("TouchOrAdd did not return the added value: v=%v", v)
	}
	if v2 := c.TouchOrAdd("a", 2); v2 != v {
		t.Errorf("TouchOrAdd did not return the existing value: v=%v; v2=%v", v, v2)
	}
	if _, exp, _ := c.GetWithExpire("a"); !exp.Equal(time.Time{}) {
		t.Errorf("TouchOrAdd did not add the default no expiration: exp=%v", exp)
	}
	if v2 := c.TouchOrAddWithExpire("a", 2, time.Second); v2 != v {
		t.Errorf("TouchOrAddWithExpire did not return the existing value: v=%v; v2=%v", v, v2)
	}
	_, expiration, _ := c.GetWithExpire("a")
	time.Sleep(time.Nanosecond)
	if expiration.After(time.Now().Add(time.Second)) {
		t.Errorf("TouchOrAddWithExpire did not set the expiration: expiration:%v", expiration)
	}
}
