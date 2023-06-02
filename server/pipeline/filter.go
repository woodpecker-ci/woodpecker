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

	"github.com/woodpecker-ci/woodpecker/pipeline"
	"github.com/woodpecker-ci/woodpecker/pipeline/frontend/yaml"
	"github.com/woodpecker-ci/woodpecker/server"
	forge_types "github.com/woodpecker-ci/woodpecker/server/forge/types"
	"github.com/woodpecker-ci/woodpecker/server/model"
)

func zeroSteps(currentPipeline *model.Pipeline, forgeYamlConfigs []*forge_types.FileMeta) bool {
	b := pipeline.StepBuilder{
		Repo:  &model.Repo{},
		Curr:  currentPipeline,
		Last:  &model.Pipeline{},
		Netrc: &model.Netrc{},
		Secs:  []*model.Secret{},
		Regs:  []*model.Registry{},
		Link:  "",
		Yamls: forgeYamlConfigs,
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
func checkIfFiltered(repo *model.Repo, p *model.Pipeline, forgeYamlConfigs []*forge_types.FileMeta) (bool, error) {
	log.Trace().Msgf("hook.branchFiltered(): pipeline branch: '%s' pipeline event: '%s' config count: %d", p.Branch, p.Event, len(forgeYamlConfigs))

	matchMetadata := pipeline.MetadataFromStruct(server.Config.Services.Forge, repo, p, nil, nil, "")

	for _, forgeYamlConfig := range forgeYamlConfigs {
		parsedPipelineConfig, err := yaml.ParseBytes(forgeYamlConfig.Data)
		if err != nil {
			log.Trace().Msgf("parse config '%s': %s", forgeYamlConfig.Name, err)
			return false, err
		}
		log.Trace().Msgf("config '%s': %#v", forgeYamlConfig.Name, parsedPipelineConfig)

		// ignore if the pipeline was filtered by matched constraints
		if match, err := parsedPipelineConfig.When.Match(matchMetadata, true); !match && err == nil {
			continue
		} else if err != nil {
			return false, err
		}

		// at least one config yielded in a valid run.
		return false, nil
	}

	// no configs yielded a valid run.
	return true, nil
}
