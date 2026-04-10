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

package rpc

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/rs/zerolog/log"

	"go.woodpecker-ci.org/woodpecker/v3/rpc"
	"go.woodpecker-ci.org/woodpecker/v3/server/model"
	"go.woodpecker-ci.org/woodpecker/v3/server/pipeline"
)

const logStreamDelayAllowed = 5 * time.Minute

func (s *RPC) checkAgentPermissionByWorkflow(_ context.Context, agent *model.Agent, strWorkflowID string, p *model.Pipeline, repo *model.Repo) error {
	var err error
	if repo == nil && p == nil {
		workflowID, err := strconv.ParseInt(strWorkflowID, 10, 64)
		if err != nil {
			return err
		}

		workflow, err := s.store.WorkflowLoad(workflowID)
		if err != nil {
			log.Error().Err(err).Msgf("cannot find workflow with id %d", workflowID)
			return err
		}

		p, err = s.store.GetPipeline(workflow.PipelineID)
		if err != nil {
			log.Error().Err(err).Msgf("cannot find pipeline with id %d", workflow.PipelineID)
			return err
		}
	}

	if repo == nil {
		repo, err = s.store.GetRepo(p.RepoID)
		if err != nil {
			log.Error().Err(err).Msgf("cannot find repo with id %d", p.RepoID)
			return err
		}
	}

	if agent.CanAccessRepo(repo) {
		return nil
	}

	log.Error().Err(ErrAgentIllegalRepo).Int64("agentID", agent.ID).Int64("repoId", repo.ID).Send()
	return fmt.Errorf("%w: agentId=%d repoID=%d", ErrAgentIllegalRepo, agent.ID, repo.ID)
}

// isActiveState returns true for states where work is in progress or not yet started.
func isActiveState(state model.StatusValue) bool {
	switch state {
	case model.StatusCreated,
		model.StatusPending,
		model.StatusRunning:
		return true
	default:
		return false
	}
}

// isDoneState returns true for terminal states where no further work will happen.
func isDoneState(state model.StatusValue) bool {
	switch state {
	case model.StatusSuccess,
		model.StatusFailure,
		model.StatusKilled,
		model.StatusCanceled,
		model.StatusSkipped,
		model.StatusError,
		model.StatusDeclined:
		return true
	default:
		return false
	}
}

// checkWorkflowAllowsStepUpdate validates whether the workflow state permits
// the given step state update. If the workflow is active (created/pending/running),
// any step update is allowed. If the workflow is in a terminal state, only
// updates that would move the step into a terminal state are permitted — this
// lets the agent report final results for steps that completed after the
// workflow was already marked done.
func checkWorkflowAllowsStepUpdate(workflowState model.StatusValue, step *model.Step, state rpc.StepState) error {
	if isActiveState(workflowState) {
		return nil
	}

	newStep, _, err := pipeline.CalcStepStatus(*step, state)
	if err != nil {
		return err
	}
	if isDoneState(newStep.State) {
		return nil
	}

	retErr := ErrAgentIllegalWorkflowReRunStateChange
	log.Error().Err(retErr).Msg("caught agent performing illegal instruction")
	return retErr
}

// checkWorkflowState checks if a workflow's own state allows it to be
// initialized or marked as done. A workflow that is already in a terminal
// state (success, failure, killed, …) must not be re-run, and a blocked
// workflow must not be started by an agent.
func checkWorkflowState(state model.StatusValue) (err error) {
	switch state {
	case model.StatusCreated,
		model.StatusPending,
		model.StatusRunning:
		return nil

	case model.StatusBlocked:
		err = ErrAgentIllegalWorkflowRun

	default:
		err = ErrAgentIllegalWorkflowReRunStateChange
	}

	if err != nil {
		log.Error().Err(err).Msg("caught agent performing illegal instruction")
	}
	return err
}

func allowAppendingLogs(currPipeline *model.Pipeline, currStep *model.Step) error {
	// As long as pipeline is running just let the agent send logs
	if currStep.State == model.StatusRunning || currPipeline.Status == model.StatusRunning {
		return nil
	}
	// else give some delay where log caches can drain and be send ... because of network outage / server restart / ...
	if time.Unix(currPipeline.Finished, 0).Add(logStreamDelayAllowed).After(time.Now()) {
		return nil
	}

	err := ErrAgentIllegalLogStreaming
	log.Error().Err(err).Msg("caught agent performing illegal instruction")
	return err
}
