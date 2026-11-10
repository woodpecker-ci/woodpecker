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
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDownloadGitPluginBinary(t *testing.T) {
	binary := []byte("plugin-git binary")
	sum := sha256.Sum256(binary)
	checksums := fmt.Sprintf("%s  linux-arm64_plugin-git\n%s  linux-arm_plugin-git\n",
		strings.Repeat("0", 64), hex.EncodeToString(sum[:]))

	tests := []struct {
		name      string
		arch      string
		version   string
		assetCode int
		assetBody []byte
		wantPath  string
		wantErr   string
	}{
		{name: "latest", arch: "arm", assetCode: http.StatusOK, assetBody: binary, wantPath: "/latest/download/linux-arm_plugin-git"},
		{name: "pinned version", arch: "arm", version: "2.12.1", assetCode: http.StatusOK, assetBody: binary, wantPath: "/download/2.12.1/linux-arm_plugin-git"},
		{name: "http error", arch: "arm", assetCode: http.StatusBadGateway, assetBody: []byte("<html>502</html>"), wantErr: "unexpected status 502"},
		{name: "checksum mismatch", arch: "arm64", assetCode: http.StatusOK, assetBody: binary, wantErr: "checksum mismatch"},
		{name: "unsupported arch", arch: "386", wantErr: "no release asset 'linux-386_plugin-git'"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var gotPath string
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if strings.HasSuffix(r.URL.Path, "/checksums.txt") {
					fmt.Fprint(w, checksums)
					return
				}
				gotPath = r.URL.Path
				w.WriteHeader(tt.assetCode)
				_, _ = w.Write(tt.assetBody)
			}))
			defer srv.Close()

			orig := pluginGitReleaseURL
			pluginGitReleaseURL = srv.URL
			defer func() { pluginGitReleaseURL = orig }()

			e := &local{os: "linux", arch: tt.arch, pluginGitVersion: tt.version}
			dest := filepath.Join(t.TempDir(), "plugin-git")
			err := e.downloadGitPluginBinary(t.Context(), dest)

			if tt.wantErr != "" {
				assert.ErrorContains(t, err, tt.wantErr)
				assert.NoFileExists(t, dest)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.wantPath, gotPath)
			content, err := os.ReadFile(dest)
			require.NoError(t, err)
			assert.Equal(t, binary, content)
		})
	}
}
