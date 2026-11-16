// Copyright 2024 Woodpecker Authors
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

package woodpecker

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPipelineList(t *testing.T) {
	tests := []struct {
		name           string
		handler        http.HandlerFunc
		opts           PipelineListOptions
		wantErr        bool
		expectedLength int
		expectedIDs    []int64
	}{
		{
			name: "success",
			handler: func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, http.MethodGet, r.Method)
				assert.Equal(t, "/api/repos/123/pipelines?after=2023-01-15T00%3A00%3A00Z&before=2023-01-16T00%3A00%3A00Z&page=2&perPage=10", r.URL.RequestURI())

				w.WriteHeader(http.StatusOK)
				_, err := fmt.Fprint(w, `[{"id":1},{"id":2}]`)
				assert.NoError(t, err)
			},
			opts: PipelineListOptions{
				ListOptions: ListOptions{
					Page:    2,
					PerPage: 10,
				},
				Before: time.Date(2023, 1, 16, 0, 0, 0, 0, time.UTC),
				After:  time.Date(2023, 1, 15, 0, 0, 0, 0, time.UTC),
			},
			expectedLength: 2,
			expectedIDs:    []int64{1, 2},
		},
		{
			name: "empty ListOptions",
			handler: func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, http.MethodGet, r.Method)
				assert.Equal(t, "/api/repos/123/pipelines", r.URL.RequestURI())

				w.WriteHeader(http.StatusOK)
				_, err := fmt.Fprint(w, `[{"id":1},{"id":2}]`)
				assert.NoError(t, err)
			},
			opts:           PipelineListOptions{},
			expectedLength: 2,
			expectedIDs:    []int64{1, 2},
		},
		{
			name: "server error",
			handler: func(w http.ResponseWriter, _ *http.Request) {
				w.WriteHeader(http.StatusInternalServerError)
			},
			opts:    PipelineListOptions{},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ts := httptest.NewServer(tt.handler)
			defer ts.Close()

			client := NewClient(ts.URL, http.DefaultClient)

			pipelines, err := client.PipelineList(123, tt.opts)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, pipelines)
				return
			}

			assert.NoError(t, err)
			assert.Len(t, pipelines, tt.expectedLength)
			for i, id := range tt.expectedIDs {
				assert.Equal(t, id, pipelines[i].ID)
			}
		})
	}
}

func TestClientDeploy(t *testing.T) {
	tests := []struct {
		name             string
		handler          http.HandlerFunc
		repoID           int64
		pipelineID       int64
		opts             DeployOptions
		wantErr          bool
		expectedPipeline *Pipeline
	}{
		{
			name: "success",
			handler: func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, http.MethodPost, r.Method)
				assert.Equal(t, "/api/repos/123/pipelines/456?event=deployment", r.URL.RequestURI())

				w.WriteHeader(http.StatusOK)
				_, err := fmt.Fprint(w, `{"id":789}`)
				assert.NoError(t, err)
			},
			repoID:     123,
			pipelineID: 456,
			opts:       DeployOptions{},
			expectedPipeline: &Pipeline{
				ID: 789,
			},
		},
		{
			name: "error",
			handler: func(w http.ResponseWriter, _ *http.Request) {
				w.WriteHeader(http.StatusInternalServerError)
			},
			repoID:     123,
			pipelineID: 456,
			opts:       DeployOptions{},
			wantErr:    true,
		},
		{
			name: "with options",
			handler: func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, http.MethodPost, r.Method)
				assert.Equal(t, "/api/repos/123/pipelines/456?deploy_to=production&event=deployment", r.URL.RequestURI())

				w.WriteHeader(http.StatusOK)
				_, err := fmt.Fprint(w, `{"id":789}`)
				assert.NoError(t, err)
			},
			repoID:     123,
			pipelineID: 456,
			opts: DeployOptions{
				DeployTo: "production",
			},
			expectedPipeline: &Pipeline{
				ID: 789,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ts := httptest.NewServer(tt.handler)
			defer ts.Close()

			client := NewClient(ts.URL, http.DefaultClient)

			pipeline, err := client.Deploy(tt.repoID, tt.pipelineID, tt.opts)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)
			assert.Equal(t, tt.expectedPipeline, pipeline)
		})
	}
}

func TestClientPipelineStart(t *testing.T) {
	tests := []struct {
		name             string
		handler          http.HandlerFunc
		repoID           int64
		pipelineID       int64
		opts             PipelineStartOptions
		wantErr          bool
		expectedPipeline *Pipeline
	}{
		{
			name: "success",
			handler: func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, http.MethodPost, r.Method)
				assert.Equal(t, "/api/repos/123/pipelines/456", r.URL.RequestURI())

				w.WriteHeader(http.StatusOK)
				_, err := fmt.Fprint(w, `{"id":789}`)
				assert.NoError(t, err)
			},
			repoID:     123,
			pipelineID: 456,
			opts:       PipelineStartOptions{},
			expectedPipeline: &Pipeline{
				ID: 789,
			},
		},
		{
			name: "error",
			handler: func(w http.ResponseWriter, _ *http.Request) {
				w.WriteHeader(http.StatusInternalServerError)
			},
			repoID:     123,
			pipelineID: 456,
			opts:       PipelineStartOptions{},
			wantErr:    true,
		},
		{
			name: "with options",
			handler: func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, http.MethodPost, r.Method)
				assert.Equal(t, "/api/repos/123/pipelines/456?foo=bar", r.URL.RequestURI())

				w.WriteHeader(http.StatusOK)
				_, err := fmt.Fprint(w, `{"id":789}`)
				assert.NoError(t, err)
			},
			repoID:     123,
			pipelineID: 456,
			opts: PipelineStartOptions{
				Params: map[string]string{"foo": "bar"},
			},
			expectedPipeline: &Pipeline{
				ID: 789,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ts := httptest.NewServer(tt.handler)
			defer ts.Close()

			client := NewClient(ts.URL, http.DefaultClient)

			pipeline, err := client.PipelineStart(tt.repoID, tt.pipelineID, tt.opts)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)
			assert.Equal(t, tt.expectedPipeline, pipeline)
		})
	}
}

