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
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	kube_core_v1 "k8s.io/api/core/v1"
	kube_meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	kube_runtime "k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes/fake"
	kube_testing "k8s.io/client-go/testing"

	"go.woodpecker-ci.org/woodpecker/v3/pipeline/backend/types"
)

func overlayTestStep() *types.Step {
	return &types.Step{
		Name:          "build",
		Image:         "golang",
		UUID:          "01he8bebctabr3kgk0qj36d2me-0",
		Type:          types.StepTypeCommands,
		WorkingDir:    "/woodpecker/src",
		WorkspaceBase: "/woodpecker",
		WorkspaceMode: types.WorkspaceModeOverlay,
		Environment:   make(map[string]string),
		Volumes:       []string{"workspace:/woodpecker", "cache:/cache"},
	}
}

func overlayTestConfig() *config {
	return &config{
		Namespace:           "woodpecker",
		StorageClass:        "fast",
		VolumeSize:          "10G",
		PermissionInitImage: "busybox:stable-musl",
	}
}

func TestApplyWorkspaceOverlayClone(t *testing.T) {
	step := overlayTestStep()
	cfg := overlayTestConfig()
	pod, err := mkPod(step, cfg, "wp-build", "linux", BackendOptions{}, taskUUID)
	require.NoError(t, err)

	applied := applyWorkspaceOverlay(pod, step, cfg, overlayClone)
	assert.True(t, applied)

	volume := pod.Spec.Volumes[0]
	assert.Equal(t, "workspace", volume.Name)
	assert.Nil(t, volume.PersistentVolumeClaim)
	require.NotNil(t, volume.Ephemeral)
	spec := volume.Ephemeral.VolumeClaimTemplate.Spec
	require.NotNil(t, spec.DataSource)
	assert.Equal(t, "PersistentVolumeClaim", spec.DataSource.Kind)
	assert.Equal(t, "workspace", spec.DataSource.Name)
	assert.Equal(t, "fast", *spec.StorageClassName)
	assert.Equal(t, "10G", spec.Resources.Requests.Storage().String())
	assert.Equal(t, []kube_core_v1.PersistentVolumeAccessMode{kube_core_v1.ReadWriteOnce}, spec.AccessModes)
	// the cache volume stays as it is
	assert.Equal(t, "cache", pod.Spec.Volumes[1].PersistentVolumeClaim.ClaimName)
	assert.Empty(t, pod.Spec.InitContainers)
}

func TestApplyWorkspaceOverlayCopy(t *testing.T) {
	step := overlayTestStep()
	cfg := overlayTestConfig()
	options := BackendOptions{SecurityContext: &SecurityContext{RunAsUser: newInt64(1000), RunAsGroup: newInt64(1000)}}
	pod, err := mkPod(step, cfg, "wp-build", "linux", options, taskUUID)
	require.NoError(t, err)

	applied := applyWorkspaceOverlay(pod, step, cfg, overlayCopy)
	assert.True(t, applied)

	assert.Equal(t, "workspace", pod.Spec.Volumes[0].Name)
	assert.NotNil(t, pod.Spec.Volumes[0].EmptyDir)
	assert.Nil(t, pod.Spec.Volumes[0].PersistentVolumeClaim)
	source := pod.Spec.Volumes[len(pod.Spec.Volumes)-1]
	assert.Equal(t, "workspace-source", source.Name)
	require.NotNil(t, source.PersistentVolumeClaim)
	assert.Equal(t, "workspace", source.PersistentVolumeClaim.ClaimName)
	assert.True(t, source.PersistentVolumeClaim.ReadOnly)

	// the copy runs first, so the permission init container sees the copied workspace
	require.NotEmpty(t, pod.Spec.InitContainers)
	init := pod.Spec.InitContainers[0]
	assert.Equal(t, "init-workspace-overlay", init.Name)
	assert.Equal(t, "busybox:stable-musl", init.Image)
	assert.Equal(t, []string{"cp", "-a", "/workspace-source/.", "/workspace-target/"}, init.Command)
	assert.Equal(t, []kube_core_v1.VolumeMount{
		{Name: "workspace-source", MountPath: "/workspace-source", ReadOnly: true},
		{Name: "workspace", MountPath: "/workspace-target"},
	}, init.VolumeMounts)
	// the copy runs as the user of the step from the pod security context, so the step owns it
	assert.Equal(t, newInt64(1000), pod.Spec.SecurityContext.RunAsUser)
	assert.Nil(t, init.SecurityContext.RunAsUser)
	assert.Equal(t, newBool(false), init.SecurityContext.AllowPrivilegeEscalation)
	assert.Equal(t, []kube_core_v1.Capability{"ALL"}, init.SecurityContext.Capabilities.Drop)
}

