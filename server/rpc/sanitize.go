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
	"errors"
	"fmt"
	"strconv"

	"github.com/rs/zerolog/log"

	"go.woodpecker-ci.org/woodpecker/v3/server/model"
)

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

	msg := fmt.Sprintf("agent '%d' is not allowed to interact with repo[%d] '%s'", agent.ID, repo.ID, repo.FullName)
	log.Error().Int64("repoId", repo.ID).Msg(msg)
	return errors.New(msg)
}

// checkPipelineState checks if an agent is allowed to change/update a workflow/pipeline state
// by the state the parent pipeline is in.
func checkPipelineState(currPipeline *model.Pipeline) (err error) {
	// check if pipeline was already run and marked finished or is blocked
	switch currPipeline.Status {
	case model.StatusCreated,
		model.StatusPending,
		model.StatusRunning:
		break

	case model.StatusBlocked:
		err = ErrAgentIllegalPipelineWorkflowRun

	default:
		err = ErrAgentIllegalPipelineWorkflowReRunStateChange
	}

	if err != nil {
		log.Error().Err(err).Msg("caught agent performing illegal instruction")
	}
	return err
}

// checkWorkflowStepStates checks if a workflow/step state or its logs can be altered
// depending on what state the workflow and step currently is in.
func checkWorkflowStepStates(currWorkflow *model.Workflow, currStep *model.Step) (err error) {
	if currWorkflow != nil {
		switch currWorkflow.State {
		case model.StatusCreated,
			model.StatusPending,
			model.StatusRunning:
			break

		case model.StatusBlocked:
			err = ErrAgentIllegalWorkflowRun

		default:
			err = ErrAgentIllegalWorkflowReRunStateChange
		}
	}

	if currStep != nil {
		switch currStep.State {
		case model.StatusCreated,
			model.StatusPending,
			model.StatusRunning:
			break

		case model.StatusBlocked:
			err = errors.Join(err, ErrAgentIllegalStepRun)

		default:
			err = errors.Join(err, ErrAgentIllegalStepReRunStateChange)
		}
	}

	if err != nil {
		log.Error().Err(err).Msg("caught agent performing illegal instruction")
	}
	return err
}

func allowAppendingLogs(currStep *model.Step) (err error) {
	if currStep.State != model.StatusRunning {
		err = ErrAgentIllegalLogStreaming
	}
	if err != nil {
		log.Error().Err(err).Msg("caught agent performing illegal instruction")
	}
	return err
}
