package disCache

import (
	"fmt"
	"log"
	"sync"
)

type Getter interface {
	Get(key string) ([]byte, error)
}

type GetterFunc func(key string) ([]byte, error)

func (f GetterFunc) Get(key string) ([]byte, error) {
	return f(key)
}

type Group struct {
	name      string
	getter    Getter
	mainCache cache
	peers     PeerPicker
}

var (
	mu     sync.RWMutex
	groups = make(map[string]*Group)
)

func (g *Group) RegisterPeers(peers PeerPicker) {
	if g.peers != nil {
		panic("RegisterPeerPicker called more than once")
	}
	g.peers = peers
}

func (g *Group) getFromPeer(peer PeerGetter, key string) (ByteView, error) {
	bytes, err := peer.Get(g.name, key)
	if err != nil {
		return ByteView{}, err
	}
	return ByteView{b: bytes}, nil
}

func NewGroup(name string, cacheBytes int64, getter Getter) *Group {
	if getter == nil {
		panic("there is no Getter. If miss, where you get value to insert?")
	}
	mu.Lock()
	defer mu.Unlock()
	g := &Group{
		name:      name,
		getter:    getter,
		mainCache: cache{cacheBytes: cacheBytes},
	}
	groups[name] = g
	return g
}

func GetGroup(name string) *Group {
	mu.RLock()
	g := groups[name]
	mu.RUnlock()
	return g
}

func (g *Group) Get(key string) (ByteView, error) {
	if key == "" {
		return ByteView{}, fmt.Errorf("key is required")
	}
	//如果命中
	if v, ok := g.mainCache.get(key); ok {
		fmt.Println("[distributedCache] hit")
		return v, nil
	}
	//没有命中，向远程节点数据，如果远程节点也没有就从本地获取
	return g.load(key)
}

//peer是一个物理节点
func (g *Group) load(key string) (value ByteView, err error) {
	if g.peers != nil {

		if peer, ok := g.peers.PickPeer(key); ok {
			if value, err = g.getFromPeer(peer, key); err == nil {
				return value, err
			}
			log.Println("[disCache] Failed to get from peer", err)
		}
	}
	return g.getLocally(key)
}

func (g *Group) getLocally(key string) (ByteView, error) {
	//使用用户的提供的参数的回调方法得到一个数据，该回调方法需要用户自己写。
	//由用户实现一个getter对象并提供它的实例或者使用一个函数并将其强制转换为GetterFunc那么就会获得我们帮它写好的Get方法不需要自己写接口的方法
	//此getter函数的get方法用于从数据库获取数据
	bytes, err := g.getter.Get(key)
	if err != nil {
		return ByteView{}, err
	}
	value := ByteView{b: cloneBytes(bytes)}
	//将用户提供的数据添加到主缓存中
	g.populateCache(key, value)
	return value, nil

}

func (g *Group) populateCache(key string, value ByteView) {
	g.mainCache.add(key, value)
}
