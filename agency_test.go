package agency

import (
	"context"
	"fmt"
	"net/http"
	_ "net/http/pprof"
	"sync"
	"testing"
	"time"
)

func init() {
	go http.ListenAndServe("127.0.0.1:10030", nil)
	m.assign("sync", &AssignConf{
		Workers:  1,
		Length:   10000,
		Interval: 0.1,
		Overflow: true,
	})
	m.assign("async", &AssignConf{
		Workers:  1000,
		Length:   10000,
		Interval: 0.1,
		Overflow: true,
	})
}

func Benchmark_Sync(b *testing.B) {
	wg := &sync.WaitGroup{}

	// wg.Add(b.N)
	dropped := 0
	for n := 0; n < 1000000; n++ {
		wg.Add(1)
		go func() {
			done := false
			ctx, _ := WithContext(context.Background(), func(ctx context.Context, out chan<- interface{}) error {
				if !done {
					done = true
					wg.Done()
				}
				time.Sleep(time.Millisecond * 1)
				return nil
			}, func() {
				dropped += 1
				if !done {
					done = true
					wg.Done()
				}
			}, Priority_Normal)
			m.emit("sync", ctx)
		}()
	}

	wg.Wait()
	fmt.Println("Benchmark_Sync", dropped)
}

func Test_Async(b *testing.T) {
	wg := &sync.WaitGroup{}

	wg.Add(100000)
	dropped := 0
	do := 0
	for n := 0; n < 100000; n++ {
		// wg.Add(1)
		go func() {
			l := &sync.Mutex{}
			done := false
			ctx, _ := WithContext(context.Background(), func(ctx context.Context, out chan<- interface{}) error {
				l.Lock()
				defer l.Unlock()
				time.Sleep(time.Second * 1)
				if !done {
					do += 1
					fmt.Println("done", do, "dropped", dropped)
					done = true
					wg.Done()
				}
				return nil
			}, func() {
				l.Lock()
				defer l.Unlock()
				if !done {
					dropped += 1
					fmt.Println("done", do, "dropped", dropped)
					done = true
					wg.Done()
				}
			}, Priority_Normal)
			m.emit("async", ctx)
		}()
	}

	wg.Wait()
}

func Test_Backoff(b *testing.T) {
	wg := &sync.WaitGroup{}

	wg.Add(1)
	dropped := 0
	do := 0
	retry := 0
	for n := 0; n < 1; n++ {
		// wg.Add(1)
		go func() {
			l := &sync.Mutex{}
			done := false
			ctx, _ := WithContext(context.Background(), func(ctx context.Context, out chan<- interface{}) error {
				l.Lock()
				defer l.Unlock()
				time.Sleep(time.Second * 1)

				if retry < 10 {
					msg := fmt.Sprintf("%v", time.Now())
					fmt.Println("retry", retry, msg)
					retry += 1
					return fmt.Errorf(msg)
				}

				if !done {
					do += 1
					fmt.Println("done", do, "dropped", dropped)
					done = true
					wg.Done()
				}
				return nil
			}, func() {
				l.Lock()
				defer l.Unlock()
				if !done {
					dropped += 1
					fmt.Println("done", do, "dropped", dropped)
					done = true
					wg.Done()
				}
			}, Priority_Normal)
			m.emit("sync", ctx.WithBackoff(&Backoff{
				Backoff:       1.0,
				BackoffFactor: 1.0,
				MaxBackoff:    60.0,
			}))
		}()
	}

	wg.Wait()
}