func TestApplyWorkspaceOverlayWithoutWorkspaceMount(t *testing.T) {
	step := overlayTestStep()
	step.Type = types.StepTypeService
	cfg := overlayTestConfig()
	disable := false
	pod, err := mkPod(step, cfg, "wp-build", "linux", BackendOptions{WorkspaceVolume: &disable}, taskUUID)
	require.NoError(t, err)

	assert.False(t, applyWorkspaceOverlay(pod, step, cfg, overlayCopy))
	assert.Empty(t, pod.Spec.InitContainers)
}

// cloneProbeClient returns a fake client with a bound workspace claim. It reports the
// claim of the clone probe as bound, or a ProvisioningFailed event for it if fail is set.
func cloneProbeClient(t *testing.T, bound, fail bool) (*fake.Clientset, *atomic.Int32) {
	t.Helper()
	client := fake.NewClientset(&kube_core_v1.PersistentVolumeClaim{
		Name:      "workspace",
		Namespace: "woodpecker",
		Status:    kube_core_v1.PersistentVolumeClaimStatus{Phase: kube_core_v1.ClaimBound},
	})
	var probeName atomic.Pointer[string]
	probeName.Store(new(""))
	probes := new(atomic.Int32)
	client.PrependReactor("create", "persistentvolumeclaims", func(action kube_testing.Action) (bool, kube_runtime.Object, error) {
		create, ok := action.(kube_testing.CreateAction)
		require.True(t, ok)
		pvc, ok := create.GetObject().(*kube_core_v1.PersistentVolumeClaim)
		require.True(t, ok)
		if strings.HasPrefix(pvc.Name, cloneProbePrefix) {
			probes.Add(1)
			probeName.Store(&pvc.Name)
		}
		return false, nil, nil
	})
	client.PrependReactor("get", "persistentvolumeclaims", func(action kube_testing.Action) (bool, kube_runtime.Object, error) {
		get, ok := action.(kube_testing.GetAction)
		require.True(t, ok)
		name := get.GetName()
		if bound && strings.HasPrefix(name, cloneProbePrefix) {
			return true, &kube_core_v1.PersistentVolumeClaim{
				Name:      name,
				Namespace: "woodpecker",
				Status:    kube_core_v1.PersistentVolumeClaimStatus{Phase: kube_core_v1.ClaimBound},
			}, nil
		}
		return false, nil, nil
	})
	client.PrependReactor("list", "events", func(kube_testing.Action) (bool, kube_runtime.Object, error) {
		if !fail {
			return true, &kube_core_v1.EventList{}, nil
		}
		return true, &kube_core_v1.EventList{Items: []kube_core_v1.Event{{
			InvolvedObject: kube_core_v1.ObjectReference{Kind: "PersistentVolumeClaim", Name: *probeName.Load()},
			Reason:         "ProvisioningFailed",
			Message:        "CSI driver does not support clone operations: controller CLONE_VOLUME capability is not reported",
		}}}, nil
	})
	return client, probes
}

func setCloneProbeTiming(t *testing.T, timeout time.Duration) {
	t.Helper()
	oldTimeout, oldInterval := cloneProbeTimeout, cloneProbeInterval
	cloneProbeTimeout, cloneProbeInterval = timeout, time.Millisecond
	t.Cleanup(func() { cloneProbeTimeout, cloneProbeInterval = oldTimeout, oldInterval })
}

