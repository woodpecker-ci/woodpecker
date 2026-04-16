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
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/gin-gonic/gin"

	"go.woodpecker-ci.org/woodpecker/v3/server"
	"go.woodpecker-ci.org/woodpecker/v3/server/pubsub"
	"go.woodpecker-ci.org/woodpecker/v3/server/pubsub/memory"
	"go.woodpecker-ci.org/woodpecker/v3/server/scheduler"
)

func TestEventStreamSSE_ConcurrentDisconnect(t *testing.T) {
	gin.SetMode(gin.TestMode)

	for range 50 {
		broker := memory.New()
		server.Config.Services.Scheduler = scheduler.NewScheduler(nil, broker)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		ctx, cancel := context.WithCancel(t.Context())
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

		// Fire concurrent publishes while cancelling the request.
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
		cancel()
		wg.Wait()
		<-done
	}
}