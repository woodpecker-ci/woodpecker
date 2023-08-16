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

	"github.com/woodpecker-ci/woodpecker/pipeline"
	"github.com/woodpecker-ci/woodpecker/pipeline/rpc"
	"github.com/woodpecker-ci/woodpecker/server"
	"github.com/woodpecker-ci/woodpecker/server/model"
)

func queuePipeline(repo *model.Repo, pipelineItems []*pipeline.Item) error {
	var tasks []*model.Task
	for _, item := range pipelineItems {
		if item.Workflow.State == model.StatusSkipped {
			continue
		}
		task := new(model.Task)
		task.ID = fmt.Sprint(item.Workflow.ID)
		task.Labels = map[string]string{}
		for k, v := range item.Labels {
			task.Labels[k] = v
		}
		task.Labels["repo"] = repo.FullName
		task.Dependencies = taskIds(item.DependsOn, pipelineItems)
		task.RunOn = item.RunsOn
		task.DepStatus = make(map[string]model.StatusValue)

		task.Data, _ = json.Marshal(rpc.Pipeline{
			ID:      fmt.Sprint(item.Workflow.ID),
			Config:  item.Config,
			Timeout: repo.Timeout,
		})

		tasks = append(tasks, task)
	}
	return server.Config.Services.Queue.PushAtOnce(context.Background(), tasks)
}

func taskIds(dependsOn []string, pipelineItems []*pipeline.Item) (taskIds []string) {
	for _, dep := range dependsOn {
		for _, pipelineItem := range pipelineItems {
			if pipelineItem.Workflow.Name == dep {
				taskIds = append(taskIds, fmt.Sprint(pipelineItem.Workflow.ID))
			}
		}
	}
	return
}
