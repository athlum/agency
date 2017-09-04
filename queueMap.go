package agency

import (
	"errors"
)

var (
	QueueNotFound = errors.New("Queue not found.")
	QueueCreated  = errors.New("Queue already exists.")

	defaultConf = &AssignConf{
		Workers:  100,
		Length:   100,
		Interval: 1.0,
		Detail:   []*QueueConf{},
	}
)

type AssignConf struct {
	Workers  int64
	Length   int64
	Interval float64
	Detail   []*QueueConf
}

type QueueConf struct {
	Overflow bool
}

type QueueMap struct {
	queues map[string]*Queue
}

func (qm *QueueMap) emit(name string, ctx *Context) error {
	if _, exists := qm.queues[name]; !exists {
		return QueueNotFound
	}
	return qm.queues[name].Insert(ctx.priority, ctx)
}

func (qm *QueueMap) assign(name string, conf *AssignConf) error {
	if _, exists := qm.queues[name]; exists {
		return QueueCreated
	}
	qm.queues[name] = newQueue(conf)
	return nil
}

var m *QueueMap

func init() {
	m = &QueueMap{
		queues: make(map[string]*Queue),
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
