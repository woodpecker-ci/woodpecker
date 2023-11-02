// Copyright 2023 Woodpecker Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package types

import (
	"github.com/woodpecker-ci/woodpecker/pipeline/frontend/yaml/constraint"
	"github.com/woodpecker-ci/woodpecker/pipeline/frontend/yaml/types/base"
)

type (
	// Workflow defines a workflow configuration.
	Workflow struct {
		When      constraint.When `yaml:"when,omitempty"`
		Workspace Workspace       `yaml:"workspace,omitempty"`
		Clone     ContainerList   `yaml:"clone,omitempty"`
		Steps     ContainerList   `yaml:"steps,omitempty"`
		Services  ContainerList   `yaml:"services,omitempty"`
		Labels    base.SliceOrMap `yaml:"labels,omitempty"`
		DependsOn []string        `yaml:"depends_on,omitempty"`
		RunsOn    []string        `yaml:"runs_on,omitempty"`
		SkipClone bool            `yaml:"skip_clone"`
		Version   int             `yaml:"version"`

		// Undocumented
		Cache    base.StringOrSlice `yaml:"cache,omitempty"`
		Networks WorkflowNetworks   `yaml:"networks,omitempty"`
		Volumes  WorkflowVolumes    `yaml:"volumes,omitempty"`

		// Deprecated
		PlatformDontUseIt string `yaml:"platform,omitempty"` // TODO: remove in next major version
		// Deprecated
		BranchesDontUseIt *constraint.List `yaml:"branches,omitempty"` // TODO: remove in next major version
		// Deprecated
		PipelineDontUseIt ContainerList `yaml:"pipeline,omitempty"` // TODO: remove in next major version
	}

	// Workspace defines a pipeline workspace.
	Workspace struct {
		Base string
		Path string
	}
)
