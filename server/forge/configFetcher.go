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

	"github.com/woodpecker-ci/woodpecker/server/forge/types"
	"github.com/woodpecker-ci/woodpecker/server/model"
	"github.com/woodpecker-ci/woodpecker/server/plugins/config"
	"github.com/woodpecker-ci/woodpecker/shared/constant"
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
	timeout         time.Duration
}

func NewConfigFetcher(forge Forge, timeout time.Duration, configExtension config.Extension, user *model.User, repo *model.Repo, pipeline *model.Pipeline) ConfigFetcher {
	return &configFetcher{
		forge:           forge,
		user:            user,
		repo:            repo,
		pipeline:        pipeline,
		configExtension: configExtension,
		timeout:         timeout,
	}
}

// Fetch pipeline config from source forge
func (cf *configFetcher) Fetch(ctx context.Context) (files []*types.FileMeta, err error) {
	log.Trace().Msgf("Start Fetching config for '%s'", cf.repo.FullName)

	// try to fetch 3 times
	for i := 0; i < 3; i++ {
		files, err = cf.fetch(ctx, time.Second*cf.timeout, strings.TrimSpace(cf.repo.Config))
		if err != nil {
			log.Trace().Err(err).Msgf("%d. try failed", i+1)
		}
		if errors.Is(err, context.DeadlineExceeded) {
			continue
		}

		if cf.configExtension != nil && cf.configExtension.IsConfigured() {
			fetchCtx, cancel := context.WithTimeout(ctx, cf.timeout)
			defer cancel() // ok here as we only try http fetching once, returning on fail and success

			log.Trace().Msgf("ConfigFetch[%s]: getting config from external http service", cf.repo.FullName)
			newConfigs, useOld, err := cf.configExtension.FetchConfig(fetchCtx, cf.repo, cf.pipeline, files)
			if err != nil {
				log.Error().Msg("Got error " + err.Error())
				return nil, fmt.Errorf("On Fetching config via http : %s", err)
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

		fileMeta, err := cf.getFirstAvailableConfig(ctx, configs, true)
		if err == nil {
			return fileMeta, err
		}

		return nil, fmt.Errorf("user defined config '%s' not found: %s", config, err)
	}

	log.Trace().Msgf("ConfigFetch[%s]: user did not defined own config, following default procedure", cf.repo.FullName)
	// for the order see shared/constants/constants.go
	fileMeta, err := cf.getFirstAvailableConfig(ctx, constant.DefaultConfigOrder[:], false)
	if err == nil {
		return fileMeta, err
	}

	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
		return []*types.FileMeta{}, fmt.Errorf("ConfigFetcher: Fallback did not find config: %s", err)
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

func (cf *configFetcher) checkPipelineFile(c context.Context, config string) (fileMeta []*types.FileMeta, found bool) {
	file, err := cf.forge.File(c, cf.user, cf.repo, cf.pipeline, config)

	if err == nil && len(file) != 0 {
		log.Trace().Msgf("ConfigFetch[%s]: found file '%s'", cf.repo.FullName, config)

		return []*types.FileMeta{{
			Name: config,
			Data: file,
		}}, true
	}

	return nil, false
}

func (cf *configFetcher) getFirstAvailableConfig(c context.Context, configs []string, userDefined bool) ([]*types.FileMeta, error) {
	userDefinedLog := ""
	if userDefined {
		userDefinedLog = "user defined"
	}

	for _, fileOrFolder := range configs {
		if strings.HasSuffix(fileOrFolder, "/") {
			// config is a folder
			files, err := cf.forge.Dir(c, cf.user, cf.repo, cf.pipeline, strings.TrimSuffix(fileOrFolder, "/"))
			// if folder is not supported we will get a "Not implemented" error and continue
			if err != nil && !errors.Is(err, types.ErrNotImplemented) {
				log.Error().Err(err).Str("repo", cf.repo.FullName).Str("user", cf.user.Login).Msg("could not get folder from forge")
			}
			files = filterPipelineFiles(files)
			if err == nil && len(files) != 0 {
				log.Trace().Msgf("ConfigFetch[%s]: found %d %s files in '%s'", cf.repo.FullName, len(files), userDefinedLog, fileOrFolder)
				return files, nil
			}
		}

		// config is a file
		if fileMeta, found := cf.checkPipelineFile(c, fileOrFolder); found {
			log.Trace().Msgf("ConfigFetch[%s]: found %s file: '%s'", cf.repo.FullName, userDefinedLog, fileOrFolder)
			return fileMeta, nil
		}
	}

	// nothing found
	return nil, fmt.Errorf("%s configs not found searched: %s", userDefinedLog, strings.Join(configs, ", "))
}
