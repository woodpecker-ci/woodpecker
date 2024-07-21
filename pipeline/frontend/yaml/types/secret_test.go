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

package types

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v3"
)

func TestUnmarshalSecrets(t *testing.T) {
	testdata := []struct {
		from string
		want []*Secret
	}{
		{
			from: "[ mysql_username, mysql_password]",
			want: []*Secret{
				{
					Source: "mysql_username",
					Target: "mysql_username",
				},
				{
					Source: "mysql_password",
					Target: "mysql_password",
				},
			},
		},
		{
			from: "[ { source: mysql_prod_username, target: mysql_username } ]",
			want: []*Secret{
				{
					Source: "mysql_prod_username",
					Target: "mysql_username",
				},
			},
		},
		{
			from: "[ { source: mysql_prod_username, target: mysql_username }, { source: redis_username, target: redis_username } ]",
			want: []*Secret{
				{
					Source: "mysql_prod_username",
					Target: "mysql_username",
				},
				{
					Source: "redis_username",
					Target: "redis_username",
				},
			},
		},
	}

	for _, test := range testdata {
		in := []byte(test.from)
		got := Secrets{}
		err := yaml.Unmarshal(in, &got)
		assert.NoError(t, err)
		assert.EqualValues(t, test.want, got.Secrets, "problem parsing secrets %q", test.from)
	}
}
