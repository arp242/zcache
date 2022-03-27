package zcache

import (
	"reflect"
	"testing"
)

func TestProxy(t *testing.T) {
	c := New[string, any](NoExpiration, 0)
	pc := NewProxy[string, string, any](c)

	has := func(v any, ok bool) {
		t.Helper()
		if !ok {
			t.Error("ok false")
		}
		if v != "vvv" {
			t.Errorf("value wrong: %q", v)
		}
	}
	not := func(v interface{}, ok bool) {
		t.Helper()
		if ok {
			t.Error("ok true")
		}
		if v != nil {
			t.Errorf("value not nil: %q", v)
		}
	}

	c.Set("k", "vvv")
	pc.Proxy("k", "p")
	has(pc.Get("p"))
	not(pc.Get("k"))

	pc.Delete("k")
	has(pc.Get("p"))
	pc.Delete("p")
	not(pc.Get("p"))

	pc.Set("main", "proxy", "vvv")
	has(pc.Get("proxy"))
	not(pc.Get("main"))

	if !reflect.DeepEqual(pc.Items(), map[string]string{"proxy": "main"}) {
		t.Error()
	}

	if k, ok := pc.Key("adsasdasd"); k != "" || ok != false {
		t.Error()
	}

	if k, ok := pc.Key("proxy"); k != "main" || ok != true {
		t.Error()
	}

	if pc.Cache() != c {
		t.Error()
	}

	pc.Flush()
	if len(pc.m) != 0 {
		t.Error()
	}
}
