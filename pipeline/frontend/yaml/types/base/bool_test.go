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

	"github.com/franela/goblin"
	"gopkg.in/yaml.v3"
)

func TestBoolTrue(t *testing.T) {
	g := goblin.Goblin(t)

	g.Describe("Yaml bool type", func() {
		g.Describe("given a yaml file", func() {
			g.It("should unmarshal true", func() {
				in := []byte("true")
				out := BoolTrue{}
				err := yaml.Unmarshal(in, &out)
				if err != nil {
					g.Fail(err)
				}
				g.Assert(out.Bool()).Equal(true)
			})

			g.It("should unmarshal false", func() {
				in := []byte("false")
				out := BoolTrue{}
				err := yaml.Unmarshal(in, &out)
				if err != nil {
					g.Fail(err)
				}
				g.Assert(out.Bool()).Equal(false)
			})

			g.It("should unmarshal true when empty", func() {
				in := []byte("")
				out := BoolTrue{}
				err := yaml.Unmarshal(in, &out)
				if err != nil {
					g.Fail(err)
				}
				g.Assert(out.Bool()).Equal(true)
			})

			g.It("should throw error when invalid", func() {
				in := []byte("abc") // string value should fail parse
				out := BoolTrue{}
				err := yaml.Unmarshal(in, &out)
				g.Assert(err).IsNotNil("expects error")
			})
		})
	})
}
