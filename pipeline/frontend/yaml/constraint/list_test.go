// Copyright 2025 Woodpecker Authors
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

package constraint

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v3"
)

func TestConstraintList(t *testing.T) {
	testdata := []struct {
		conf string
		with string
		want bool
	}{
		// string value
		{
			conf: "main",
			with: "develop",
			want: false,
		},
		{
			conf: "main",
			with: "main",
			want: true,
		},
		{
			conf: "feature/*",
			with: "feature/foo",
			want: true,
		},
		// slice value
		{
			conf: "[ main, feature/* ]",
			with: "develop",
			want: false,
		},
		{
			conf: "[ main, feature/* ]",
			with: "main",
			want: true,
		},
		{
			conf: "[ main, feature/* ]",
			with: "feature/foo",
			want: true,
		},
		// includes block
		{
			conf: "include: main",
			with: "develop",
			want: false,
		},
		{
			conf: "include: main",
			with: "main",
			want: true,
		},
		{
			conf: "include: feature/*",
			with: "main",
			want: false,
		},
		{
			conf: "include: feature/*",
			with: "feature/foo",
			want: true,
		},
		{
			conf: "include: [ main, feature/* ]",
			with: "develop",
			want: false,
		},
		{
			conf: "include: [ main, feature/* ]",
			with: "main",
			want: true,
		},
		{
			conf: "include: [ main, feature/* ]",
			with: "feature/foo",
			want: true,
		},
		// excludes block
		{
			conf: "exclude: main",
			with: "develop",
			want: true,
		},
		{
			conf: "exclude: main",
			with: "main",
			want: false,
		},
		{
			conf: "exclude: feature/*",
			with: "main",
			want: true,
		},
		{
			conf: "exclude: feature/*",
			with: "feature/foo",
			want: false,
		},
		{
			conf: "exclude: [ main, develop ]",
			with: "main",
			want: false,
		},
		{
			conf: "exclude: [ feature/*, bar ]",
			with: "main",
			want: true,
		},
		{
			conf: "exclude: [ feature/*, bar ]",
			with: "feature/foo",
			want: false,
		},
		// include and exclude blocks
		{
			conf: "{ include: [ main, feature/* ], exclude: [ develop ] }",
			with: "main",
			want: true,
		},
		{
			conf: "{ include: [ main, feature/* ], exclude: [ feature/bar ] }",
			with: "feature/bar",
			want: false,
		},
		{
			conf: "{ include: [ main, feature/* ], exclude: [ main, develop ] }",
			with: "main",
			want: false,
		},
		// empty blocks
		{
			conf: "",
			with: "main",
			want: true,
		},
	}
	for _, test := range testdata {
		c := parseConstraintList(t, test.conf)
		assert.Equal(t, test.want, c.Match(test.with))
	}
}

func parseConstraintList(t *testing.T, s string) *List {
	c := &List{}
	assert.NoError(t, yaml.Unmarshal([]byte(s), c))
	return c
}
