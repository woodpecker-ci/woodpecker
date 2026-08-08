// Copyright 2026 Woodpecker Authors
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

package scheduler

import (
	"context"

	"go.woodpecker-ci.org/woodpecker/v3/server/model"
)

func (p *impl) ReprioritizeRepo(c context.Context, repo *model.Repo, rules []model.QueuePriorityRule) error {
	priorities := map[string]int{}
	info := p.q.Info(c)
	collect := func(tasks []*model.Task) error {
		for _, task := range tasks {
			if task.RepoID != repo.ID {
				continue
			}
			pipeline, err := p.store.GetPipeline(task.PipelineID)
			if err != nil {
				return err
			}
			priorities[task.ID] = model.QueuePriority(repo, pipeline, rules)
		}
		return nil
	}
	if err := collect(info.Pending); err != nil {
		return err
	}
	if err := collect(info.WaitingOnDeps); err != nil {
		return err
	}
	if len(priorities) == 0 {
		return nil
	}
	return p.q.Reprioritize(c, priorities)
}
