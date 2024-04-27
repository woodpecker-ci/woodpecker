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
