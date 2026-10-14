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

package update

import (
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"go.woodpecker-ci.org/woodpecker/v3/version"
)

func TestCheckForUpdate(t *testing.T) {
	version.Version = "1.0.0"
	fixtureHandler := func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/version.json" {
			http.NotFound(w, r)
			return
		}

		_, _ = io.WriteString(w, `{"latest": "1.0.1", "next": "1.0.2", "rc": "1.0.3"}`)
	}
	ts := httptest.NewServer(http.HandlerFunc(fixtureHandler))
	defer ts.Close()

	newVersion, err := checkForUpdate(t.Context(), ts.URL+"/version.json", false)
	require.NoError(t, err)
	require.NotNil(t, newVersion)
	assert.Equal(t, "1.0.1", newVersion.Version)
}

func TestDownloadNewVersion(t *testing.T) {
	downloadFilePath := "/woodpecker-cli_linux_amd64.tar.gz"

	fixtureHandler := func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != downloadFilePath {
			http.NotFound(w, r)
			return
		}

		_, _ = io.WriteString(w, `blob`)
	}
	ts := httptest.NewServer(http.HandlerFunc(fixtureHandler))
	defer ts.Close()

	file, err := downloadNewVersion(t.Context(), ts.URL+downloadFilePath)
	require.NoError(t, err)
	require.NotEmpty(t, file)

	_ = os.Remove(file)
}
