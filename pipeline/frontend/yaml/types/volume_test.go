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

func TestMarshalVolumes(t *testing.T) {
	volumes := []struct {
		volumes  Volumes
		expected string
	}{
		{
			volumes: Volumes{},
			expected: `[]
`,
		},
		{
			volumes: Volumes{
				Volumes: []*Volume{
					{
						Destination: "/in/the/container",
					},
				},
			},
			expected: `- /in/the/container
`,
		},
		{
			volumes: Volumes{
				Volumes: []*Volume{
					{
						Source:      "./a/path",
						Destination: "/in/the/container",
						AccessMode:  "ro",
					},
				},
			},
			expected: `- ./a/path:/in/the/container:ro
`,
		},
		{
			volumes: Volumes{
				Volumes: []*Volume{
					{
						Source:      "./a/path",
						Destination: "/in/the/container",
					},
				},
			},
			expected: `- ./a/path:/in/the/container
`,
		},
		{
			volumes: Volumes{
				Volumes: []*Volume{
					{
						Source:      "./a/path",
						Destination: "/in/the/container",
					},
					{
						Source:      "named",
						Destination: "/in/the/container",
					},
				},
			},
			expected: `- ./a/path:/in/the/container
- named:/in/the/container
`,
		},
	}
	for _, volume := range volumes {
		bytes, err := yaml.Marshal(volume.volumes)
		assert.NoError(t, err)
		assert.Equal(t, volume.expected, string(bytes), "should be equal")
	}
}

func TestUnmarshalVolumes(t *testing.T) {
	volumes := []struct {
		yaml     string
		expected *Volumes
	}{
		{
			yaml: `- ./a/path:/in/the/container`,
			expected: &Volumes{
				Volumes: []*Volume{
					{
						Source:      "./a/path",
						Destination: "/in/the/container",
					},
				},
			},
		},
		{
			yaml: `- /in/the/container`,
			expected: &Volumes{
				Volumes: []*Volume{
					{
						Destination: "/in/the/container",
					},
				},
			},
		},
		{
			yaml: `- /a/path:/in/the/container:ro`,
			expected: &Volumes{
				Volumes: []*Volume{
					{
						Source:      "/a/path",
						Destination: "/in/the/container",
						AccessMode:  "ro",
					},
				},
			},
		},
		{
			yaml: `- /a/path:/in/the/container
- named:/somewhere/in/the/container`,
			expected: &Volumes{
				Volumes: []*Volume{
					{
						Source:      "/a/path",
						Destination: "/in/the/container",
					},
					{
						Source:      "named",
						Destination: "/somewhere/in/the/container",
					},
				},
			},
		},
	}
	for _, volume := range volumes {
		actual := &Volumes{}
		err := yaml.Unmarshal([]byte(volume.yaml), actual)
		assert.NoError(t, err)
		assert.Equal(t, volume.expected, actual, "should be equal")
	}
}
