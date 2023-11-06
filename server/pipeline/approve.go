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

	"github.com/woodpecker-ci/woodpecker/pipeline/errors"
	forge_types "github.com/woodpecker-ci/woodpecker/server/forge/types"
	"github.com/woodpecker-ci/woodpecker/server/model"
	"github.com/woodpecker-ci/woodpecker/server/store"
)

// Approve update the status to pending for a blocked pipeline because of a gated repo
// and start them afterward
func Approve(ctx context.Context, store store.Store, currentPipeline *model.Pipeline, user *model.User, repo *model.Repo) (*model.Pipeline, error) {
	if currentPipeline.Status != model.StatusBlocked {
		return nil, ErrBadRequest{Msg: fmt.Sprintf("cannot approve a pipeline with status %s", currentPipeline.Status)}
	}

	// fetch the pipeline file from the database
	configs, err := store.ConfigsForPipeline(currentPipeline.ID)
	if err != nil {
		msg := fmt.Sprintf("failure to get pipeline config for %s", repo.FullName)
		log.Error().Err(err).Msg(msg)
		return nil, ErrNotFound{Msg: msg}
	}
	var yamls []*forge_types.FileMeta
	for _, y := range configs {
		yamls = append(yamls, &forge_types.FileMeta{Data: y.Data, Name: y.Name})
	}

	if currentPipeline.Workflows, err = store.WorkflowGetTree(currentPipeline); err != nil {
		return nil, fmt.Errorf("error: loading workflows. %w", err)
	}

	if currentPipeline, err = UpdateToStatusPending(store, *currentPipeline, user.Login); err != nil {
		return nil, fmt.Errorf("error updating pipeline. %w", err)
	}

	for _, wf := range currentPipeline.Workflows {
		if wf.State != model.StatusBlocked {
			continue
		}
		wf.State = model.StatusPending
		if err := store.WorkflowUpdate(wf); err != nil {
			return nil, fmt.Errorf("error updating workflow. %w", err)
		}

		for _, step := range wf.Children {
			if step.State != model.StatusBlocked {
				continue
			}
			step.State = model.StatusPending
			if err := store.StepUpdate(step); err != nil {
				return nil, fmt.Errorf("error updating step. %w", err)
			}
		}
	}

	pipelineItems, err := parsePipeline(store, currentPipeline, user, repo, yamls, nil)
	if errors.HasBlockingErrors(err) {
		msg := fmt.Sprintf("failure to parsePipeline for %s", repo.FullName)
		log.Error().Err(err).Msg(msg)
		return nil, err
	}

	// TODO improve this
	for _, item := range pipelineItems {
		for _, wf := range currentPipeline.Workflows {
			if item.Workflow.Name == wf.Name {
				item.Workflow = wf
				for _, stage := range item.Config.Stages {
					for _, step := range stage.Steps {
						for _, storeStep := range wf.Children {
							if storeStep.Name == step.Alias {
								step.UUID = storeStep.UUID
								break
							}
						}
					}
				}

				break
			}
		}
	}

	publishPipeline(ctx, currentPipeline, repo, user)

	currentPipeline, err = start(ctx, store, currentPipeline, user, repo, pipelineItems)
	if err != nil {
		msg := fmt.Sprintf("failure to start pipeline for %s: %v", repo.FullName, err)
		log.Error().Err(err).Msg(msg)
		return nil, fmt.Errorf(msg)
	}

	return currentPipeline, nil
}
