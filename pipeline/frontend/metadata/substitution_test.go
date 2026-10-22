// Copyright 2023 Woodpecker Authors
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

package metadata

import (
	"testing"

	"codeberg.org/6543/xyaml/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEnvVarSubst(t *testing.T) {
	result, err := EnvVarSubst(`steps:
		step1:
			image: ${HELLO_IMAGE}
			command: echo ${NEWLINE}`, map[string]string{"HELLO_IMAGE": "hello-world", "NEWLINE": "some env\nwith newline"})
	assert.NoError(t, err)
	assert.EqualValues(t, `steps:
		step1:
			image: hello-world
			command: echo "some env\nwith newline"`, result)
}

// TestEnvVarSubstNoStructureInjection guards against a user-controlled value
// (commit message, branch name, PR title, …) injecting new YAML structure
// through a line break. YAML 1.1 treats LF, CR, NEL, LS and PS as line breaks,
// so any of them left bare in a substituted plain scalar would end the scalar
// and let the rest be parsed as sibling nodes. All must be quoted.
func TestEnvVarSubstNoStructureInjection(t *testing.T) {
	breaks := map[string]string{
		"LF":  "\n",
		"CR":  "\r",
		"NEL": "\u0085",
		"LS":  " ",
		"PS":  " ",
	}
	for name, br := range breaks {
		t.Run(name, func(t *testing.T) {
			payload := "value" + br + "injected: owned"
			result, err := EnvVarSubst("key: ${EVIL}", map[string]string{"EVIL": payload})
			require.NoError(t, err)

			var parsed map[string]any
			require.NoError(t, xyaml.Unmarshal([]byte(result), &parsed))

			_, injected := parsed["injected"]
			assert.False(t, injected, "break char %q injected a top-level key: %q", br, result)
			assert.Equal(t, payload, parsed["key"], "value must round-trip intact")
		})
	}
}
