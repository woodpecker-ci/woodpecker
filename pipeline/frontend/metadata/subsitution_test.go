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

	"github.com/stretchr/testify/assert"
)

func TestEnvVarSubst(t *testing.T) {
	testCases := []struct {
		name    string
		yaml    string
		environ map[string]string
		want    string
	}{{
		name: "simple substitution",
		yaml: `steps:
		step1:
			image: ${HELLO_IMAGE}`,
		environ: map[string]string{"HELLO_IMAGE": "hello-world"},
		want: `steps:
		step1:
			image: hello-world`,
	}}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			result, err := EnvVarSubst(testCase.yaml, testCase.environ)
			assert.NoError(t, err)
			assert.EqualValues(t, testCase.want, result)
		})
	}
}
