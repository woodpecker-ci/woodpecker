package logging

import (
	"context"
	"sync"
	"testing"
	"time"
)

func TestLogging(t *testing.T) {
	var (
		wg sync.WaitGroup

		testPath  = "test"
		testEntry = &Entry{
			Data: []byte("test"),
		}
	)

	ctx, cancel := context.WithCancel(
		context.Background(),
	)

	logger := New()
	logger.Open(ctx, testPath)
	go func() {
		logger.Tail(ctx, testPath, func(entry ...*Entry) { wg.Done() })
	}()
	go func() {
		logger.Tail(ctx, testPath, func(entry ...*Entry) { wg.Done() })
	}()

	<-time.After(time.Millisecond)

	wg.Add(4)
	go func() {
		logger.Write(ctx, testPath, testEntry)
		logger.Write(ctx, testPath, testEntry)
	}()

	wg.Wait()

	wg.Add(1)
	go func() {
		logger.Tail(ctx, testPath, func(entry ...*Entry) { wg.Done() })
	}()

	<-time.After(time.Millisecond)

	wg.Wait()
	cancel()
}