func TestWorkspaceOverlayModeCloneSupported(t *testing.T) {
	setCloneProbeTiming(t, time.Second)
	client, probes := cloneProbeClient(t, true, false)
	engine := &kube{client: client, config: overlayTestConfig()}

	mode, err := engine.workspaceOverlayMode(t.Context(), "woodpecker", "workspace")
	require.NoError(t, err)
	assert.Equal(t, overlayClone, mode)

	// the result is kept per storage class, so there is no second probe
	mode, err = engine.workspaceOverlayMode(t.Context(), "woodpecker", "workspace")
	require.NoError(t, err)
	assert.Equal(t, overlayClone, mode)
	assert.Equal(t, int32(1), probes.Load())

	// the probe claim and pod are removed again
	pvcs, err := client.CoreV1().PersistentVolumeClaims("woodpecker").List(t.Context(), kube_meta_v1.ListOptions{})
	require.NoError(t, err)
	assert.Len(t, pvcs.Items, 1)
	pods, err := client.CoreV1().Pods("woodpecker").List(t.Context(), kube_meta_v1.ListOptions{})
	require.NoError(t, err)
	assert.Empty(t, pods.Items)
}

func TestWorkspaceOverlayModeCloneUnsupported(t *testing.T) {
	setCloneProbeTiming(t, time.Second)
	client, probes := cloneProbeClient(t, false, true)
	engine := &kube{client: client, config: overlayTestConfig()}

	mode, err := engine.workspaceOverlayMode(t.Context(), "woodpecker", "workspace")
	require.NoError(t, err)
	assert.Equal(t, overlayCopy, mode)
	mode, err = engine.workspaceOverlayMode(t.Context(), "woodpecker", "workspace")
	require.NoError(t, err)
	assert.Equal(t, overlayCopy, mode)
	assert.Equal(t, int32(1), probes.Load())
}

func TestWorkspaceOverlayModeProbeTimeout(t *testing.T) {
	setCloneProbeTiming(t, 50*time.Millisecond)
	client, _ := cloneProbeClient(t, false, false)
	engine := &kube{client: client, config: overlayTestConfig()}

	mode, err := engine.workspaceOverlayMode(t.Context(), "woodpecker", "workspace")
	require.NoError(t, err)
	assert.Equal(t, overlayCopy, mode)
}

func TestWorkspaceOverlayModeSourceNotBound(t *testing.T) {
	setCloneProbeTiming(t, time.Second)
	client := fake.NewClientset(&kube_core_v1.PersistentVolumeClaim{
		Name:      "workspace",
		Namespace: "woodpecker",
		Status:    kube_core_v1.PersistentVolumeClaimStatus{Phase: kube_core_v1.ClaimPending},
	})
	engine := &kube{client: client, config: overlayTestConfig()}

	mode, err := engine.workspaceOverlayMode(t.Context(), "woodpecker", "workspace")
	require.NoError(t, err)
	assert.Equal(t, overlayCopy, mode)
	// an unbound source says nothing about the driver, so nothing is kept
	assert.Empty(t, engine.cloneProbe.results)
}

func TestStartStepWorkspaceOverlay(t *testing.T) {
	setCloneProbeTiming(t, time.Second)
	client, _ := cloneProbeClient(t, false, true)
	engine := &kube{client: client, config: overlayTestConfig()}
	step := overlayTestStep()

	require.NoError(t, engine.StartStep(t.Context(), step, taskUUID))

	podName, err := stepToPodName(step)
	require.NoError(t, err)
	pod, err := client.CoreV1().Pods("woodpecker").Get(t.Context(), podName, kube_meta_v1.GetOptions{})
	require.NoError(t, err)
	assert.NotNil(t, pod.Spec.Volumes[0].EmptyDir)
	require.NotEmpty(t, pod.Spec.InitContainers)
	assert.Equal(t, "init-workspace-overlay", pod.Spec.InitContainers[0].Name)
}
