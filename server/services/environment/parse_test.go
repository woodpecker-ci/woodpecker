// Copyright 2024 Woodpecker Authors
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

package environment

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParse(t *testing.T) {
	service := Parse([]string{})
	env, err := service.EnvironList(nil)
	assert.NoError(t, err)
	assert.Empty(t, env)

	service = Parse([]string{"ENV:value"})
	env, err = service.EnvironList(nil)
	assert.NoError(t, err)
	assert.Len(t, env, 1)
	assert.Equal(t, env[0].Name, "ENV")
	assert.Equal(t, env[0].Value, "value")

	service = Parse([]string{"ENV:value", "ENV2:value2"})
	env, err = service.EnvironList(nil)
	assert.NoError(t, err)
	assert.Len(t, env, 2)

	service = Parse([]string{"ENV:value", "ENV2:value2", "ENV3_WITHOUT_VALUE"})
	env, err = service.EnvironList(nil)
	assert.NoError(t, err)
	assert.Len(t, env, 2)
}
