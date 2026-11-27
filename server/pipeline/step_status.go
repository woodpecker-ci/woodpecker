// Copyright 2022 Woodpecker Authors
// Copyright 2019 mhmxs
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

package pipeline

import (
	"context"

	"github.com/rs/zerolog/log"

	"go.woodpecker-ci.org/woodpecker/v3/pipeline/status"
	"go.woodpecker-ci.org/woodpecker/v3/rpc"
	"go.woodpecker-ci.org/woodpecker/v3/server"
	"go.woodpecker-ci.org/woodpecker/v3/server/model"
	"go.woodpecker-ci.org/woodpecker/v3/server/store"
)

// statusStep returns the part of the step its status is calculated from.
func statusStep(step *model.Step) status.Step {
	return status.Step{
		State:    status.Value(step.State),
		Failure:  step.Failure,
		Started:  step.Started,
		Finished: step.Finished,
		ExitCode: step.ExitCode,
		Error:    step.Error,
	}
}

// setStatusStep writes the calculated status back to the step.
func setStatusStep(step *model.Step, s status.Step) {
	step.State = model.StatusValue(s.State)
	step.Started = s.Started
	step.Finished = s.Finished
	step.ExitCode = s.ExitCode
	step.Error = s.Error
}

func CalcStepStatus(step model.Step, state rpc.StepState) (_ *model.Step, cancelPipelineFromStep bool, _ error) {
	log.Debug().Str("StepUUID", step.UUID).Msgf("Update step %#v state %#v", step, state)

	updated, cancelPipelineFromStep, err := statusStep(&step).Update(state)
	if err != nil {
		return nil, false, err
	}
	setStatusStep(&step, updated)

	return &step, cancelPipelineFromStep, nil
}

// UpdateStepStatus updates step status based on agent reports via RPC.
func UpdateStepStatus(ctx context.Context, store store.Store, step *model.Step, state rpc.StepState) error {
	log.Debug().Str("StepUUID", step.UUID).Msgf("Update step %#v state %#v", *step, state)

	updatedStep, shouldCancelPipelineFromStep, err := CalcStepStatus(*step, state)
	if err != nil {
		return err
	}
	*step = *updatedStep // update step for external callers

	if shouldCancelPipelineFromStep {
		if err := cancelPipelineFromStep(ctx, store, step); err != nil {
			return err
		}
	}
	return store.StepUpdate(step)
}

func cancelPipelineFromStep(ctx context.Context, store store.Store, step *model.Step) error {
	pipeline, err := store.GetPipeline(step.PipelineID)
	if err != nil {
		return err
	}

	repo, err := store.GetRepo(pipeline.RepoID)
	if err != nil {
		return err
	}

	repoUser, err := store.GetUser(repo.UserID)
	if err != nil {
		return err
	}

	_forge, err := server.Config.Services.Manager.ForgeFromRepo(repo)
	if err != nil {
		return err
	}
	return Cancel(ctx, _forge, store, repo, repoUser, pipeline, &model.CancelInfo{
		CanceledByStep: step.Name,
	})
}

func UpdateStepToStatusSkipped(store store.Store, step model.Step, finished int64, state model.StatusValue) (*model.Step, error) {
	setStatusStep(&step, statusStep(&step).End(status.Value(state), finished))
	return &step, store.StepUpdate(&step)
}
