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

// TODO(770): pipeline filter should not belong here

import (
	"github.com/rs/zerolog/log"

	"github.com/woodpecker-ci/woodpecker/pipeline/frontend/yaml"
	"github.com/woodpecker-ci/woodpecker/server/model"
	"github.com/woodpecker-ci/woodpecker/server/remote"
	"github.com/woodpecker-ci/woodpecker/server/shared"
)

func zeroSteps(build *model.Build, remoteYamlConfigs []*remote.FileMeta) bool {
	b := shared.ProcBuilder{
		Repo:  &model.Repo{},
		Curr:  build,
		Last:  &model.Build{},
		Netrc: &model.Netrc{},
		Secs:  []*model.Secret{},
		Regs:  []*model.Registry{},
		Link:  "",
		Yamls: remoteYamlConfigs,
	}

	buildItems, err := b.Build()
	if err != nil {
		return false
	}
	if len(buildItems) == 0 {
		return true
	}

	return false
}

// TODO: parse yaml once and not for each filter function
func branchFiltered(build *model.Build, remoteYamlConfigs []*remote.FileMeta) (bool, error) {
	log.Trace().Msgf("hook.branchFiltered(): build branch: '%s' build event: '%s' config count: %d", build.Branch, build.Event, len(remoteYamlConfigs))

	if build.Event == model.EventTag || build.Event == model.EventDeploy {
		return false, nil
	}

	for _, remoteYamlConfig := range remoteYamlConfigs {
		parsedPipelineConfig, err := yaml.ParseBytes(remoteYamlConfig.Data)
		if err != nil {
			log.Trace().Msgf("parse config '%s': %s", remoteYamlConfig.Name, err)
			return false, err
		}
		log.Trace().Msgf("config '%s': %#v", remoteYamlConfig.Name, parsedPipelineConfig)

		if parsedPipelineConfig.Branches.Match(build.Branch) {
			return false, nil
		}
	}

	return true, nil
}
