package agency

import (
	"time"
)

type Backoff struct {
	Backoff       float64
	BackoffFactor float64
	MaxBackoff    float64
}

func (b *Backoff) next(base float64, wave int) (float64, int) {
	return nextBackoff(base, b.BackoffFactor, b.Backoff, b.MaxBackoff, wave)
}

type BackoffState struct {
	Backoff float64
	Retry   int
}

func (bs *BackoffState) sleep(b *Backoff) {
	if bs.Retry > 0 {
		bs.Backoff, bs.Retry = b.next(bs.Backoff, bs.Retry)
		time.Sleep(duration(bs.Backoff))
	} else {
		bs.Retry += 1
	}
}
