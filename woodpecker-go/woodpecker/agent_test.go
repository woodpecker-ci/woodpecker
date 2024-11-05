package woodpecker

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestClient_AgentCreate(t *testing.T) {
	tests := []struct {
		name     string
		handler  http.HandlerFunc
		input    *Agent
		expected *Agent
		wantErr  bool
	}{
		{
			name: "success",
			handler: func(w http.ResponseWriter, r *http.Request) {
				if r.Method != http.MethodPost {
					w.WriteHeader(http.StatusMethodNotAllowed)
					return
				}
				w.WriteHeader(http.StatusCreated)
				_, err := fmt.Fprint(w, `{"id":1,"name":"new_agent","backend":"local","capacity":2,"version":"1.0.0"}`)
				assert.NoError(t, err)
			},
			input:    &Agent{Name: "new_agent", Backend: "local", Capacity: 2, Version: "1.0.0"},
			expected: &Agent{ID: 1, Name: "new_agent", Backend: "local", Capacity: 2, Version: "1.0.0"},
			wantErr:  false,
		},
		{
			name: "invalid input",
			handler: func(w http.ResponseWriter, r *http.Request) {
				if r.Method != http.MethodPost {
					w.WriteHeader(http.StatusMethodNotAllowed)
					return
				}
				w.WriteHeader(http.StatusBadRequest)
			},
			input:    &Agent{},
			expected: nil,
			wantErr:  true,
		},
		{
			name: "server error",
			handler: func(w http.ResponseWriter, r *http.Request) {
				if r.Method != http.MethodPost {
					w.WriteHeader(http.StatusMethodNotAllowed)
					return
				}
				w.WriteHeader(http.StatusInternalServerError)
			},
			input:    &Agent{Name: "new_agent", Backend: "local", Capacity: 2, Version: "1.0.0"},
			expected: nil,
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ts := httptest.NewServer(tt.handler)
			defer ts.Close()

			client := NewClient(ts.URL, http.DefaultClient)
			agent, err := client.AgentCreate(tt.input)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)
			assert.Equal(t, agent, tt.expected)
		})
	}
}

func TestClient_AgentList(t *testing.T) {
	tests := []struct {
		name     string
		handler  http.HandlerFunc
		expected []*Agent
		wantErr  bool
	}{
		{
			name: "success",
			handler: func(w http.ResponseWriter, _ *http.Request) {
				w.WriteHeader(http.StatusOK)
				_, err := fmt.Fprint(w, `[
					{
						"id": 1,
						"name": "agent-1",
						"backend": "local",
						"capacity": 2,
						"version": "1.0.0"
					},
					{
						"id": 2,
						"name": "agent-2",
						"backend": "kubernetes",
						"capacity": 4,
						"version": "1.0.0"
					}
				]`)
				assert.NoError(t, err)
			},
			expected: []*Agent{
				{
					ID:       1,
					Name:     "agent-1",
					Backend:  "local",
					Capacity: 2,
					Version:  "1.0.0",
				},
				{
					ID:       2,
					Name:     "agent-2",
					Backend:  "kubernetes",
					Capacity: 4,
					Version:  "1.0.0",
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
			agents, err := client.AgentList()

			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)
			assert.Equal(t, tt.expected, agents)
		})
	}
}

