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

package build

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/woodpecker-ci/woodpecker/pipeline/rpc"
	"github.com/woodpecker-ci/woodpecker/server"
	"github.com/woodpecker-ci/woodpecker/server/model"
	"github.com/woodpecker-ci/woodpecker/server/queue"
	"github.com/woodpecker-ci/woodpecker/server/shared"
)

func queueBuild(build *model.Build, repo *model.Repo, buildItems []*shared.BuildItem) error {
	var tasks []*queue.Task
	for _, item := range buildItems {
		if item.Proc.State == model.StatusSkipped {
			continue
		}
		task := new(queue.Task)
		task.ID = fmt.Sprint(item.Proc.ID)
		task.Labels = map[string]string{}
		for k, v := range item.Labels {
			task.Labels[k] = v
		}
		task.Labels["platform"] = item.Platform
		task.Labels["repo"] = repo.FullName
		task.Dependencies = taskIds(item.DependsOn, buildItems)
		task.RunOn = item.RunsOn
		task.DepStatus = make(map[string]string)

		task.Data, _ = json.Marshal(rpc.Pipeline{
			ID:      fmt.Sprint(item.Proc.ID),
			Config:  item.Config,
			Timeout: repo.Timeout,
		})

		if err := server.Config.Services.Logs.Open(context.Background(), task.ID); err != nil {
			return err
		}
		tasks = append(tasks, task)
	}
	return server.Config.Services.Queue.PushAtOnce(context.Background(), tasks)
}

func taskIds(dependsOn []string, buildItems []*shared.BuildItem) (taskIds []string) {
	for _, dep := range dependsOn {
		for _, buildItem := range buildItems {
			if buildItem.Proc.Name == dep {
				taskIds = append(taskIds, fmt.Sprint(buildItem.Proc.ID))
			}
		}
	}
	return
}
