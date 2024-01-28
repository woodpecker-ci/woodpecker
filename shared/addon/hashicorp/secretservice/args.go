package secretservice

import (
	"go.woodpecker-ci.org/woodpecker/v2/server/model"
)

type argumentsFindDelete struct {
	Repo *model.Repo `json:"repo"`
	Name string      `json:"name"`
}

type argumentsCreateUpdate struct {
	Repo   *model.Repo   `json:"repo"`
	Secret *model.Secret `json:"secret"`
}

type argumentsList struct {
	Repo        *model.Repo        `json:"repo"`
	ListOptions *model.ListOptions `json:"list_options"`
}

type argumentsListPipeline struct {
	Repo        *model.Repo        `json:"repo"`
	Pipeline    *model.Pipeline    `json:"pipeline"`
	ListOptions *model.ListOptions `json:"list_options"`
}

type argumentsOrgFindDelete struct {
	OrgID int64  `json:"org_id"`
	Name  string `json:"name"`
}

type argumentsOrgCreateUpdate struct {
	OrgID  int64         `json:"org_id"`
	Secret *model.Secret `json:"secret"`
}

type argumentsOrgList struct {
	OrgID       int64              `json:"org_id"`
	ListOptions *model.ListOptions `json:"list_options"`
}
