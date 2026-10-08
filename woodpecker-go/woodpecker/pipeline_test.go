// Copyright 2026 Woodpecker Authors
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
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestClientWorkflowAgent(t *testing.T) {
	tests := []struct {
		name       string
		handler    http.HandlerFunc
		expected   *AgentSnapshot
		wantStatus int
	}{
		{
			name: "success",
			handler: func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, http.MethodGet, r.Method)
				assert.Equal(t, "/api/repos/1/pipelines/2/workflows/3/agent", r.URL.Path)
				w.WriteHeader(http.StatusOK)
				_, err := fmt.Fprint(w, `{"id":7,"org_id":-1,"name":"builder-07","platform":"linux/amd64","backend":"docker","custom_labels":{"zone":"eu"}}`)
				assert.NoError(t, err)
			},
			expected: &AgentSnapshot{
				ID:           7,
				OrgID:        -1,
				Name:         "builder-07",
				Platform:     "linux/amd64",
				Backend:      "docker",
				CustomLabels: map[string]string{"zone": "eu"},
			},
		},
		{
			name: "not allowed to see the agent",
			handler: func(w http.ResponseWriter, _ *http.Request) {
				w.WriteHeader(http.StatusForbidden)
			},
			wantStatus: http.StatusForbidden,
		},
		{
			name: "no agent recorded",
			handler: func(w http.ResponseWriter, _ *http.Request) {
				w.WriteHeader(http.StatusNotFound)
			},
			wantStatus: http.StatusNotFound,
		},
		{
			name: "invalid response",
			handler: func(w http.ResponseWriter, _ *http.Request) {
				w.WriteHeader(http.StatusOK)
				_, err := fmt.Fprint(w, `invalid json`)
				assert.NoError(t, err)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ts := httptest.NewServer(tt.handler)
			defer ts.Close()

			client := NewClient(ts.URL, http.DefaultClient)
			agent, err := client.WorkflowAgent(1, 2, 3)

			if tt.expected == nil {
				require.Error(t, err)
				if tt.wantStatus != 0 {
					var clientErr *ClientError
					require.True(t, errors.As(err, &clientErr), "callers need the status to tell 403 from 404")
					assert.Equal(t, tt.wantStatus, clientErr.StatusCode)
				}
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.expected, agent)
		})
	}
}
