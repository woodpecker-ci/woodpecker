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
	"encoding/json"
	"fmt"
	"maps"
	"time"

	"go.woodpecker-ci.org/woodpecker/v3/pipeline/frontend/builder"
	"go.woodpecker-ci.org/woodpecker/v3/rpc"
	"go.woodpecker-ci.org/woodpecker/v3/server/model"
)

// pipelineTasks builds the queue tasks for a pipeline's workflow items.
// Enqueuing happens via the scheduler (see scheduler.StartPipeline).
func pipelineTasks(repo *model.Repo, activePipeline *model.Pipeline, pipelineItems []*builder.Item) ([]*model.Task, error) {
	var tasks []*model.Task
	for _, item := range pipelineItems {
		task := &model.Task{
			ID:         fmt.Sprint(item.Workflow.ID),
			PID:        item.Workflow.PID,
			Name:       item.Workflow.Name,
			Labels:     make(map[string]string),
			PipelineID: activePipeline.ID,
			RepoID:     repo.ID,
			Created:    activePipeline.Created,
		}
		// fall back to the current time if the pipeline has no creation
		// timestamp, so the queue always has a defined ordering key.
		if task.Created == 0 {
			task.Created = time.Now().Unix()
		}
		maps.Copy(task.Labels, item.Labels)
		err := task.ApplyLabelsFromRepo(repo)
		if err != nil {
			return nil, err
		}
		task.Dependencies = getTaskDependencies(item.DependsOn.Names(), pipelineItems)
		task.RunOn = item.RunsOn
		task.DepStatus = make(map[string]model.StatusValue)

		// Set up the concurrency limit if the workflow opted in.
		if item.ConcurrencyLimit > 0 {
			task.ConcurrencyLimit = item.ConcurrencyLimit

			// If no group assigned, each workflow is it's own unique group,
			// else we use defined group unique per repo.
			if item.ConcurrencyGroup == "" {
				task.ConcurrencyGroup = fmt.Sprintf("%d/%s/", repo.ID, item.Workflow.Name)
			} else {
				task.ConcurrencyGroup = fmt.Sprintf("%d//%s", repo.ID, item.ConcurrencyGroup)
			}
		}

		task.Data, err = json.Marshal(rpc.Workflow{
			ID:      fmt.Sprint(item.Workflow.ID),
			Config:  item.Config,
			Timeout: repo.Timeout,
		})
		if err != nil {
			return nil, err
		}

		tasks = append(tasks, task)
	}
	return tasks, nil
}

func getTaskDependencies(dependsOn []string, items []*builder.Item) (taskIDs []string) {
	for _, dep := range dependsOn {
		for _, pipelineItem := range items {
			if pipelineItem.Workflow.Name == dep {
				taskIDs = append(taskIDs, fmt.Sprint(pipelineItem.Workflow.ID))
			}
		}
	}
	return taskIDs
}
