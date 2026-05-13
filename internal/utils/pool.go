package utils

import "sync"

type poolMember interface {
	Reset()
}

type Pool[T poolMember] struct {
	store []T
	mu    sync.Mutex
	new   func() T
}

func New[T poolMember](new func() T) *Pool[T] {
	return &Pool[T]{store: []T{}, mu: sync.Mutex{}, new: new}
}

func (p *Pool[T]) Put(x T) {
	p.mu.Lock()
	x.Reset()
	p.store = append(p.store, x)
	p.mu.Unlock()
}

func (p *Pool[T]) Get() T {
	p.mu.Lock()
	defer p.mu.Unlock()
	n := len(p.store)
	if n == 0 {
		return p.new()
	}
	x := p.store[n-1]
	p.store = p.store[:n-1]
	return x
}
