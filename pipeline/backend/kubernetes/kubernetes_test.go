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
	kube_meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"

	"go.woodpecker-ci.org/woodpecker/v3/pipeline/backend/types"
)

func TestGettingConfig(t *testing.T) {
	engine := kube{
		config: &config{
			Namespace:              "default",
			StorageClass:           "hdd",
			VolumeSize:             "1G",
			StorageRwx:             false,
			PodLabels:              map[string]string{"l1": "v1"},
			PodAnnotations:         map[string]string{"a1": "v1"},
			ImagePullSecretNames:   []string{"regcred"},
			DefaultSecurityContext: SecurityContext{RunAsNonRoot: newBool(false)},
		},
	}
	config := engine.getConfig()
	config.Namespace = "wp"
	config.StorageClass = "ssd"
	config.StorageRwx = true
	config.PodLabels = nil
	config.PodAnnotations["a2"] = "v2"
	config.ImagePullSecretNames = append(config.ImagePullSecretNames, "docker.io")
	config.DefaultSecurityContext.RunAsNonRoot = newBool(true)

	assert.Equal(t, "default", engine.config.Namespace)
	assert.Equal(t, "hdd", engine.config.StorageClass)
	assert.Equal(t, "1G", engine.config.VolumeSize)
	assert.False(t, engine.config.StorageRwx)
	assert.Len(t, engine.config.PodLabels, 1)
	assert.Len(t, engine.config.PodAnnotations, 1)
	assert.Len(t, engine.config.ImagePullSecretNames, 1)
	assert.False(t, *engine.config.DefaultSecurityContext.RunAsNonRoot)
}

func TestSetupWorkflow(t *testing.T) {
	namespace := "foo"
	volumeName := "volume-name"
	volumePath := volumeName + ":/woodpecker"
	networkName := "test-network"
	taskUUID := "11301"

	engine := kube{
		config: &config{
			Namespace:              namespace,
			StorageClass:           "hdd",
			VolumeSize:             "1G",
			StorageRwx:             false,
			PodLabels:              map[string]string{"l1": "v1"},
			PodAnnotations:         map[string]string{"a1": "v1"},
			ImagePullSecretNames:   []string{"regcred"},
			DefaultSecurityContext: SecurityContext{RunAsNonRoot: newBool(false)},
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

	err := engine.SetupWorkflow(context.Background(), conf, taskUUID)
	assert.NoError(t, err, "SetupWorkflow should not error with minimal config and fake client")

	_, err = engine.client.CoreV1().PersistentVolumeClaims(namespace).Get(context.Background(), "volume-name", kube_meta_v1.GetOptions{})
	assert.NoError(t, err, "persistent volume should be created during workflow setup")

	_, err = engine.client.CoreV1().Services(namespace).Get(context.Background(), "wp-hsvc-"+taskUUID, kube_meta_v1.GetOptions{})
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
	err := cmd.Run(context.Background(), []string{"test"})
	require.NoError(t, err)
}

func TestSecctxNonrootFromCliContext(t *testing.T) {
	t.Setenv("WOODPECKER_BACKEND_K8S_SECCTX_NONROOT", "true")

	cmd := &cli.Command{
		Flags: Flags,
		Action: func(ctx context.Context, c *cli.Command) error {
			ctx = context.WithValue(ctx, types.CliCommand, c)
			config, err := configFromCliContext(ctx)

			require.NoError(t, err)
			require.NotNil(t, config)

			// Verify security context was parsed
			require.NotNil(t, config.EnforcedSecurityContext)
			assert.True(t, *config.EnforcedSecurityContext.RunAsNonRoot)
			return nil
		},
	}
	err := cmd.Run(context.Background(), []string{"test"})
	require.NoError(t, err)
}

func TestSecurityContextFromCliContext(t *testing.T) {
	t.Setenv("WOODPECKER_BACKEND_K8S_DEFAULT_SECCTX", `{
		"runAsUser":1000,
		"runAsGroup":1000,
		"fsGroup":1000,
		"fsGroupChangePolicy": "OnRootMismatch"
	}`)
	t.Setenv("WOODPECKER_BACKEND_K8S_ENFORCED_SECCTX", `{
		"privileged":false,
		"runAsNonRoot":true,
		"allowPrivilegeEscalation":false,
		"seccompProfile": {"type": "RuntimeDefault"},
		"capabilities": {"drop": ["ALL"]}
	}`)

	cmd := &cli.Command{
		Flags: Flags,
		Action: func(ctx context.Context, c *cli.Command) error {
			ctx = context.WithValue(ctx, types.CliCommand, c)
			config, err := configFromCliContext(ctx)

			require.NoError(t, err)
			require.NotNil(t, config)

			// Verify security context was parsed
			require.NotNil(t, config.DefaultSecurityContext)
			require.NotNil(t, config.EnforcedSecurityContext)

			assert.Equal(t, (int64)(1000), *config.DefaultSecurityContext.RunAsUser)
			assert.Equal(t, (int64)(1000), *config.DefaultSecurityContext.RunAsGroup)
			assert.Equal(t, (int64)(1000), *config.DefaultSecurityContext.FSGroup)
			assert.Equal(t, kube_core_v1.PodFSGroupChangePolicy("OnRootMismatch"), *config.DefaultSecurityContext.FsGroupChangePolicy)

			assert.False(t, *config.EnforcedSecurityContext.Privileged)
			assert.True(t, *config.EnforcedSecurityContext.RunAsNonRoot)
			assert.False(t, *config.EnforcedSecurityContext.AllowPrivilegeEscalation)
			assert.Equal(t, SecProfileType("RuntimeDefault"), config.EnforcedSecurityContext.SeccompProfile.Type)
			assert.Equal(t, []string{"ALL"}, config.EnforcedSecurityContext.Capabilities.Drop)

			return nil
		},
	}
	err := cmd.Run(context.Background(), []string{"test"})
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
		ObjectMeta: kube_meta_v1.ObjectMeta{
			Name:      podName,
			Namespace: namespace,
		},
		Status: kube_core_v1.PodStatus{
			Phase: kube_core_v1.PodPending,
		},
	}
	_, err = client.CoreV1().Pods(namespace).Create(
		context.Background(), pod, kube_meta_v1.CreateOptions{},
	)
	require.NoError(t, err)
	return podName
}

func TestWaitStepReturnsOnContextCancel(t *testing.T) {
	client := fake.NewClientset()
	engine := makeEngine(client)
	step := makeStep("ctx-cancel-01")
	namespace := "test-ns"

	createPod(t, client, step, namespace)

	ctx, cancel := context.WithCancelCause(context.Background())

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
		wg.Add(1)
		go func() {
			defer wg.Done()

			ctx, cancel := context.WithCancelCause(context.Background())

			go func() {
				_, _ = engine.WaitStep(ctx, steps[i], fmt.Sprintf("task-%d", i))
			}()

			time.Sleep(200 * time.Millisecond)
			cancel(nil)
		}()
	}
	wg.Wait()

	time.Sleep(1 * time.Second)

	afterCancelGoroutines := runtime.NumGoroutine()
	leaked := afterCancelGoroutines - baselineGoroutines

	assert.Less(t, leaked, numSteps,
		"goroutines leaked after canceling %d WaitStep calls: got %d leaked",
		numSteps, leaked)
}
