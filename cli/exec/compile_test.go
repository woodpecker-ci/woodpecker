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

package exec

import (
	"io"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	backend_types "go.woodpecker-ci.org/woodpecker/v3/pipeline/backend/types"
	"go.woodpecker-ci.org/woodpecker/v3/pipeline/compile"
	"go.woodpecker-ci.org/woodpecker/v3/pipeline/frontend/builder"
)

func collect(t *testing.T, step *backend_types.Step, output string) *compileCollector {
	t.Helper()

	collector := newCompileCollector()
	require.NoError(t, collector.logger()(step, io.NopCloser(strings.NewReader(output))))

	return collector
}

func TestCompileCollectorCollectsAResponse(t *testing.T) {
	block, err := compile.EncodeResponse([]compile.Config{{Name: "build", Data: []byte("steps: {}\n")}})
	require.NoError(t, err)

	step := &backend_types.Step{UUID: "step-1", Name: "generate", Type: backend_types.StepTypeCommands}
	configs, err := collect(t, step, "generating\n"+block).emitted()
	require.NoError(t, err)

	require.Len(t, configs, 1)
	assert.Equal(t, "build", configs[0].Name)
}

func TestCompileCollectorIgnoresTheCloneStep(t *testing.T) {
	block, err := compile.EncodeResponse([]compile.Config{{Name: "build", Data: []byte("steps: {}\n")}})
	require.NoError(t, err)

	// Same rule as on an agent: only steps the user declared under compile: may
	// emit a response.
	step := &backend_types.Step{UUID: "clone", Name: "clone", Type: backend_types.StepTypeClone}
	configs, err := collect(t, step, block).emitted()
	require.NoError(t, err)
	assert.Empty(t, configs)
}

func TestCompileCollectorReportsAMissingResponse(t *testing.T) {
	step := &backend_types.Step{UUID: "step-1", Name: "generate", Type: backend_types.StepTypeCommands}
	_, err := collect(t, step, "I forgot to emit anything\n").emitted()
	assert.ErrorIs(t, err, compile.ErrNoResponse)
}

func TestMergeCompiled(t *testing.T) {
	// The yaml file name carries a path and an extension, but an emitted config
	// name does not. They have to be matched on the sanitized name or a
	// generator rewriting its own config would add a second workflow instead.
	yamls := []*builder.YamlFile{
		{Name: ".woodpecker/build.yaml", Data: []byte("compile: {}\n")},
		{Name: ".woodpecker/deploy.yaml", Data: []byte("steps: {}\n")},
	}

	merged, err := mergeCompiled(yamls, []compile.Config{
		{Name: "build", Data: []byte("steps:\n  test:\n    image: golang\n")},
	})
	require.NoError(t, err)

	require.Len(t, merged, 2)
	assert.Equal(t, "build", merged[0].Name)
	assert.Equal(t, "steps:\n  test:\n    image: golang\n", string(merged[0].Data))
	assert.Equal(t, "deploy", merged[1].Name)
}

func TestMergeCompiledRejectsANestedCompileSection(t *testing.T) {
	yamls := []*builder.YamlFile{{Name: "build.yaml", Data: []byte("compile: {}\n")}}

	_, err := mergeCompiled(yamls, []compile.Config{
		{Name: "build", Data: []byte("compile:\n  again:\n    image: alpine\n")},
	})
	assert.ErrorContains(t, err, "declares a compile section")
}
