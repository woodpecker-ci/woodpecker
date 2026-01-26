// Copyright 2026 Woodpecker Authors
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

package kubernetes

import (
	"context"
	"encoding/json"
	"strconv"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"

	"go.woodpecker-ci.org/woodpecker/v3/pipeline/backend/types"
)

func TestStateConfigMapName(t *testing.T) {
	assert.Equal(t, "wp-state-task-123", stateConfigMapName("task-123"))
	assert.Equal(t, "wp-state-abc", stateConfigMapName("abc"))
}

func TestCreateWorkflowState(t *testing.T) {
	namespace := "test-ns"
	taskUUID := "task-123"
	timeout := 10 * time.Minute

	engine := &kube{
		config: &config{
			Namespace:            namespace,
			StateRecoveryEnabled: true,
		},
		client: fake.NewClientset(),
	}

	conf := &types.Config{
		Volume: "test-volume",
		Stages: []*types.Stage{
			{
				Steps: []*types.Step{
					{UUID: "step-1", Name: "clone", OrgID: 1},
					{UUID: "step-2", Name: "build", OrgID: 1},
				},
			},
			{
				Steps: []*types.Step{
					{UUID: "step-3", Name: "test", OrgID: 1},
				},
			},
		},
	}

	err := engine.createWorkflowState(context.Background(), conf, taskUUID, timeout)
	require.NoError(t, err)

	// Verify ConfigMap was created
	cm, err := engine.client.CoreV1().ConfigMaps(namespace).Get(context.Background(), stateConfigMapName(taskUUID), meta_v1.GetOptions{})
	require.NoError(t, err)
	assert.NotNil(t, cm)

	assert.Equal(t, taskUUID, cm.Labels[TaskUUIDLabel])
	assert.Equal(t, "active", cm.Labels[WorkflowStateLabel])

	// Verify data
	assert.Equal(t, namespace, cm.Data[stateKeyNamespace])
	assert.Equal(t, "1", cm.Data[stateKeyOrgID])
	assert.Equal(t, "test-volume", cm.Data[stateKeyVolume])

	// Verify steps
	var steps map[string]*StepState
	err = json.Unmarshal([]byte(cm.Data[stateKeySteps]), &steps)
	require.NoError(t, err)
	assert.Len(t, steps, 3)

	assert.Equal(t, types.StatusPending, steps["step-1"].Status)
	assert.Equal(t, types.StatusPending, steps["step-2"].Status)
	assert.Equal(t, types.StatusPending, steps["step-3"].Status)
}

func TestCreateWorkflowStateAlreadyExists(t *testing.T) {
	namespace := "test-ns"
	taskUUID := "task-123"

	engine := &kube{
		config: &config{
			Namespace: namespace,
		},
		client: fake.NewClientset(),
	}

	conf := &types.Config{
		Volume: "test-volume",
		Stages: []*types.Stage{
			{
				Steps: []*types.Step{
					{UUID: "step-1", Name: "clone", OrgID: 1},
				},
			},
		},
	}

	// Create first time
	err := engine.createWorkflowState(context.Background(), conf, taskUUID, 10*time.Minute)
	require.NoError(t, err)

	// Create again - should not error
	err = engine.createWorkflowState(context.Background(), conf, taskUUID, 10*time.Minute)
	assert.NoError(t, err)
}

func TestCreateWorkflowStateInvalidConfig(t *testing.T) {
	engine := &kube{
		config: &config{
			Namespace: "test-ns",
		},
		client: fake.NewClientset(),
	}

	// nil config
	err := engine.createWorkflowState(context.Background(), nil, "task-123", 10*time.Minute)
	assert.Error(t, err)

	// empty stages
	err = engine.createWorkflowState(context.Background(), &types.Config{Stages: []*types.Stage{}}, "task-123", 10*time.Minute)
	assert.Error(t, err)

	// empty steps
	err = engine.createWorkflowState(context.Background(), &types.Config{Stages: []*types.Stage{{Steps: []*types.Step{}}}}, "task-123", 10*time.Minute)
	assert.Error(t, err)
}

