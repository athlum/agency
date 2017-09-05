package agency

import (
	"context"
)

const (
	Priority_Important Priority = 2 - iota
	Priority_Normal
	Priority_Low
)

type (
	Handler func(context.Context, chan<- interface{}) error

	Context struct {
		context.Context
		handler  Handler
		out      chan interface{}
		priority Priority
		backoff  *Backoff
	}

	Priority int
)

func WithContext(ctx context.Context, handler Handler, priority Priority) (*Context, chan<- interface{}) {
	c := &Context{
		Context:  ctx,
		handler:  handler,
		out:      make(chan interface{}),
		priority: priority,
	}
	return c, c.out
}

func (c *Context) WithBackoff(b *Backoff) *Context {
	c.backoff = b
	return c
}

func (c *Context) Do() error {
	return c.handler(c.Context, c.out)
}
