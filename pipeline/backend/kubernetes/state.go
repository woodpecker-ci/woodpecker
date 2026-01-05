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
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/rs/zerolog/log"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/tools/leaderelection"
	"k8s.io/client-go/tools/leaderelection/resourcelock"
	"k8s.io/client-go/util/retry"

	"go.woodpecker-ci.org/woodpecker/v3/pipeline/backend/types"
)

const (
	stateConfigMapPrefix = "wp-state-"

	WorkflowStateLabel = "woodpecker-ci.org/workflow-state"
	TTLExpiresAtLabel  = "woodpecker-ci.org/ttl-expires-at"

	stateKeySteps = "steps"

	stateKeyNamespace = "namespace"
	stateKeyOrgID     = "org-id"
	stateKeyVolume    = "volume"
	stateKeyTimeout   = "timeout"
	stateKeyStarted   = "started"
)

// StepState represents the minimal state of a step needed for recovery.
type StepState struct {
	Status   types.StepStatus `json:"status"`
	ExitCode int              `json:"exit_code,omitempty"`
	Started  int64            `json:"started,omitempty"`
	Finished int64            `json:"finished,omitempty"`
}

type WorkflowState struct {
	Volume  string                `json:"volume"`
	Steps   map[string]*StepState `json:"steps"`
	Timeout int64                 `json:"timeout"`
	Started int64                 `json:"started"`
}

func stateConfigMapName(taskUUID string) string {
	return stateConfigMapPrefix + taskUUID
}

// createWorkflowState creates the initial state ConfigMap for a workflow.
// This should be called at the start of SetupWorkflow.
func (e *kube) createWorkflowState(ctx context.Context, conf *types.Config, taskUUID string, timeout time.Duration) error {
	if conf == nil || len(conf.Stages) == 0 || len(conf.Stages[0].Steps) == 0 {
		return fmt.Errorf("invalid workflow config")
	}

	namespace := e.config.GetNamespace(conf.Stages[0].Steps[0].OrgID)
	orgID := conf.Stages[0].Steps[0].OrgID

	// Build initial step states - all pending
	steps := make(map[string]*StepState)
	for _, stage := range conf.Stages {
		for _, step := range stage.Steps {
			steps[step.UUID] = &StepState{
				Status: types.StatusPending,
			}
		}
	}

	stepsJSON, err := json.Marshal(steps)
	if err != nil {
		return fmt.Errorf("failed to marshal steps state: %w", err)
	}

	now := time.Now().Unix()
	timeoutTimestamp := now + int64(timeout.Seconds())
	ttlExpiresAt := timeoutTimestamp + int64(time.Hour.Seconds())

	cm := &v1.ConfigMap{
		ObjectMeta: meta_v1.ObjectMeta{
			Name:      stateConfigMapName(taskUUID),
			Namespace: namespace,
			Labels: map[string]string{
				TaskUUIDLabel:      taskUUID,
				WorkflowStateLabel: "active",
			},
			Annotations: map[string]string{
				TTLExpiresAtLabel: strconv.FormatInt(ttlExpiresAt, 10),
			},
		},
		Data: map[string]string{
			stateKeySteps:     string(stepsJSON),
			stateKeyNamespace: namespace,
			stateKeyOrgID:     strconv.FormatInt(orgID, 10),
			stateKeyVolume:    conf.Volume,
			stateKeyTimeout:   strconv.FormatInt(timeoutTimestamp, 10),
			stateKeyStarted:   strconv.FormatInt(now, 10),
		},
	}

	_, err = e.client.CoreV1().ConfigMaps(namespace).Create(ctx, cm, meta_v1.CreateOptions{})
	if err != nil {
		if errors.IsAlreadyExists(err) {
			// ConfigMap already exists - this might be a recovery scenario
			// or a race condition. Log and continue.
			log.Warn().Str("taskUUID", taskUUID).Msg("workflow state ConfigMap already exists")
			return nil
		}
		return fmt.Errorf("failed to create workflow state ConfigMap: %w", err)
	}

	log.Debug().Str("taskUUID", taskUUID).Str("namespace", namespace).Msg("created workflow state ConfigMap")
	return nil
}

