package zcache

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"reflect"
	"sort"
	"sync"
	"testing"
	"time"
)

func wantKeys(t *testing.T, tc *Cache, want []string, dontWant []string) {
	t.Helper()

	for _, k := range want {
		_, ok := tc.Get(k)
		if !ok {
			t.Errorf("key not found: %q", k)
		}
	}

	for _, k := range dontWant {
		v, ok := tc.Get(k)
		if ok {
			t.Errorf("key %q found with value %v", k, v)
		}
		if v != nil {
			t.Error("v is not nil:", v)
		}
	}
}

type TestStruct struct {
	Num      int
	Children []*TestStruct
}

func TestCache(t *testing.T) {
	tc := New(DefaultExpiration, 0)

	a, found := tc.Get("a")
	if found || a != nil {
		t.Error("Getting A found value that shouldn't exist:", a)
	}

	b, found := tc.Get("b")
	if found || b != nil {
		t.Error("Getting B found value that shouldn't exist:", b)
	}

	c, found := tc.Get("c")
	if found || c != nil {
		t.Error("Getting C found value that shouldn't exist:", c)
	}

	tc.Set("a", 1, DefaultExpiration)
	tc.Set("b", "b", DefaultExpiration)
	tc.Set("c", 3.5, DefaultExpiration)

	v, found := tc.Get("a")
	if !found {
		t.Error("a was not found while getting a2")
	}
	if v == nil {
		t.Error("v for a is nil")
	} else if a2 := v.(int); a2+2 != 3 {
		t.Error("a2 (which should be 1) plus 2 does not equal 3; value:", a2)
	}

	v, found = tc.Get("b")
	if !found {
		t.Error("b was not found while getting b2")
	}
	if v == nil {
		t.Error("v for b is nil")
	} else if b2 := v.(string); b2+"B" != "bB" {
		t.Error("b2 (which should be b) plus B does not equal bB; value:", b2)
	}

	v, found = tc.Get("c")
	if !found {
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

	tc := New(50*time.Millisecond, 1*time.Millisecond)
	tc.Set("a", 1, DefaultExpiration)
	tc.Set("b", 2, NoExpiration)
	tc.Set("c", 3, 20*time.Millisecond)
	tc.Set("d", 4, 70*time.Millisecond)

	<-time.After(25 * time.Millisecond)
	_, found = tc.Get("c")
	if found {
		t.Error("Found c when it should have been automatically deleted")
	}

	<-time.After(30 * time.Millisecond)
	_, found = tc.Get("a")
	if found {
		t.Error("Found a when it should have been automatically deleted")
	}

	_, found = tc.Get("b")
	if !found {
		t.Error("Did not find b even though it was set to never expire")
	}

	_, found = tc.Get("d")
	if !found {
		t.Error("Did not find d even though it was set to expire later than the default")
	}

	<-time.After(20 * time.Millisecond)
	_, found = tc.Get("d")
	if found {
		t.Error("Found d when it should have been automatically deleted (later than the default)")
	}
}

func TestNewFrom(t *testing.T) {
	m := map[string]Item{
		"a": {
			Object:     1,
			Expiration: 0,
		},
		"b": {
			Object:     2,
			Expiration: 0,
		},
	}
	tc := NewFrom(DefaultExpiration, 0, m)
	a, found := tc.Get("a")
	if !found {
		t.Fatal("Did not find a")
	}
	if a.(int) != 1 {
		t.Fatal("a is not 1")
	}
	b, found := tc.Get("b")
	if !found {
		t.Fatal("Did not find b")
	}
	if b.(int) != 2 {
		t.Fatal("b is not 2")
	}
}

func TestStorePointerToStruct(t *testing.T) {
	tc := New(DefaultExpiration, 0)
	tc.Set("foo", &TestStruct{Num: 1}, DefaultExpiration)
	v, found := tc.Get("foo")
	if !found {
		t.Fatal("*TestStruct was not found for foo")
	}
	foo := v.(*TestStruct)
	foo.Num++

	y, found := tc.Get("foo")
	if !found {
		t.Fatal("*TestStruct was not found for foo (second time)")
	}
	bar := y.(*TestStruct)
	if bar.Num != 2 {
		t.Fatal("TestStruct.Num is not 2")
	}
}

func TestOnEvicted(t *testing.T) {
	tc := New(DefaultExpiration, 0)
	tc.Set("foo", 3, DefaultExpiration)
	if tc.onEvicted != nil {
		t.Fatal("tc.onEvicted is not nil")
	}
	works := false
	tc.OnEvicted(func(k string, v interface{}) {
		if k == "foo" && v.(int) == 3 {
			works = true
		}
		tc.Set("bar", 4, DefaultExpiration)
	})
	tc.Delete("foo")
	v, _ := tc.Get("bar")
	if !works {
		t.Error("works bool not true")
	}
	if v.(int) != 4 {
		t.Error("bar was not 4")
	}
}

func TestCacheSerialization(t *testing.T) {
	tc := New(DefaultExpiration, 0)
	testFillAndSerialize(t, tc)

	// Check if gob.Register behaves properly even after multiple gob.Register
	// on c.Items (many of which will be the same type)
	testFillAndSerialize(t, tc)
}

func testFillAndSerialize(t *testing.T, tc *Cache) {
	tc.Set("a", "a", DefaultExpiration)
	tc.Set("b", "b", DefaultExpiration)
	tc.Set("c", "c", DefaultExpiration)
	tc.Set("expired", "foo", 1*time.Millisecond)
	tc.Set("*struct", &TestStruct{Num: 1}, DefaultExpiration)
	tc.Set("[]struct", []TestStruct{
		{Num: 2},
		{Num: 3},
	}, DefaultExpiration)
	tc.Set("[]*struct", []*TestStruct{
		{Num: 4},
		{Num: 5},
	}, DefaultExpiration)
	tc.Set("structception", &TestStruct{
		Num: 42,
		Children: []*TestStruct{
			{Num: 6174},
			{Num: 4716},
		},
	}, DefaultExpiration)

	fp := &bytes.Buffer{}
	err := tc.Save(fp)
	if err != nil {
		t.Fatal("Couldn't save cache to fp:", err)
	}

	oc := New(DefaultExpiration, 0)
	err = oc.Load(fp)
	if err != nil {
		t.Fatal("Couldn't load cache from fp:", err)
	}

	a, found := oc.Get("a")
	if !found {
		t.Error("a was not found")
	}
	if a.(string) != "a" {
		t.Error("a is not a")
	}

	b, found := oc.Get("b")
	if !found {
		t.Error("b was not found")
	}
	if b.(string) != "b" {
		t.Error("b is not b")
	}

	c, found := oc.Get("c")
	if !found {
		t.Error("c was not found")
	}
	if c.(string) != "c" {
		t.Error("c is not c")
	}

	<-time.After(5 * time.Millisecond)
	_, found = oc.Get("expired")
	if found {
		t.Error("expired was found")
	}

	s1, found := oc.Get("*struct")
	if !found {
		t.Error("*struct was not found")
	}
	if s1.(*TestStruct).Num != 1 {
		t.Error("*struct.Num is not 1")
	}

	s2, found := oc.Get("[]struct")
	if !found {
		t.Error("[]struct was not found")
	}
	s2r := s2.([]TestStruct)
	if len(s2r) != 2 {
		t.Error("Length of s2r is not 2")
	}
	if s2r[0].Num != 2 {
		t.Error("s2r[0].Num is not 2")
	}
	if s2r[1].Num != 3 {
		t.Error("s2r[1].Num is not 3")
	}

	s3, found := oc.get("[]*struct")
	if !found {
		t.Error("[]*struct was not found")
	}
	s3r := s3.([]*TestStruct)
	if len(s3r) != 2 {
		t.Error("Length of s3r is not 2")
	}
	if s3r[0].Num != 4 {
		t.Error("s3r[0].Num is not 4")
	}
	if s3r[1].Num != 5 {
		t.Error("s3r[1].Num is not 5")
	}

	s4, found := oc.get("structception")
	if !found {
		t.Error("structception was not found")
	}
	s4r := s4.(*TestStruct)
	if len(s4r.Children) != 2 {
		t.Error("Length of s4r.Children is not 2")
	}
	if s4r.Children[0].Num != 6174 {
		t.Error("s4r.Children[0].Num is not 6174")
	}
	if s4r.Children[1].Num != 4716 {
		t.Error("s4r.Children[1].Num is not 4716")
	}
}

func TestFileSerialization(t *testing.T) {
	tc := New(DefaultExpiration, 0)
	tc.Add("a", "a", DefaultExpiration)
	tc.Add("b", "b", DefaultExpiration)
	f, err := ioutil.TempFile("", "go-cache-cache.dat")
	if err != nil {
		t.Fatal("Couldn't create cache file:", err)
	}
	fname := f.Name()
	f.Close()
	tc.SaveFile(fname)

	oc := New(DefaultExpiration, 0)
	oc.Add("a", "aa", 0) // this should not be overwritten
	err = oc.LoadFile(fname)
	if err != nil {
		t.Error(err)
	}
	a, found := oc.Get("a")
	if !found {
		t.Error("a was not found")
	}
	astr := a.(string)
	if astr != "aa" {
		if astr == "a" {
			t.Error("a was overwritten")
		} else {
			t.Error("a is not aa")
		}
	}
	b, found := oc.Get("b")
	if !found {
		t.Error("b was not found")
	}
	if b.(string) != "b" {
		t.Error("b is not b")
	}
}

func TestSerializeUnserializable(t *testing.T) {
	tc := New(DefaultExpiration, 0)
	ch := make(chan bool, 1)
	ch <- true
	tc.Set("chan", ch, DefaultExpiration)
	fp := &bytes.Buffer{}
	err := tc.Save(fp) // this should fail gracefully
	if err.Error() != "gob NewTypeObject can't handle type: chan bool" {
		t.Error("Error from Save was not gob NewTypeObject can't handle type chan bool:", err)
	}
}

func TestTouch(t *testing.T) {
	tc := New(DefaultExpiration, 0)

	tc.Set("a", "b", 5*time.Second)
	_, first, _ := tc.GetWithExpiration("a")
	v, ok := tc.Touch("a", 10*time.Second)
	if !ok {
		t.Fatal("!ok")
	}
	_, second, _ := tc.GetWithExpiration("a")
	if v.(string) != "b" {
		t.Error("wrong value")
	}

	if first.Equal(second) {
		t.Errorf("not updated\nfirst:  %s\nsecond: %s", first, second)
	}
}

func TestGetWithExpiration(t *testing.T) {
	tc := New(DefaultExpiration, 0)

	a, expiration, ok := tc.GetWithExpiration("a")
	if ok || a != nil || !expiration.IsZero() {
		t.Error("Getting A found value that shouldn't exist:", a)
	}

	b, expiration, ok := tc.GetWithExpiration("b")
	if ok || b != nil || !expiration.IsZero() {
		t.Error("Getting B found value that shouldn't exist:", b)
	}

	c, expiration, ok := tc.GetWithExpiration("c")
	if ok || c != nil || !expiration.IsZero() {
		t.Error("Getting C found value that shouldn't exist:", c)
	}

	tc.Set("a", 1, DefaultExpiration)
	tc.Set("b", "b", DefaultExpiration)
	tc.Set("c", 3.5, DefaultExpiration)
	tc.Set("d", 1, NoExpiration)
	tc.Set("e", 1, 50*time.Millisecond)

	v, expiration, ok := tc.GetWithExpiration("a")
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

	v, expiration, ok = tc.GetWithExpiration("b")
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

	v, expiration, ok = tc.GetWithExpiration("c")
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

	v, expiration, ok = tc.GetWithExpiration("d")
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

	v, expiration, ok = tc.GetWithExpiration("e")
	if !ok {
		t.Error("e was not found while getting e2")
	}
	if v == nil {
		t.Error("v for e is nil")
	} else if e2 := v.(int); e2+2 != 3 {
		t.Error("e (which should be 1) plus 2 does not equal 3; value:", e2)
	}
	if expiration.UnixNano() != tc.items["e"].Expiration {
		t.Error("expiration for e is not the correct time")
	}
	if expiration.UnixNano() < time.Now().UnixNano() {
		t.Error("expiration for e is in the past")
	}
}

func TestGetStale(t *testing.T) {
	tc := New(5*time.Millisecond, 0)

	tc.SetDefault("x", "y")

	v, exp, ok := tc.GetStale("x")
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

	v, ok = tc.Get("x")
	if ok || v != nil {
		t.Fatalf("Get retrieved expired item: %v", v)
	}

	v, exp, ok = tc.GetStale("x")
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
	tc := New(DefaultExpiration, 0)
	err := tc.Add("foo", "bar", DefaultExpiration)
	if err != nil {
		t.Error("Couldn't add foo even though it shouldn't exist")
	}
	err = tc.Add("foo", "baz", DefaultExpiration)
	if err == nil {
		t.Error("Successfully added another foo when it should have returned an error")
	}
}

func TestReplace(t *testing.T) {
	tc := New(DefaultExpiration, 0)
	err := tc.Replace("foo", "bar", DefaultExpiration)
	if err == nil {
		t.Error("Replaced foo when it shouldn't exist")
	}
	tc.Set("foo", "bar", DefaultExpiration)
	err = tc.Replace("foo", "bar", DefaultExpiration)
	if err != nil {
		t.Error("Couldn't replace existing key foo")
	}
}

func TestDelete(t *testing.T) {
	tc := New(DefaultExpiration, 0)

	tc.Set("foo", "bar", DefaultExpiration)
	tc.Delete("foo")
	wantKeys(t, tc, []string{}, []string{"foo"})
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
	tc := New(DefaultExpiration, 0)

	var onEvict onEvictTest
	tc.OnEvicted(onEvict.add)

	tc.Set("foo", "val", DefaultExpiration)

	v, ok := tc.Pop("foo")
	wantKeys(t, tc, []string{}, []string{"foo"})
	if !ok {
		t.Error("ok is false")
	}
	if v.(string) != "val" {
		t.Errorf("wrong value: %v", v)
	}

	v, ok = tc.Pop("nonexistent")
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
	tc := New(DefaultExpiration, 0)

	tc.Set("k", []string{"x"}, DefaultExpiration)
	ok := tc.Modify("k", func(v interface{}) interface{} {
		vv := v.([]string)
		vv = append(vv, "y")
		return vv
	})
	if !ok {
		t.Error("ok is false")
	}
	v, _ := tc.Get("k")
	if fmt.Sprintf("%v", v) != `[x y]` {
		t.Errorf("value wrong: %v", v)
	}

	ok = tc.Modify("doesntexist", func(v interface{}) interface{} {
		t.Error("should not be called")
		return nil
	})
	if ok {
		t.Error("ok is true")
	}

	tc.Modify("k", func(v interface{}) interface{} { return nil })
	v, ok = tc.Get("k")
	if !ok {
		t.Error("ok not set")
	}
	if v != nil {
		t.Error("v not nil")
	}
}

func TestItems(t *testing.T) {
	tc := New(DefaultExpiration, 1*time.Millisecond)
	tc.Set("foo", "1", DefaultExpiration)
	tc.Set("bar", "2", DefaultExpiration)
	tc.Set("baz", "3", DefaultExpiration)
	tc.Set("exp", "4", 1)
	time.Sleep(2 * time.Millisecond)
	if n := tc.ItemCount(); n != 3 {
		t.Errorf("Item count is not 3: %d", n)
	}

	keys := tc.Keys()
	sort.Strings(keys)
	if fmt.Sprintf("%v", keys) != "[bar baz foo]" {
		t.Errorf("%v", keys)
	}

	want := map[string]Item{
		"foo": {Object: "1"},
		"bar": {Object: "2"},
		"baz": {Object: "3"},
	}
	if !reflect.DeepEqual(tc.Items(), want) {
		t.Errorf("%v", tc.Items())
	}
}

func TestFlush(t *testing.T) {
	tc := New(DefaultExpiration, 0)
	tc.Set("foo", "bar", DefaultExpiration)
	tc.Set("baz", "yes", DefaultExpiration)
	tc.Flush()
	v, found := tc.Get("foo")
	if found {
		t.Error("foo was found, but it should have been deleted")
	}
	if v != nil {
		t.Error("v is not nil:", v)
	}
	v, found = tc.Get("baz")
	if found {
		t.Error("baz was found, but it should have been deleted")
	}
	if v != nil {
		t.Error("v is not nil:", v)
	}
}

func TestDeleteAll(t *testing.T) {
	tc := New(DefaultExpiration, 0)
	tc.Set("foo", 3, DefaultExpiration)
	if tc.onEvicted != nil {
		t.Fatal("tc.onEvicted is not nil")
	}
	works := false
	tc.OnEvicted(func(k string, v interface{}) {
		if k == "foo" && v.(int) == 3 {
			works = true
		}
	})
	tc.DeleteAll()
	if !works {
		t.Error("works bool not true")
	}
}

func TestDeleteFunc(t *testing.T) {
	tc := New(NoExpiration, 0)
	tc.Set("foo", 3, DefaultExpiration)
	tc.Set("bar", 4, DefaultExpiration)

	works := false
	tc.OnEvicted(func(k string, v interface{}) {
		if k == "foo" && v.(int) == 3 {
			works = true
		}
	})

	tc.DeleteFunc(func(k string, v Item) (bool, bool) {
		return k == "foo" && v.Object.(int) == 3, false
	})

	if !works {
		t.Error("onEvicted isn't called for 'foo'")
	}

	_, found := tc.Get("bar")
	if !found {
		t.Error("bar shouldn't be removed from the cache")
	}

	tc.Set("boo", 5, DefaultExpiration)

	count := tc.ItemCount()

	// Only one item should be deleted here
	tc.DeleteFunc(func(k string, v Item) (bool, bool) {
		return true, true
	})

	if tc.ItemCount() != count-1 {
		t.Errorf("unexpected number of items in the cache. item count expected %d, found %d", count-1, tc.ItemCount())
	}
}

func TestGetOrSet(t *testing.T) {
	getOrSetOnce = once{}

	tc := New(DefaultExpiration, 0)
	tc.SetDefault("a", "aa")
	vv := tc.GetOrSet("a", func() (interface{}, time.Duration) {
		t.Error("this should not be run")
		return "set", DefaultExpiration
	})
	if vv != "aa" {
		t.Errorf("vv is not aa: %v", vv)
	}

	var (
		v, v2 interface{}
		wg    sync.WaitGroup
	)
	wg.Add(2)
	go func() {
		defer wg.Done()
		v = tc.GetOrSet("x", func() (interface{}, time.Duration) {
			time.Sleep(50 * time.Millisecond)
			return "first", DefaultExpiration
		})
	}()
	time.Sleep(10 * time.Millisecond)
	go func() {
		defer wg.Done()
		v2 = tc.GetOrSet("x", func() (interface{}, time.Duration) {
			t.Error("this should not be run")
			return "second", DefaultExpiration
		})
	}()
	wg.Wait()

	if v != "first" {
		t.Errorf("v is not first: %v", v)
	}
	if v2 != "first" {
		t.Errorf("v2 is not first: %v", v2)
	}

	time.Sleep(10 * time.Millisecond) // TODO: until I fix once
	tc.Delete("x")
	v = tc.GetOrSet("x", func() (interface{}, time.Duration) {
		return "after", DefaultExpiration
	})
	if v != "after" {
		t.Errorf("v is not after: %v", v)
	}

}