func TestClient_PipelineLast(t *testing.T) {
	tests := []struct {
		name     string
		handler  http.HandlerFunc
		repoID   int64
		opts     PipelineLastOptions
		expected *Pipeline
		wantErr  bool
	}{
		{
			name: "success",
			handler: func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, "/api/repos/1/pipelines/latest?branch=main", r.URL.Path+"?"+r.URL.RawQuery)
				w.WriteHeader(http.StatusOK)
				_, err := fmt.Fprint(w, `{"id":1,"number":1,"status":"success","event":"push","branch":"main"}`)
				assert.NoError(t, err)
			},
			repoID: 1,
			opts:   PipelineLastOptions{Branch: "main"},
			expected: &Pipeline{
				ID:     1,
				Number: 1,
				Status: "success",
				Event:  "push",
				Branch: "main",
			},
			wantErr: false,
		},
		{
			name: "server error",
			handler: func(w http.ResponseWriter, _ *http.Request) {
				w.WriteHeader(http.StatusInternalServerError)
			},
			repoID:   1,
			opts:     PipelineLastOptions{},
			expected: nil,
			wantErr:  true,
		},
		{
			name: "invalid response",
			handler: func(w http.ResponseWriter, _ *http.Request) {
				w.WriteHeader(http.StatusOK)
				_, err := fmt.Fprint(w, `invalid json`)
				assert.NoError(t, err)
			},
			repoID:   1,
			opts:     PipelineLastOptions{},
			expected: nil,
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ts := httptest.NewServer(tt.handler)
			defer ts.Close()

			client := NewClient(ts.URL, http.DefaultClient)
			pipeline, err := client.PipelineLast(tt.repoID, tt.opts)

			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)
			assert.Equal(t, tt.expected, pipeline)
		})
	}
}

func TestClientRepoPost(t *testing.T) {
	tests := []struct {
		name     string
		handler  http.HandlerFunc
		opts     RepoPostOptions
		expected *Repo
		wantErr  bool
	}{
		{
			name: "success",
			handler: func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, http.MethodPost, r.Method)
				assert.Equal(t, "/api/repos?forge_remote_id=10", r.URL.RequestURI())

				w.WriteHeader(http.StatusOK)
				_, err := fmt.Fprint(w, `{"id":1,"name":"test","owner":"owner","full_name":"owner/test","forge_remote_id":"10"}`)
				assert.NoError(t, err)
			},
			opts: RepoPostOptions{
				ForgeRemoteID: 10,
			},
			expected: &Repo{
				ID:            1,
				ForgeRemoteID: "10",
				Name:          "test",
				Owner:         "owner",
				FullName:      "owner/test",
			},
			wantErr: false,
		},
		{
			name: "server error",
			handler: func(w http.ResponseWriter, _ *http.Request) {
				w.WriteHeader(http.StatusInternalServerError)
			},
			opts:     RepoPostOptions{},
			expected: nil,
			wantErr:  true,
		},
		{
			name: "invalid response",
			handler: func(w http.ResponseWriter, _ *http.Request) {
				w.WriteHeader(http.StatusOK)
				_, err := fmt.Fprint(w, `invalid json`)
				assert.NoError(t, err)
			},
			opts:     RepoPostOptions{},
			expected: nil,
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ts := httptest.NewServer(tt.handler)
			defer ts.Close()

			client := NewClient(ts.URL, http.DefaultClient)
			repo, err := client.RepoPost(tt.opts)

			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)
			assert.Equal(t, tt.expected, repo)
		})
	}
}

func TestClientRepoMove(t *testing.T) {
	tests := []struct {
		name    string
		handler http.HandlerFunc
		repoID  int64
		opts    RepoMoveOptions
		wantErr bool
	}{
		{
			name: "success",
			handler: func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, http.MethodPost, r.Method)
				assert.Equal(t, "/api/repos/123/move?to=new_owner", r.URL.RequestURI())
				w.WriteHeader(http.StatusOK)
			},
			repoID: 123,
			opts: RepoMoveOptions{
				To: "new_owner",
			},
			wantErr: false,
		},
		{
			name: "server error",
			handler: func(w http.ResponseWriter, _ *http.Request) {
				w.WriteHeader(http.StatusInternalServerError)
			},
			repoID:  123,
			opts:    RepoMoveOptions{},
			wantErr: true,
		},
		{
			name: "invalid options",
			handler: func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, http.MethodPost, r.Method)
				w.WriteHeader(http.StatusBadRequest)
			},
			repoID:  123,
			opts:    RepoMoveOptions{},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ts := httptest.NewServer(tt.handler)
			defer ts.Close()

			client := NewClient(ts.URL, http.DefaultClient)
			err := client.RepoMove(tt.repoID, tt.opts)

			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)
		})
	}
}

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
