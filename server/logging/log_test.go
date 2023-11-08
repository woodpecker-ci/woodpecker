package logging

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"go.woodpecker-ci.org/woodpecker/server/model"
)

func TestLogging(t *testing.T) {
	var (
		wg sync.WaitGroup

		testStepID = int64(123)
		testEntry  = &model.LogEntry{
			Data: []byte("test"),
		}
	)

	ctx, cancel := context.WithCancelCause(
		context.Background(),
	)

	logger := New()
	assert.NoError(t, logger.Open(ctx, testStepID))
	go func() {
		assert.NoError(t, logger.Tail(ctx, testStepID, func(entry ...*model.LogEntry) { wg.Done() }))
	}()
	go func() {
		assert.NoError(t, logger.Tail(ctx, testStepID, func(entry ...*model.LogEntry) { wg.Done() }))
	}()

	<-time.After(500 * time.Millisecond)

	wg.Add(4)
	go func() {
		assert.NoError(t, logger.Write(ctx, testStepID, testEntry))
		assert.NoError(t, logger.Write(ctx, testStepID, testEntry))
	}()

	wg.Wait()

	wg.Add(1)
	go func() {
		assert.NoError(t, logger.Tail(ctx, testStepID, func(entry ...*model.LogEntry) { wg.Done() }))
	}()

	<-time.After(500 * time.Millisecond)

	wg.Wait()
	cancel(nil)
}
