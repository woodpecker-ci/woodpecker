// Copyright 2024 Woodpecker Authors
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

type StructStringOrInt struct {
	Foo StringOrInt
}

func TestStringOrIntYaml(t *testing.T) {
	for _, str := range []string{`{foo: 10}`, `{foo: "10"}`} {
		s := StructStringOrInt{}
		assert.NoError(t, yaml.Unmarshal([]byte(str), &s))

		assert.Equal(t, StringOrInt(10), s.Foo)

		d, err := yaml.Marshal(&s)
		assert.NoError(t, err)

		s2 := StructStringOrInt{}
		assert.NoError(t, yaml.Unmarshal(d, &s2))

		assert.Equal(t, StringOrInt(10), s2.Foo)
	}
}
