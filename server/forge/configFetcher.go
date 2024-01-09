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

package forge

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/rs/zerolog/log"

	"go.woodpecker-ci.org/woodpecker/v2/server/forge/types"
	"go.woodpecker-ci.org/woodpecker/v2/server/model"
	"go.woodpecker-ci.org/woodpecker/v2/server/plugins/config"
	"go.woodpecker-ci.org/woodpecker/v2/shared/constant"
)

type ConfigFetcher interface {
	Fetch(ctx context.Context) (files []*types.FileMeta, err error)
}

type configFetcher struct {
	forge           Forge
	user            *model.User
	repo            *model.Repo
	pipeline        *model.Pipeline
	configExtension config.Extension
	configPath      string
	timeout         time.Duration
}

func NewConfigFetcher(forge Forge, timeout time.Duration, configExtension config.Extension, user *model.User, repo *model.Repo, pipeline *model.Pipeline) ConfigFetcher {
	return &configFetcher{
		forge:           forge,
		user:            user,
		repo:            repo,
		pipeline:        pipeline,
		configExtension: configExtension,
		configPath:      repo.Config,
		timeout:         timeout,
	}
}

// Fetch pipeline config from source forge
func (cf *configFetcher) Fetch(ctx context.Context) (files []*types.FileMeta, err error) {
	log.Trace().Msgf("Start Fetching config for '%s'", cf.repo.FullName)

	// try to fetch 3 times
	for i := 0; i < 3; i++ {
		files, err = cf.fetch(ctx, cf.timeout, strings.TrimSpace(cf.configPath))
		if err != nil {
			log.Trace().Err(err).Msgf("%d. try failed", i+1)
		}
		if errors.Is(err, context.DeadlineExceeded) {
			continue
		}

		if cf.configExtension != nil {
			fetchCtx, cancel := context.WithTimeout(ctx, cf.timeout)
			defer cancel() // ok here as we only try http fetching once, returning on fail and success

			log.Trace().Msgf("ConfigFetch[%s]: getting config from external http service", cf.repo.FullName)
			netrc, err := cf.forge.Netrc(cf.user, cf.repo)
			if err != nil {
				return nil, fmt.Errorf("could not get Netrc data from forge: %w", err)
			}

			newConfigs, useOld, err := cf.configExtension.FetchConfig(fetchCtx, cf.repo, cf.pipeline, files, netrc)
			if err != nil {
				log.Error().Err(err).Msg("could not fetch config via http")
				return nil, fmt.Errorf("could not fetch config via http: %w", err)
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
func (cf *configFetcher) fetch(c context.Context, timeout time.Duration, config string) ([]*types.FileMeta, error) {
	ctx, cancel := context.WithTimeout(c, timeout)
	defer cancel()

	if len(config) > 0 {
		log.Trace().Msgf("ConfigFetch[%s]: use user config '%s'", cf.repo.FullName, config)

		// could be adapted to allow the user to supply a list like we do in the defaults
		configs := []string{config}

		fileMeta, err := cf.getFirstAvailableConfig(ctx, configs)
		if err == nil {
			return fileMeta, err
		}

		return nil, fmt.Errorf("user defined config '%s' not found: %w", config, err)
	}

	log.Trace().Msgf("ConfigFetch[%s]: user did not define own config, following default procedure", cf.repo.FullName)
	// for the order see shared/constants/constants.go
	fileMeta, err := cf.getFirstAvailableConfig(ctx, constant.DefaultConfigOrder[:])
	if err == nil {
		return fileMeta, err
	}

	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
		return []*types.FileMeta{}, fmt.Errorf("ConfigFetcher: Fallback did not find config: %w", err)
	}
}

func filterPipelineFiles(files []*types.FileMeta) []*types.FileMeta {
	var res []*types.FileMeta

	for _, file := range files {
		if strings.HasSuffix(file.Name, ".yml") || strings.HasSuffix(file.Name, ".yaml") {
			res = append(res, file)
		}
	}

	return res
}

func (cf *configFetcher) checkPipelineFile(c context.Context, config string) ([]*types.FileMeta, error) {
	file, err := cf.forge.File(c, cf.user, cf.repo, cf.pipeline, config)

	if err == nil && len(file) != 0 {
		log.Trace().Msgf("ConfigFetch[%s]: found file '%s'", cf.repo.FullName, config)

		return []*types.FileMeta{{
			Name: config,
			Data: file,
		}}, nil
	}

	return nil, err
}

func (cf *configFetcher) getFirstAvailableConfig(c context.Context, configs []string) ([]*types.FileMeta, error) {
	var forgeErr []error
	for _, fileOrFolder := range configs {
		if strings.HasSuffix(fileOrFolder, "/") {
			// config is a folder
			files, err := cf.forge.Dir(c, cf.user, cf.repo, cf.pipeline, strings.TrimSuffix(fileOrFolder, "/"))
			// if folder is not supported we will get a "Not implemented" error and continue
			if err != nil {
				if !(errors.Is(err, types.ErrNotImplemented) || errors.Is(err, &types.ErrConfigNotFound{})) {
					log.Error().Err(err).Str("repo", cf.repo.FullName).Str("user", cf.user.Login).Msg("could not get folder from forge")
					forgeErr = append(forgeErr, err)
				}
				continue
			}
			files = filterPipelineFiles(files)
			if len(files) != 0 {
				log.Trace().Msgf("ConfigFetch[%s]: found %d files in '%s'", cf.repo.FullName, len(files), fileOrFolder)
				return files, nil
			}
		}

		// config is a file
		if fileMeta, err := cf.checkPipelineFile(c, fileOrFolder); err == nil {
			log.Trace().Msgf("ConfigFetch[%s]: found file: '%s'", cf.repo.FullName, fileOrFolder)
			return fileMeta, nil
		} else if !errors.Is(err, &types.ErrConfigNotFound{}) {
			forgeErr = append(forgeErr, err)
		}
	}

	// got unexpected errors
	if len(forgeErr) != 0 {
		return nil, errors.Join(forgeErr...)
	}

	// nothing found
	return nil, &types.ErrConfigNotFound{Configs: configs}
}
