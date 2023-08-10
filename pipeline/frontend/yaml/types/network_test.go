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

	"gopkg.in/yaml.v3"

	"github.com/stretchr/testify/assert"
)

func TestMarshalNetworks(t *testing.T) {
	networks := []struct {
		networks Networks
		expected string
	}{
		{
			networks: Networks{},
			expected: "{}\n",
		},
		{
			networks: Networks{
				Networks: []*Network{
					{
						Name: "network1",
					},
					{
						Name: "network2",
					},
				},
			},
			expected: `network1: {}
network2: {}
`,
		},
		{
			networks: Networks{
				Networks: []*Network{
					{
						Name:    "network1",
						Aliases: []string{"alias1", "alias2"},
					},
					{
						Name: "network2",
					},
				},
			},
			expected: `network1:
    aliases:
        - alias1
        - alias2
network2: {}
`,
		},
		{
			networks: Networks{
				Networks: []*Network{
					{
						Name:    "network1",
						Aliases: []string{"alias1", "alias2"},
					},
					{
						Name:        "network2",
						IPv4Address: "172.16.238.10",
						IPv6Address: "2001:3984:3989::10",
					},
				},
			},
			expected: `network1:
    aliases:
        - alias1
        - alias2
network2:
    ipv4_address: 172.16.238.10
    ipv6_address: 2001:3984:3989::10
`,
		},
	}
	for _, network := range networks {
		bytes, err := yaml.Marshal(network.networks)
		assert.Nil(t, err)
		assert.Equal(t, network.expected, string(bytes), "should be equal")
	}
}

func TestUnmarshalNetworks(t *testing.T) {
	networks := []struct {
		yaml     string
		expected *Networks
	}{
		{
			yaml: `- network1
- network2`,
			expected: &Networks{
				Networks: []*Network{
					{
						Name: "network1",
					},
					{
						Name: "network2",
					},
				},
			},
		},
		{
			yaml: `network1:`,
			expected: &Networks{
				Networks: []*Network{
					{
						Name: "network1",
					},
				},
			},
		},
		{
			yaml: `network1: {}`,
			expected: &Networks{
				Networks: []*Network{
					{
						Name: "network1",
					},
				},
			},
		},
		{
			yaml: `network1:
  aliases:
    - alias1
    - alias2`,
			expected: &Networks{
				Networks: []*Network{
					{
						Name:    "network1",
						Aliases: []string{"alias1", "alias2"},
					},
				},
			},
		},
		{
			yaml: `network1:
  aliases:
    - alias1
    - alias2
  ipv4_address: 172.16.238.10
  ipv6_address: 2001:3984:3989::10`,
			expected: &Networks{
				Networks: []*Network{
					{
						Name:        "network1",
						Aliases:     []string{"alias1", "alias2"},
						IPv4Address: "172.16.238.10",
						IPv6Address: "2001:3984:3989::10",
					},
				},
			},
		},
	}
	for _, network := range networks {
		actual := &Networks{}
		err := yaml.Unmarshal([]byte(network.yaml), actual)
		assert.NoError(t, err)
		assert.EqualValues(t, network.expected, actual)
	}
}
