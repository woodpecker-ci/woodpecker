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
	"context"
	"fmt"
	"runtime"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/urfave/cli/v3"
	kube_core_v1 "k8s.io/api/core/v1"
	kube_errors "k8s.io/apimachinery/pkg/api/errors"
	kube_meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	kube_runtime "k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/kubernetes/fake"
	kube_testing "k8s.io/client-go/testing"

	"go.woodpecker-ci.org/woodpecker/v3/pipeline/backend/types"
)

func TestGettingConfig(t *testing.T) {
	engine := kube{
		config: &config{
			Namespace:            "default",
			StorageClass:         "hdd",
			VolumeSize:           "1G",
			StorageRwx:           false,
			PodLabels:            map[string]string{"l1": "v1"},
			PodAnnotations:       map[string]string{"a1": "v1"},
			ImagePullSecretNames: []string{"regcred"},
			SecurityContext:      SecurityContextConfig{RunAsNonRoot: false},
		},
	}
	config := engine.getConfig()
	config.Namespace = "wp"
	config.StorageClass = "ssd"
	config.StorageRwx = true
	config.PodLabels = nil
	config.PodAnnotations["a2"] = "v2"
	config.ImagePullSecretNames = append(config.ImagePullSecretNames, "docker.io")
	config.SecurityContext.RunAsNonRoot = true

	assert.Equal(t, "default", engine.config.Namespace)
	assert.Equal(t, "hdd", engine.config.StorageClass)
	assert.Equal(t, "1G", engine.config.VolumeSize)
	assert.False(t, engine.config.StorageRwx)
	assert.Len(t, engine.config.PodLabels, 1)
	assert.Len(t, engine.config.PodAnnotations, 1)
	assert.Len(t, engine.config.ImagePullSecretNames, 1)
	assert.False(t, engine.config.SecurityContext.RunAsNonRoot)
}

func TestSetupWorkflow(t *testing.T) {
	namespace := "foo"
	volumeName := "volume-name"
	volumePath := volumeName + ":/woodpecker"
	networkName := "test-network"
	taskUUID := "11301"

	engine := kube{
		config: &config{
			Namespace:            namespace,
			StorageClass:         "hdd",
			VolumeSize:           "1G",
			StorageRwx:           false,
			PodLabels:            map[string]string{"l1": "v1"},
			PodAnnotations:       map[string]string{"a1": "v1"},
			ImagePullSecretNames: []string{"regcred"},
			SecurityContext:      SecurityContextConfig{RunAsNonRoot: false},
		},
		client: fake.NewClientset(),
	}

	serviceWithPorts := types.Step{
		OrgID:    42,
		Name:     "service",
		UUID:     "123",
		Type:     types.StepTypeService,
		Volumes:  []string{volumePath},
		Networks: []types.Conn{{Name: networkName, Aliases: []string{"alias"}}},
		Ports: []types.Port{
			{Number: 8080, Protocol: "tcp"},
		},
	}

	conf := &types.Config{
		Volume:  volumePath,
		Network: networkName,
		Stages: []*types.Stage{
			{
				Steps: []*types.Step{
					&serviceWithPorts,
					{
						OrgID:    42,
						UUID:     "234",
						Name:     "service2",
						Type:     types.StepTypeService,
						Volumes:  []string{volumePath},
						Networks: []types.Conn{{Name: networkName, Aliases: []string{"alias"}}},
					},
				},
			},
			{
				Steps: []*types.Step{
					{
						OrgID:    42,
						UUID:     "456",
						Name:     "step-1",
						Volumes:  []string{volumePath},
						Networks: []types.Conn{{Name: networkName, Aliases: []string{"alias"}}},
					},
				},
			},
		},
	}

	err := engine.SetupWorkflow(t.Context(), conf, taskUUID)
	assert.NoError(t, err, "SetupWorkflow should not error with minimal config and fake client")

	_, err = engine.client.CoreV1().PersistentVolumeClaims(namespace).Get(t.Context(), "volume-name", kube_meta_v1.GetOptions{})
	assert.NoError(t, err, "persistent volume should be created during workflow setup")

	_, err = engine.client.CoreV1().Services(namespace).Get(t.Context(), "wp-hsvc-"+taskUUID, kube_meta_v1.GetOptions{})
	assert.NoError(t, err, "headless service should be created during workflow setup")
}