// updateStepState updates a single step's state in the ConfigMap.
// If the update fails due to conflict, it retries with fresh data.
func (e *kube) updateStepState(ctx context.Context, taskUUID, namespace, stepUUID string, updateFn func(*StepState)) error {
	cmName := stateConfigMapName(taskUUID)

	return retry.RetryOnConflict(retry.DefaultRetry, func() error {
		cm, err := e.client.CoreV1().ConfigMaps(namespace).Get(ctx, cmName, meta_v1.GetOptions{})
		if err != nil {
			if errors.IsNotFound(err) {
				// ConfigMap doesn't exist - workflow might have been cleaned up
				log.Warn().Str("taskUUID", taskUUID).Msg("workflow state ConfigMap not found, skipping update")
				return nil
			}
			return fmt.Errorf("failed to get workflow state ConfigMap: %w", err)
		}

		// Parse current steps state
		var steps map[string]*StepState
		if err := json.Unmarshal([]byte(cm.Data[stateKeySteps]), &steps); err != nil {
			return fmt.Errorf("failed to unmarshal steps state: %w", err)
		}

		// Get or create step state
		step, exists := steps[stepUUID]
		if !exists {
			step = &StepState{Status: types.StatusPending}
			steps[stepUUID] = step
		}

		// Apply the update
		updateFn(step)

		// Marshal back
		stepsJSON, err := json.Marshal(steps)
		if err != nil {
			return fmt.Errorf("failed to marshal steps state: %w", err)
		}
		cm.Data[stateKeySteps] = string(stepsJSON)

		_, err = e.client.CoreV1().ConfigMaps(namespace).Update(ctx, cm, meta_v1.UpdateOptions{})
		if err != nil {
			return fmt.Errorf("failed to update workflow state ConfigMap: %w", err)
		}

		return nil
	})
}

// markStepStarted marks a step as running with start timestamp.
func (e *kube) markStepStarted(ctx context.Context, taskUUID, namespace string, step *types.Step) error {
	return e.updateStepState(ctx, taskUUID, namespace, step.UUID, func(s *StepState) {
		s.Status = types.StatusRunning
		s.Started = time.Now().Unix()
	})
}

// markStepCompleted marks a step as completed with exit code and finish timestamp.
func (e *kube) markStepCompleted(ctx context.Context, taskUUID, namespace, stepUUID string, exitCode int) error {
	return e.updateStepState(ctx, taskUUID, namespace, stepUUID, func(s *StepState) {
		if exitCode == 0 {
			s.Status = types.StatusSuccess
		} else {
			s.Status = types.StatusFailed
		}
		s.ExitCode = exitCode
		s.Finished = time.Now().Unix()
	})
}

// markStepSkipped marks a step as skipped.
func (e *kube) markStepSkipped(ctx context.Context, taskUUID, namespace, stepUUID string) error {
	return e.updateStepState(ctx, taskUUID, namespace, stepUUID, func(s *StepState) {
		s.Status = types.StatusSkipped
	})
}

// deleteWorkflowState deletes the state ConfigMap for a workflow.
// This should be called at the end of DestroyWorkflow.
func (e *kube) deleteWorkflowState(ctx context.Context, taskUUID, namespace string) error {
	cmName := stateConfigMapName(taskUUID)

	err := e.client.CoreV1().ConfigMaps(namespace).Delete(ctx, cmName, defaultDeleteOptions)
	if err != nil {
		if errors.IsNotFound(err) {
			// Already deleted, that's fine
			return nil
		}
		return fmt.Errorf("failed to delete workflow state ConfigMap: %w", err)
	}

	log.Debug().Str("taskUUID", taskUUID).Msg("deleted workflow state ConfigMap")
	return nil
}

// getWorkflowState retrieves the workflow state from ConfigMap.
func (e *kube) getWorkflowState(ctx context.Context, taskUUID, namespace string) (*WorkflowState, error) {
	cmName := stateConfigMapName(taskUUID)

	cm, err := e.client.CoreV1().ConfigMaps(namespace).Get(ctx, cmName, meta_v1.GetOptions{})
	if err != nil {
		return nil, err
	}

	var steps map[string]*StepState
	if err := json.Unmarshal([]byte(cm.Data[stateKeySteps]), &steps); err != nil {
		return nil, fmt.Errorf("failed to unmarshal steps state: %w", err)
	}

	timeout, _ := strconv.ParseInt(cm.Data[stateKeyTimeout], 10, 64)
	started, _ := strconv.ParseInt(cm.Data[stateKeyStarted], 10, 64)

	return &WorkflowState{
		Volume:  cm.Data[stateKeyVolume],
		Steps:   steps,
		Timeout: timeout,
		Started: started,
	}, nil
}

