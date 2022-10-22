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

package shared

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/rs/zerolog/log"
	"github.com/woodpecker-ci/woodpecker/server/plugins/config"

	"github.com/woodpecker-ci/woodpecker/server/model"
	"github.com/woodpecker-ci/woodpecker/server/remote"
)

type ConfigFetcher interface {
	Fetch(ctx context.Context) (files []*remote.FileMeta, err error)
}

// TODO(974) move to new package

type configFetcher struct {
	remote          remote.Remote
	user            *model.User
	repo            *model.Repo
	pipeline        *model.Pipeline
	configExtension config.Extension
}

func NewConfigFetcher(remote remote.Remote, configExtension config.Extension, user *model.User, repo *model.Repo, pipeline *model.Pipeline) ConfigFetcher {
	return &configFetcher{
		remote:          remote,
		user:            user,
		repo:            repo,
		pipeline:        pipeline,
		configExtension: configExtension,
	}
}

// configFetchTimeout determine seconds the configFetcher wait until cancel fetch process
var configFetchTimeout = time.Second * 3

// Fetch pipeline config from source forge
func (cf *configFetcher) Fetch(ctx context.Context) (files []*remote.FileMeta, err error) {
	log.Trace().Msgf("Start Fetching config for '%s'", cf.repo.FullName)

	// try to fetch 3 times
	for i := 0; i < 3; i++ {
		files, err = cf.fetch(ctx, configFetchTimeout, strings.TrimSpace(cf.repo.Config))
		if err != nil {
			log.Trace().Err(err).Msgf("%d. try failed", i+1)
		}
		if errors.Is(err, context.DeadlineExceeded) {
			continue
		}

		if cf.configExtension != nil && cf.configExtension.IsConfigured() {
			fetchCtx, cancel := context.WithTimeout(ctx, configFetchTimeout)
			defer cancel() // ok here as we only try http fetching once, returning on fail and success

			log.Trace().Msgf("ConfigFetch[%s]: getting config from external http service", cf.repo.FullName)
			newConfigs, useOld, err := cf.configExtension.FetchConfig(fetchCtx, cf.repo, cf.pipeline, files)
			if err != nil {
				log.Error().Msg("Got error " + err.Error())
				return nil, fmt.Errorf("On Fetching config via http : %w", err)
			}

			if !useOld {
				return newConfigs, nil
			}
		}

		return
	}
	return
}

// fetch config by timeout
// TODO: deduplicate code
func (cf *configFetcher) fetch(c context.Context, timeout time.Duration, config string) ([]*remote.FileMeta, error) {
	ctx, cancel := context.WithTimeout(c, timeout)
	defer cancel()

	if len(config) > 0 {
		log.Trace().Msgf("ConfigFetch[%s]: use user config '%s'", cf.repo.FullName, config)
		// either a file
		if !strings.HasSuffix(config, "/") {
			file, err := cf.remote.File(ctx, cf.user, cf.repo, cf.pipeline, config)
			if err == nil && len(file) != 0 {
				log.Trace().Msgf("ConfigFetch[%s]: found file '%s'", cf.repo.FullName, config)
				return []*remote.FileMeta{{
					Name: config,
					Data: file,
				}}, nil
			}
		}

		// or a folder
		files, err := cf.remote.Dir(ctx, cf.user, cf.repo, cf.pipeline, strings.TrimSuffix(config, "/"))
		if err == nil && len(files) != 0 {
			log.Trace().Msgf("ConfigFetch[%s]: found %d files in '%s'", cf.repo.FullName, len(files), config)
			return filterPipelineFiles(files), nil
		}

		return nil, fmt.Errorf("config '%s' not found: %w", config, err)
	}

	log.Trace().Msgf("ConfigFetch[%s]: user did not defined own config follow default procedure", cf.repo.FullName)
	// no user defined config so try .woodpecker/*.yml -> .woodpecker.yml -> .drone.yml

	// test .woodpecker/ folder
	// if folder is not supported we will get a "Not implemented" error and continue
	config = ".woodpecker"
	files, err := cf.remote.Dir(ctx, cf.user, cf.repo, cf.pipeline, config)
	files = filterPipelineFiles(files)
	if err == nil && len(files) != 0 {
		log.Trace().Msgf("ConfigFetch[%s]: found %d files in '%s'", cf.repo.FullName, len(files), config)
		return files, nil
	}

	config = ".woodpecker.yml"
	file, err := cf.remote.File(ctx, cf.user, cf.repo, cf.pipeline, config)
	if err == nil && len(file) != 0 {
		log.Trace().Msgf("ConfigFetch[%s]: found file '%s'", cf.repo.FullName, config)
		return []*remote.FileMeta{{
			Name: config,
			Data: file,
		}}, nil
	}

	config = ".drone.yml"
	file, err = cf.remote.File(ctx, cf.user, cf.repo, cf.pipeline, config)
	if err == nil && len(file) != 0 {
		log.Trace().Msgf("ConfigFetch[%s]: found file '%s'", cf.repo.FullName, config)
		return []*remote.FileMeta{{
			Name: config,
			Data: file,
		}}, nil
	}

	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
		return []*remote.FileMeta{}, fmt.Errorf("ConfigFetcher: Fallback did not found config: %w", err)
	}
}

func filterPipelineFiles(files []*remote.FileMeta) []*remote.FileMeta {
	var res []*remote.FileMeta

	for _, file := range files {
		if strings.HasSuffix(file.Name, ".yml") {
			res = append(res, file)
		}
	}

	return res
}
