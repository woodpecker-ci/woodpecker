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

	"github.com/oklog/ulid/v2"

	"go.woodpecker-ci.org/woodpecker/v3/server"
	"go.woodpecker-ci.org/woodpecker/v3/server/model"
	"go.woodpecker-ci.org/woodpecker/v3/server/pubsub"
)

// publishToTopic publishes message to UI clients.
func publishToTopic(c context.Context, pipeline *model.Pipeline, repo *model.Repo) (err error) {
	message := pubsub.Message{ID: ulid.Make().String()}
	message.Data, err = json.Marshal(model.Event{
		Repo:     *repo,
		Pipeline: *pipeline,
	})
	if err != nil {
		return fmt.Errorf("can't marshal JSON: %w", err)
	}

	subTopics := make(map[string]struct{})
	// if repo is public, push to public topic
	if !repo.IsSCMPrivate {
		subTopics[pubsub.PublicTopic] = struct{}{}
	}
	// publish to repo specific topic
	subTopics[pubsub.GetRepoTopic(repo)] = struct{}{}

	return server.Config.Services.Scheduler.Publish(c, subTopics, message)
}
