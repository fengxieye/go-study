package gcache

import (
	"cache/gcache/singleflight"
	"fmt"
	"log"
	"sync"
)

//回调函数
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
	peers  PeerPicker
	//
	loader *singleflight.Group
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
		loader: &singleflight.Group{},
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

func Get(groupName string, key string) (ByteView, error) {
	group := groups[groupName]
	if group == nil {
		log.Println("group not exist")
		return ByteView{}, fmt.Errorf("group not exist")
	}
	return group.Get(key)
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
	viewi, err := g.loader.Do(key, func() (i interface{}, err error) {
		if g.peers != nil {
			if peer, ok := g.peers.PickPeer(key); ok {
				if value, err = g.getFromPeer(peer, key); err == nil {
					return value, nil
				}
				log.Println("failed to get from peer, err:", err)
			}
		}
		return g.getLocally(key)
	})

	if err != nil {
		return viewi.(ByteView), nil
	}

	return
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

func (g *Group) RegisterPeers(peers PeerPicker) {
	if g.peers != nil {
		panic("RegisterPeers more than once")
	}
	g.peers = peers
}

func (g *Group) getFromPeer(peer PeerGetter, key string) (ByteView, error) {
	bytes, err := peer.Get(g.name, key)
	if err != nil {
		return ByteView{}, err
	}

	return ByteView{bytes}, nil
}
