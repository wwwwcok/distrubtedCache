package distrubtedCache

import (
	"distrubtedCache/lru"
	"sync"
)

type cache struct {
	mu sync.Mutex
	//	ch         chan struct{}
	lru        *lru.Cache
	cacheBytes int64
}

func (c *cache) add(key string, val ByteView) {
	c.mu.Lock()
	defer c.mu.Unlock()
	//<-c.ch

	//延迟初始化,即第一次使用该对象才会初始化
	if c.lru == nil {
		c.lru = lru.New(c.cacheBytes, nil)
	}
	c.lru.Add(key, val)
	//c.ch <- struct{}{}
}

func (c *cache) get(key string) (ByteView, bool) {
	//<-c.ch
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.lru == nil {
		return ByteView{}, false
	}

	if v, ok := c.lru.Get(key); ok {
		return v.(ByteView), ok
	}
	return ByteView{}, false
	//c.ch<-struct{}{}
}