func (e *kube) listWoodpeckerNamespaces(ctx context.Context) ([]string, error) {
	namespaces, err := e.client.CoreV1().Namespaces().List(ctx, meta_v1.ListOptions{})
	if err != nil {
		return nil, err
	}

	prefix := e.config.Namespace + "-"
	var result []string
	result = append(result, e.config.Namespace)

	for _, ns := range namespaces.Items {
		if strings.HasPrefix(ns.Name, prefix) {
			result = append(result, ns.Name)
		}
	}
	return result, nil
}

func (e *kube) RecordStepStarted(ctx context.Context, taskUUID string, step *types.Step) error {
	if !e.config.StateRecoveryEnabled {
		return nil
	}
	namespace := e.config.GetNamespace(step.OrgID)
	return e.markStepStarted(ctx, taskUUID, namespace, step)
}

func (e *kube) RecordStepCompleted(ctx context.Context, taskUUID string, step *types.Step, exitCode int) error {
	if !e.config.StateRecoveryEnabled {
		return nil
	}
	namespace := e.config.GetNamespace(step.OrgID)
	return e.markStepCompleted(ctx, taskUUID, namespace, step.UUID, exitCode)
}

func (e *kube) RecordStepSkipped(ctx context.Context, taskUUID string, step *types.Step) error {
	if !e.config.StateRecoveryEnabled {
		return nil
	}
	namespace := e.config.GetNamespace(step.OrgID)
	return e.markStepSkipped(ctx, taskUUID, namespace, step.UUID)
}

func (e *kube) getRemainingTimeout(ctx context.Context, taskUUID, namespace string) (int64, error) {
	cmName := stateConfigMapName(taskUUID)

	cm, err := e.client.CoreV1().ConfigMaps(namespace).Get(ctx, cmName, meta_v1.GetOptions{})
	if err != nil {
		return 0, fmt.Errorf("failed to get workflow state ConfigMap: %w", err)
	}

	timeoutStr, exists := cm.Data[stateKeyTimeout]
	if !exists {
		return 0, fmt.Errorf("timeout not found in workflow state")
	}

	remainingTimeout, err := strconv.ParseInt(timeoutStr, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("failed to parse remaining timeout: %w", err)
	}

	return remainingTimeout, nil
}

func (e *kube) GetStepStatus(ctx context.Context, taskUUID, stepUUID string) (types.StepStatus, error) {
	if !e.config.StateRecoveryEnabled {
		return types.StatusUnknown, nil
	}

	namespacesToCheck := []string{e.config.Namespace}
	if e.config.EnableNamespacePerOrg {
		namespaces, err := e.listWoodpeckerNamespaces(ctx)
		if err == nil {
			namespacesToCheck = namespaces
		}
	}

	for _, namespace := range namespacesToCheck {
		state, err := e.getWorkflowState(ctx, taskUUID, namespace)
		if err != nil {
			if errors.IsNotFound(err) {
				continue
			}
			return types.StatusUnknown, err
		}

		if stepState, exists := state.Steps[stepUUID]; exists {
			return stepState.Status, nil
		}
		return types.StatusUnknown, nil
	}

	return types.StatusUnknown, nil
}

// CleanupExpiredStates runs a periodic cleanup loop with leader election to remove expired
// state ConfigMaps across all Woodpecker namespaces. Only one agent in the cluster will
// perform cleanup at a time to avoid duplicate API calls. This goroutine runs until the
// context is canceled.
func (e *kube) CleanupExpiredStates(ctx context.Context) {
	if !e.config.StateRecoveryEnabled {
		return
	}

	// Get agent identity for leader election
	hostname, err := os.Hostname()
	if err != nil {
		log.Warn().Err(err).Msg("failed to get hostname for leader election, using fallback")
		hostname = fmt.Sprintf("agent-%d", time.Now().Unix())
	}
	identity := fmt.Sprintf("%s_%d", hostname, os.Getpid())

	log.Info().
		Str("identity", identity).
		Str("lease_name", leaderElectionResourceLockName).
		Str("namespace", e.config.Namespace).
		Msg("starting state cleanup with leader election")

	// Ensure the base namespace exists for the leader election lease.
	// When namespace-per-org is enabled, the base namespace might not exist yet
	// since only org-specific namespaces are created during workflow setup.
	if err := mkNamespace(ctx, e.client.CoreV1().Namespaces(), e.config.Namespace); err != nil {
		log.Error().Err(err).Str("namespace", e.config.Namespace).Msg("failed to create namespace for leader election lease")
		return
	}

	// Create resource lock for leader election in the main woodpecker namespace
	lock := &resourcelock.LeaseLock{
		LeaseMeta: meta_v1.ObjectMeta{
			Name:      leaderElectionResourceLockName,
			Namespace: e.config.Namespace,
		},
		Client: e.client.CoordinationV1(),
		LockConfig: resourcelock.ResourceLockConfig{
			Identity: identity,
		},
	}

	// Run cleanup when this agent becomes leader
	runCleanup := func(ctx context.Context) {
		log.Info().Str("identity", identity).Msg("became leader for state cleanup")

		ticker := time.NewTicker(defaultStateCleanupInterval)
		defer ticker.Stop()

		// Run cleanup immediately on becoming leader
		e.performCleanup(ctx)

		// Then run periodically
		for {
			select {
			case <-ctx.Done():
				log.Info().Str("identity", identity).Msg("stopping state cleanup (context canceled)")
				return
			case <-ticker.C:
				e.performCleanup(ctx)
			}
		}
	}

	// Start leader election
	leaderelection.RunOrDie(ctx, leaderelection.LeaderElectionConfig{
		Lock:            lock,
		ReleaseOnCancel: true,
		LeaseDuration:   leaderElectionLeaseDuration,
		RenewDeadline:   leaderElectionRenewDeadline,
		RetryPeriod:     leaderElectionRetryPeriod,
		Callbacks: leaderelection.LeaderCallbacks{
			OnStartedLeading: runCleanup,
			OnStoppedLeading: func() {
				log.Info().Str("identity", identity).Msg("lost leadership for state cleanup")
			},
			OnNewLeader: func(currentLeader string) {
				if currentLeader != identity {
					log.Debug().
						Str("identity", identity).
						Str("leader", currentLeader).
						Msg("current leader for state cleanup")
				}
			},
		},
	})
}

