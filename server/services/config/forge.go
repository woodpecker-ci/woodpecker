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
	"fmt"
	"strings"
	"time"

	"github.com/rs/zerolog/log"

	"go.woodpecker-ci.org/woodpecker/v2/server/forge"
	"go.woodpecker-ci.org/woodpecker/v2/server/forge/types"
	"go.woodpecker-ci.org/woodpecker/v2/server/model"
	"go.woodpecker-ci.org/woodpecker/v2/shared/constant"
)

type forgeFetcher struct {
	timeout    time.Duration
	retryCount uint
}

func NewForge(timeout time.Duration, retries uint) Service {
	return &forgeFetcher{
		timeout:    timeout,
		retryCount: retries,
	}
}

func (f *forgeFetcher) Fetch(ctx context.Context, forge forge.Forge, user *model.User, repo *model.Repo, pipeline *model.Pipeline, oldConfigData []*types.FileMeta, restart bool) (files []*types.FileMeta, err error) {
	// skip fetching if we are restarting and have the old config
	if restart && len(oldConfigData) > 0 {
		return oldConfigData, nil
	}

	ffc := &forgeFetcherContext{
		forge:    forge,
		user:     user,
		repo:     repo,
		pipeline: pipeline,
		timeout:  f.timeout,
	}

	// try to fetch multiple times
	for i := 0; i < int(f.retryCount); i++ {
		files, err = ffc.fetch(ctx, strings.TrimSpace(repo.Config))
		if err != nil {
			log.Trace().Err(err).Msgf("Attempt #%d failed", i+1)
		}
		if errors.Is(err, context.DeadlineExceeded) {
			continue
		}
	}

	return
}

type forgeFetcherContext struct {
	forge    forge.Forge
	user     *model.User
	repo     *model.Repo
	pipeline *model.Pipeline
	timeout  time.Duration
}

// fetch attempts to fetch the configuration file(s) for the given config string.
func (f *forgeFetcherContext) fetch(c context.Context, config string) ([]*types.FileMeta, error) {
	ctx, cancel := context.WithTimeout(c, f.timeout)
	defer cancel()

	if len(config) > 0 {
		log.Trace().Msgf("configFetcher[%s]: use user config '%s'", f.repo.FullName, config)

		// could be adapted to allow the user to supply a list like we do in the defaults
		configs := []string{config}

		fileMetas, err := f.getFirstAvailableConfig(ctx, configs)
		if err == nil {
			return fileMetas, nil
		}

		return nil, fmt.Errorf("user defined config '%s' not found: %w", config, err)
	}

	log.Trace().Msgf("configFetcher[%s]: user did not define own config, following default procedure", f.repo.FullName)
	// for the order see shared/constants/constants.go
	fileMetas, err := f.getFirstAvailableConfig(ctx, constant.DefaultConfigOrder[:])
	if err == nil {
		return fileMetas, nil
	}

	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
		return nil, fmt.Errorf("configFetcher: fallback did not find config: %w", err)
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

func (f *forgeFetcherContext) checkPipelineFile(c context.Context, config string) ([]*types.FileMeta, error) {
	file, err := f.forge.File(c, f.user, f.repo, f.pipeline, config)

	if err == nil && len(file) != 0 {
		log.Trace().Msgf("configFetcher[%s]: found file '%s'", f.repo.FullName, config)

		return []*types.FileMeta{{
			Name: config,
			Data: file,
		}}, nil
	}

	return nil, err
}

func (f *forgeFetcherContext) getFirstAvailableConfig(c context.Context, configs []string) ([]*types.FileMeta, error) {
	var forgeErr []error
	for _, fileOrFolder := range configs {
		if strings.HasSuffix(fileOrFolder, "/") {
			// config is a folder
			log.Trace().Msgf("fetching %s from forge", fileOrFolder)
			files, err := f.forge.Dir(c, f.user, f.repo, f.pipeline, strings.TrimSuffix(fileOrFolder, "/"))
			// if folder is not supported we will get a "Not implemented" error and continue
			if err != nil {
				if !(errors.Is(err, types.ErrNotImplemented) || errors.Is(err, &types.ErrConfigNotFound{})) {
					log.Error().Err(err).Str("repo", f.repo.FullName).Str("user", f.user.Login).Msgf("could not get folder from forge: %s", err)
					forgeErr = append(forgeErr, err)
				}
				continue
			}
			files = filterPipelineFiles(files)
			if len(files) != 0 {
				log.Trace().Msgf("configFetcher[%s]: found %d files in '%s'", f.repo.FullName, len(files), fileOrFolder)
				return files, nil
			}
		}

		// config is a file
		if fileMeta, err := f.checkPipelineFile(c, fileOrFolder); err == nil {
			log.Trace().Msgf("configFetcher[%s]: found file: '%s'", f.repo.FullName, fileOrFolder)
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