func TestClient_Agent(t *testing.T) {
	tests := []struct {
		name     string
		handler  http.HandlerFunc
		agentID  int64
		expected *Agent
		wantErr  bool
	}{
		{
			name: "success",
			handler: func(w http.ResponseWriter, r *http.Request) {
				if r.Method != http.MethodGet {
					w.WriteHeader(http.StatusMethodNotAllowed)
					return
				}
				w.WriteHeader(http.StatusOK)
				_, err := fmt.Fprint(w, `{"id":1,"name":"agent-1","backend":"local","capacity":2,"version":"1.0.0"}`)
				assert.NoError(t, err)
			},
			agentID:  1,
			expected: &Agent{ID: 1, Name: "agent-1", Backend: "local", Capacity: 2, Version: "1.0.0"},
			wantErr:  false,
		},
		{
			name: "not found",
			handler: func(w http.ResponseWriter, r *http.Request) {
				if r.Method != http.MethodGet {
					w.WriteHeader(http.StatusMethodNotAllowed)
					return
				}
				w.WriteHeader(http.StatusNotFound)
			},
			agentID:  999,
			expected: nil,
			wantErr:  true,
		},
		{
			name: "server error",
			handler: func(w http.ResponseWriter, r *http.Request) {
				if r.Method != http.MethodGet {
					w.WriteHeader(http.StatusMethodNotAllowed)
					return
				}
				w.WriteHeader(http.StatusInternalServerError)
			},
			agentID:  1,
			expected: nil,
			wantErr:  true,
		},
		{
			name: "invalid response",
			handler: func(w http.ResponseWriter, r *http.Request) {
				if r.Method != http.MethodGet {
					w.WriteHeader(http.StatusMethodNotAllowed)
					return
				}
				w.WriteHeader(http.StatusOK)
				_, err := fmt.Fprint(w, `invalid json`)
				assert.NoError(t, err)
			},
			agentID:  1,
			expected: nil,
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ts := httptest.NewServer(tt.handler)
			defer ts.Close()

			client := NewClient(ts.URL, http.DefaultClient)
			agent, err := client.Agent(tt.agentID)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)
			assert.Equal(t, tt.expected, agent)
		})
	}
}

func TestClient_AgentUpdate(t *testing.T) {
	tests := []struct {
		name     string
		handler  http.HandlerFunc
		input    *Agent
		expected *Agent
		wantErr  bool
	}{
		{
			name: "success",
			handler: func(w http.ResponseWriter, r *http.Request) {
				if r.Method != http.MethodPatch {
					w.WriteHeader(http.StatusMethodNotAllowed)
					return
				}
				w.WriteHeader(http.StatusOK)
				_, err := fmt.Fprint(w, `{"id":1,"name":"updated_agent"}`)
				assert.NoError(t, err)
			},
			input:    &Agent{ID: 1, Name: "existing_agent"},
			expected: &Agent{ID: 1, Name: "updated_agent"},
			wantErr:  false,
		},
		{
			name: "not found",
			handler: func(w http.ResponseWriter, r *http.Request) {
				if r.Method != http.MethodPatch {
					w.WriteHeader(http.StatusMethodNotAllowed)
					return
				}
				w.WriteHeader(http.StatusNotFound)
			},
			input:    &Agent{ID: 999, Name: "nonexistent_agent"},
			expected: nil,
			wantErr:  true,
		},
		{
			name: "invalid input",
			handler: func(w http.ResponseWriter, r *http.Request) {
				if r.Method != http.MethodPatch {
					w.WriteHeader(http.StatusMethodNotAllowed)
					return
				}
				w.WriteHeader(http.StatusBadRequest)
			},
			input:    &Agent{},
			expected: nil,
			wantErr:  true,
		},
		{
			name: "server error",
			handler: func(w http.ResponseWriter, r *http.Request) {
				if r.Method != http.MethodPatch {
					w.WriteHeader(http.StatusMethodNotAllowed)
					return
				}
				w.WriteHeader(http.StatusInternalServerError)
			},
			input:    &Agent{ID: 1, Name: "existing_agent"},
			expected: nil,
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ts := httptest.NewServer(tt.handler)
			defer ts.Close()

			client := NewClient(ts.URL, http.DefaultClient)
			agent, err := client.AgentUpdate(tt.input)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)
			assert.Equal(t, agent, tt.expected)
		})
	}
}

