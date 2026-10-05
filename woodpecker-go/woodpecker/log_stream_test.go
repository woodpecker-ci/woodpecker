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

package woodpecker

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestStepLogStream(t *testing.T) {
	stream := func(t *testing.T, events string) Client {
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "/api/stream/logs/1/2/3", r.URL.Path)
			w.Header().Set("Content-Type", "text/event-stream")
			fmt.Fprint(w, events)
			_ = http.NewResponseController(w).Flush()
			// a real stream stays open until the client leaves
			if events == "" {
				<-r.Context().Done()
			}
		}))
		t.Cleanup(ts.Close)
		return NewClient(ts.URL, http.DefaultClient)
	}

	collect := func(entries *[]*LogEntry) func(*LogEntry) {
		return func(entry *LogEntry) { *entries = append(*entries, entry) }
	}

	t.Run("entries until end of log", func(t *testing.T) {
		client := stream(t, ": ping\n\n"+
			"id: 1\ndata: {\"step_id\":3,\"line\":0,\"time\":1,\"data\":\"aGVsbG8=\"}\n\n"+
			": ping\n\n"+
			"id: 2\ndata: {\"step_id\":3,\"line\":1,\"time\":2,\"data\":\"d29ybGQ=\"}\n\n"+
			"event: eof\ndata: eof\n\n"+
			"data: {\"line\":99}\n\n")

		var entries []*LogEntry
		require.NoError(t, client.StepLogStream(t.Context(), 1, 2, 3, collect(&entries)))

		require.Len(t, entries, 2, "nothing is read after the end of the log")
		assert.Equal(t, &LogEntry{StepID: 3, Line: 0, Time: 1, Data: []byte("hello")}, entries[0])
		assert.Equal(t, &LogEntry{StepID: 3, Line: 1, Time: 2, Data: []byte("world")}, entries[1])
	})

	t.Run("error reported by server", func(t *testing.T) {
		client := stream(t, ": ping\n\nevent: error\ndata: step not running (anymore)\n\n")

		var entries []*LogEntry
		err := client.StepLogStream(t.Context(), 1, 2, 3, collect(&entries))
		require.EqualError(t, err, "step not running (anymore)")
		assert.Empty(t, entries)
	})

	t.Run("an event name is only valid for one event", func(t *testing.T) {
		client := stream(t, "event: other\ndata: ignored\n\ndata: {\"line\":5}\n\nevent: eof\ndata: eof\n\n")

		var entries []*LogEntry
		require.NoError(t, client.StepLogStream(t.Context(), 1, 2, 3, collect(&entries)))
		require.Len(t, entries, 1)
		assert.Equal(t, 5, entries[0].Line)
	})

	t.Run("stream cut off", func(t *testing.T) {
		client := stream(t, "data: {\"line\":0}\n\n")

		var entries []*LogEntry
		err := client.StepLogStream(t.Context(), 1, 2, 3, collect(&entries))
		require.Error(t, err, "a log that did not end is incomplete")
		assert.Len(t, entries, 1)
	})

	t.Run("http error", func(t *testing.T) {
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			http.Error(w, "no such repo", http.StatusNotFound)
		}))
		defer ts.Close()

		err := NewClient(ts.URL, http.DefaultClient).StepLogStream(t.Context(), 1, 2, 3, func(*LogEntry) {})
		var clientErr *ClientError
		require.ErrorAs(t, err, &clientErr)
		assert.Equal(t, http.StatusNotFound, clientErr.StatusCode)
	})

	t.Run("canceled", func(t *testing.T) {
		client := stream(t, "")

		ctx, cancel := context.WithCancelCause(t.Context())
		cancel(nil)
		require.ErrorIs(t, client.StepLogStream(ctx, 1, 2, 3, func(*LogEntry) {}), context.Canceled)
	})
}
