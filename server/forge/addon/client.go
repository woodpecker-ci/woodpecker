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
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/rpc"
	"os/exec"
	"sync"

	"github.com/hashicorp/go-plugin"
	"github.com/rs/zerolog/log"

	"go.woodpecker-ci.org/woodpecker/v3/pipeline/frontend/metadata"
	"go.woodpecker-ci.org/woodpecker/v3/server/forge"
	"go.woodpecker-ci.org/woodpecker/v3/server/forge/types"
	"go.woodpecker-ci.org/woodpecker/v3/server/model"
	"go.woodpecker-ci.org/woodpecker/v3/shared/logger"
)

// make sure RPC implements forge.Forge.
var _ forge.Forge = new(RPC)

// ErrAddon is an error returned by the addon itself, not by the addon system.
// Only the message of the error crosses the RPC boundary. Err is set to the
// error the server checks for, if the addon sent its kind.
type ErrAddon struct {
	Message string
	Err     error
}

func (err *ErrAddon) Error() string {
	return err.Message
}

func (err *ErrAddon) Unwrap() error {
	return err.Err
}

func (*ErrAddon) Is(target error) bool {
	_, ok := target.(*ErrAddon)
	return ok
}

// addonError turns an error returned by the addon into an ErrAddon and keeps
// errors of the addon system as they are.
func addonError(err error) error {
	if serverErr, ok := errors.AsType[rpc.ServerError](err); ok {
		return &ErrAddon{Message: string(serverErr)}
	}
	return err
}

func Load(file string) (forge.Forge, error) {
	client := plugin.NewClient(&plugin.ClientConfig{
		HandshakeConfig: HandshakeConfig,
		Plugins: map[string]plugin.Plugin{
			pluginKey: &Plugin{},
		},
		Cmd: exec.Command(file),
		Logger: &logger.AddonClientLogger{
			Logger: log.With().Str("addon", file).Logger(),
		},
	})
	success := false
	defer func() {
		if !success {
			client.Kill()
		}
	}()

	rpcClient, err := client.Client()
	if err != nil {
		return nil, err
	}

	raw, err := rpcClient.Dispense(pluginKey)
	if err != nil {
		return nil, err
	}

	extension, ok := raw.(*RPC)
	if !ok {
		return nil, fmt.Errorf("invalid forge addon client %T", raw)
	}
	state := &processState{current: &process{rpc: extension.client, client: client, done: make(chan struct{})}}
	extension.client = &processSlot{state: state}
	extension.slot = state
	success = true
	return extension, nil
}

type RPC struct {
	client interface{ Call(string, any, any) error }
	slot   *processState
}

type process struct {
	mu      sync.Mutex
	rpc     interface{ Call(string, any, any) error }
	client  *plugin.Client
	done    chan struct{}
	active  int
	closing bool
}

type processState struct {
	mu      sync.Mutex
	current *process
}

type processSlot struct {
	state *processState
}

func (s *processSlot) Call(method string, args, reply any) error {
	state := s.state
	state.mu.Lock()
	p := state.current
	if p == nil {
		state.mu.Unlock()
		return errors.New("forge addon is closed")
	}
	p.mu.Lock()
	if p.closing {
		p.mu.Unlock()
		state.mu.Unlock()
		return errors.New("forge addon is closing")
	}
	p.active++
	p.mu.Unlock()
	state.mu.Unlock()

	defer func() {
		p.mu.Lock()
		p.active--
		if p.closing && p.active == 0 {
			close(p.done)
		}
		p.mu.Unlock()
	}()
	return p.rpc.Call(method, args, reply)
}

func (g *RPC) Close() error {
	if g == nil || g.slot == nil {
		return nil
	}

	state := g.slot
	state.mu.Lock()
	p := state.current
	if p == nil {
		state.mu.Unlock()
		return nil
	}
	state.current = nil
	p.mu.Lock()
	p.closing = true
	if p.active == 0 {
		close(p.done)
	}
	p.mu.Unlock()
	state.mu.Unlock()

	<-p.done
	var resp []byte
	err := p.rpc.Call("Plugin.Close", []byte{}, &resp)
	p.client.Kill()
	return addonError(err)
}

