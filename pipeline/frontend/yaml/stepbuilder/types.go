package stepbuilder

import (
	backend_types "go.woodpecker-ci.org/woodpecker/v3/pipeline/backend/types"
)

type Item struct {
	Workflow  *Workflow
	Labels    map[string]string
	DependsOn []string
	RunsOn    []string
	Pending   bool
	Config    *backend_types.Config
}

type Workflow struct {
	ID      int64             `json:"id"`
	PID     int               `json:"pid"`
	Name    string            `json:"name"`
	Environ map[string]string `json:"environ,omitempty"`
	AxisID  int               `json:"-"`
}
