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
		fixtureHandler http.HandlerFunc
		opts           PipelineListOptions
		wantErr        bool
		expectedLength int
		expectedIDs    []int64
	}{
		{
			name: "success",
			fixtureHandler: func(w http.ResponseWriter, r *http.Request) {
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
			fixtureHandler: func(w http.ResponseWriter, r *http.Request) {
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
			name: "error",
			fixtureHandler: func(w http.ResponseWriter, _ *http.Request) {
				w.WriteHeader(http.StatusInternalServerError)
			},
			opts:    PipelineListOptions{},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ts := httptest.NewServer(tt.fixtureHandler)
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
