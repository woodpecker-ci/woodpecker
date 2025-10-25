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

func TestConstraintPath(t *testing.T) {
	testdata := []struct {
		conf    string
		with    []string
		message string
		want    bool
	}{
		{
			conf: "",
			with: []string{"CHANGELOG.md", "README.md"},
			want: true,
		},
		{
			conf: "CHANGELOG.md",
			with: []string{"CHANGELOG.md", "README.md"},
			want: true,
		},
		{
			conf: "'*.md'",
			with: []string{"CHANGELOG.md", "README.md"},
			want: true,
		},
		{
			conf: "['*.md']",
			with: []string{"CHANGELOG.md", "README.md"},
			want: true,
		},
		{
			conf: "'docs/*'",
			with: []string{"docs/README.md"},
			want: true,
		},
		{
			conf: "'docs/*'",
			with: []string{"docs/sub/README.md"},
			want: false,
		},
		{
			conf: "'docs/**'",
			with: []string{"docs/README.md", "docs/sub/README.md", "docs/sub-sub/README.md"},
			want: true,
		},
		{
			conf: "'docs/**'",
			with: []string{"README.md"},
			want: false,
		},
		{
			conf: "{ include: [ README.md ] }",
			with: []string{"CHANGELOG.md"},
			want: false,
		},
		{
			conf: "{ exclude: [ README.md ] }",
			with: []string{"design.md"},
			want: true,
		},
		// include and exclude blocks
		{
			conf: "{ include: [ '*.md', '*.ini' ], exclude: [ CHANGELOG.md ] }",
			with: []string{"README.md"},
			want: true,
		},
		{
			conf: "{ include: [ '*.md' ], exclude: [ CHANGELOG.md ] }",
			with: []string{"CHANGELOG.md"},
			want: false,
		},
		{
			conf: "{ include: [ '*.md' ], exclude: [ CHANGELOG.md ] }",
			with: []string{"README.md", "CHANGELOG.md"},
			want: true,
		},
		{
			conf: "{ exclude: [ CHANGELOG.md ] }",
			with: []string{"README.md", "CHANGELOG.md"},
			want: true,
		},
		{
			conf: "{ exclude: [ CHANGELOG.md, docs/**/*.md ] }",
			with: []string{"docs/main.md", "CHANGELOG.md"},
			want: false,
		},
		{
			conf: "{ exclude: [ CHANGELOG.md, docs/**/*.md ] }",
			with: []string{"docs/main.md", "CHANGELOG.md", "README.md"},
			want: true,
		},
		// commit message ignore matches
		{
			conf:    "{ include: [ README.md ], ignore_message: '[ALL]' }",
			with:    []string{"CHANGELOG.md"},
			message: "Build them [ALL]",
			want:    true,
		},
		{
			conf:    "{ exclude: [ '*.php' ], ignore_message: '[ALL]' }",
			with:    []string{"myfile.php"},
			message: "Build them [ALL]",
			want:    true,
		},
		{
			conf:    "{ ignore_message: '[ALL]' }",
			with:    []string{},
			message: "Build them [ALL]",
			want:    true,
		},
		// empty commit
		{
			conf: "{ include: [ README.md ] }",
			with: []string{},
			want: true,
		},
		{
			conf: "{ include: [ README.md ], on_empty: false }",
			with: []string{},
			want: false,
		},
		{
			conf: "{ include: [ README.md ], on_empty: true }",
			with: []string{},
			want: true,
		},
	}
	for _, test := range testdata {
		c := parseConstraintPath(t, test.conf)
		assert.Equal(t, test.want, c.Match(test.with, test.message))
	}
}

func parseConstraintPath(t *testing.T, s string) *Path {
	c := &Path{}
	assert.NoError(t, yaml.Unmarshal([]byte(s), c))
	return c
}