// performCleanup scans all Woodpecker namespaces and removes expired state ConfigMaps.
func (e *kube) performCleanup(ctx context.Context) {
	log.Debug().Msg("performing periodic state cleanup")

	namespaces, err := e.listWoodpeckerNamespaces(ctx)
	if err != nil {
		log.Warn().Err(err).Msg("failed to list Woodpecker namespaces for cleanup")
		return
	}

	for _, namespace := range namespaces {
		e.cleanupExpiredStateConfigMaps(ctx, namespace)
	}
}

// cleanupExpiredStateConfigMaps removes state ConfigMaps that have exceeded their TTL.
// TTL is set to workflow timeout + 1 hour buffer, allowing time for recovery operations
// after workflow completion or failure. This is called periodically by the leader agent.
func (e *kube) cleanupExpiredStateConfigMaps(ctx context.Context, namespace string) {
	now := time.Now().Unix()

	cms, err := e.client.CoreV1().ConfigMaps(namespace).List(ctx, meta_v1.ListOptions{
		LabelSelector: fmt.Sprintf("%s=active", WorkflowStateLabel),
	})
	if err != nil {
		log.Warn().Err(err).Str("namespace", namespace).Msg("failed to list state ConfigMaps for cleanup")
		return
	}

	if len(cms.Items) > 0 {
		log.Debug().
			Str("namespace", namespace).
			Int("count", len(cms.Items)).
			Int64("now", now).
			Msg("checking state ConfigMaps for cleanup")
	}

	for _, cm := range cms.Items {
		// Check if TTL has expired (TTL = workflow timeout + 1 hour buffer)
		ttlStr, exists := cm.Annotations[TTLExpiresAtLabel]
		if !exists {
			log.Debug().
				Str("configmap", cm.Name).
				Str("namespace", namespace).
				Msg("state ConfigMap has no TTL annotation, skipping cleanup")
			continue
		}

		ttl, err := strconv.ParseInt(ttlStr, 10, 64)
		if err != nil {
			log.Warn().
				Err(err).
				Str("configmap", cm.Name).
				Str("namespace", namespace).
				Msg("failed to parse TTL annotation, skipping cleanup")
			continue
		}

		if now >= ttl {
			// TTL expired - delete the ConfigMap
			log.Info().
				Str("configmap", cm.Name).
				Str("namespace", namespace).
				Msg("cleaning up expired state ConfigMap")

			if err := e.client.CoreV1().ConfigMaps(namespace).Delete(ctx, cm.Name, defaultDeleteOptions); err != nil {
				if !errors.IsNotFound(err) {
					log.Warn().Err(err).Str("configmap", cm.Name).Msg("failed to delete state ConfigMap")
				}
			}
		} else {
			// TTL not yet expired - keep the ConfigMap
			ttlRemaining := ttl - now
			log.Debug().
				Str("configmap", cm.Name).
				Str("namespace", namespace).
				Int64("ttl_remaining_sec", ttlRemaining).
				Msg("state ConfigMap not expired yet")
		}
	}
}
