package agency

import (
	"errors"
	"git.elenet.me/appos/appos-scheduler/configuration"
	"git.elenet.me/appos/appos-scheduler/utils"
	"sync"
	"time"
)

var (
	QueueFull     = errors.New("Queue is full.")
	ErrorPriority = errors.New("Invalid priority.")
)

type task struct {
	queue    *bufferQueue
	ctx      *Context
	acquired bool
	index    int64
}

type bufferQueue struct {
	queue    *Queue
	index    int64
	tasks    []*task
	acquired []*task
}

func newBufferQueue(queue *Queue) *bufferQueue {
	bq := &bufferQueue{
		queue: queue,
		tasks: []*task{},
	}
	return bq
}

func (bq *bufferQueue) insert(ctx *Context) error {
	bq.index += 1
	bq.tasks = append(bq.tasks, &task{
		queue: bq,
		ctx:   ctx,
		index: bq.index,
	})
	return nil
}

func (bq *bufferQueue) acquire() *task {
	for _, qt := range bq.acquired {
		if !qt.acquired {
			return qt
		}
	}

	if len(bq.tasks) == 0 {
		return nil
	}

	t := bq.tasks[0]
	t.acquired = true
	bq.tasks = bq.tasks[1:]
	bq.acquired = append(bq.acquired, t)
	return t
}

func (bq *bufferQueue) ack(t *task, err error) {
	t.acquired = false
	if err == nil {
		for i, tt := range bq.acquired {
			if t.index == tt.index {
				bq.acquired = append(bq.acquired[:i], bq.acquired[i+1:])
				bq.queue.ack()
				break
			}
		}
	}
}

type Queue struct {
	*sync.Mutex
	length int64
	queues []*bufferQueue
}

func newQueue(conf *AssignConf) *Queue {
	q := &Queue{
		Mutex:  &sync.Mutex{},
		length: conf.Length,
		queues: make([]*bufferQueue, 3),
	}

	for p := 0; p < 3; p += 1 {
		overflow := false
		if len(conf.Detail) > p {
			overflow = conf.Detail[p].Overflow
		}
		q.queues[p] = newBufferQueue(q, overflow)
	}

	interval := utils.Duration(conf.Interval)
	for n := 0; n < conf.Workers; n += 1 {
		go q.worker(interval)
	}
	return q
}

func (q *Queue) worker(interval time.Duration) {
	for {
		t := q.acquire()
		if t == nil {
			time.Sleep(interval)
		}

		t.queue.ack(t, t.ctx.Do())
	}
}

func (q *Queue) Insert(p Priority, ctx *Context) error {
	q.Lock()
	defer q.Unlock()

	if p > Priority_Important {
		return ErrorPriority
	}

	if q.count == q.length {
		return QueueFull
	}

	q.count += 1
	return q.queues[p].insert(ctx)
}

func (q *Queue) acquire() *task {
	q.Lock()
	defer q.Unlock()

	for p := 2; p > -1; p += 1 {
		if t := q.queues[p].acquire(); t != nil {
			return t
		}
	}

	return nil
}

func (q *Queue) ack() {
	q.Lock()
	defer q.Unlock()

	q.count -= 1
}
