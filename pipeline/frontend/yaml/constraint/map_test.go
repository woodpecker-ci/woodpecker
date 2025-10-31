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

func TestConstraintMap(t *testing.T) {
	testdata := []struct {
		conf string
		with map[string]string
		want bool
	}{
		{
			conf: "GOLANG: 1.7",
			with: map[string]string{"GOLANG": "1.7"},
			want: true,
		},
		{
			conf: "GOLANG: tip",
			with: map[string]string{"GOLANG": "1.7"},
			want: false,
		},
		{
			conf: "{ GOLANG: 1.7, REDIS: 3.1 }",
			with: map[string]string{"GOLANG": "1.7", "REDIS": "3.1", "MYSQL": "5.6"},
			want: true,
		},
		{
			conf: "{ GOLANG: 1.7, REDIS: 3.1 }",
			with: map[string]string{"GOLANG": "1.7", "REDIS": "3.0"},
			want: false,
		},
		{
			conf: "{ GOLANG: 1.7, REDIS: 3.* }",
			with: map[string]string{"GOLANG": "1.7", "REDIS": "3.0"},
			want: true,
		},
		{
			conf: "{ GOLANG: 1.7, BRANCH: release/**/test }",
			with: map[string]string{"GOLANG": "1.7", "BRANCH": "release/v1.12.1//test"},
			want: true,
		},
		{
			conf: "{ GOLANG: 1.7, BRANCH: release/**/test }",
			with: map[string]string{"GOLANG": "1.7", "BRANCH": "release/v1.12.1/qest"},
			want: false,
		},
		// include syntax
		{
			conf: "include: { GOLANG: 1.7 }",
			with: map[string]string{"GOLANG": "1.7"},
			want: true,
		},
		{
			conf: "include: { GOLANG: tip }",
			with: map[string]string{"GOLANG": "1.7"},
			want: false,
		},
		{
			conf: "include: { GOLANG: 1.7, REDIS: 3.1 }",
			with: map[string]string{"GOLANG": "1.7", "REDIS": "3.1", "MYSQL": "5.6"},
			want: true,
		},
		{
			conf: "include: { GOLANG: 1.7, REDIS: 3.1 }",
			with: map[string]string{"GOLANG": "1.7", "REDIS": "3.0"},
			want: false,
		},
		// exclude syntax
		{
			conf: "exclude: { GOLANG: 1.7 }",
			with: map[string]string{"GOLANG": "1.7"},
			want: false,
		},
		{
			conf: "exclude: { GOLANG: tip }",
			with: map[string]string{"GOLANG": "1.7"},
			want: true,
		},
		{
			conf: "exclude: { GOLANG: 1.7, REDIS: 3.1 }",
			with: map[string]string{"GOLANG": "1.7", "REDIS": "3.1", "MYSQL": "5.6"},
			want: false,
		},
		{
			conf: "exclude: { GOLANG: 1.7, REDIS: 3.1 }",
			with: map[string]string{"GOLANG": "1.7", "REDIS": "3.0"},
			want: true,
		},
		// exclude AND include values
		{
			conf: "{ include: { GOLANG: 1.7 }, exclude: { GOLANG: 1.7 } }",
			with: map[string]string{"GOLANG": "1.7"},
			want: false,
		},
		// blanks
		{
			conf: "",
			with: map[string]string{"GOLANG": "1.7", "REDIS": "3.0"},
			want: true,
		},
		{
			conf: "GOLANG: 1.7",
			with: map[string]string{},
			want: false,
		},
		{
			conf: "{ GOLANG: 1.7, REDIS: 3.0 }",
			with: map[string]string{},
			want: false,
		},
		{
			conf: "include: { GOLANG: 1.7, REDIS: 3.1 }",
			with: map[string]string{},
			want: false,
		},
		{
			conf: "exclude: { GOLANG: 1.7, REDIS: 3.1 }",
			with: map[string]string{},
			want: true,
		},
	}
	for _, test := range testdata {
		c := parseConstraintMap(t, test.conf)
		assert.Equal(t, test.want, c.Match(test.with), "config: '%s', with: '%s'", test.conf, test.with)
	}
}

func parseConstraintMap(t *testing.T, s string) *Map {
	c := &Map{}
	assert.NoError(t, yaml.Unmarshal([]byte(s), c))
	return c
}
