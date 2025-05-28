// Copyright 2018 Drone.IO Inc.
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

package fixtures

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

// NewServer setup a mock server for testing purposes.
func NewServer(t *testing.T) *httptest.Server {
	mux := http.NewServeMux()
	server := httptest.NewServer(mux)

	// handle requests and serve mock data
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		t.Logf("gitlab forge mock server: [%s] %s", r.Method, r.URL.Path)
		// evaluate the path to serve a dummy data file

		// TODO: find source of "/api/v4/" requests
		// assert.EqualValues(t, "go-gitlab", r.Header.Get("user-agent"), "on request: "+r.URL.Path)

		switch r.URL.Path {
		case "/api/v4/projects":
			if r.FormValue("archived") == "false" {
				w.Write(notArchivedProjectsPayload)
			} else {
				w.Write(allProjectsPayload)
			}

			return
		case "/api/v4/projects/diaspora/diaspora-client":
			w.Write(project4Payload)
			return
		case "/api/v4/projects/brightbox/puppet":
		case "/api/v4/projects/6":
			w.Write(project6Payload)
			return
		case "/api/v4/projects/4/hooks":
			switch r.Method {
			case http.MethodGet:
				w.Write(project4PayloadHooks)
			case http.MethodPost:
				w.Write(project4PayloadHook)
				w.WriteHeader(201)
			}
			return
		case "/api/v4/projects/4/hooks/10717088":
			w.WriteHeader(201)
			return
		case "/api/v4/projects/4/members/all/3":
			w.Write(project4PayloadMembers)
			return
		case "/api/v4/projects/diaspora/diaspora-client/members/all/3":
			w.Write(project4PayloadMembers)
			return
		case "/api/v4/projects/6/members/all/3":
			w.Write(project6PayloadMembers)
			return
		case "/oauth/token":
			w.Write(accessTokenPayload)
			return
		case "/api/v4/user":
			w.Write(currentUserPayload)
			return
		}

		// else return a 404
		http.NotFound(w, r)
	})

	// return the server to the client which
	// will need to know the base URL path
	return server
}
