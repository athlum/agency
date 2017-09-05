package agency

import (
	"context"
	"sync"
	"testing"
	"time"
)

func init() {
	m.assign("sync", &AssignConf{
		Workers:  1,
		Length:   1000000,
		Interval: 0.1,
		Overflow: true,
	})
	m.assign("async", &AssignConf{
		Workers:  1000,
		Length:   1000000,
		Interval: 0.1,
		Overflow: true,
	})
}

func Benchmark_Sync(b *testing.B) {
	wg := &sync.WaitGroup{}

	// wg.Add(b.N)
	for n := 0; n < b.N; n++ {
		wg.Add(1)
		go func() {
			ctx, _ := WithContext(context.Background(), func(ctx context.Context, out chan<- interface{}) error {
				defer wg.Done()
				time.Sleep(time.Millisecond * 1)
				return nil
			}, Priority_Normal)
			m.emit("sync", ctx)
		}()
	}

	wg.Wait()
}

func Benchmark_Async(b *testing.B) {
	wg := &sync.WaitGroup{}

	wg.Add(b.N)
	for n := 0; n < b.N; n++ {
		// wg.Add(1)
		go func() {
			ctx, _ := WithContext(context.Background(), func(ctx context.Context, out chan<- interface{}) error {
				defer wg.Done()
				time.Sleep(time.Millisecond * 1)
				return nil
			}, Priority_Normal)
			m.emit("async", ctx)
		}()
	}

	wg.Wait()
}
