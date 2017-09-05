package agency

import (
	"time"
)

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
