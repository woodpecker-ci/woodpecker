// Copyright 2026 Woodpecker Authors
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

package local

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDownloadGitPluginBinary(t *testing.T) {
	binary := []byte("plugin-git binary")

	tests := []struct {
		name      string
		arch      string
		version   string
		assetCode int
		wantAsset string
		wantErr   string
	}{
		{name: "latest", arch: "arm64", assetCode: http.StatusOK, wantAsset: "/assets/linux-arm64_plugin-git"},
		{name: "pinned version", arch: "arm64", version: "2.12.1", assetCode: http.StatusOK, wantAsset: "/assets/linux-arm64_plugin-git"},
		{name: "arm does not match arm64", arch: "arm", assetCode: http.StatusOK, wantAsset: "/assets/linux-arm_plugin-git"},
		{name: "unknown version", arch: "arm64", version: "0.0.0", wantErr: "plugin-git release '0.0.0' does not exist"},
		{name: "unsupported arch", arch: "386", wantErr: "plugin-git release 'latest' has no binary for linux/386"},
		{name: "asset download error", arch: "arm64", assetCode: http.StatusBadGateway, wantErr: "unexpected status 502"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var gotAsset string
			mux := http.NewServeMux()
			srv := httptest.NewServer(mux)
			defer srv.Close()

			release := func(w http.ResponseWriter, _ *http.Request) {
				// arm64 is listed before arm, like in the real releases
				fmt.Fprintf(w, `{"assets":[
					{"name":"checksums.txt","browser_download_url":"%[1]s/assets/checksums.txt"},
					{"name":"linux-amd64_plugin-git","browser_download_url":"%[1]s/assets/linux-amd64_plugin-git"},
					{"name":"linux-arm64_plugin-git","browser_download_url":"%[1]s/assets/linux-arm64_plugin-git"},
					{"name":"linux-arm_plugin-git","browser_download_url":"%[1]s/assets/linux-arm_plugin-git"}
				]}`, srv.URL)
			}
			mux.HandleFunc("/releases/latest", release)
			mux.HandleFunc("/releases/tags/2.12.1", release)
			mux.HandleFunc("/assets/", func(w http.ResponseWriter, r *http.Request) {
				gotAsset = r.URL.Path
				w.WriteHeader(tt.assetCode)
				_, _ = w.Write(binary)
			})

			orig := pluginGitReleaseAPI
			pluginGitReleaseAPI = srv.URL + "/releases"
			defer func() { pluginGitReleaseAPI = orig }()

			e := &local{os: "linux", arch: tt.arch, pluginGitVersion: tt.version}
			dest := filepath.Join(t.TempDir(), "plugin-git")
			err := e.downloadGitPluginBinary(t.Context(), dest)

			if tt.wantErr != "" {
				assert.ErrorContains(t, err, tt.wantErr)
				assert.NoFileExists(t, dest)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.wantAsset, gotAsset)
			content, err := os.ReadFile(dest)
			require.NoError(t, err)
			assert.Equal(t, binary, content)
		})
	}
}
