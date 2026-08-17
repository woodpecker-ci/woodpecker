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

package compile

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func names(configs []Config) []string {
	out := make([]string, 0, len(configs))
	for _, config := range configs {
		out = append(out, config.Name)
	}

	return out
}

func TestMerge(t *testing.T) {
	source := []Config{
		{Name: "build", Data: []byte("source build")},
		{Name: "deploy", Data: []byte("source deploy")},
	}

	tests := []struct {
		name    string
		emitted []Config
		want    []string
	}{
		{
			name: "nothing emitted leaves the set alone",
			want: []string{"build", "deploy"},
		},
		{
			name:    "overwrite keeps the position",
			emitted: []Config{{Name: "build", Data: []byte("compiled build")}},
			want:    []string{"build", "deploy"},
		},
		{
			name:    "added configs are appended in emission order",
			emitted: []Config{{Name: "lint", Data: []byte("x")}, {Name: "docs", Data: []byte("x")}},
			want:    []string{"build", "deploy", "lint", "docs"},
		},
		{
			name:    "an empty payload removes",
			emitted: []Config{{Name: "deploy"}},
			want:    []string{"build"},
		},
		{
			name:    "removing everything is allowed",
			emitted: []Config{{Name: "build"}, {Name: "deploy"}},
			want:    []string{},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			merged, err := Merge(source, test.emitted)
			require.NoError(t, err)
			assert.Equal(t, test.want, names(merged))
		})
	}
}

func TestMergeOverwritesTheEmittingConfig(t *testing.T) {
	// Permutation is the natural case: a compile workflow reads its own yaml,
	// transforms it and emits it back under the same name.
	source := []Config{{Name: "build", Data: []byte("compile:\n  generate:\n    image: alpine\n")}}

	merged, err := Merge(source, []Config{{Name: "build", Data: []byte("steps:\n  test:\n    image: golang\n")}})
	require.NoError(t, err)

	require.Len(t, merged, 1)
	assert.Equal(t, "steps:\n  test:\n    image: golang\n", string(merged[0].Data))
}

func TestMergeRejectsRemovingAnUnknownConfig(t *testing.T) {
	_, err := Merge([]Config{{Name: "build", Data: []byte("x")}}, []Config{{Name: "typo"}})
	assert.ErrorIs(t, err, ErrRemoveUnknownConfig)
}

func TestMergeRejectsADuplicateName(t *testing.T) {
	_, err := Merge(nil, []Config{
		{Name: "build", Data: []byte("first")},
		{Name: "build", Data: []byte("second")},
	})
	assert.ErrorIs(t, err, ErrDuplicateName)
}

func TestMergeDoesNotMutateTheSource(t *testing.T) {
	source := []Config{{Name: "build", Data: []byte("source")}}

	_, err := Merge(source, []Config{{Name: "build", Data: []byte("compiled")}})
	require.NoError(t, err)

	assert.Equal(t, "source", string(source[0].Data))
}

func TestLintEmitted(t *testing.T) {
	tests := []struct {
		name    string
		emitted []Config
		want    string // empty = no error expected
	}{
		{
			name:    "a plain name is fine",
			emitted: []Config{{Name: "build", Data: []byte("steps:\n  test:\n    image: golang\n")}},
		},
		{
			name:    "a removal carries no data to inspect",
			emitted: []Config{{Name: "build"}},
		},
		{
			name:    "empty name",
			emitted: []Config{{Name: "", Data: []byte("steps: {}")}},
			want:    "empty name",
		},
		{
			name:    "path separator",
			emitted: []Config{{Name: "sub/build", Data: []byte("steps: {}")}},
			want:    "path separator",
		},
		{
			name:    "traversal",
			emitted: []Config{{Name: "..build", Data: []byte("steps: {}")}},
			want:    `must not contain ".."`,
		},
		{
			name:    "padded name",
			emitted: []Config{{Name: " build ", Data: []byte("steps: {}")}},
			want:    "whitespace",
		},
		{
			name: "duplicate name",
			emitted: []Config{
				{Name: "build", Data: []byte("steps: {}")},
				{Name: "build", Data: []byte("steps: {}")},
			},
			want: "duplicate config name",
		},
		{
			name:    "broken yaml",
			emitted: []Config{{Name: "build", Data: []byte("steps: [")}},
			want:    "not valid yaml",
		},
		{
			name:    "nested compile",
			emitted: []Config{{Name: "build", Data: []byte("compile:\n  again:\n    image: alpine\n")}},
			want:    "declares a compile section",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := LintEmitted(test.emitted)
			if test.want == "" {
				assert.NoError(t, err)
				return
			}
			assert.ErrorContains(t, err, test.want)
		})
	}
}

func TestLintEmittedCaps(t *testing.T) {
	emitted := []Config{
		{Name: "a", Data: []byte("steps: {}")},
		{Name: "b", Data: []byte("steps: {}")},
	}

	assert.ErrorContains(t, LintEmitted(emitted, WithMaxConfigs(1)), "the limit is 1")

	big := []Config{{Name: "a", Data: []byte(strings.Repeat("#", 128))}}
	assert.ErrorContains(t, LintEmitted(big, WithMaxSize(64)), "more than 64 bytes")
}
