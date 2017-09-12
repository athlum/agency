package agency

import (
	"errors"
	"sync"
	"time"
)

var (
	QueueFull     = errors.New("Queue is full.")
	ErrorPriority = errors.New("Invalid priority.")
)

type task struct {
	*sync.RWMutex
	queue        *bufferQueue
	ctx          *Context
	acquired     bool
	index        int64
	backoffState *BackoffState
	removed      bool
}

func (t *task) do() error {
	if t.ctx.backoff != nil {
		t.backoffState.sleep(t.ctx.backoff)
	}
	return t.ctx.Do()
}

func (t *task) isRemoved() bool {
	t.RLock()
	defer t.RUnlock()
	return t.removed
}

func (t *task) setRemove(b bool) {
	t.Lock()
	defer t.Unlock()

	t.removed = b
	if b {
		go t.ctx.Dropped()
	}
}

type bufferQueue struct {
	queue    *Queue
	index    int64
	tasks    *tasks
	acquired *tasks
}

func newBufferQueue(queue *Queue) *bufferQueue {
	bq := &bufferQueue{
		queue:    queue,
		tasks:    taskSlice(make([]*task, 0, queue.length)),
		acquired: taskSlice(make([]*task, 0, queue.length)),
	}
	return bq
}

func (bq *bufferQueue) drop() bool {
	var t *task
	if bq.acquired.length() > 0 {
		t = bq.acquired.pop()
	} else if bq.tasks.length() > 0 {
		t = bq.tasks.pop()
	}
	if t != nil {
		t.setRemove(true)
		return true
	}
	return false
}

func (bq *bufferQueue) insert(ctx *Context) error {
	bq.index += 1
	bq.tasks.push(&task{
		RWMutex:      &sync.RWMutex{},
		queue:        bq,
		ctx:          ctx,
		index:        bq.index,
		backoffState: &BackoffState{},
	})
	return nil
}

func (bq *bufferQueue) acquire() *task {
	if bq.acquired.length() > 0 {
		for _, qt := range bq.acquired.list() {
			if !qt.acquired {
				qt.acquired = true
				return qt
			}
		}
	}

	if bq.tasks.length() == 0 {
		return nil
	}

	t := bq.tasks.pop()
	t.acquired = true
	bq.acquired.push(t)
	return t
}

func (bq *bufferQueue) ack(t *task, err error) {
	bq.queue.ack(func() {
		t.acquired = false
		if err == nil {
			for i, qt := range bq.acquired.list() {
				if qt.index == t.index {
					bq.acquired.remove(i)
					break
				}
			}
		}
	})
}

type Queue struct {
	*sync.Mutex
	length   int64
	count    int64
	overflow bool
	queues   []*bufferQueue
}

func newQueue(conf *AssignConf) *Queue {
	q := &Queue{
		Mutex:    &sync.Mutex{},
		length:   conf.Length,
		overflow: conf.Overflow,
	}
	q.queues = []*bufferQueue{
		newBufferQueue(q),
		newBufferQueue(q),
		newBufferQueue(q),
	}

	interval := duration(conf.Interval)
	for n := int64(0); n < conf.Workers; n += 1 {
		go q.worker(interval)
	}
	return q
}

func (q *Queue) worker(interval time.Duration) {
	for {
		t := q.acquire()
		if t == nil || t.isRemoved() {
			time.Sleep(interval)
			continue
		}
		err := t.do()
		if !t.isRemoved() {
			t.queue.ack(t, err)
		}
	}
}

func (q *Queue) Insert(p Priority, ctx *Context) error {
	q.Lock()
	defer q.Unlock()

	if p > Priority_Important {
		return ErrorPriority
	}

	isFull := q.count == q.length
	if isFull && !q.overflow {
		return QueueFull
	}

	if isFull && q.overflow {
		var dropped bool
		ip := int(p)
		for tp := 0; tp <= ip; tp += 1 {
			if dropped = q.queues[tp].drop(); !dropped && tp == ip {
				return QueueFull
			} else if dropped {
				q.count -= 1
				break
			}
		}
	}

	q.count += 1
	return q.queues[p].insert(ctx)
}

func (q *Queue) acquire() *task {
	q.Lock()
	defer q.Unlock()

	for p := 2; p > -1; p -= 1 {
		if t := q.queues[p].acquire(); t != nil {
			return t
		}
	}

	return nil
}

func (q *Queue) ack(f func()) {
	q.Lock()
	defer q.Unlock()

	if f != nil {
		f()
	}
	q.count -= 1
}
