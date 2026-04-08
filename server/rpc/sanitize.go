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

	"go.woodpecker-ci.org/woodpecker/v3/server/model"
)

const logStreamDelayAllowed = 5 * time.Minute

func (s *RPC) checkAgentPermissionByWorkflow(_ context.Context, agent *model.Agent, strWorkflowID string, pipeline *model.Pipeline, repo *model.Repo) error {
	var err error
	if repo == nil && pipeline == nil {
		workflowID, err := strconv.ParseInt(strWorkflowID, 10, 64)
		if err != nil {
			return err
		}

		workflow, err := s.store.WorkflowLoad(workflowID)
		if err != nil {
			log.Error().Err(err).Msgf("cannot find workflow with id %d", workflowID)
			return err
		}

		pipeline, err = s.store.GetPipeline(workflow.PipelineID)
		if err != nil {
			log.Error().Err(err).Msgf("cannot find pipeline with id %d", workflow.PipelineID)
			return err
		}
	}

	if repo == nil {
		repo, err = s.store.GetRepo(pipeline.RepoID)
		if err != nil {
			log.Error().Err(err).Msgf("cannot find repo with id %d", pipeline.RepoID)
			return err
		}
	}

	if agent.CanAccessRepo(repo) {
		return nil
	}

	log.Error().Err(ErrAgentIllegalRepo).Int64("agentID", agent.ID).Int64("repoId", repo.ID).Send()
	return fmt.Errorf("%w: agentId=%d repoID=%d", ErrAgentIllegalRepo, agent.ID, repo.ID)
}

// checkParentState checks if an agent is allowed to change/update a workflow/step state
// by the state the parent pipeline/workflow.
func checkParentState(parentState, childState model.StatusValue, isStep bool) (err error) {
	// check if pipeline was already run and marked finished or is blocked
	switch parentState {
	case model.StatusCreated,
		model.StatusPending,
		model.StatusRunning:
		return nil

	case model.StatusBlocked:
		if isStep {
			err = ErrAgentIllegalWorkflowRun
		} else {
			err = ErrAgentIllegalPipelineWorkflowRun
		}

	case model.StatusCanceled,
		model.StatusFailure,
		model.StatusKilled:

		switch childState {
		case model.StatusCanceled,
			model.StatusKilled,
			model.StatusSkipped,
			model.StatusFailure,
			model.StatusSuccess:
			return nil

		default:
			if isStep {
				err = ErrAgentIllegalWorkflowReRunStateChange
			} else {
				err = ErrAgentIllegalPipelineWorkflowReRunStateChange
			}
		}

	default:
		if isStep {
			err = ErrAgentIllegalWorkflowReRunStateChange
		} else {
			err = ErrAgentIllegalPipelineWorkflowReRunStateChange
		}
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
