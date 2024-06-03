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

package addon

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/rpc"
	"os/exec"

	"github.com/hashicorp/go-plugin"
	"github.com/rs/zerolog/log"

	"go.woodpecker-ci.org/woodpecker/v2/server/forge"
	"go.woodpecker-ci.org/woodpecker/v2/server/forge/types"
	"go.woodpecker-ci.org/woodpecker/v2/server/model"
)

// make sure RPC implements forge.Forge.
var _ forge.Forge = new(RPC)

func Load(file string) (forge.Forge, error) {
	client := plugin.NewClient(&plugin.ClientConfig{
		HandshakeConfig: HandshakeConfig,
		Plugins: map[string]plugin.Plugin{
			pluginKey: &Plugin{},
		},
		Cmd: exec.Command(file),
		Logger: &clientLogger{
			logger: log.With().Str("addon", file).Logger(),
		},
	})
	// TODO: defer client.Kill()

	rpcClient, err := client.Client()
	if err != nil {
		return nil, err
	}

	raw, err := rpcClient.Dispense(pluginKey)
	if err != nil {
		return nil, err
	}

	extension, _ := raw.(forge.Forge)
	return extension, nil
}

type RPC struct {
	client *rpc.Client
}

func (g *RPC) Name() string {
	var resp string
	_ = g.client.Call("Plugin.Name", nil, &resp)
	return resp
}

func (g *RPC) URL() string {
	var resp string
	_ = g.client.Call("Plugin.URL", nil, &resp)
	return resp
}

func (g *RPC) Login(_ context.Context, r *types.OAuthRequest) (*model.User, string, error) {
	args, err := json.Marshal(r)
	if err != nil {
		return nil, "", err
	}
	var jsonResp []byte
	err = g.client.Call("Plugin.Login", args, &jsonResp)
	if err != nil {
		return nil, "", err
	}

	var resp responseLogin
	err = json.Unmarshal(jsonResp, &resp)
	if err != nil {
		return nil, "", err
	}

	return resp.User.asModel(), resp.RedirectURL, nil
}

func (g *RPC) Auth(_ context.Context, token, secret string) (string, error) {
	args, err := json.Marshal(&argumentsAuth{
		Token:  token,
		Secret: secret,
	})
	if err != nil {
		return "", err
	}
	var resp string
	return resp, g.client.Call("Plugin.Auth", args, &resp)
}

func (g *RPC) Teams(_ context.Context, u *model.User) ([]*model.Team, error) {
	args, err := json.Marshal(modelUserFromModel(u))
	if err != nil {
		return nil, err
	}
	var jsonResp []byte
	err = g.client.Call("Plugin.Teams", args, &jsonResp)
	if err != nil {
		return nil, err
	}

	var resp []*model.Team
	return resp, json.Unmarshal(jsonResp, &resp)
}

func (g *RPC) Repo(_ context.Context, u *model.User, remoteID model.ForgeRemoteID, owner, name string) (*model.Repo, error) {
	args, err := json.Marshal(&argumentsRepo{
		U:        modelUserFromModel(u),
		RemoteID: remoteID,
		Owner:    owner,
		Name:     name,
	})
	if err != nil {
		return nil, err
	}
	var jsonResp []byte
	err = g.client.Call("Plugin.Repo", args, &jsonResp)
	if err != nil {
		return nil, err
	}

	var resp *modelRepo
	err = json.Unmarshal(jsonResp, resp)
	if err != nil {
		return nil, err
	}
	return resp.asModel(), nil
}

func (g *RPC) Repos(_ context.Context, u *model.User) ([]*model.Repo, error) {
	args, err := json.Marshal(modelUserFromModel(u))
	if err != nil {
		return nil, err
	}
	var jsonResp []byte
	err = g.client.Call("Plugin.Repos", args, &jsonResp)
	if err != nil {
		return nil, err
	}

	var resp []*modelRepo
	err = json.Unmarshal(jsonResp, &resp)
	if err != nil {
		return nil, err
	}
	var modelRepos []*model.Repo
	for _, repo := range resp {
		modelRepos = append(modelRepos, repo.asModel())
	}
	return modelRepos, nil
}

func (g *RPC) File(_ context.Context, u *model.User, r *model.Repo, b *model.Pipeline, f string) ([]byte, error) {
	args, err := json.Marshal(&argumentsFileDir{
		U: modelUserFromModel(u),
		R: modelRepoFromModel(r),
		B: b,
		F: f,
	})
	if err != nil {
		return nil, err
	}
	var resp []byte
	return resp, g.client.Call("Plugin.File", args, &resp)
}

func (g *RPC) Dir(_ context.Context, u *model.User, r *model.Repo, b *model.Pipeline, f string) ([]*types.FileMeta, error) {
	args, err := json.Marshal(&argumentsFileDir{
		U: modelUserFromModel(u),
		R: modelRepoFromModel(r),
		B: b,
		F: f,
	})
	if err != nil {
		return nil, err
	}
	var jsonResp []byte
	err = g.client.Call("Plugin.Dir", args, &jsonResp)
	if err != nil {
		return nil, err
	}
	var resp []*types.FileMeta
	return resp, json.Unmarshal(jsonResp, &resp)
}

