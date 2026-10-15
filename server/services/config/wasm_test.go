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

package config_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	forge_types "go.woodpecker-ci.org/woodpecker/v3/server/forge/types"
	"go.woodpecker-ci.org/woodpecker/v3/server/model"
	"go.woodpecker-ci.org/woodpecker/v3/server/services/config"
	"go.woodpecker-ci.org/woodpecker/v3/server/services/utils/wasm"
	"go.woodpecker-ci.org/woodpecker/v3/server/services/utils/wasm/wasmtest"
)

func newWasmService(t *testing.T, pkg string) config.Service {
	t.Helper()

	runner, err := wasm.NewRunner(t.Context(), wasmtest.Build(t, pkg), wasm.DefaultLimits())
	require.NoError(t, err)
	t.Cleanup(func() { assert.NoError(t, runner.Close(context.Background())) })

	return config.NewWasm(runner)
}

func TestFetchFromWasmExtension(t *testing.T) {
	repo := &model.Repo{Name: "woodpecker", FullName: "woodpecker-ci/woodpecker"}
	pipeline := &model.Pipeline{Event: model.EventPush}
	// the forge and the user are not passed on purpose, the extension must not need them
	fetch := func(t *testing.T, files ...*forge_types.FileMeta) ([]*forge_types.FileMeta, error) {
		return newWasmService(t, wasmtest.ConfigGuest).Fetch(t.Context(), nil, nil, repo, pipeline, files, false)
	}

	t.Run("converts with the data of the request", func(t *testing.T) {
		got, err := fetch(t,
			&forge_types.FileMeta{Name: ".woodpecker/build.steps", Data: []byte("compile\non-{event}-in-{repo}\n")},
			&forge_types.FileMeta{Name: ".woodpecker/lint.yaml", Data: []byte("steps: []")},
		)
		require.NoError(t, err)

		require.Len(t, got, 2)
		assert.Equal(t, ".woodpecker/build.yaml", got[0].Name)
		assert.YAMLEq(t, `
steps:
  - name: compile
    image: dummy
    commands: [echo compile]
  - name: on-push-in-woodpecker
    image: dummy
    commands: [echo on-push-in-woodpecker]
`, string(got[0].Data))
		assert.Equal(t, &forge_types.FileMeta{Name: ".woodpecker/lint.yaml", Data: []byte("steps: []")}, got[1])
	})

	t.Run("keeps the config if the extension has nothing to say", func(t *testing.T) {
		files := []*forge_types.FileMeta{{Name: ".woodpecker.yaml", Data: []byte("steps: []")}}

		got, err := fetch(t, files...)
		require.NoError(t, err)
		assert.Equal(t, files, got)
	})

	t.Run("fails closed and tells why", func(t *testing.T) {
		got, err := fetch(t, &forge_types.FileMeta{Name: "a.steps", Data: []byte("!fail line 1: unknown thing")})
		require.ErrorContains(t, err, "line 1: unknown thing")
		assert.Nil(t, got, "the unconverted config must not be used as fallback")
	})

	t.Run("rejects what is not a response", func(t *testing.T) {
		for answer, reason := range map[string]string{
			`not json`:                  "invalid response",
			`{"configs":[]} and more`:   "invalid response",
			`{"configs":[null]}`:        "empty config",
			`{}`:                        "response without configs",
			`{"repo":{},"pipeline":{}}`: "response without configs",
		} {
			got, err := fetch(t, &forge_types.FileMeta{Name: "a.steps", Data: []byte("!raw " + answer)})
			require.ErrorContains(t, err, reason, answer)
			assert.Nil(t, got, answer)
		}
	})

	t.Run("accepts an answer without any workflow", func(t *testing.T) {
		got, err := fetch(t, &forge_types.FileMeta{Name: "a.steps", Data: []byte(`!raw {"configs":[]}`)})
		require.NoError(t, err)
		assert.Empty(t, got)
		assert.NotNil(t, got)
	})
}