func TestMarkStepStarted(t *testing.T) {
	namespace := "test-ns"
	taskUUID := "task-123"

	engine := &kube{
		config: &config{
			Namespace: namespace,
		},
		client: fake.NewClientset(),
	}

	// Create initial state
	conf := &types.Config{
		Volume: "test-volume",
		Stages: []*types.Stage{
			{
				Steps: []*types.Step{
					{UUID: "step-1", Name: "clone", OrgID: 1},
				},
			},
		},
	}
	err := engine.createWorkflowState(context.Background(), conf, taskUUID, 10*time.Minute)
	require.NoError(t, err)

	// Mark step started
	step := &types.Step{UUID: "step-1", Name: "clone"}
	err = engine.markStepStarted(context.Background(), taskUUID, namespace, step)
	require.NoError(t, err)

	// Verify state
	state, err := engine.getWorkflowState(context.Background(), taskUUID, namespace)
	require.NoError(t, err)
	assert.Equal(t, types.StatusRunning, state.Steps["step-1"].Status)
	assert.Greater(t, state.Steps["step-1"].Started, int64(0))
}

func TestMarkStepCompleted(t *testing.T) {
	namespace := "test-ns"
	taskUUID := "task-123"

	engine := &kube{
		config: &config{
			Namespace: namespace,
		},
		client: fake.NewClientset(),
	}

	// Create initial state
	conf := &types.Config{
		Volume: "test-volume",
		Stages: []*types.Stage{
			{
				Steps: []*types.Step{
					{UUID: "step-1", Name: "clone", OrgID: 1},
				},
			},
		},
	}
	err := engine.createWorkflowState(context.Background(), conf, taskUUID, 10*time.Minute)
	require.NoError(t, err)

	// Mark step completed with success
	err = engine.markStepCompleted(context.Background(), taskUUID, namespace, "step-1", 0)
	require.NoError(t, err)

	state, err := engine.getWorkflowState(context.Background(), taskUUID, namespace)
	require.NoError(t, err)
	assert.Equal(t, types.StatusSuccess, state.Steps["step-1"].Status)
	assert.Equal(t, 0, state.Steps["step-1"].ExitCode)
	assert.Greater(t, state.Steps["step-1"].Finished, int64(0))
}

func TestMarkStepCompletedWithFailure(t *testing.T) {
	namespace := "test-ns"
	taskUUID := "task-123"

	engine := &kube{
		config: &config{
			Namespace: namespace,
		},
		client: fake.NewClientset(),
	}

	conf := &types.Config{
		Volume: "test-volume",
		Stages: []*types.Stage{
			{
				Steps: []*types.Step{
					{UUID: "step-1", Name: "build", OrgID: 1},
				},
			},
		},
	}
	err := engine.createWorkflowState(context.Background(), conf, taskUUID, 10*time.Minute)
	require.NoError(t, err)

	// Mark step completed with failure
	err = engine.markStepCompleted(context.Background(), taskUUID, namespace, "step-1", 1)
	require.NoError(t, err)

	state, err := engine.getWorkflowState(context.Background(), taskUUID, namespace)
	require.NoError(t, err)
	assert.Equal(t, types.StatusFailed, state.Steps["step-1"].Status)
	assert.Equal(t, 1, state.Steps["step-1"].ExitCode)
}

func TestMarkStepSkipped(t *testing.T) {
	namespace := "test-ns"
	taskUUID := "task-123"

	engine := &kube{
		config: &config{
			Namespace: namespace,
		},
		client: fake.NewClientset(),
	}

	conf := &types.Config{
		Volume: "test-volume",
		Stages: []*types.Stage{
			{
				Steps: []*types.Step{
					{UUID: "step-1", Name: "deploy", OrgID: 1},
				},
			},
		},
	}
	err := engine.createWorkflowState(context.Background(), conf, taskUUID, 10*time.Minute)
	require.NoError(t, err)

	err = engine.markStepSkipped(context.Background(), taskUUID, namespace, "step-1")
	require.NoError(t, err)

	state, err := engine.getWorkflowState(context.Background(), taskUUID, namespace)
	require.NoError(t, err)
	assert.Equal(t, types.StatusSkipped, state.Steps["step-1"].Status)
}

func TestDeleteWorkflowState(t *testing.T) {
	namespace := "test-ns"
	taskUUID := "task-123"

	engine := &kube{
		config: &config{
			Namespace: namespace,
		},
		client: fake.NewClientset(),
	}

	conf := &types.Config{
		Volume: "test-volume",
		Stages: []*types.Stage{
			{Steps: []*types.Step{{UUID: "step-1", OrgID: 1}}},
		},
	}
	err := engine.createWorkflowState(context.Background(), conf, taskUUID, 10*time.Minute)
	require.NoError(t, err)

	// Delete
	err = engine.deleteWorkflowState(context.Background(), taskUUID, namespace)
	require.NoError(t, err)

	// Verify it's gone
	_, err = engine.client.CoreV1().ConfigMaps(namespace).Get(context.Background(), stateConfigMapName(taskUUID), meta_v1.GetOptions{})
	assert.True(t, err != nil)
}

