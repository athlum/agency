package agency

import (
	"errors"
	"sync"
)

var (
	QueueNotFound = errors.New("Queue not found.")
	QueueCreated  = errors.New("Queue already exists.")

	defaultConf = &AssignConf{
		Workers:  100,
		Length:   100,
		Interval: 1.0,
		Overflow: false,
	}
)

type AssignConf struct {
	Workers  int64
	Length   int64
	Interval float64
	Overflow bool
}

type QueueMap struct {
	*sync.RWMutex
	queues map[string]*Queue
}

func (qm *QueueMap) emit(name string, ctx *Context) error {
	qm.RLock()
	if _, exists := qm.queues[name]; !exists {
		return QueueNotFound
	}
	qm.RUnlock()
	return qm.queues[name].Insert(ctx.priority, ctx)
}

func (qm *QueueMap) assign(name string, conf *AssignConf) error {
	qm.Lock()
	defer qm.Unlock()
	if _, exists := qm.queues[name]; exists {
		return QueueCreated
	}
	qm.queues[name] = newQueue(conf)
	return nil
}

func (qm *QueueMap) state(name string) *State {
	qm.RLock()
	q, exists := qm.queues[name]
	qm.RUnlock()

	if exists {
		return q.State()
	}
	return nil
}

var m *QueueMap

func init() {
	m = &QueueMap{
		RWMutex: &sync.RWMutex{},
		queues:  make(map[string]*Queue),
	}
}

func Emit(name string, ctx *Context) error {
	return m.emit(name, ctx)
}

func Assign(name string, conf *AssignConf) error {
	if conf == nil {
		conf = defaultConf
	}
	return m.assign(name, conf)
}
