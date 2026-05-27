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