func TestDeleteWorkflowStateNotFound(t *testing.T) {
	engine := &kube{
		config: &config{
			Namespace: "test-ns",
		},
		client: fake.NewClientset(),
	}

	// Deleting non-existent ConfigMap should not error
	err := engine.deleteWorkflowState(context.Background(), "nonexistent", "test-ns")
	assert.NoError(t, err)
}

func TestGetWorkflowState(t *testing.T) {
	namespace := "test-ns"
	taskUUID := "task-123"

	engine := &kube{
		config: &config{
			Namespace: namespace,
		},
		client: fake.NewClientset(),
	}

	conf := &types.Config{
		Volume: "test-volume",
		Stages: []*types.Stage{
			{
				Steps: []*types.Step{
					{UUID: "step-1", Name: "clone", OrgID: 42},
					{UUID: "step-2", Name: "build", OrgID: 42},
				},
			},
		},
	}
	err := engine.createWorkflowState(context.Background(), conf, taskUUID, 10*time.Minute)
	require.NoError(t, err)

	state, err := engine.getWorkflowState(context.Background(), taskUUID, namespace)
	require.NoError(t, err)

	// TaskUUID, Namespace, OrgID are no longer stored in WorkflowState for size optimization
	// They are tracked separately in the ConfigMap metadata and calling context
	assert.Equal(t, "test-volume", state.Volume)
	assert.Len(t, state.Steps, 2)
	assert.Greater(t, state.Timeout, int64(0))
	assert.Greater(t, state.Started, int64(0))
}

func TestRecordStepStartedDisabled(t *testing.T) {
	engine := &kube{
		config: &config{
			StateRecoveryEnabled: false,
		},
		client: fake.NewClientset(),
	}

	err := engine.RecordStepStarted(context.Background(), "task-1", &types.Step{UUID: "step-1"})
	assert.NoError(t, err) // Should no-op without error
}

func TestRecordStepCompletedDisabled(t *testing.T) {
	engine := &kube{
		config: &config{
			StateRecoveryEnabled: false,
		},
		client: fake.NewClientset(),
	}

	err := engine.RecordStepCompleted(context.Background(), "task-1", &types.Step{UUID: "step-1"}, 0)
	assert.NoError(t, err) // Should no-op without error
}

func TestRecordStepSkippedDisabled(t *testing.T) {
	engine := &kube{
		config: &config{
			StateRecoveryEnabled: false,
		},
		client: fake.NewClientset(),
	}

	err := engine.RecordStepSkipped(context.Background(), "task-1", &types.Step{UUID: "step-1"})
	assert.NoError(t, err) // Should no-op without error
}

func TestUpdateStepStateNotFound(t *testing.T) {
	engine := &kube{
		config: &config{
			Namespace: "test-ns",
		},
		client: fake.NewClientset(),
	}

	// Updating non-existent ConfigMap should not error (graceful handling)
	err := engine.updateStepState(context.Background(), "nonexistent", "test-ns", "step-1", func(s *StepState) {
		s.Status = types.StatusRunning
	})
	assert.NoError(t, err)
}

func TestStepStatusConstants(t *testing.T) {
	assert.Equal(t, "pending", types.StatusPending.String())
	assert.Equal(t, "running", types.StatusRunning.String())
	assert.Equal(t, "success", types.StatusSuccess.String())
	assert.Equal(t, "failed", types.StatusFailed.String())
	assert.Equal(t, "skipped", types.StatusSkipped.String())
}

func TestGetStepStatus(t *testing.T) {
	namespace := "test-ns"
	taskUUID := "task-123"

	engine := &kube{
		config: &config{
			Namespace:            namespace,
			StateRecoveryEnabled: true,
		},
		client: fake.NewClientset(),
	}

	conf := &types.Config{
		Volume: "vol-1",
		Stages: []*types.Stage{
			{Steps: []*types.Step{
				{UUID: "step-1", Name: "clone", OrgID: 1},
				{UUID: "step-2", Name: "build", OrgID: 1},
			}},
		},
	}
	err := engine.createWorkflowState(context.Background(), conf, taskUUID, 10*time.Minute)
	require.NoError(t, err)

	err = engine.markStepCompleted(context.Background(), taskUUID, namespace, "step-1", 0)
	require.NoError(t, err)
	err = engine.markStepStarted(context.Background(), taskUUID, namespace, &types.Step{UUID: "step-2", Name: "build"})
	require.NoError(t, err)

	status, err := engine.GetStepStatus(context.Background(), taskUUID, "step-1")
	require.NoError(t, err)
	assert.Equal(t, types.StatusSuccess, status)

	status, err = engine.GetStepStatus(context.Background(), taskUUID, "step-2")
	require.NoError(t, err)
	assert.Equal(t, types.StatusRunning, status)

	status, err = engine.GetStepStatus(context.Background(), taskUUID, "nonexistent")
	require.NoError(t, err)
	assert.Equal(t, types.StatusUnknown, status)
}

