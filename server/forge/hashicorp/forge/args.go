package forgeaddon

import (
	"go.woodpecker-ci.org/woodpecker/v2/server/model"
)

type argumentsAuth struct {
	Token  string `json:"token"`
	Secret string `json:"secret"`
}

type argumentsRepo struct {
	U        *model.User         `json:"u"`
	RemoteID model.ForgeRemoteID `json:"remote_id"`
	Owner    string              `json:"owner"`
	Name     string              `json:"name"`
}

type argumentsFileDir struct {
	U *model.User     `json:"u"`
	R *model.Repo     `json:"r"`
	B *model.Pipeline `json:"b"`
	F string          `json:"f"`
}

type argumentsStatus struct {
	U *model.User     `json:"u"`
	R *model.Repo     `json:"r"`
	B *model.Pipeline `json:"b"`
	P *model.Workflow `json:"p"`
}

type argumentsNetrc struct {
	U *model.User `json:"u"`
	R *model.Repo `json:"r"`
}

type argumentsActivateDeactivate struct {
	U    *model.User `json:"u"`
	R    *model.Repo `json:"r"`
	Link string      `json:"link"`
}

type argumentsBranchesPullRequests struct {
	U *model.User        `json:"u"`
	R *model.Repo        `json:"r"`
	P *model.ListOptions `json:"p"`
}

type argumentsBranchHead struct {
	U      *model.User `json:"u"`
	R      *model.Repo `json:"r"`
	Branch string      `json:"branch"`
}

type argumentsOrgMembershipOrg struct {
	U   *model.User `json:"u"`
	Org string      `json:"org"`
}

type responseHook struct {
	Repo     *model.Repo     `json:"repo"`
	Pipeline *model.Pipeline `json:"pipeline"`
}
