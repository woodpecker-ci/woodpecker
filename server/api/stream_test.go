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
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

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
	broker := memory.New()
	server.Config.Services.Scheduler = scheduler.NewScheduler(t.Context(), nil, nil, broker)
	t.Cleanup(func() { server.Config.Services.Scheduler = nil })

	for i := range 50 {
		t.Run(fmt.Sprint(i), func(t *testing.T) {
			t.Parallel()
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
				wg.Go(func() {
					_ = broker.Publish(ctx, topic, pubsub.Message{
						Data: []byte(`{"pipeline":1}`),
					})
				})
			}

			// Simulate client disconnect mid-publish.
			cancel(nil)
			wg.Wait()
			<-done
		})
	}
}

func setupLogStreamContext(t *testing.T) (*httptest.ResponseRecorder, *gin.Context, context.CancelCauseFunc) {
	t.Helper()

	const stepID int64 = 42
	const pipelineID int64 = 10

	mockStore := store_mocks.NewMockStore(t)
	mockStore.On("GetPipelineNumber", mock.Anything, mock.Anything).
		Return(&model.Pipeline{ID: pipelineID}, nil)
	mockStore.On("StepLoad", mock.Anything, mock.Anything).
		Return(&model.Step{
			ID:         stepID,
			PipelineID: pipelineID,
			State:      model.StatusRunning,
		}, nil)

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

	logService := logging.New()
	server.Config.Services.Logs = logService
	t.Cleanup(func() { server.Config.Services.Logs = nil })

	const stepID int64 = 42

	for i := range 50 {
		t.Run(fmt.Sprint(i), func(t *testing.T) {
			t.Parallel()
			done := make(chan struct{})

			_, c, cancel := setupLogStreamContext(t)

			go func() {
				defer close(done)
				LogStreamSSE(c)
			}()

			// Let LogStreamSSE open the stream and start tailing.
			time.Sleep(20 * time.Millisecond)

			// Fire concurrent log writes while canceling the request.
			var wg sync.WaitGroup
			for i := range 20 {
				wg.Go(func() {
					_ = logService.Write(t.Context(), stepID, []*model.LogEntry{
						{Line: i, Data: []byte("log line")},
					})
				})
			}

			// Simulate client disconnect mid-write.
			cancel(nil)
			wg.Wait()
			<-done
		})
	}
}

// The handlers must not leave a goroutine behind that still reads
// server.Config after the handler returned: the t.Cleanup of the stress tests
// above resets the service, and the race detector reports the unsynchronized
// read/write pair (issue #6766). Swapping the service around every request
// reproduces that reliably.
func TestEventStreamSSEDisconnectedClient(t *testing.T) {
	gin.SetMode(gin.TestMode)

	for range 50 {
		server.Config.Services.Scheduler = scheduler.NewScheduler(t.Context(), nil, nil, memory.New())

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		ctx, cancel := context.WithCancelCause(t.Context())
		req, _ := http.NewRequestWithContext(ctx, http.MethodGet, "/stream/events", nil)
		c.Request = req

		// The client is already gone when the handler starts.
		cancel(nil)
		EventStreamSSE(c)
		assert.Contains(t, w.Body.String(), ": ping\n\n")

		server.Config.Services.Scheduler = nil
	}
}

// A client without login must only get events of public projects,
// no matter if the forge repo is public or private.
func TestEventStreamSSEWithoutLogin(t *testing.T) {
	gin.SetMode(gin.TestMode)
	sched := scheduler.NewScheduler(t.Context(), nil, nil, memory.New())
	server.Config.Services.Scheduler = sched
	t.Cleanup(func() { server.Config.Services.Scheduler = nil })

	e := gin.New()
	e.GET("/stream/events", EventStreamSSE)
	srv := httptest.NewServer(e)
	t.Cleanup(srv.Close)

	req, err := http.NewRequestWithContext(t.Context(), http.MethodGet, srv.URL+"/stream/events", nil)
	require.NoError(t, err)
	resp, err := srv.Client().Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	require.Equal(t, http.StatusOK, resp.StatusCode)

	repoIDs := make(chan int64)
	go func(body io.Reader) {
		defer close(repoIDs)
		scanner := bufio.NewScanner(body)
		for scanner.Scan() {
			data, ok := strings.CutPrefix(scanner.Text(), "data: ")
			if !ok {
				continue
			}
			var event model.Event
			if json.Unmarshal([]byte(data), &event) == nil {
				repoIDs <- event.Repo.ID
			}
		}
	}(resp.Body)

	publish := func(id int64, visibility model.RepoVisibility, scmPrivate bool) {
		repo := &model.Repo{ID: id, Visibility: visibility, IsSCMPrivate: scmPrivate}
		require.NoError(t, sched.PublishPipelineEvent(t.Context(), repo, &model.Pipeline{RepoID: id}))
	}

	// The stream subscribes after the first ping, so publish until an event arrives.
	ticker := time.NewTicker(10 * time.Millisecond)
	defer ticker.Stop()
	timeout := time.After(5 * time.Second)
	for subscribed := false; !subscribed; {
		select {
		case <-repoIDs:
			subscribed = true
		case <-ticker.C:
			publish(1, model.VisibilityPublic, false)
		case <-timeout:
			t.Fatal("stream did not subscribe")
		}
	}

	publish(2, model.VisibilityPublic, true)
	publish(3, model.VisibilityInternal, false)
	publish(4, model.VisibilityInternal, true)
	publish(5, model.VisibilityPrivate, false)
	publish(6, model.VisibilityPrivate, true)

	got := make(map[int64]bool)
	for done := false; !done; {
		select {
		case id := <-repoIDs:
			got[id] = true
		case <-time.After(300 * time.Millisecond):
			done = true
		}
	}
	delete(got, 1) // late events of the subscribe loop
	assert.Equal(t, map[int64]bool{2: true}, got)
}

func TestLogStreamSSEDisconnectedClient(t *testing.T) {
	gin.SetMode(gin.TestMode)

	for range 50 {
		server.Config.Services.Logs = logging.New()

		w, c, cancel := setupLogStreamContext(t)

		// The client is already gone when the handler starts.
		cancel(nil)
		LogStreamSSE(c)
		assert.Contains(t, w.Body.String(), ": ping\n\n")

		server.Config.Services.Logs = nil
	}
}
