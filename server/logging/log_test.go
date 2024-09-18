// Copyright 2023 Woodpecker Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package logging

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"go.woodpecker-ci.org/woodpecker/v2/server/model"
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

	receiver := make(LogChan, 10)
	defer close(receiver)

	go func() {
		for range receiver {
			wg.Done()
		}
	}()

	logger := New()
	assert.NoError(t, logger.Open(ctx, testStepID))
	go func() {
		assert.NoError(t, logger.Tail(ctx, testStepID, receiver))
	}()
	go func() {
		assert.NoError(t, logger.Tail(ctx, testStepID, receiver))
	}()

	<-time.After(500 * time.Millisecond)

	wg.Add(4)
	go func() {
		assert.NoError(t, logger.Write(ctx, testStepID, []*model.LogEntry{testEntry}))
		assert.NoError(t, logger.Write(ctx, testStepID, []*model.LogEntry{testEntry}))
	}()

	wg.Wait()

	wg.Add(1)
	go func() {
		assert.NoError(t, logger.Tail(ctx, testStepID, receiver))
	}()

	<-time.After(500 * time.Millisecond)

	wg.Wait()
	cancel(nil)
}
