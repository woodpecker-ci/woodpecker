package woodpecker

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestClient_UserList(t *testing.T) {
	tests := []struct {
		name     string
		handler  http.HandlerFunc
		expected []*User
		wantErr  bool
	}{
		{
			name: "success",
			handler: func(w http.ResponseWriter, _ *http.Request) {
				w.WriteHeader(http.StatusOK)
				_, err := fmt.Fprint(w, `[{"id":1,"login":"user1"},{"id":2,"login":"user2"}]`)
				assert.NoError(t, err)
			},
			expected: []*User{{ID: 1, Login: "user1"}, {ID: 2, Login: "user2"}},
			wantErr:  false,
		},
		{
			name: "empty response",
			handler: func(w http.ResponseWriter, _ *http.Request) {
				w.WriteHeader(http.StatusOK)
				_, err := fmt.Fprint(w, `[]`)
				assert.NoError(t, err)
			},
			expected: []*User{},
			wantErr:  false,
		},
		{
			name: "server error",
			handler: func(w http.ResponseWriter, _ *http.Request) {
				w.WriteHeader(http.StatusInternalServerError)
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
			users, err := client.UserList()

			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)
			assert.Equal(t, users, tt.expected)
		})
	}
}

func TestClient_UserPost(t *testing.T) {
	tests := []struct {
		name     string
		handler  http.HandlerFunc
		input    *User
		expected *User
		wantErr  bool
	}{
		{
			name: "success",
			handler: func(w http.ResponseWriter, _ *http.Request) {
				w.WriteHeader(http.StatusCreated)
				_, err := fmt.Fprint(w, `{"id":1,"login":"new_user"}`)
				assert.NoError(t, err)
			},
			input:    &User{Login: "new_user"},
			expected: &User{ID: 1, Login: "new_user"},
			wantErr:  false,
		},
		{
			name: "invalid input",
			handler: func(w http.ResponseWriter, _ *http.Request) {
				w.WriteHeader(http.StatusBadRequest)
			},
			input:    &User{},
			expected: nil,
			wantErr:  true,
		},
		{
			name: "server error",
			handler: func(w http.ResponseWriter, _ *http.Request) {
				w.WriteHeader(http.StatusInternalServerError)
			},
			input:    &User{Login: "new_user"},
			expected: nil,
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ts := httptest.NewServer(tt.handler)
			defer ts.Close()

			client := NewClient(ts.URL, http.DefaultClient)
			user, err := client.UserPost(tt.input)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)
			assert.Equal(t, user, tt.expected)
		})
	}
}

func TestClient_UserPatch(t *testing.T) {
	tests := []struct {
		name     string
		handler  http.HandlerFunc
		input    *User
		expected *User
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
				_, err := fmt.Fprint(w, `{"id":1,"login":"updated_user"}`)
				assert.NoError(t, err)
			},
			input:    &User{ID: 1, Login: "existing_user"},
			expected: &User{ID: 1, Login: "updated_user"},
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
			input:    &User{ID: 999, Login: "nonexistent_user"},
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
			input:    &User{},
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
			input:    &User{ID: 1, Login: "existing_user"},
			expected: nil,
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ts := httptest.NewServer(tt.handler)
			defer ts.Close()

			client := NewClient(ts.URL, http.DefaultClient)
			user, err := client.UserPatch(tt.input)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)
			assert.Equal(t, user, tt.expected)
		})
	}
}

func TestClient_UserDel(t *testing.T) {
	tests := []struct {
		name    string
		handler http.HandlerFunc
		login   string
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
			login:   "existing_user",
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
			login:   "nonexistent_user",
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
			login:   "existing_user",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ts := httptest.NewServer(tt.handler)
			defer ts.Close()

			client := NewClient(ts.URL, http.DefaultClient)
			err := client.UserDel(tt.login)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)
		})
	}
}

func TestClient_RepoList(t *testing.T) {
	tests := []struct {
		name     string
		handler  http.HandlerFunc
		opt      RepoListOptions
		expected []*Repo
		wantErr  bool
	}{
		{
			name: "success",
			handler: func(w http.ResponseWriter, _ *http.Request) {
				w.WriteHeader(http.StatusOK)
				_, err := fmt.Fprint(w, `[{"id":1,"name":"repo1"},{"id":2,"name":"repo2"}]`)
				assert.NoError(t, err)
			},
			opt:      RepoListOptions{},
			expected: []*Repo{{ID: 1, Name: "repo1"}, {ID: 2, Name: "repo2"}},
			wantErr:  false,
		},
		{
			name: "empty response",
			handler: func(w http.ResponseWriter, _ *http.Request) {
				w.WriteHeader(http.StatusOK)
				_, err := fmt.Fprint(w, `[]`)
				assert.NoError(t, err)
			},
			opt:      RepoListOptions{},
			expected: []*Repo{},
			wantErr:  false,
		},
		{
			name: "server error",
			handler: func(w http.ResponseWriter, _ *http.Request) {
				w.WriteHeader(http.StatusInternalServerError)
			},
			opt:      RepoListOptions{},
			expected: nil,
			wantErr:  true,
		},
		{
			name: "with options",
			handler: func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, "/api/user/repos?all=true", r.URL.RequestURI())
				w.WriteHeader(http.StatusOK)
				_, err := fmt.Fprint(w, `[]`)
				assert.NoError(t, err)
			},
			opt:      RepoListOptions{All: true},
			expected: []*Repo{},
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ts := httptest.NewServer(tt.handler)
			defer ts.Close()

			client := NewClient(ts.URL, http.DefaultClient)
			repos, err := client.RepoList(tt.opt)

			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)
			assert.Equal(t, tt.expected, repos)
		})
	}
}
