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
	Dropped func()

	Context struct {
		context.Context
		handler  Handler
		dropped  Dropped
		out      chan interface{}
		priority Priority
		backoff  *Backoff
		err      error
	}

	Priority int
)

func WithContext(ctx context.Context, handler Handler, dropped Dropped, priority Priority) (*Context, chan<- interface{}) {
	c := &Context{
		Context:  ctx,
		handler:  handler,
		dropped:  dropped,
		out:      make(chan interface{}),
		priority: priority,
	}
	return c, c.out
}

func (c *Context) clear() {
	c.handler = nil
}

func (c *Context) WithBackoff(b *Backoff) *Context {
	c.backoff = b
	return c
}

func (c *Context) Dropped() {
	if c.dropped != nil {
		c.dropped()
	}
}

func (c *Context) Do() error {
	if c.handler != nil {
		c.err = c.handler(c.Context, c.out)
	}
	return c.err
}

func (c *Context) Err() error {
	if c.err != nil {
		return c.err
	}
	return c.Context.Err()
}
