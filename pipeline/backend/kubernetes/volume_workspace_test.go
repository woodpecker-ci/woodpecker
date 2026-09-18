// Copyright 2026 Woodpecker Authors
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
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	kube_core_v1 "k8s.io/api/core/v1"
	kube_meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"

	"go.woodpecker-ci.org/woodpecker/v3/pipeline/backend/types"
)

// TestWorkspaceVolumeUsesWorkflowPVC ensures that a step which mounts a
// workspace volume under a caller-chosen name references the PVC the backend
// actually creates for the workflow volume, and that unrelated volumes are
// left alone.
func TestWorkspaceVolumeUsesWorkflowPVC(t *testing.T) {
	namespace := "foo"
	workflowVolume := "wp_workflow_0_default"

	engine := kube{
		config: &config{
			Namespace:    namespace,
			StorageClass: "hdd",
			VolumeSize:   "1G",
		},
		client: fake.NewClientset(),
	}

	// Mirrors the volumes of `woodpecker-cli exec` run in local mode: a
	// workspace volume named by the caller plus the local repository path.
	step := &types.Step{
		UUID:          "01he8bebctabr3kgk0qj36d2me",
		Name:          "test",
		Type:          types.StepTypeCommands,
		OrgID:         42,
		WorkspaceBase: "/woodpecker",
		Volumes: []string{
			"wp_cli_prefix_default:/woodpecker",
			"/home/user/repo:/woodpecker/src",
			"woodpecker-cache:/cache",
			"/host/data:/data",
		},
	}
	conf := &types.Config{
		Volume: workflowVolume,
		Stages: []*types.Stage{{Steps: []*types.Step{step}}},
	}

	require.NoError(t, engine.SetupWorkflow(context.Background(), conf, taskUUID))

	_, err := engine.client.CoreV1().PersistentVolumeClaims(namespace).Get(
		context.Background(), "wp-workflow-0-default", kube_meta_v1.GetOptions{},
	)
	require.NoError(t, err, "the workflow PVC should be created during workflow setup")

	require.NoError(t, engine.StartStep(context.Background(), step, taskUUID))

	podName, err := stepToPodName(step)
	require.NoError(t, err)
	pod, err := engine.client.CoreV1().Pods(namespace).Get(context.Background(), podName, kube_meta_v1.GetOptions{})
	require.NoError(t, err)

	claims := map[string]string{}
	for _, volume := range pod.Spec.Volumes {
		if volume.PersistentVolumeClaim != nil {
			claims[volume.Name] = volume.PersistentVolumeClaim.ClaimName
		}
	}

	assert.Equal(t, "wp-workflow-0-default", claims["wp-workflow-0-default"],
		"the workspace volume must reference the PVC created for the workflow volume")
	assert.NotContains(t, claims, "wp-cli-prefix-default",
		"the caller-named workspace volume must not be referenced as a PVC")
	assert.NotContains(t, claims, "home-user-repo",
		"the local repository path cannot be mounted inside the workspace and must not be referenced as a PVC")

	// Ordinary volumes are not workspace volumes and stay untouched.
	assert.Equal(t, "woodpecker-cache", claims["woodpecker-cache"])
	assert.Equal(t, "host-data", claims["host-data"])

	assert.Contains(t, pod.Spec.Containers[0].VolumeMounts, kube_core_v1.VolumeMount{
		Name:      "wp-workflow-0-default",
		MountPath: "/woodpecker",
	})
}
