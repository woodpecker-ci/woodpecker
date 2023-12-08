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

import (
	"crypto/sha256"
	"fmt"

	"go.woodpecker-ci.org/woodpecker/v2/pipeline"
	forge_types "go.woodpecker-ci.org/woodpecker/v2/server/forge/types"
	"go.woodpecker-ci.org/woodpecker/v2/server/model"
	"go.woodpecker-ci.org/woodpecker/v2/server/store"
)

func findOrPersistPipelineConfig(store store.Store, currentPipeline *model.Pipeline, forgeYamlConfig *forge_types.FileMeta) (*model.Config, error) {
	sha := fmt.Sprintf("%x", sha256.Sum256(forgeYamlConfig.Data))
	conf, err := store.ConfigFindIdentical(currentPipeline.RepoID, sha)
	if err != nil {
		conf = &model.Config{
			RepoID: currentPipeline.RepoID,
			Data:   forgeYamlConfig.Data,
			Hash:   sha,
			Name:   pipeline.SanitizePath(forgeYamlConfig.Name),
		}
		err = store.ConfigCreate(conf)
		if err != nil {
			// retry in case we receive two hooks at the same time
			conf, err = store.ConfigFindIdentical(currentPipeline.RepoID, sha)
			if err != nil {
				return nil, err
			}
		}
	}

	return conf, nil
}