func TestAffinityFromCliContext(t *testing.T) {
	t.Setenv("WOODPECKER_BACKEND_K8S_NAMESPACE", "")
	t.Setenv("WOODPECKER_BACKEND_K8S_POD_AFFINITY", `{
		"podAffinity": {
			"requiredDuringSchedulingIgnoredDuringExecution": [
			{
				"labelSelector": {},
				"matchLabelKeys": [
				"woodpecker-ci.org/task-uuid"
				],
				"topologyKey": "kubernetes.io/hostname"
			}
			]
		}
		}`)
	t.Setenv("WOODPECKER_BACKEND_K8S_POD_AFFINITY_ALLOW_FROM_STEP", "false")

	cmd := &cli.Command{
		Flags: Flags,
		Action: func(ctx context.Context, c *cli.Command) error {
			ctx = context.WithValue(ctx, types.CliCommand, c)
			config, err := configFromCliContext(ctx)

			require.NoError(t, err)
			require.NotNil(t, config)
			assert.False(t, config.PodAffinityAllowFromStep)

			// Verify affinity was parsed
			require.NotNil(t, config.PodAffinity)
			require.NotNil(t, config.PodAffinity.PodAffinity)
			require.Len(t, config.PodAffinity.PodAffinity.RequiredDuringSchedulingIgnoredDuringExecution, 1)

			term := config.PodAffinity.PodAffinity.RequiredDuringSchedulingIgnoredDuringExecution[0]
			assert.Equal(t, "kubernetes.io/hostname", term.TopologyKey)
			assert.Equal(t, []string{"woodpecker-ci.org/task-uuid"}, term.MatchLabelKeys)

			return nil
		},
	}
	err := cmd.Run(t.Context(), []string{"test"})
	require.NoError(t, err)
}

func makeStep(uuid string) *types.Step {
	return &types.Step{
		UUID:  uuid,
		Name:  "step-" + uuid,
		OrgID: 1,
	}
}

func makeEngine(client *fake.Clientset) *kube {
	return &kube{
		client: client,
		config: &config{
			Namespace: "test-ns",
		},
	}
}

func createPod(
	t *testing.T,
	client *fake.Clientset,
	step *types.Step,
	namespace string,
) string {
	t.Helper()
	podName, err := stepToPodName(step)
	require.NoError(t, err)

	pod := &kube_core_v1.Pod{
		Name:      podName,
		Namespace: namespace,
		Status: kube_core_v1.PodStatus{
			Phase: kube_core_v1.PodPending,
		},
	}
	_, err = client.CoreV1().Pods(namespace).Create(t.Context(), pod, kube_meta_v1.CreateOptions{})
	require.NoError(t, err)
	return podName
}

func TestStartStepReturnsBackendOptionsParseError(t *testing.T) {
	client := fake.NewClientset()
	engine := kube{
		client: client,
		config: &config{
			Namespace: "default",
		},
	}

	step := &types.Step{
		Name:  "bad-backend-options",
		Image: "alpine",
		UUID:  "01he8bebctabr3kgk0qj36d2me-0",
		// a plain string cannot be decoded into the kubernetes BackendOptions struct
		BackendOptions: map[string]any{"kubernetes": "not-a-map"},
	}

	err := engine.StartStep(t.Context(), step, "11301")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "could not parse backend options")

	// the step must not be started if the backend options cannot be parsed
	assert.Empty(t, client.Actions())
}

func TestWaitStepReturnsOnContextCancel(t *testing.T) {
	client := fake.NewClientset()
	engine := makeEngine(client)
	step := makeStep("ctx-cancel-01")
	namespace := "test-ns"

	createPod(t, client, step, namespace)

	ctx, cancel := context.WithCancelCause(t.Context())

	type result struct {
		state *types.State
		err   error
	}
	ch := make(chan result, 1)

	go func() {
		s, err := engine.WaitStep(ctx, step, "task-1")
		ch <- result{s, err}
	}()

	// Give the informer time to start and begin watching.
	time.Sleep(200 * time.Millisecond)

	cancel(nil)

	select {
	case r := <-ch:
		assert.Nil(t, r.state)
		assert.ErrorIs(t, r.err, context.Canceled)
	case <-time.After(3 * time.Second):
		t.Fatal("WaitStep did not return after context cancellation")
	}
}