func TestClient_AgentDelete(t *testing.T) {
	tests := []struct {
		name    string
		handler http.HandlerFunc
		agentID int64
		wantErr bool
	}{
		{
			name: "success",
			handler: func(w http.ResponseWriter, r *http.Request) {
				if r.Method != http.MethodDelete {
					w.WriteHeader(http.StatusMethodNotAllowed)
					return
				}
				w.WriteHeader(http.StatusOK)
			},
			agentID: 1,
			wantErr: false,
		},
		{
			name: "not found",
			handler: func(w http.ResponseWriter, r *http.Request) {
				if r.Method != http.MethodDelete {
					w.WriteHeader(http.StatusMethodNotAllowed)
					return
				}
				w.WriteHeader(http.StatusNotFound)
			},
			agentID: 999,
			wantErr: true,
		},
		{
			name: "server error",
			handler: func(w http.ResponseWriter, r *http.Request) {
				if r.Method != http.MethodDelete {
					w.WriteHeader(http.StatusMethodNotAllowed)
					return
				}
				w.WriteHeader(http.StatusInternalServerError)
			},
			agentID: 1,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ts := httptest.NewServer(tt.handler)
			defer ts.Close()

			client := NewClient(ts.URL, http.DefaultClient)
			err := client.AgentDelete(tt.agentID)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)
		})
	}
}

func TestClient_AgentTasksList(t *testing.T) {
	tests := []struct {
		name     string
		handler  http.HandlerFunc
		agentID  int64
		expected []*Task
		wantErr  bool
	}{
		{
			name: "success",
			handler: func(w http.ResponseWriter, r *http.Request) {
				if r.Method != http.MethodGet {
					w.WriteHeader(http.StatusMethodNotAllowed)
					return
				}
				w.WriteHeader(http.StatusOK)
				_, err := fmt.Fprint(w, `[
					{
						"id": "4696",
						"data": "",
						"labels": {
							"platform": "linux/amd64",
							"repo": "woodpecker-ci/woodpecker"
						}
					},
					{
						"id": "4697",
						"data": "",
						"labels": {
							"platform": "linux/arm64",
							"repo": "woodpecker-ci/woodpecker"
						}
					}
				]`)
				assert.NoError(t, err)
			},
			agentID: 1,
			expected: []*Task{
				{
					ID: "4696",
					Labels: map[string]string{
						"platform": "linux/amd64",
						"repo":     "woodpecker-ci/woodpecker",
					},
				},
				{
					ID: "4697",
					Labels: map[string]string{
						"platform": "linux/arm64",
						"repo":     "woodpecker-ci/woodpecker",
					},
				},
			},
			wantErr: false,
		},
		{
			name: "not found",
			handler: func(w http.ResponseWriter, r *http.Request) {
				if r.Method != http.MethodGet {
					w.WriteHeader(http.StatusMethodNotAllowed)
					return
				}
				w.WriteHeader(http.StatusNotFound)
			},
			agentID:  999,
			expected: nil,
			wantErr:  true,
		},
		{
			name: "server error",
			handler: func(w http.ResponseWriter, r *http.Request) {
				if r.Method != http.MethodGet {
					w.WriteHeader(http.StatusMethodNotAllowed)
					return
				}
				w.WriteHeader(http.StatusInternalServerError)
			},
			agentID:  1,
			expected: nil,
			wantErr:  true,
		},
		{
			name: "invalid response",
			handler: func(w http.ResponseWriter, r *http.Request) {
				if r.Method != http.MethodGet {
					w.WriteHeader(http.StatusMethodNotAllowed)
					return
				}
				w.WriteHeader(http.StatusOK)
				_, err := fmt.Fprint(w, `invalid json`)
				assert.NoError(t, err)
			},
			agentID:  1,
			expected: nil,
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ts := httptest.NewServer(tt.handler)
			defer ts.Close()

			client := NewClient(ts.URL, http.DefaultClient)
			tasks, err := client.AgentTasksList(tt.agentID)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)
			assert.Equal(t, tt.expected, tasks)
		})
	}
}
