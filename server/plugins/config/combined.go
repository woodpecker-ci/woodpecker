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
	"errors"
	"strings"
	"time"

	"github.com/rs/zerolog/log"

	"go.woodpecker-ci.org/woodpecker/v2/server/forge"
	"go.woodpecker-ci.org/woodpecker/v2/server/forge/types"
	"go.woodpecker-ci.org/woodpecker/v2/server/model"
)

type forgeFetcher struct {
	timeout time.Duration
}

func NewForge(timeout time.Duration) Service {
	return &forgeFetcher{
		timeout: timeout,
	}
}

func (f *forgeFetcher) Fetch(ctx context.Context, forge forge.Forge, user *model.User, repo *model.Repo, pipeline *model.Pipeline) ([]*types.FileMeta, error) {
	var files []*types.FileMeta
	var err error

	// if cf.configExtension != nil {
	// 	fetchCtx, cancel := context.WithTimeout(ctx, cf.timeout)
	// 	defer cancel() // ok here as we only try http fetching once, returning on fail and success

	// 	log.Trace().Msgf("configFetcher[%s]: getting config from external http service", cf.repo.FullName)
	// 	netrc, err := cf.forge.Netrc(cf.user, cf.repo)
	// 	if err != nil {
	// 		return nil, fmt.Errorf("could not get Netrc data from forge: %w", err)
	// 	}

	// 	newConfigs, useOld, err := cf.configExtension.FetchConfig(fetchCtx, cf.repo, cf.pipeline, files, netrc)
	// 	if err != nil {
	// 		log.Error().Err(err).Msg("could not fetch config via http")
	// 		return nil, fmt.Errorf("could not fetch config via http: %w", err)
	// 	}

	// 	if !useOld {
	// 		return newConfigs, nil
	// 	}
	// }

	// try to fetch 3 times
	configFetcher := &forgeFetcherContext{
		forge:    forge,
		user:     user,
		repo:     repo,
		pipeline: pipeline,
		timeout:  f.timeout,
	}
	for i := 0; i < 3; i++ {
		files, err = configFetcher.fetch(ctx, strings.TrimSpace(repo.Config))
		if err != nil {
			log.Trace().Err(err).Msgf("%d. try failed", i+1)
		}
		if errors.Is(err, context.DeadlineExceeded) {
			continue
		}
	}

	return files, err
}
