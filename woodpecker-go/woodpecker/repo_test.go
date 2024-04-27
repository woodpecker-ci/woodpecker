package woodpecker

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
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
