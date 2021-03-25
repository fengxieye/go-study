package main

import (
	"cache/gcache"
	"flag"
	"fmt"
	"log"
	"net/http"
)

var db = map[string]string{
	"stu1": "100",
	"stu2": "90",
	"stu3": "80",
}

var db2 = map[string]string{
	"stu1": "50",
	"stu2": "60",
	"stu3": "90",
}

func createGroup() *gcache.Group {
	return gcache.NewGroup("scores", 100, gcache.GetterFunc(
		func(key string) ([]byte, error) {
			log.Println("search key:", key)
			if v, ok := db[key]; ok {
				return []byte(v), nil
			}
			return nil, fmt.Errorf("%s not exist", key)
		}))
}

func createGroup2() *gcache.Group {
	return gcache.NewGroup("scores2", 100, gcache.GetterFunc(
		func(key string) ([]byte, error) {
			log.Println("search key:", key)
			if v, ok := db2[key]; ok {
				return []byte(v), nil
			}
			return nil, fmt.Errorf("%s not exist", key)
		}))
}

func startCacheServer(addr string, addrs []string, group ...*gcache.Group) {
	peers := gcache.NewHTTPPool(addr)
	peers.Set(addrs...)
	for _, v := range group {
		v.RegisterPeers(peers)
	}

	log.Println("gcache is running at", addr)
	log.Fatal(http.ListenAndServe(addr[7:], peers))
}

func startAPIServer(addr string) {
	http.Handle("/api", http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			key := r.URL.Query().Get("key")
			groupName := r.URL.Query().Get("group")
			view, err := gcache.Get(groupName, key)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}
			w.Header().Set("Content-Type", "text/plain")
			w.Write(view.ByteSlice())

		}))

	log.Println("fontend server is running at", addr)
	log.Fatal(http.ListenAndServe(addr[7:], nil))
}

/*
下一步可以改成读配置文件：修改成不同的组选择不同的节点
*/

func main() {
	var port int
	var api bool
	var groupNames string
	flag.IntVar(&port, "port", 8001, "Geecache server port")
	flag.BoolVar(&api, "api", false, "Start a api server?")
	flag.StringVar(&groupNames, "groups", "", "Start a api server?")
	flag.Parse()

	apiAddr := "http://localhost:9999"
	addrMap := map[int]string{
		8001: "http://localhost:8001",
		8002: "http://localhost:8002",
		8003: "http://localhost:8003",
	}

	var addrs []string
	for _, v := range addrMap {
		addrs = append(addrs, v)
	}

	group := createGroup()
	group2 := createGroup2()

	if api {
		go startAPIServer(apiAddr)
	}
	startCacheServer(addrMap[port], []string(addrs), group, group2)
}
