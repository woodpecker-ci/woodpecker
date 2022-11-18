// Copyright 2022 Woodpecker Authors
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

	"github.com/rs/zerolog/log"
	"github.com/woodpecker-ci/woodpecker/server"
	"github.com/woodpecker-ci/woodpecker/server/model"
	"github.com/woodpecker-ci/woodpecker/server/store"
)

// Skip update the status to skip for pending pipeline
func Skip(ctx context.Context, store store.Store, pipeline *model.Pipeline, user *model.User, repo *model.Repo) (*model.Pipeline, error) {
	// if pipeline.Status != model.StatusPending {
	// 	return nil, fmt.Errorf("cannot skip a pipeline with status %s", pipeline.Status)
	// }

	var err error
	if pipeline.Steps, err = store.StepList(pipeline); err != nil {
		log.Error().Err(err).Msg("can not get step list from store")
	}

	for _, step := range pipeline.Steps {
		fmt.Printf("each step: %v\n", step.State)
	}

	_, err = UpdateToStatusSkipped(store, *pipeline, user.Login)
	if err != nil {
		return nil, fmt.Errorf("error updating pipeline. %s", err)
	}

	if pipeline.Steps, err = model.Tree(pipeline.Steps); err != nil {
		log.Error().Err(err).Msg("can not build tree from step list")
	}

	if err := updatePipelineStatus(ctx, pipeline, repo, user); err != nil {
		log.Error().Err(err).Msg("updateBuildStatus")
	}

	if err := publishToTopic(ctx, pipeline, repo); err != nil {
		log.Error().Err(err).Msg("publishToTopic")
	}

	return pipeline, nil
}

func SkipStep(ctx context.Context, store store.Store, pipeline *model.Pipeline, stepPid int, user *model.User, repo *model.Repo) (*model.Pipeline, error) {
	// fmt.Printf("Trying to skip pid: %d\n", stepPid)
	var err error
	if pipeline.Steps, err = store.StepList(pipeline); err != nil {
		log.Error().Err(err).Msg("can not get step list from store")
	}

	for _, step := range pipeline.Steps {
		if step.PID == stepPid {
			fmt.Printf("Trying to skip step %d %d %v\n", step.ID, step.PID, step.State)
			// if err = server.Config.Services.Queue.Done(ctx, fmt.Sprint(step.ID), model.StatusSkipped); err != nil {
			// 	log.Error().Err(err).Msgf("queue: skip: %v", step.ID)
			// }
			if err = server.Config.Services.Queue.EvictCurrent(ctx, fmt.Sprint(step.ID), model.StatusSkipped); err != nil {
				log.Error().Err(err).Msgf("queue: evict: %v", step.ID)
			}
			if _, err = UpdateStepToStatusSkipped(store, *step, 0); err != nil {
				log.Error().Msgf("error: done: cannot update step_id %d state: %s", step.ID, err)
			}
		}
	}

	if pipeline.Steps, err = store.StepList(pipeline); err != nil {
		log.Error().Err(err).Msg("can not get step list from store")
	}

	if pipeline.Steps, err = model.Tree(pipeline.Steps); err != nil {
		log.Error().Err(err).Msg("can not build tree from step list")
	}

	for _, step := range pipeline.Steps {
		fmt.Printf("Updated step status: %d %v %v\n", step.ID, step.Name, step.State)
	}

	if err := publishToTopic(ctx, pipeline, repo); err != nil {
		log.Error().Err(err).Msg("publishToTopic")
	}
	return pipeline, nil
}
