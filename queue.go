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
	queue        *bufferQueue
	ctx          *Context
	acquired     bool
	index        int64
	backoffState *BackoffState
}

func (t *task) do() error {
	if t.ctx.backoff != nil {
		t.backoffState.sleep(t.ctx.backoff)
	}
	return t.ctx.Do()
}

type bufferQueue struct {
	queue    *Queue
	index    int64
	tasks    []*task
	acquired []*task
}

func newBufferQueue(queue *Queue) *bufferQueue {
	bq := &bufferQueue{
		queue:    queue,
		tasks:    make([]*task, 0, queue.length),
		acquired: make([]*task, 0, queue.length),
	}
	return bq
}

func (bq *bufferQueue) insert(ctx *Context) error {
	bq.index += 1
	bq.tasks = append(bq.tasks, &task{
		queue:        bq,
		ctx:          ctx,
		index:        bq.index,
		backoffState: &BackoffState{},
	})
	return nil
}

func (bq *bufferQueue) acquire() *task {
	if len(bq.acquired) > 0 {
		for _, qt := range bq.acquired {
			if !qt.acquired {
				return qt
			}
		}
	}

	if len(bq.tasks) == 0 {
		return nil
	}

	t := bq.tasks[0]
	bq.tasks = bq.tasks[1:]
	t.acquired = true
	bq.acquired = append(bq.acquired, t)
	return t
}

func (bq *bufferQueue) ack(t *task, err error) {
	bq.queue.ack(func() {
		t.acquired = false
		if err == nil {
			for i, qt := range bq.acquired {
				if qt.index == t.index {
					bq.acquired = append(bq.acquired[:i], bq.acquired[i+1:]...)
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
		if t == nil {
			time.Sleep(interval)
			continue
		}

		t.queue.ack(t, t.do())
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