func TestWaitStepReturnsOnAlreadyDeletedPod(t *testing.T) {
	client := fake.NewClientset()
	engine := makeEngine(client)
	step := makeStep("pod-delete-02")
	namespace := "test-ns"

	podName := createPod(t, client, step, namespace)

	// Delete before WaitStep starts
	err := client.CoreV1().Pods(namespace).Delete(t.Context(), podName, kube_meta_v1.DeleteOptions{})
	require.NoError(t, err)

	type result struct {
		state *types.State
		err   error
	}
	ch := make(chan result, 1)
	go func() {
		s, err := engine.WaitStep(t.Context(), step, "task-1")
		ch <- result{s, err}
	}()

	select {
	case r := <-ch:
		require.Error(t, r.err)
		assert.True(t, kube_errors.IsNotFound(r.err))
		assert.Nil(t, r.state)
	case <-time.After(3 * time.Second):
		t.Fatal("WaitStep did not return for already-deleted pod")
	}
}

func TestWaitStepReturnsOnPodDeletedBeforeInformerSync(t *testing.T) {
	client := fake.NewClientset()
	engine := makeEngine(client)
	step := makeStep("pod-delete-03")
	namespace := "test-ns"

	podName := createPod(t, client, step, namespace)

	// Delete the pod inside the informer's initial List, so it never enters the
	// cache and no delete event can arrive; only the guard Get can end the wait.
	deleted := false
	client.PrependReactor("list", "pods", func(action kube_testing.Action) (bool, kube_runtime.Object, error) {
		if deleted {
			return false, nil, nil
		}
		deleted = true
		if err := client.Tracker().Delete(action.GetResource(), namespace, podName); err != nil {
			return true, nil, err
		}
		return false, nil, nil
	})

	ctx, cancel := context.WithTimeout(t.Context(), 3*time.Second)
	defer cancel()

	state, err := engine.WaitStep(ctx, step, "task-1")
	require.Error(t, err)
	assert.True(t, kube_errors.IsNotFound(err))
	require.True(t, deleted, "the informer's List should have removed the pod")
	assert.Nil(t, state)
}

func TestWaitStepReturnsOnAlreadyTerminatedPod(t *testing.T) {
	client := fake.NewClientset()
	engine := makeEngine(client)
	step := makeStep("pod-done-01")
	namespace := "test-ns"

	podName, err := stepToPodName(step)
	require.NoError(t, err)
	_, err = client.CoreV1().Pods(namespace).Create(t.Context(), &kube_core_v1.Pod{
		Name:      podName,
		Namespace: namespace,
		Status: kube_core_v1.PodStatus{
			Phase: kube_core_v1.PodFailed,
			ContainerStatuses: []kube_core_v1.ContainerStatus{{
				State: kube_core_v1.ContainerState{
					Terminated: &kube_core_v1.ContainerStateTerminated{ExitCode: 1},
				},
			}},
		},
	}, kube_meta_v1.CreateOptions{})
	require.NoError(t, err)

	// Below defaultResyncDuration, so a resync Update cannot end the wait instead.
	ctx, cancel := context.WithTimeout(t.Context(), 3*time.Second)
	defer cancel()

	state, err := engine.WaitStep(ctx, step, "task-1")
	require.NoError(t, err)
	require.NotNil(t, state)
	assert.True(t, state.Exited)
	assert.Equal(t, 1, state.ExitCode)
}

func TestWaitStepNoGoroutineLeak(t *testing.T) {
	client := fake.NewClientset()
	engine := makeEngine(client)
	namespace := "test-ns"
	numSteps := 10

	steps := make([]*types.Step, numSteps)
	for i := range numSteps {
		steps[i] = makeStep(fmt.Sprintf("leak-%02d", i))
		createPod(t, client, steps[i], namespace)
	}

	runtime.GC()
	time.Sleep(100 * time.Millisecond)
	baselineGoroutines := runtime.NumGoroutine()

	var wg sync.WaitGroup
	for i := range numSteps {
		wg.Go(func() {
			ctx, cancel := context.WithCancelCause(t.Context())

			go func() {
				_, _ = engine.WaitStep(ctx, steps[i], fmt.Sprintf("task-%d", i))
			}()

			time.Sleep(200 * time.Millisecond)
			cancel(nil)
		})
	}
	wg.Wait()

	time.Sleep(1 * time.Second)

	afterCancelGoroutines := runtime.NumGoroutine()
	leaked := afterCancelGoroutines - baselineGoroutines

	assert.Less(t, leaked, numSteps,
		"goroutines leaked after canceling %d WaitStep calls: got %d leaked",
		numSteps, leaked)
}

