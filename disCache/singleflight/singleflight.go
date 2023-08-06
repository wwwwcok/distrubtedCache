package singleflight

import "sync"

type call struct {
	wg  sync.WaitGroup
	val interface{}
	err error
}

type Group struct {
	mu sync.Mutex //保护请求映射
	m  map[string]*call
}

func (g *Group) Do(key string, fn func() (interface{}, error)) (interface{}, error) {
	g.mu.Lock()
	if g.m == nil {
		g.m = make(map[string]*call)
	}
	// key有对应的value，说明有相同的key只在执行，当前的请求需要等待。
	if c, ok := g.m[key]; ok {
		g.mu.Unlock()
		c.wg.Wait()
		return c.val, c.err
	}
	//
	c := new(call)
	c.wg.Add(1)
	//请求前先加锁
	g.m[key] = c
	//使用完map解锁
	g.mu.Unlock()

	c.val, c.err = fn() //调用,发起并接收请求
	c.wg.Done()         //结束后减一

	g.mu.Lock()
	delete(g.m, key) //更新 g.m
	g.mu.Unlock()
	return c.val, c.err

}
