// Copyright 2025 Woodpecker Authors
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

//go:build external_web

package web

import (
	"fmt"
	"io"
	"net/http"
	"os"
)

var webUIRoot string // do not forget to set it at build time

func HTTPFS() (http.FileSystem, error) {
	if stat, err := os.Stat(webUIRoot); err != nil {
		return nil, fmt.Errorf("compiled in webui root path '%s' does not exist: %w", webUIRoot, err)
	} else if !stat.IsDir() {
		return nil, fmt.Errorf("compiled in webui root path '%s' exist but is no directory", webUIRoot)
	}
	return http.Dir(webUIRoot), nil
}

func Lookup(path string) (buf []byte, err error) {
	httpFS, err := HTTPFS()
	if err != nil {
		return nil, err
	}
	file, err := httpFS.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	buf, err = io.ReadAll(file)
	if err != nil {
		return nil, err
	}

	return buf, nil
}
