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
