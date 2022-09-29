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

	"github.com/woodpecker-ci/woodpecker/pipeline/frontend"
	"github.com/woodpecker-ci/woodpecker/pipeline/frontend/yaml"
	"github.com/woodpecker-ci/woodpecker/server/model"
	"github.com/woodpecker-ci/woodpecker/server/remote"
	"github.com/woodpecker-ci/woodpecker/server/shared"
)

func zeroSteps(pipeline *model.Pipeline, remoteYamlConfigs []*remote.FileMeta) bool {
	b := shared.ProcBuilder{
		Repo:  &model.Repo{},
		Curr:  pipeline,
		Last:  &model.Pipeline{},
		Netrc: &model.Netrc{},
		Secs:  []*model.Secret{},
		Regs:  []*model.Registry{},
		Link:  "",
		Yamls: remoteYamlConfigs,
	}

	pipelineItems, err := b.Build()
	if err != nil {
		return false
	}
	if len(pipelineItems) == 0 {
		return true
	}

	return false
}

// TODO: parse yaml once and not for each filter function
// Check if at least one pipeline step will be execute otherwise we will just ignore this webhook
func checkIfFiltered(pipeline *model.Pipeline, remoteYamlConfigs []*remote.FileMeta) (bool, error) {
	log.Trace().Msgf("hook.branchFiltered(): pipeline branch: '%s' pipeline event: '%s' config count: %d", pipeline.Branch, pipeline.Event, len(remoteYamlConfigs))

	matchMetadata := frontend.Metadata{
		Curr: frontend.Pipeline{
			Event: string(pipeline.Event),
			Commit: frontend.Commit{
				Branch: pipeline.Branch,
			},
		},
	}

	for _, remoteYamlConfig := range remoteYamlConfigs {
		parsedPipelineConfig, err := yaml.ParseBytes(remoteYamlConfig.Data)
		if err != nil {
			log.Trace().Msgf("parse config '%s': %s", remoteYamlConfig.Name, err)
			return false, err
		}
		log.Trace().Msgf("config '%s': %#v", remoteYamlConfig.Name, parsedPipelineConfig)

		// ignore if the pipeline was filtered by matched constraints
		if !parsedPipelineConfig.When.Match(matchMetadata, true) {
			continue
		}

		// ignore if the pipeline was filtered by the branch (legacy)
		if !parsedPipelineConfig.Branches.Match(pipeline.Branch) {
			continue
		}

		// at least one config yielded in a valid run.
		return false, nil
	}

	// no configs yielded a valid run.
	return true, nil
}
