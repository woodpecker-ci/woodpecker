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

func TestUnmarshalNetwork(t *testing.T) {
	testdata := []struct {
		from string
		want WorkflowNetwork
	}{
		{
			from: "{ name: foo, driver: bar }",
			want: WorkflowNetwork{
				Name:   "foo",
				Driver: "bar",
			},
		},
		{
			from: "{ name: foo, driver: bar, driver_opts: { baz: qux } }",
			want: WorkflowNetwork{
				Name:   "foo",
				Driver: "bar",
				DriverOpts: map[string]string{
					"baz": "qux",
				},
			},
		},
	}

	for _, test := range testdata {
		in := []byte(test.from)
		got := WorkflowNetwork{}
		err := yaml.Unmarshal(in, &got)
		assert.NoError(t, err)
		assert.EqualValues(t, test.want, got, "problem parsing network %q", test.from)
	}
}

func TestUnmarshalWorkflowNetworks(t *testing.T) {
	testdata := []struct {
		from string
		want []*WorkflowNetwork
	}{
		{
			from: "foo: { driver: bar }",
			want: []*WorkflowNetwork{
				{
					Name:   "foo",
					Driver: "bar",
				},
			},
		},
		{
			from: "foo: { name: baz }",
			want: []*WorkflowNetwork{
				{
					Name:   "baz",
					Driver: "bridge",
				},
			},
		},
		{
			from: "foo: { name: baz, driver: bar }",
			want: []*WorkflowNetwork{
				{
					Name:   "baz",
					Driver: "bar",
				},
			},
		},
	}

	for _, test := range testdata {
		in := []byte(test.from)
		got := WorkflowNetworks{}
		err := yaml.Unmarshal(in, &got)
		assert.NoError(t, err)
		assert.EqualValues(t, test.want, got.WorkflowNetworks, "problem parsing network %q", test.from)
	}
}

func TestUnmarshalNetworkErr(t *testing.T) {
	testdata := []string{
		"foo: { name: [ foo, bar] }",
		"- foo",
	}

	for _, test := range testdata {
		in := []byte(test)
		err := yaml.Unmarshal(in, new(WorkflowNetworks))
		assert.Error(t, err, "wanted error for networks %q", test)
	}
}
