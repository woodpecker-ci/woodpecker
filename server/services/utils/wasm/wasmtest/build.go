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

//go:build test

// Package wasmtest builds the wasm modules that are used as test fixtures.
package wasmtest

import (
	"os"
	"os/exec"
	"path/filepath"
	"sync"
	"testing"

	"github.com/stretchr/testify/require"
)

const (
	// ProbeGuest is a module to test the sandbox itself.
	ProbeGuest = "go.woodpecker-ci.org/woodpecker/v3/server/services/utils/wasm/testdata/guest"
	// ConfigGuest is a minimal configuration extension.
	ConfigGuest = "go.woodpecker-ci.org/woodpecker/v3/server/services/config/testdata/guest"
)

var (
	lock  sync.Mutex
	built = map[string][]byte{}
)

// Build compiles the main package to a WASI module and returns it.
// A package is only compiled once per test binary.
func Build(tb testing.TB, pkg string) []byte {
	tb.Helper()
	lock.Lock()
	defer lock.Unlock()

	if binary, ok := built[pkg]; ok {
		return binary
	}

	out := filepath.Join(tb.TempDir(), "guest.wasm")
	cmd := exec.CommandContext(tb.Context(), "go", "build", "-o", out, pkg)
	cmd.Env = append(os.Environ(), "GOOS=wasip1", "GOARCH=wasm", "CGO_ENABLED=0", "GOFLAGS=")
	output, err := cmd.CombinedOutput()
	require.NoError(tb, err, "build wasm module %s: %s", pkg, output)

	binary, err := os.ReadFile(out)
	require.NoError(tb, err)
	built[pkg] = binary

	return binary
}

// File is like Build, but returns the path of a file that contains the module.
func File(tb testing.TB, pkg string) string {
	tb.Helper()

	path := filepath.Join(tb.TempDir(), "extension.wasm")
	require.NoError(tb, os.WriteFile(path, Build(tb, pkg), 0o600))

	return path
}
