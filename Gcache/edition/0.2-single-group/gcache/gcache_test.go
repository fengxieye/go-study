package gcache

import (
	"fmt"
	"log"
	"testing"
)

var db = map[string]string{
	"stu1": "100",
	"stu2": "90",
	"stu3": "80",
}

func TestGroup_Get(t *testing.T) {
	loadCount := make(map[string]int, len(db))
	group := NewGroup("scores", 100, GetterFunc(
		func(key string) ([]byte, error) {
			log.Println("search key:", key)
			if v, ok := db[key]; ok {
				if _, ok := loadCount[key]; !ok {
					loadCount[key] = 0
				}
				loadCount[key] += 1
				return []byte(v), nil
			}
			return nil, fmt.Errorf("%s not exist", key)
		}))

	for k, v := range db {
		if view, err := group.Get(k); err != nil || view.String() != v {
			t.Fatal("fail get value, key:", k)
		}
		if _, err := group.Get(k); err != nil || loadCount[k] > 1 {
			t.Fatal("cache not save")
		}
	}

	if view, err := group.Get("unknown"); err == nil {
		t.Fatalf("the value of unknow should be empty, but %s got", view)
	}
}