func (g *RPC) Name() string {
	var resp string
	err := g.client.Call("Plugin.Name", []byte{}, &resp)
	if err != nil {
		log.Error().Err(err).Msg("addon Plugin.Name call failed")
	}
	return resp
}

// Config returns the metadata of the configured forge.
func (g *RPC) Config() metadata.Forge {
	var config metadata.Forge
	var jsonResp []byte
	err := g.client.Call("Plugin.Config", []byte{}, &jsonResp)
	if err == nil {
		err = json.Unmarshal(jsonResp, &config)
	}
	if err != nil {
		log.Error().Err(err).Msg("addon Plugin.Config call failed")
	}
	return config
}

func (g *RPC) Login(_ context.Context, r *types.OAuthRequest) (*model.User, string, error) {
	args, err := json.Marshal(r)
	if err != nil {
		return nil, "", err
	}
	var jsonResp []byte
	err = g.client.Call("Plugin.Login", args, &jsonResp)
	if err != nil {
		return nil, "", addonError(err)
	}

	var resp responseLogin
	err = json.Unmarshal(jsonResp, &resp)
	if err != nil {
		return nil, "", err
	}

	return resp.User.asModel(), resp.RedirectURL, nil
}

func (g *RPC) Teams(_ context.Context, u *model.User, p *model.ListOptions) ([]*model.Team, error) {
	args, err := json.Marshal(&argumentsTeams{
		U: modelUserFromModel(u),
		P: p,
	})
	if err != nil {
		return nil, err
	}
	var jsonResp []byte
	err = g.client.Call("Plugin.Teams", args, &jsonResp)
	if err != nil {
		return nil, addonError(err)
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
		return nil, addonError(err)
	}

	var resp modelRepo
	err = json.Unmarshal(jsonResp, &resp)
	if err != nil {
		return nil, err
	}
	return resp.asModel(), nil
}

func (g *RPC) Repos(_ context.Context, u *model.User, p *model.ListOptions) ([]*model.Repo, error) {
	args, err := json.Marshal(&argumentsRepos{
		U: modelUserFromModel(u),
		P: p,
	})
	if err != nil {
		return nil, err
	}
	var jsonResp []byte
	err = g.client.Call("Plugin.Repos", args, &jsonResp)
	if err != nil {
		return nil, addonError(err)
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
	return resp, addonError(
		g.client.Call("Plugin.File", args, &resp),
	)
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
		return nil, addonError(err)
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
	return addonError(
		g.client.Call("Plugin.Status", args, &jsonResp),
	)
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
		return nil, addonError(err)
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
	return addonError(
		g.client.Call("Plugin.Activate", args, &jsonResp),
	)
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
	return addonError(
		g.client.Call("Plugin.Deactivate", args, &jsonResp),
	)
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
		return nil, addonError(err)
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
		return nil, addonError(err)
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
		return nil, addonError(err)
	}
	var resp []*model.PullRequest
	return resp, json.Unmarshal(jsonResp, &resp)
}

func (g *RPC) Hook(_ context.Context, r *http.Request, secret string) (*model.Repo, *model.Pipeline, error) {
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
		Secret: secret,
	})
	if err != nil {
		return nil, nil, err
	}
	var jsonResp []byte
	err = g.client.Call("Plugin.Hook", args, &jsonResp)
	if err != nil {
		return nil, nil, addonError(err)
	}
	var resp responseHook
	err = json.Unmarshal(jsonResp, &resp)
	if err != nil {
		return nil, nil, err
	}
	if resp.ErrorKind != "" {
		return nil, nil, decodeErrorKinds(resp.ErrorKind, resp.ErrorMessage)
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
		return nil, addonError(err)
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
		return nil, addonError(err)
	}
	var resp *model.Org
	return resp, json.Unmarshal(jsonResp, &resp)
}
