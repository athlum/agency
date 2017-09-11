package agency

import (
	"time"
)

type tasks []*task

func taskSlice(tl []*task) *tasks {
	ts := tasks(tl)
	return &ts
}

func (ts *tasks) remove(index int) *task {
	e := (*ts)[index]
	copy((*ts)[:index], (*ts)[index+1:])
	(*ts)[len(*ts)-1] = nil
	*ts = (*ts)[:len(*ts)-1]
	return e
}

func (ts *tasks) pop() *task {
	return ts.remove(0)
}

func (ts *tasks) push(t *task) *tasks {
	*ts = append(*ts, t)
	return ts
}

func (ts *tasks) length() int {
	return len(*ts)
}

func (ts *tasks) list() []*task {
	return *ts
}

func duration(d float64) time.Duration {
	return time.Millisecond * time.Duration(d*1000)
}

func nextBackoff(backoff, factor, unit, max float64, backoffCount int) (float64, int) {
	if factor == 0.0 {
		return 0.0, backoffCount + 1
	}

	nextBackOff := backoff + float64(backoffCount)*factor*unit
	if nextBackOff > max {
		return max, backoffCount
	}
	return nextBackOff, backoffCount + 1
}
