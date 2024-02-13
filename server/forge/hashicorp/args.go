// Copyright 2024 Woodpecker Authors
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

package hashicorp

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

type responseLogin struct {
	User        *model.User `json:"user"`
	RedirectURL string      `json:"redirect_url"`
}

type httpRequest struct {
	Method string              `json:"method"`
	URL    string              `json:"url"`
	Header map[string][]string `json:"header"`
	Form   map[string][]string `json:"form"`
	Body   []byte              `json:"body"`
}
