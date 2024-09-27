package woodpecker

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestClient_QueueInfo(t *testing.T) {
	tests := []struct {
		name     string
		handler  http.HandlerFunc
		expected *Info
		wantErr  bool
	}{
		{
			name: "success",
			handler: func(w http.ResponseWriter, _ *http.Request) {
				w.WriteHeader(http.StatusOK)
				_, err := fmt.Fprint(w, `{
					"pending": null,
					"running": [
							{
									"id": "4696",
									"data": "",
									"labels": {
											"platform": "linux/amd64",
											"repo": "woodpecker-ci/woodpecker"
									},
									"Dependencies": [],
									"DepStatus": {},
									"RunOn": null
							}
					],
					"stats": {
						"worker_count": 2,
						"pending_count": 0,
						"waiting_on_deps_count": 0,
						"running_count": 0,
						"completed_count": 0
					},
					"Paused": false
				}`)
				assert.NoError(t, err)
			},
			expected: &Info{
				Running: []Task{
					{
						ID: "4696",
						Labels: map[string]string{
							"platform": "linux/amd64",
							"repo":     "woodpecker-ci/woodpecker",
						},
						Dependencies: []string{},
						DepStatus:    nil,
						RunOn:        nil,
					},
				},
				Stats: struct {
					Workers       int `json:"worker_count"`
					Pending       int `json:"pending_count"`
					WaitingOnDeps int `json:"waiting_on_deps_count"`
					Running       int `json:"running_count"`
					Complete      int `json:"completed_count"`
				}{
					Workers:       2,
					Pending:       0,
					WaitingOnDeps: 0,
					Running:       0,
					Complete:      0,
				},
			},
			wantErr: false,
		},
		{
			name: "server error",
			handler: func(w http.ResponseWriter, _ *http.Request) {
				w.WriteHeader(http.StatusInternalServerError)
			},
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
			expected: nil,
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ts := httptest.NewServer(tt.handler)
			defer ts.Close()

			client := NewClient(ts.URL, http.DefaultClient)
			info, err := client.QueueInfo()

			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)
			assert.Equal(t, tt.expected, info)
		})
	}
}
