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

//go:build kind

// These tests run the real Kubernetes backend against a live cluster (a local
// `kind` cluster in CI/dev) and induce a real pod eviction to prove that
// WaitStep detects an infrastructure failure from the pod's DisruptionTarget
// condition — the boundary the in-process e2e suite cannot exercise.
//
// Run with:
//
//	kind create cluster --name wp-e2e
//	go test -tags kind -run TestKind -timeout 180s ./pipeline/backend/kubernetes/
package kubernetes

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	kube_core_v1 "k8s.io/api/core/v1"
	kube_policy_v1 "k8s.io/api/policy/v1"
	kube_errors "k8s.io/apimachinery/pkg/api/errors"
	kube_meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"

	"go.woodpecker-ci.org/woodpecker/v3/pipeline/backend/types"
)

// kindClient builds a clientset from the ambient kubeconfig, pinned to the
// kind-wp-e2e context. Skips the test if no cluster is reachable.
func kindClient(t *testing.T) kubernetes.Interface {
	t.Helper()
	rules := clientcmd.NewDefaultClientConfigLoadingRules()
	overrides := &clientcmd.ConfigOverrides{CurrentContext: "kind-wp-e2e"}
	cfg, err := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(rules, overrides).ClientConfig()
	if err != nil {
		t.Skipf("no kubeconfig / cluster available: %v", err)
	}
	cs, err := kubernetes.NewForConfig(cfg)
	require.NoError(t, err)
	if _, err := cs.Discovery().ServerVersion(); err != nil {
		t.Skipf("kind cluster not reachable: %v", err)
	}
	return cs
}

// TestKindInfraFailureOnEviction creates a real pod, evicts it via the
// Eviction API (which sets the pod's DisruptionTarget condition exactly as a
// spot-node preemption does), and asserts the backend's WaitStep returns a
// State flagged as an infrastructure failure.
func TestKindInfraFailureOnEviction(t *testing.T) {
	client := kindClient(t)
	ctx := context.Background()
	const ns = "wp-kind-e2e"

	if _, err := client.CoreV1().Namespaces().Create(ctx, &kube_core_v1.Namespace{
		ObjectMeta: kube_meta_v1.ObjectMeta{Name: ns},
	}, kube_meta_v1.CreateOptions{}); err != nil && !kube_errors.IsAlreadyExists(err) {
		require.NoError(t, err, "create namespace")
	}
	t.Cleanup(func() {
		_ = client.CoreV1().Namespaces().Delete(context.Background(), ns, kube_meta_v1.DeleteOptions{})
	})

	engine := &kube{
		client: client,
		config: &config{Namespace: ns},
		goos:   "linux",
	}

	step := &types.Step{UUID: "01he8bebctabr3kgk0qj36d2me-0", Name: "preempted", Type: types.StepTypeCommands}
	podName, err := stepToPodName(step)
	require.NoError(t, err)

	// Ignore SIGTERM so the pod lingers in Terminating (with DisruptionTarget
	// set) for the whole grace period, then receives SIGKILL (137). This
	// mirrors a workload that doesn't shut down promptly on a preemption.
	grace := int64(20)
	pod := &kube_core_v1.Pod{
		ObjectMeta: kube_meta_v1.ObjectMeta{Name: podName, Namespace: ns},
		Spec: kube_core_v1.PodSpec{
			RestartPolicy:                 kube_core_v1.RestartPolicyNever,
			TerminationGracePeriodSeconds: &grace,
			Containers: []kube_core_v1.Container{{
				Name:    "wp-step",
				Image:   "busybox:1.36",
				Command: []string{"sh", "-c", "trap '' TERM; sleep 600"},
			}},
		},
	}
	_, err = client.CoreV1().Pods(ns).Create(ctx, pod, kube_meta_v1.CreateOptions{})
	require.NoError(t, err)

	require.Eventually(t, func() bool {
		p, gErr := client.CoreV1().Pods(ns).Get(ctx, podName, kube_meta_v1.GetOptions{})
		return gErr == nil && p.Status.Phase == kube_core_v1.PodRunning
	}, 120*time.Second, time.Second, "pod never reached Running")

	type result struct {
		state *types.State
		err   error
	}
	done := make(chan result, 1)
	go func() {
		s, e := engine.WaitStep(ctx, step, "task-kind")
		done <- result{s, e}
	}()

	// Let WaitStep's informer sync, then evict — the same DisruptionTarget
	// signal a spot preemption produces.
	time.Sleep(3 * time.Second)
	require.NoError(t, client.PolicyV1().Evictions(ns).Evict(ctx, &kube_policy_v1.Eviction{
		ObjectMeta: kube_meta_v1.ObjectMeta{Name: podName, Namespace: ns},
	}), "evict pod")

	select {
	case r := <-done:
		require.NoError(t, r.err, "WaitStep returned error")
		require.NotNil(t, r.state)
		assert.True(t, r.state.InfraFailure,
			"WaitStep must flag an evicted (DisruptionTarget) pod as an infra failure; got %+v", r.state)
		assert.True(t, r.state.Exited)
	case <-time.After(120 * time.Second):
		t.Fatal("WaitStep did not return within 120s after eviction")
	}
}
