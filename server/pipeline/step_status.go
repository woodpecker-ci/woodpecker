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
	"fmt"
	"time"

	"github.com/rs/zerolog/log"

	"go.woodpecker-ci.org/woodpecker/v3/rpc"
	"go.woodpecker-ci.org/woodpecker/v3/server"
	"go.woodpecker-ci.org/woodpecker/v3/server/model"
	"go.woodpecker-ci.org/woodpecker/v3/server/store"
)

// UpdateStepStatus updates step status based on agent reports via RPC.
func UpdateStepStatus(ctx context.Context, store store.Store, step *model.Step, state rpc.StepState) error {
	log.Debug().Str("StepUUID", step.UUID).Msgf("Update step %#v state %#v", *step, state)

	switch step.State {
	case model.StatusPending:
		// Transition from pending to running when started
		if state.Finished == 0 {
			step.State = model.StatusRunning
		}
		step.Started = state.Started
		if step.Started == 0 {
			step.Started = time.Now().Unix()
		}

		// Handle direct transition to finished if step setup error happened
		if state.Exited || state.Error != "" {
			step.Finished = state.Finished
			if step.Finished == 0 {
				step.Finished = time.Now().Unix()
			}
			step.ExitCode = state.ExitCode
			step.Error = state.Error

			if state.ExitCode == 0 && state.Error == "" {
				step.State = model.StatusSuccess
			} else {
				step.State = model.StatusFailure

				if step.Failure == model.FailureCancel {
					// cancel the pipeline
					err := cancelPipelineFromStep(ctx, store, step)
					if err != nil {
						return err
					}
				}
			}
		}

	case model.StatusRunning:
		// Already running, check if it finished
		if state.Exited || state.Error != "" {
			step.Finished = state.Finished
			if step.Finished == 0 {
				step.Finished = time.Now().Unix()
			}
			step.ExitCode = state.ExitCode
			step.Error = state.Error

			if state.ExitCode == 0 && state.Error == "" {
				step.State = model.StatusSuccess
			} else {
				step.State = model.StatusFailure

				if step.Failure == model.FailureCancel {
					// cancel the pipeline
					err := cancelPipelineFromStep(ctx, store, step)
					if err != nil {
						return err
					}
				}
			}
		}

	default:
		return fmt.Errorf("step has state %s and does not expect rpc state updates", step.State)
	}

	// Handle cancellation across both cases
	if state.Canceled && step.State != model.StatusKilled {
		step.State = model.StatusKilled
		if step.Finished == 0 {
			step.Finished = time.Now().Unix()
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

func UpdateStepToStatusSkipped(store store.Store, step model.Step, finished int64) (*model.Step, error) {
	step.State = model.StatusSkipped
	if step.Started != 0 {
		step.State = model.StatusSuccess // for daemons that are killed
		step.Finished = finished
	}
	return &step, store.StepUpdate(&step)
}
