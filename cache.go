package imgconvserver

import "context"

type Cache interface {
	Start() error
	Get(ctx context.Context, key string) (interface{}, error)
	Delete(key string) bool
}

type Func func(ctx context.Context, key string) (interface{}, error)

type result struct {
	value interface{}
	err   error
}

type request struct {
	key      string
	response chan<- result
	ctx      context.Context
}

type entry struct {
	res    result
	expire int
	ready  chan struct{}
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

func (cache) Start() error {
	panic("implement me")
}

func (c *cache) Get(ctx context.Context, key string) (interface{}, error) {
	response := make(chan result)
	c.requests <- request{key, response, ctx}
	select {
	case res := <-resopnse:
		return res.value, res.err
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}

func (cache) Delete(key string) bool {
	return true
}

func (c *cache) server(f Func) {
	memo := make(map[string]*entry)
	for req := range c.requests {
		e := memo[req.key]
		if e == nil {
			e = &entry{ready: make(chan struct{})}
			memo[req.key] = e
			<-e.ready

		}
	}
}
