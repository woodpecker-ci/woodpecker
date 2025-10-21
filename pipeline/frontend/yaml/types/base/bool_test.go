// Copyright 2023 Woodpecker Authors
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

package base

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v3"
)

func TestBoolTrue(t *testing.T) {
	t.Run("unmarshal true", func(t *testing.T) {
		in := []byte("true")
		out := BoolTrue{}
		err := yaml.Unmarshal(in, &out)
		assert.NoError(t, err)
		assert.True(t, out.Bool())
	})

	t.Run("unmarshal false", func(t *testing.T) {
		in := []byte("false")
		out := BoolTrue{}
		err := yaml.Unmarshal(in, &out)
		assert.NoError(t, err)
		assert.False(t, out.Bool())
	})

	t.Run("unmarshal true when empty", func(t *testing.T) {
		in := []byte("")
		out := BoolTrue{}
		err := yaml.Unmarshal(in, &out)
		assert.NoError(t, err)
		assert.True(t, out.Bool())
	})

	t.Run("throw error when invalid", func(t *testing.T) {
		in := []byte("abc") // string value should fail parse
		out := BoolTrue{}
		err := yaml.Unmarshal(in, &out)
		assert.Error(t, err)
	})
}
