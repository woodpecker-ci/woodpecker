// Copyright 2024 Woodpecker Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package kubernetes

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"go.woodpecker-ci.org/woodpecker/v3/pipeline/backend/types"
)

func TestNodeSelectorAllowFromStep(t *testing.T) {
	step := &types.Step{
		Name:  "ns-test",
		Image: "alpine",
		UUID:  "01he8bebctabr3kgk0qj36d2me-0",
	}

	// When disabled (default), a step-provided node selector must be ignored.
	pod, err := mkPod(step, &config{
		Namespace: "woodpecker",
	}, "wp-01he8bebctabr3kgk0qj36d2me-0", "linux/amd64", BackendOptions{
		NodeSelector: map[string]string{"attacker-target": "sensitive-node"},
	}, taskUUID)
	assert.NoError(t, err)
	assert.NotContains(t, pod.Spec.NodeSelector, "attacker-target")

	// When explicitly enabled by the admin, the step value is honored.
	pod, err = mkPod(step, &config{
		Namespace:                    "woodpecker",
		PodNodeSelectorAllowFromStep: true,
	}, "wp-01he8bebctabr3kgk0qj36d2me-0", "linux/amd64", BackendOptions{
		NodeSelector: map[string]string{"attacker-target": "sensitive-node"},
	}, taskUUID)
	assert.NoError(t, err)
	assert.Equal(t, "sensitive-node", pod.Spec.NodeSelector["attacker-target"])
}
