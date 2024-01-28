package registryservice

import (
	"go.woodpecker-ci.org/woodpecker/v2/server/model"
)

type argumentsFindDelete struct {
	Repo *model.Repo `json:"repo"`
	Name string      `json:"name"`
}

type argumentsCreateUpdate struct {
	Repo     *model.Repo     `json:"repo"`
	Registry *model.Registry `json:"registry"`
}

type argumentsList struct {
	Repo        *model.Repo        `json:"repo"`
	ListOptions *model.ListOptions `json:"list_options"`
}
