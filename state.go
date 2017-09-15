package agency

import (
	"encoding/json"
	"time"
)

type durationCounter struct {
	lived time.Duration
	value float64
}

func (dc *durationCounter) push(n float64, d time.Duration) float64 {
	dc.value = (dc.value*float64(dc.lived) + n*float64(time.Second)) / float64(d)
	dc.lived = d
	return dc.value
}

func (dc *durationCounter) MarshalJSON() ([]byte, error) {
	return json.Marshal(dc.value)
}

type State struct {
	Start                 time.Time
	Insert                int64
	InsertedPerSecond     *durationCounter
	Acquired              int64
	AcquiredPerSecond     *durationCounter
	Acknowledged          int64
	AcknowledgedPerSecond *durationCounter
	Dropped               int64
	DroppedPerSecond      *durationCounter
}

func NewState() *State {
	return &State{
		Start:                 time.Now(),
		InsertedPerSecond:     &durationCounter{},
		AcquiredPerSecond:     &durationCounter{},
		AcknowledgedPerSecond: &durationCounter{},
		DroppedPerSecond:      &durationCounter{},
	}
}

func (s *State) Lived() time.Duration {
	return time.Now().Sub(s.Start)
}

func (s *State) insert() {
	s.Insert += 1
	s.InsertedPerSecond.push(1, s.Lived())
}

func (s *State) acq() {
	s.Acquired += 1
	s.AcquiredPerSecond.push(1, s.Lived())
}

func (s *State) ack() {
	s.Acknowledged += 1
	s.AcknowledgedPerSecond.push(1, s.Lived())
}

func (s *State) dropped() {
	s.Dropped += 1
	s.DroppedPerSecond.push(1, s.Lived())
}

func (s *State) MarshalJSON() ([]byte, error) {
	o := &struct {
		Start                 time.Time
		Insert                int64
		InsertedPerSecond     *durationCounter
		Acquired              int64
		AcquiredPerSecond     *durationCounter
		Acknowledged          int64
		AcknowledgedPerSecond *durationCounter
		Dropped               int64
		DroppedPerSecond      *durationCounter
		Lived                 time.Duration
	}{
		Start:                 s.Start,
		InsertedPerSecond:     s.InsertedPerSecond,
		AcquiredPerSecond:     s.AcquiredPerSecond,
		AcknowledgedPerSecond: s.AcknowledgedPerSecond,
		DroppedPerSecond:      s.DroppedPerSecond,
		Insert:                s.Insert,
		Acquired:              s.Acquired,
		Acknowledged:          s.Acknowledged,
		Dropped:               s.Dropped,
		Lived:                 s.Lived(),
	}
	return json.Marshal(o)
}
