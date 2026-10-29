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
	"archive/tar"
	"compress/gzip"
	"io"
	"io/fs"
	"os"
)

const tarDirectoryMode fs.FileMode = 0o755

func UnTar(dst string, r io.Reader) error {
	gzr, err := gzip.NewReader(r)
	if err != nil {
		return err
	}
	defer gzr.Close()

	// all file access goes through root, so no entry can write outside of dst
	root, err := os.OpenRoot(dst)
	if err != nil {
		return err
	}
	defer root.Close()

	tr := tar.NewReader(gzr)

	for {
		header, err := tr.Next()

		switch {
		case err == io.EOF:
			return nil

		case err != nil:
			return err

		case header == nil:
			continue
		}

		switch header.Typeflag {
		case tar.TypeDir:
			if _, err := root.Stat(header.Name); err != nil {
				if err := root.MkdirAll(header.Name, tarDirectoryMode); err != nil {
					return err
				}
			}

		case tar.TypeReg:
			f, err := root.OpenFile(header.Name, os.O_CREATE|os.O_RDWR, os.FileMode(header.Mode)&os.ModePerm)
			if err != nil {
				return err
			}

			if _, err := io.Copy(f, tr); err != nil {
				return err
			}

			f.Close()
		}
	}
}