// TestWaitStepPodDisappears uses the real informer List/Watch and deletion
// handler. The guard Get emits the event before subsequent Gets see NotFound.
func TestWaitStepPodDisappears(t *testing.T) {
	for _, tt := range []struct {
		name       string
		terminated bool
		exitCode   int32
		reason     string
		initial    bool
		service    bool
		guardOnly  bool
	}{
		{name: "UnobservedDeletion"},
		{name: "ObservedSuccess", terminated: true},
		{name: "ObservedFailure", terminated: true, exitCode: 42},
		{name: "ObservedOOM", terminated: true, exitCode: 137, reason: "OOMKilled"},
		{name: "InitialObservedSuccess", terminated: true, initial: true},
		{name: "InitialObservedFailure", terminated: true, exitCode: 42, initial: true},
		{name: "InitialObservedOOM", terminated: true, exitCode: 137, reason: "OOMKilled", initial: true},
		{name: "ServiceObservedSuccess", terminated: true, service: true},
		{name: "ServiceObservedFailure", terminated: true, exitCode: 42, service: true},
		{name: "ServiceObservedOOM", terminated: true, exitCode: 137, reason: "OOMKilled", service: true},
		{name: "GuardObservedSuccess", terminated: true, guardOnly: true},
		{name: "GuardObservedFailure", terminated: true, exitCode: 42, guardOnly: true},
		{name: "GuardObservedOOM", terminated: true, exitCode: 137, reason: "OOMKilled", guardOnly: true},
	} {
		t.Run(tt.name, func(t *testing.T) {
			client := fake.NewClientset()
			engine := makeEngine(client)
			step := makeStep("disappears")
			if tt.service {
				step.Type = types.StepTypeService
			}
			podName := createPod(t, client, step, "test-ns")
			pod, err := client.CoreV1().Pods("test-ns").Get(t.Context(), podName, kube_meta_v1.GetOptions{})
			require.NoError(t, err)
			runningPod := pod.DeepCopy()
			if tt.terminated {
				pod.Status.Phase = kube_core_v1.PodFailed
				if tt.exitCode == 0 {
					pod.Status.Phase = kube_core_v1.PodSucceeded
				}
				pod.Status.ContainerStatuses = []kube_core_v1.ContainerStatus{{
					State: kube_core_v1.ContainerState{Terminated: &kube_core_v1.ContainerStateTerminated{
						ExitCode: tt.exitCode, Reason: tt.reason,
					}},
				}}
			}
			if tt.initial {
				_, err = client.CoreV1().Pods("test-ns").Update(t.Context(), pod, kube_meta_v1.UpdateOptions{})
				require.NoError(t, err)
			}
			watcher := watch.NewRaceFreeFake()
			defer watcher.Stop()
			watching := make(chan struct{})
			client.PrependWatchReactor("pods", func(kube_testing.Action) (bool, watch.Interface, error) {
				close(watching)
				return true, watcher, nil
			})
			gets := 0
			client.PrependReactor("get", "pods", func(action kube_testing.Action) (bool, kube_runtime.Object, error) {
				gets++
				if gets == 1 && !tt.initial {
					// The informer must be watching before we emit deletion.
					<-watching
					if tt.terminated && !tt.guardOnly {
						watcher.Modify(pod.DeepCopy())
					}
					watcher.Delete(runningPod.DeepCopy())
					watcher.Delete(runningPod.DeepCopy())
					if tt.guardOnly {
						return true, pod.DeepCopy(), nil
					}
					return true, runningPod.DeepCopy(), nil
				}
				return true, nil, kube_errors.NewNotFound(action.GetResource().GroupResource(), podName)
			})
			ctx, cancel := context.WithTimeout(t.Context(), 3*time.Second)
			defer cancel()
			state, err := engine.WaitStep(ctx, step, "task-1")
			require.Positive(t, gets, "WaitStep must perform the API check")
			if !tt.terminated {
				require.Error(t, err)
				assert.True(t, kube_errors.IsNotFound(err))
				assert.Nil(t, state)
				return
			}
			require.NoError(t, err)
			require.NotNil(t, state)
			assert.Equal(t, &types.State{Exited: true, ExitCode: int(tt.exitCode), OOMKilled: tt.reason == "OOMKilled"}, state)
		})
	}
}

