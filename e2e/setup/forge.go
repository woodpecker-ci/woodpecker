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

package setup

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/mock"

	forge_mocks "go.woodpecker-ci.org/woodpecker/v3/server/forge/mocks"
	forge_types "go.woodpecker-ci.org/woodpecker/v3/server/forge/types"
	"go.woodpecker-ci.org/woodpecker/v3/server/model"
)

// newMockForge builds a MockForge that serves the given files for any
// config-fetch call, no-ops status reporting, and stubs all other methods safely.
//
// Single-workflow (len(files)==1, name ".woodpecker.yaml"): File() returns the
// raw YAML bytes; Dir() is not called but is stubbed for safety.
//
// Multi-workflow (len(files)>1, names ".woodpecker/foo.yaml"): File() returns
// empty (causing the config service to fall through to Dir()); Dir() returns
// all files.
func newMockForge(t *testing.T, files []*forge_types.FileMeta) *forge_mocks.MockForge {
	t.Helper()
	m := forge_mocks.NewMockForge(t)

	// Identity.
	m.On("Name").Return("mock").Maybe()
	m.On("URL").Return("https://forge.example.test").Maybe()

	if len(files) == 1 {
		// Single-workflow: config service calls File(".woodpecker.yaml").
		m.On("File",
			mock.Anything, mock.Anything, mock.Anything, mock.Anything, ".woodpecker.yaml",
		).Return(files[0].Data, nil).Maybe()

		m.On("Dir",
			mock.Anything, mock.Anything, mock.Anything, mock.Anything, ".woodpecker",
		).Return(files, nil).Maybe()
	} else {
		// Multi-workflow: config service calls Dir(".woodpecker").
		// File() must return empty so the service falls through to Dir().
		m.On("File",
			mock.Anything, mock.Anything, mock.Anything, mock.Anything, ".woodpecker.yaml",
		).Return([]byte(nil), nil).Maybe()
		m.On("Dir",
			mock.Anything, mock.Anything, mock.Anything, mock.Anything, ".woodpecker",
		).Return(files, nil).Maybe()
	}

	// Status reporting back to forge — no-op.
	m.On("Status",
		mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything,
	).Return(nil).Maybe()

	// Netrc for clone steps.
	m.On("Netrc",
		mock.Anything, mock.Anything,
	).Return(&model.Netrc{}, nil).Maybe()

	return m
}

// compile-time import guard.
var _ *http.Request
