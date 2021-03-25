package lru

import (
	"reflect"
	"testing"
)

type eleTest string

func (d eleTest) Len() int {
	return len(d)
}

func TestCache_Add(t *testing.T) {
	lru := New(int64(0), nil)
	lru.Add("key1", eleTest("1111"))
	if v, ok := lru.Get("key1"); !ok || string(v.(eleTest)) != "1111" {
		t.Fatalf("cache key1=1111 err")
	}
	if _, ok := lru.Get("key2"); ok {
		t.Fatalf("get not exist value key2")
	}
}

func TestCache_RemoveOldest(t *testing.T) {
	k1, k2, k3 := "key1", "key2", "k3"
	v1, v2, v3 := "value1", "value2", "value3"

	max := int64(len(k1 + k2 + v1 + v2))
	lru := New(max, nil)
	lru.Add(k1, eleTest(v1))
	lru.Add(k2, eleTest(v2))
	lru.Add(k3, eleTest(v3))

	if _, ok := lru.Get("key1"); ok || lru.Len() != 2 {
		t.Fatalf("remove oldest key1 fail")
	}
}

func TestCache_OnEvicted(t *testing.T) {
	keys := []string{}
	onEvicted := func(key string, value Value) {
		keys = append(keys, key)
	}
	k1, k2, k3 := "key1", "key2", "k3"
	v1, v2, v3 := "value1", "value2", "value3"

	max := int64(len(k1 + v1))
	lru := New(max, onEvicted)
	lru.Add(k1, eleTest(v1))
	lru.Add(k2, eleTest(v2))
	lru.Add(k3, eleTest(v3))

	expect := []string{"key1", "key2"}

	if !reflect.DeepEqual(expect, keys) {
		t.Fatalf("OnEvicted err, expect keys : %s", expect)
	}
}
