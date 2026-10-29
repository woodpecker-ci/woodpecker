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

package update

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func tarGz(t *testing.T, headers ...*tar.Header) *bytes.Buffer {
	t.Helper()
	buf := new(bytes.Buffer)
	gz := gzip.NewWriter(buf)
	tw := tar.NewWriter(gz)
	for _, h := range headers {
		require.NoError(t, tw.WriteHeader(h))
		_, err := tw.Write(make([]byte, h.Size))
		require.NoError(t, err)
	}
	require.NoError(t, tw.Close())
	require.NoError(t, gz.Close())
	return buf
}

func TestUnTar(t *testing.T) {
	t.Run("extracts files and dirs", func(t *testing.T) {
		dst := t.TempDir()
		err := UnTar(dst, tarGz(t,
			&tar.Header{Name: "bin/", Typeflag: tar.TypeDir, Mode: 0o755},
			&tar.Header{Name: "bin/woodpecker-cli", Typeflag: tar.TypeReg, Mode: 0o755, Size: 4},
		))
		require.NoError(t, err)
		assert.FileExists(t, filepath.Join(dst, "bin", "woodpecker-cli"))
	})

	for _, name := range []string{"../escaped", "a/../../escaped"} {
		t.Run("rejects "+name, func(t *testing.T) {
			parent := t.TempDir()
			dst := filepath.Join(parent, "dst")
			require.NoError(t, os.Mkdir(dst, 0o755))

			err := UnTar(dst, tarGz(t, &tar.Header{Name: name, Typeflag: tar.TypeReg, Mode: 0o644, Size: 4}))
			assert.Error(t, err)
			assert.NoFileExists(t, filepath.Join(parent, "escaped"))
		})
	}
}