func TestGetStepStatusDisabled(t *testing.T) {
	engine := &kube{
		config: &config{
			Namespace:            "test-ns",
			StateRecoveryEnabled: false,
		},
		client: fake.NewClientset(),
	}

	status, err := engine.GetStepStatus(context.Background(), "task-1", "step-1")
	require.NoError(t, err)
	assert.Equal(t, types.StatusUnknown, status)
}

func TestGetStepStatusWorkflowNotFound(t *testing.T) {
	engine := &kube{
		config: &config{
			Namespace:            "test-ns",
			StateRecoveryEnabled: true,
		},
		client: fake.NewClientset(),
	}

	status, err := engine.GetStepStatus(context.Background(), "nonexistent", "step-1")
	require.NoError(t, err)
	assert.Equal(t, types.StatusUnknown, status)
}

func TestCleanupExpiredStateConfigMaps(t *testing.T) {
	namespace := "test-ns"
	now := time.Now().Unix()

	engine := &kube{
		config: &config{
			Namespace:            namespace,
			StateRecoveryEnabled: true,
		},
		client: fake.NewClientset(),
	}

	// Create ConfigMaps with different expiry states
	configMaps := []struct {
		name        string
		ttl         int64
		shouldExist bool
	}{
		{"wp-state-expired-1", now - 3600, false}, // Expired 1 hour ago
		{"wp-state-expired-2", now - 1, false},    // Expired 1 second ago
		{"wp-state-valid-1", now + 3600, true},    // Expires in 1 hour
		{"wp-state-valid-2", now + 86400, true},   // Expires in 1 day
		{"wp-state-no-ttl", 0, true},              // No TTL annotation
		{"wp-state-invalid-ttl", 0, true},         // Invalid TTL annotation
		{"wp-state-just-expired", now, false},     // Expires exactly now
	}

	for _, tc := range configMaps {
		cm := &v1.ConfigMap{
			ObjectMeta: meta_v1.ObjectMeta{
				Name:      tc.name,
				Namespace: namespace,
				Labels: map[string]string{
					WorkflowStateLabel: "active",
				},
				Annotations: map[string]string{},
			},
			Data: map[string]string{
				"steps": "{}",
			},
		}

		switch tc.name {
		case "wp-state-no-ttl":
			// Don't add TTL annotation
		case "wp-state-invalid-ttl":
			cm.Annotations[TTLExpiresAtLabel] = "invalid"
		default:
			cm.Annotations[TTLExpiresAtLabel] = strconv.FormatInt(tc.ttl, 10)
		}

		_, err := engine.client.CoreV1().ConfigMaps(namespace).Create(context.Background(), cm, meta_v1.CreateOptions{})
		require.NoError(t, err)
	}

	// Create a ConfigMap without the workflow-state label (should be ignored)
	nonStateCM := &v1.ConfigMap{
		ObjectMeta: meta_v1.ObjectMeta{
			Name:      "wp-state-other",
			Namespace: namespace,
			Labels:    map[string]string{},
		},
	}
	_, err := engine.client.CoreV1().ConfigMaps(namespace).Create(context.Background(), nonStateCM, meta_v1.CreateOptions{})
	require.NoError(t, err)

	// Run cleanup
	engine.cleanupExpiredStateConfigMaps(context.Background(), namespace)

	// Verify results
	for _, tc := range configMaps {
		_, err := engine.client.CoreV1().ConfigMaps(namespace).Get(context.Background(), tc.name, meta_v1.GetOptions{})
		if tc.shouldExist {
			assert.NoError(t, err, "ConfigMap %s should still exist", tc.name)
		} else {
			assert.True(t, errors.IsNotFound(err), "ConfigMap %s should be deleted", tc.name)
		}
	}

	// Verify non-state ConfigMap is untouched
	_, err = engine.client.CoreV1().ConfigMaps(namespace).Get(context.Background(), "wp-state-other", meta_v1.GetOptions{})
	assert.NoError(t, err, "Non-state ConfigMap should be untouched")
}