func TestWaitStepDelayedTerminatedStatus(t *testing.T) {
	client := fake.NewClientset()
	engine := makeEngine(client)
	step := makeStep("delayed-status")
	podName := createPod(t, client, step, "test-ns")
	pod, err := client.CoreV1().Pods("test-ns").Get(t.Context(), podName, kube_meta_v1.GetOptions{})
	require.NoError(t, err)
	pod.Status.Phase = kube_core_v1.PodSucceeded
	_, err = client.CoreV1().Pods("test-ns").Update(t.Context(), pod, kube_meta_v1.UpdateOptions{})
	require.NoError(t, err)
	gets := 0
	client.PrependReactor("get", "pods", func(kube_testing.Action) (bool, kube_runtime.Object, error) {
		gets++
		result := pod.DeepCopy()
		if gets == 3 {
			result.Status.ContainerStatuses = []kube_core_v1.ContainerStatus{{}}
		}
		if gets >= 4 {
			result.Status.ContainerStatuses = []kube_core_v1.ContainerStatus{{
				State: kube_core_v1.ContainerState{Terminated: &kube_core_v1.ContainerStateTerminated{ExitCode: 42}},
			}}
		}
		return true, result, nil
	})
	ctx, cancel := context.WithTimeout(t.Context(), 5*time.Second)
	defer cancel()
	state, err := engine.WaitStep(ctx, step, "task-1")
	require.NoError(t, err)
	require.NotNil(t, state)
	assert.Equal(t, 42, state.ExitCode)
	assert.Equal(t, 4, gets, "the final status must come from the backoff retry")
}

func TestWaitStepServiceCancellation(t *testing.T) {
	client := fake.NewClientset()
	engine := makeEngine(client)
	step := makeStep("service-cancel")
	step.Type = types.StepTypeService
	createPod(t, client, step, "test-ns")
	ctx, cancel := context.WithCancelCause(t.Context())
	defer cancel(nil)
	client.PrependReactor("get", "pods", func(kube_testing.Action) (bool, kube_runtime.Object, error) {
		cancel(nil)
		return false, nil, nil
	})
	state, err := engine.WaitStep(ctx, step, "task-1")
	assert.Nil(t, state)
	assert.ErrorIs(t, err, context.Canceled)
}

// TestWaitStepRelistDeletion makes the informer discover a missing pod during
// relist, exercising its DeletedFinalStateUnknown path after an expired watch.
func TestWaitStepRelistDeletion(t *testing.T) {
	client := fake.NewClientset()
	engine := makeEngine(client)
	step := makeStep("relist-delete")
	podName := createPod(t, client, step, "test-ns")
	watcher := watch.NewRaceFreeFake()
	defer watcher.Stop()
	watching := make(chan struct{})
	var watchingOnce sync.Once
	client.PrependWatchReactor("pods", func(kube_testing.Action) (bool, watch.Interface, error) {
		watchingOnce.Do(func() { close(watching) })
		return true, watcher, nil
	})
	lists := 0
	client.PrependReactor("list", "pods", func(kube_testing.Action) (bool, kube_runtime.Object, error) {
		lists++
		if lists > 1 {
			return true, &kube_core_v1.PodList{}, nil
		}
		return false, nil, nil
	})
	gets := 0
	client.PrependReactor("get", "pods", func(action kube_testing.Action) (bool, kube_runtime.Object, error) {
		gets++
		if gets == 1 {
			<-watching
			watcher.Error(&kube_meta_v1.Status{Status: kube_meta_v1.StatusFailure, Reason: kube_meta_v1.StatusReasonExpired, Code: 410})
			return false, nil, nil
		}
		return true, nil, kube_errors.NewNotFound(action.GetResource().GroupResource(), podName)
	})
	ctx, cancel := context.WithTimeout(t.Context(), 10*time.Second)
	defer cancel()
	state, err := engine.WaitStep(ctx, step, "task-1")
	require.Error(t, err)
	assert.True(t, kube_errors.IsNotFound(err))
	assert.Nil(t, state)
	assert.GreaterOrEqual(t, lists, 2, "the informer must relist to discover deletion")
	assert.Equal(t, 2, gets, "completion must reach the post-event Get")
}
