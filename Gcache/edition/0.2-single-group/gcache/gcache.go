package gcache

import (
	"fmt"
	"log"
	"sync"
)

type Getter interface {
	Get(key string) ([]byte, error)
}

type GetterFunc func(key string) ([]byte, error)

//接口型函数
func (f GetterFunc) Get(key string) ([]byte, error) {
	return f(key)
}

type Group struct {
	name   string
	getter Getter
	cache  cache
}

var (
	mu     sync.RWMutex
	groups = map[string]*Group{}
)

func NewGroup(name string, maxBytes int64, getter Getter) *Group {
	if getter == nil {
		panic("nil gettrt")
	}
	mu.Lock()
	defer mu.Unlock()

	g := &Group{
		name:   name,
		getter: getter,
		cache:  cache{maxBytes: maxBytes},
	}
	groups[name] = g
	return g
}

func GetGroup(name string) *Group {
	mu.RLock()
	defer mu.RUnlock()
	g := groups[name]
	return g
}

func (g *Group) Get(key string) (ByteView, error) {
	if key == "" {
		return ByteView{}, fmt.Errorf("key is empty")
	}

	if v, ok := g.cache.get(key); ok {
		log.Println("get cache value, group:", g.name, ",key:", key, ",value:", v.b)
		return v, nil
	}

	//not exist
	return g.load(key)
}

func (g *Group) load(key string) (value ByteView, err error) {
	return g.getLocally(key)
}

func (g *Group) getLocally(key string) (value ByteView, err error) {
	bytes, err := g.getter.Get(key)
	if err != nil {
		return ByteView{}, err
	}

	value = ByteView{cloneBytes(bytes)}
	g.populateCache(key, value)
	log.Println("load value, group:", g.name, ",key:", key, ",value:", bytes)
	return value, nil
}

func (g *Group) populateCache(key string, value ByteView) {
	g.cache.add(key, value)
}
