package logging

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
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
	assert.NoError(t, logger.Open(ctx, testPath))
	go func() {
		assert.NoError(t, logger.Tail(ctx, testPath, func(entry ...*Entry) { wg.Done() }))
	}()
	go func() {
		assert.NoError(t, logger.Tail(ctx, testPath, func(entry ...*Entry) { wg.Done() }))
	}()

	<-time.After(500 * time.Millisecond)

	wg.Add(4)
	go func() {
		assert.NoError(t, logger.Write(ctx, testPath, testEntry))
		assert.NoError(t, logger.Write(ctx, testPath, testEntry))
	}()

	wg.Wait()

	wg.Add(1)
	go func() {
		assert.NoError(t, logger.Tail(ctx, testPath, func(entry ...*Entry) { wg.Done() }))
	}()

	<-time.After(500 * time.Millisecond)

	wg.Wait()
	cancel()
}
