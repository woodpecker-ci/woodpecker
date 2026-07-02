package kubernetes

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"go.woodpecker-ci.org/woodpecker/v3/pipeline/backend/types"
)

// A step's nodeSelector is only applied when the admin enables
// PodNodeSelectorAllowFromStep, matching the other *AllowFromStep pod fields.
func TestNodeSelectorAllowFromStep(t *testing.T) {
	mk := func(allow bool) map[string]string {
		pod, err := mkPod(
			&types.Step{Name: "build", Image: "golang:1.24", UUID: "01he8bebctabr3kgk0qj36d2me-0"},
			&config{Namespace: "woodpecker", PodNodeSelectorAllowFromStep: allow},
			"wp-01he8bebctabr3kgk0qj36d2me-0",
			"linux/amd64",
			BackendOptions{NodeSelector: map[string]string{"attacker-target": "sensitive-node"}},
			taskUUID,
		)
		require.NoError(t, err)
		return pod.Spec.NodeSelector
	}

	// disabled: the step value is ignored.
	assert.NotContains(t, mk(false), "attacker-target")

	// enabled: the step value is applied.
	assert.Equal(t, "sensitive-node", mk(true)["attacker-target"])
}
