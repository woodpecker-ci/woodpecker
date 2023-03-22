// Copyright 2022 Woodpecker Authors
// Copyright 2018 Drone.IO Inc.
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

package bitbucketserver

import (
	"encoding/json"
	"net/http"

	"github.com/woodpecker-ci/woodpecker/server/forge/bitbucketserver/internal"
	"github.com/woodpecker-ci/woodpecker/server/model"
)

// parseHook parses a Bitbucket hook from an http.Request request and returns
// Repo and Pipeline detail. TODO: find a way to support PR hooks
func parseHook(r *http.Request, baseURL string) (*model.Repo, *model.Pipeline, error) {
	hook := new(internal.PostHook)
	if err := json.NewDecoder(r.Body).Decode(hook); err != nil {
		return nil, nil, err
	}
	pipeline := convertPushHook(hook, baseURL)
	repo := convertRepo(&hook.Repository, &model.Perm{})

	return repo, pipeline, nil
}
