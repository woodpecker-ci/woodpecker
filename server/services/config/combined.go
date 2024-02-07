// Copyright 2024 Woodpecker Authors
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

package config

import (
	"context"

	"github.com/rs/zerolog/log"

	"go.woodpecker-ci.org/woodpecker/v2/server/forge"
	"go.woodpecker-ci.org/woodpecker/v2/server/forge/types"
	"go.woodpecker-ci.org/woodpecker/v2/server/model"
)

type combined struct {
	forgeService Service
	httpService  Service
}

func NewCombined(forgeService, httpService Service) Service {
	return &combined{
		forgeService: forgeService,
		httpService:  httpService,
	}
}

func (c *combined) Fetch(ctx context.Context, forge forge.Forge, user *model.User, repo *model.Repo, pipeline *model.Pipeline) (files []*types.FileMeta, err error) {
	files, err = c.forgeService.Fetch(ctx, forge, user, repo, pipeline)
	if err != nil {
		return nil, err
	}

	if c.httpService != nil {
		// TODO(anbraten): This is a hack to get the current configs into the http service
		_httpService, ok := c.httpService.(*http)
		if !ok {
			log.Err(err).Msg("http service is not of type http")
			return files, nil
		}
		_httpService.currentConfigs = files

		httpFiles, err := c.httpService.Fetch(ctx, forge, user, repo, pipeline)
		if err != nil {
			log.Err(err).Msg("failed to fetch config from http service using forge config instead")
			return files, nil
		}

		files = httpFiles
	}

	return files, nil
}
