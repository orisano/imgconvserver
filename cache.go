package imgconvserver

import (
	"context"
	"time"
)

type Cache interface {
	Get(ctx context.Context, key string, expire int) (interface{}, error)
	Delete(key string) bool
}

type Func func(ctx context.Context, key string) (interface{}, error)

type result struct {
	value interface{}
	err   error
}

type requestType int

const (
	get requestType = iota
	del
)

type request struct {
	key      string
	response chan<- result
	expire   int
	tp       requestType
	ctx      context.Context
}

type entry struct {
	res   result
	ready chan struct{}
}

func (e *entry) call(ctx context.Context, f Func, key string) {
	value, err := f(ctx, key)
	e.res.value, e.res.err = value, err
	close(e.ready)
}

func (e *entry) deliver(response chan<- result) {
	<-e.ready
	response <- e.res
}

type cache struct {
	requests chan request
}

func New(f Func) Cache {
	cache := &cache{requests: make(chan request)}
	go cache.server(f)
	return cache
}

func (c *cache) Get(ctx context.Context, key string, expire int) (interface{}, error) {
	response := make(chan result)
	c.requests <- request{key, response, expire, get, ctx}
	select {
	case res := <-response:
		return res.value, res.err
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}

func (c *cache) Delete(key string) bool {
	response := make(chan result)
	c.requests <- request{key, response, -1, del, nil}
	<-response
	return true
}

func (c *cache) server(f Func) {
	memo := make(map[string]*entry)
	for req := range c.requests {
		switch req.tp {
		case get:
			e := memo[req.key]
			if e == nil {
				e = &entry{ready: make(chan struct{})}
				memo[req.key] = e
				go e.call(req.ctx, f, req.key)
				go func(req request) {
					select {
					case <-req.ctx.Done():
						c.Delete(req.key)
					case <-e.ready:
						if req.expire > 0 {
							time.Sleep(time.Millisecond * time.Duration(req.expire))
							c.Delete(req.key)
						}
					}
				}(req)
			}
			go e.deliver(req.response)
		case del:
			e, ok := memo[req.key]
			if !ok && e.res.value == nil {
				close(req.response)
				continue
			}
			delete(memo, req.key)
			close(req.response)
		}
	}
}
