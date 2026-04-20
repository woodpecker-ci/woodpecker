// Copyright 2026 Woodpecker Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package api

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/mock"

	"go.woodpecker-ci.org/woodpecker/v3/server"
	"go.woodpecker-ci.org/woodpecker/v3/server/logging"
	"go.woodpecker-ci.org/woodpecker/v3/server/model"
	"go.woodpecker-ci.org/woodpecker/v3/server/pubsub"
	"go.woodpecker-ci.org/woodpecker/v3/server/pubsub/memory"
	"go.woodpecker-ci.org/woodpecker/v3/server/scheduler"
	store_mocks "go.woodpecker-ci.org/woodpecker/v3/server/store/mocks"
)

func TestEventStreamSSEConcurrentDisconnect(t *testing.T) {
	gin.SetMode(gin.TestMode)

	for range 50 {
		broker := memory.New()
		server.Config.Services.Scheduler = scheduler.NewScheduler(nil, broker)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		ctx, cancel := context.WithCancelCause(t.Context())
		req, _ := http.NewRequestWithContext(ctx, http.MethodGet, "/stream/events", nil)
		c.Request = req

		topic := map[string]struct{}{pubsub.PublicTopic: {}}

		done := make(chan struct{})
		go func() {
			defer close(done)
			EventStreamSSE(c)
		}()

		// Let the event handler subscribe
		time.Sleep(20 * time.Millisecond)

		// Fire concurrent publishes while canceling the request.
		var wg sync.WaitGroup
		for range 20 {
			wg.Add(1)
			go func() {
				defer wg.Done()
				_ = broker.Publish(ctx, topic, pubsub.Message{
					Data: []byte(`{"pipeline":1}`),
				})
			}()
		}

		// Simulate client disconnect mid-publish.
		cancel(nil)
		wg.Wait()
		<-done
	}
}

func setupLogStreamContext(t *testing.T, logService logging.Log) (*httptest.ResponseRecorder, *gin.Context, context.CancelCauseFunc) {
	t.Helper()

	const stepID int64 = 42
	const pipelineID int64 = 10

	mockStore := store_mocks.NewMockStore(t)
	mockStore.On("GetPipelineNumber", mock.Anything, mock.Anything).
		Return(&model.Pipeline{ID: pipelineID}, nil)
	mockStore.On("StepLoad", mock.Anything).
		Return(&model.Step{
			ID:         stepID,
			PipelineID: pipelineID,
			State:      model.StatusRunning,
		}, nil)

	server.Config.Services.Logs = logService

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	ctx, cancel := context.WithCancelCause(t.Context())
	req, _ := http.NewRequestWithContext(ctx, http.MethodGet, "/stream/logs/1/1/42", nil)
	c.Request = req
	c.Params = gin.Params{
		{Key: "repo_id", Value: "1"},
		{Key: "pipeline", Value: "1"},
		{Key: "step_id", Value: "42"},
	}
	c.Set("repo", &model.Repo{ID: 1, FullName: "owner/repo"})
	c.Set("store", mockStore)

	return w, c, cancel
}

func TestLogStreamSSEConcurrentDisconnect(t *testing.T) {
	gin.SetMode(gin.TestMode)

	const stepID int64 = 42

	for range 50 {
		logService := logging.New()
		_, c, cancel := setupLogStreamContext(t, logService)

		done := make(chan struct{})
		go func() {
			defer close(done)
			LogStreamSSE(c)
		}()

		// Let LogStreamSSE open the stream and start tailing.
		time.Sleep(20 * time.Millisecond)

		// Fire concurrent log writes while canceling the request.
		var wg sync.WaitGroup
		for i := range 20 {
			wg.Add(1)
			go func() {
				defer wg.Done()
				_ = logService.Write(t.Context(), stepID, []*model.LogEntry{
					{Line: i, Data: []byte("log line")},
				})
			}()
		}

		// Simulate client disconnect mid-write.
		cancel(nil)
		wg.Wait()
		<-done
	}
}
