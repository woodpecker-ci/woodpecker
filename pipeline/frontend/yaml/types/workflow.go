package types

import (
	"go.woodpecker-ci.org/woodpecker/pipeline/frontend/yaml/constraint"
	"go.woodpecker-ci.org/woodpecker/pipeline/frontend/yaml/types/base"
)

type (
	// Workflow defines a workflow configuration.
	Workflow struct {
		When      constraint.When `yaml:"when,omitempty"`
		Platform  string          `yaml:"platform,omitempty"`
		Workspace Workspace       `yaml:"workspace,omitempty"`
		Clone     ContainerList   `yaml:"clone,omitempty"`
		Steps     ContainerList   `yaml:"steps,omitempty"`
		Services  ContainerList   `yaml:"services,omitempty"`
		Labels    base.SliceOrMap `yaml:"labels,omitempty"`
		DependsOn []string        `yaml:"depends_on,omitempty"`
		RunsOn    []string        `yaml:"runs_on,omitempty"`
		SkipClone bool            `yaml:"skip_clone"`
		// Undocumented
		Cache    base.StringOrSlice `yaml:"cache,omitempty"`
		Networks WorkflowNetworks   `yaml:"networks,omitempty"`
		Volumes  WorkflowVolumes    `yaml:"volumes,omitempty"`
		// Deprecated
		BranchesDontUseIt *constraint.List `yaml:"branches,omitempty"`
		// Deprecated
		PipelineDontUseIt ContainerList `yaml:"pipeline,omitempty"`
	}

	// Workspace defines a pipeline workspace.
	Workspace struct {
		Base string
		Path string
	}
)
