package stepbuilder

import (
	backend_types "go.woodpecker-ci.org/woodpecker/v3/pipeline/backend/types"
	"go.woodpecker-ci.org/woodpecker/v3/server/model"
)

type Item struct {
	Workflow  *model.Workflow // TODO: get rid of server type in this package
	Labels    map[string]string
	DependsOn []string
	RunsOn    []string
	Config    *backend_types.Config
}
