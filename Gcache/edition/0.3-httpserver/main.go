package main

import (
	"cache/gcache"
	"fmt"
	"log"
	"net/http"
)

var db = map[string]string{
	"stu1": "100",
	"stu2": "90",
	"stu3": "80",
}

func main() {
	gcache.NewGroup("scores", 100, gcache.GetterFunc(
		func(key string) ([]byte, error) {
			log.Println("search key:", key)
			if v, ok := db[key]; ok {
				return []byte(v), nil
			}
			return nil, fmt.Errorf("%s not exist", key)
		}))

	addr := "localhost:9999"
	peers := gcache.NewHTTPPool(addr)
	log.Println("geecache is running at", addr)
	log.Fatal(http.ListenAndServe(addr, peers))
}
