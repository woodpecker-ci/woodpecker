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
	"encoding/json"
	"fmt"
	"maps"

	"go.woodpecker-ci.org/woodpecker/v2/pipeline/rpc"
	"go.woodpecker-ci.org/woodpecker/v2/server"
	"go.woodpecker-ci.org/woodpecker/v2/server/model"
	"go.woodpecker-ci.org/woodpecker/v2/server/pipeline/stepbuilder"
)

func queuePipeline(ctx context.Context, repo *model.Repo, pipelineItems []*stepbuilder.Item) error {
	var tasks []*model.Task
	for _, item := range pipelineItems {
		if item.Workflow.State == model.StatusSkipped {
			continue
		}
		task := &model.Task{
			ID:     fmt.Sprint(item.Workflow.ID),
			Labels: make(map[string]string),
		}
		maps.Copy(task.Labels, item.Labels)
		err := task.ApplyLabelsFromRepo(repo)
		if err != nil {
			return err
		}
		task.Dependencies = taskIDs(item.DependsOn, pipelineItems)
		task.RunOn = item.RunsOn
		task.DepStatus = make(map[string]model.StatusValue)

		task.Data, err = json.Marshal(rpc.Workflow{
			ID:      fmt.Sprint(item.Workflow.ID),
			Config:  item.Config,
			Timeout: repo.Timeout,
		})
		if err != nil {
			return err
		}

		tasks = append(tasks, task)
	}
	return server.Config.Services.Queue.PushAtOnce(ctx, tasks)
}

func taskIDs(dependsOn []string, pipelineItems []*stepbuilder.Item) (taskIDs []string) {
	for _, dep := range dependsOn {
		for _, pipelineItem := range pipelineItems {
			if pipelineItem.Workflow.Name == dep {
				taskIDs = append(taskIDs, fmt.Sprint(pipelineItem.Workflow.ID))
			}
		}
	}
	return
}
