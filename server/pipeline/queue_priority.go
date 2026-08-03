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

package pipeline

import (
	"context"
	"errors"
	"fmt"

	"go.woodpecker-ci.org/woodpecker/v3/server/forge"
	forge_types "go.woodpecker-ci.org/woodpecker/v3/server/forge/types"
	"go.woodpecker-ci.org/woodpecker/v3/server/model"
)

const queuePriorityConfigPath = ".woodpecker/queue-priority"

func loadQueuePriorityRules(ctx context.Context, forge forge.Forge, user *model.User, repo *model.Repo, pipeline *model.Pipeline) ([]model.QueuePriorityRule, error) {
	if repo.Branch == "" {
		return nil, nil
	}

	defaultHead, err := forge.BranchHead(ctx, user, repo, repo.Branch)
	if err != nil {
		return nil, fmt.Errorf("could not resolve default branch %q for queue priority config: %w", repo.Branch, err)
	}

	configPipeline := *pipeline
	configPipeline.Commit = defaultHead.SHA
	data, err := forge.File(ctx, user, repo, &configPipeline, queuePriorityConfigPath)
	if errors.Is(err, &forge_types.ErrConfigNotFound{}) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("could not load queue priority config %q: %w", queuePriorityConfigPath, err)
	}

	rules, err := model.ParseQueuePriorityRuleFile(data)
	if err != nil {
		return nil, fmt.Errorf("could not parse queue priority config %q: %w", queuePriorityConfigPath, err)
	}
	return rules, nil
}
