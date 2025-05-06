package stepbuilder

import (
	backend_types "go.woodpecker-ci.org/woodpecker/v3/pipeline/backend/types"
	"go.woodpecker-ci.org/woodpecker/v3/server/model"
)

type Item struct {
	Workflow  *Workflow
	Labels    map[string]string
	DependsOn []string
	RunsOn    []string
	Config    *backend_types.Config
}

type Workflow struct {
	ID    int64             `json:"id"`
	PID   int               `json:"pid"`
	Name  string            `json:"name"`
	State model.StatusValue `json:"state"` // TODO
	// State   string            `json:"state"` // TODO
	Environ map[string]string `json:"environ,omitempty"`
	AxisID  int               `json:"-"`
}
