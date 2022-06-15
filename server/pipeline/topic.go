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
	"strconv"

	"github.com/woodpecker-ci/woodpecker/server"
	"github.com/woodpecker-ci/woodpecker/server/model"
	"github.com/woodpecker-ci/woodpecker/server/pubsub"
)

// publishToTopic publishes message to UI clients
func publishToTopic(c context.Context, build *model.Build, repo *model.Repo) (err error) {
	message := pubsub.Message{
		Labels: map[string]string{
			"repo":    repo.FullName,
			"private": strconv.FormatBool(repo.IsSCMPrivate),
		},
	}
	buildCopy := *build
	if buildCopy.Procs, err = model.Tree(buildCopy.Procs); err != nil {
		return err
	}

	message.Data, _ = json.Marshal(model.Event{
		Repo:  *repo,
		Build: buildCopy,
	})
	return server.Config.Services.Pubsub.Publish(c, "topic/events", message)
}