func (g *RPC) Status(_ context.Context, u *model.User, r *model.Repo, b *model.Pipeline, p *model.Workflow) error {
	args, err := json.Marshal(&argumentsStatus{
		U: modelUserFromModel(u),
		R: modelRepoFromModel(r),
		B: b,
		P: p,
	})
	if err != nil {
		return err
	}
	var jsonResp []byte
	return g.client.Call("Plugin.Status", args, &jsonResp)
}

func (g *RPC) Netrc(u *model.User, r *model.Repo) (*model.Netrc, error) {
	args, err := json.Marshal(&argumentsNetrc{
		U: modelUserFromModel(u),
		R: modelRepoFromModel(r),
	})
	if err != nil {
		return nil, err
	}
	var jsonResp []byte
	err = g.client.Call("Plugin.Netrc", args, &jsonResp)
	if err != nil {
		return nil, err
	}
	var resp *model.Netrc
	return resp, json.Unmarshal(jsonResp, &resp)
}

func (g *RPC) Activate(_ context.Context, u *model.User, r *model.Repo, link string) error {
	args, err := json.Marshal(&argumentsActivateDeactivate{
		U:    modelUserFromModel(u),
		R:    modelRepoFromModel(r),
		Link: link,
	})
	if err != nil {
		return err
	}
	var jsonResp []byte
	return g.client.Call("Plugin.Activate", args, &jsonResp)
}

func (g *RPC) Deactivate(_ context.Context, u *model.User, r *model.Repo, link string) error {
	args, err := json.Marshal(&argumentsActivateDeactivate{
		U:    modelUserFromModel(u),
		R:    modelRepoFromModel(r),
		Link: link,
	})
	if err != nil {
		return err
	}
	var jsonResp []byte
	return g.client.Call("Plugin.Deactivate", args, &jsonResp)
}

func (g *RPC) Branches(_ context.Context, u *model.User, r *model.Repo, p *model.ListOptions) ([]string, error) {
	args, err := json.Marshal(&argumentsBranchesPullRequests{
		U: modelUserFromModel(u),
		R: modelRepoFromModel(r),
		P: p,
	})
	if err != nil {
		return nil, err
	}
	var jsonResp []byte
	err = g.client.Call("Plugin.Branches", args, &jsonResp)
	if err != nil {
		return nil, err
	}
	var resp []string
	return resp, json.Unmarshal(jsonResp, &resp)
}

func (g *RPC) BranchHead(_ context.Context, u *model.User, r *model.Repo, branch string) (*model.Commit, error) {
	args, err := json.Marshal(&argumentsBranchHead{
		U:      modelUserFromModel(u),
		R:      modelRepoFromModel(r),
		Branch: branch,
	})
	if err != nil {
		return nil, err
	}
	var jsonResp []byte
	err = g.client.Call("Plugin.BranchHead", args, &jsonResp)
	if err != nil {
		return nil, err
	}
	var resp *model.Commit
	return resp, json.Unmarshal(jsonResp, &resp)
}

func (g *RPC) PullRequests(_ context.Context, u *model.User, r *model.Repo, p *model.ListOptions) ([]*model.PullRequest, error) {
	args, err := json.Marshal(&argumentsBranchesPullRequests{
		U: modelUserFromModel(u),
		R: modelRepoFromModel(r),
		P: p,
	})
	if err != nil {
		return nil, err
	}
	var jsonResp []byte
	err = g.client.Call("Plugin.PullRequests", args, &jsonResp)
	if err != nil {
		return nil, err
	}
	var resp []*model.PullRequest
	return resp, json.Unmarshal(jsonResp, &resp)
}

func (g *RPC) Hook(_ context.Context, r *http.Request) (*model.Repo, *model.Pipeline, error) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		return nil, nil, err
	}
	args, err := json.Marshal(&httpRequest{
		Method: r.Method,
		URL:    r.URL.String(),
		Header: r.Header,
		Form:   r.Form,
		Body:   body,
	})
	if err != nil {
		return nil, nil, err
	}
	var jsonResp []byte
	err = g.client.Call("Plugin.Hook", args, &jsonResp)
	if err != nil {
		return nil, nil, err
	}
	var resp responseHook
	err = json.Unmarshal(jsonResp, &resp)
	if err != nil {
		return nil, nil, err
	}
	return resp.Repo.asModel(), resp.Pipeline, nil
}

func (g *RPC) OrgMembership(_ context.Context, u *model.User, org string) (*model.OrgPerm, error) {
	args, err := json.Marshal(&argumentsOrgMembershipOrg{
		U:   modelUserFromModel(u),
		Org: org,
	})
	if err != nil {
		return nil, err
	}
	var jsonResp []byte
	err = g.client.Call("Plugin.OrgMembership", args, &jsonResp)
	if err != nil {
		return nil, err
	}
	var resp *model.OrgPerm
	return resp, json.Unmarshal(jsonResp, &resp)
}

func (g *RPC) Org(_ context.Context, u *model.User, org string) (*model.Org, error) {
	args, err := json.Marshal(&argumentsOrgMembershipOrg{
		U:   modelUserFromModel(u),
		Org: org,
	})
	if err != nil {
		return nil, err
	}
	var jsonResp []byte
	err = g.client.Call("Plugin.Org", args, &jsonResp)
	if err != nil {
		return nil, err
	}
	var resp *model.Org
	return resp, json.Unmarshal(jsonResp, &resp)
}
