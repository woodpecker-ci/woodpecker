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
	"bytes"
	"context"
	"encoding/json"
	"net/http"

	"github.com/hashicorp/go-plugin"

	"go.woodpecker-ci.org/woodpecker/v2/server/forge"
	"go.woodpecker-ci.org/woodpecker/v2/server/forge/types"
)

func Serve(impl forge.Forge) {
	plugin.Serve(&plugin.ServeConfig{
		HandshakeConfig: HandshakeConfig,
		Plugins: map[string]plugin.Plugin{
			pluginKey: &Plugin{Impl: impl},
		},
	})
}

func mkCtx() context.Context {
	return context.Background()
}

type RPCServer struct {
	Impl forge.Forge
}

func (s *RPCServer) Name(_ []byte, resp *string) error {
	*resp = s.Impl.Name()
	return nil
}

func (s *RPCServer) URL(_ []byte, resp *string) error {
	*resp = s.Impl.URL()
	return nil
}

func (s *RPCServer) Teams(args []byte, resp *[]byte) error {
	var a *modelUser
	err := json.Unmarshal(args, a)
	if err != nil {
		return err
	}
	teams, err := s.Impl.Teams(mkCtx(), a.asModel())
	if err != nil {
		return err
	}
	*resp, err = json.Marshal(teams)
	return err
}

func (s *RPCServer) Repo(args []byte, resp *[]byte) error {
	var a argumentsRepo
	err := json.Unmarshal(args, &a)
	if err != nil {
		return err
	}
	repos, err := s.Impl.Repo(mkCtx(), a.U.asModel(), a.RemoteID, a.Owner, a.Name)
	if err != nil {
		return err
	}
	*resp, err = json.Marshal(modelRepoFromModel(repos))
	return err
}

func (s *RPCServer) Repos(args []byte, resp *[]byte) error {
	var a *modelUser
	err := json.Unmarshal(args, a)
	if err != nil {
		return err
	}
	repos, err := s.Impl.Repos(mkCtx(), a.asModel())
	if err != nil {
		return err
	}
	var modelRepos []*modelRepo
	for _, repo := range repos {
		modelRepos = append(modelRepos, modelRepoFromModel(repo))
	}
	*resp, err = json.Marshal(modelRepos)
	return err
}

func (s *RPCServer) File(args []byte, resp *[]byte) error {
	var a argumentsFileDir
	err := json.Unmarshal(args, &a)
	if err != nil {
		return err
	}
	*resp, err = s.Impl.File(mkCtx(), a.U.asModel(), a.R.asModel(), a.B, a.F)
	return err
}

func (s *RPCServer) Dir(args []byte, resp *[]byte) error {
	var a argumentsFileDir
	err := json.Unmarshal(args, &a)
	if err != nil {
		return err
	}
	meta, err := s.Impl.Dir(mkCtx(), a.U.asModel(), a.R.asModel(), a.B, a.F)
	if err != nil {
		return err
	}
	*resp, err = json.Marshal(meta)
	return err
}

func (s *RPCServer) Status(args []byte, resp *[]byte) error {
	var a argumentsStatus
	err := json.Unmarshal(args, &a)
	if err != nil {
		return err
	}
	*resp = []byte{}
	return s.Impl.Status(mkCtx(), a.U.asModel(), a.R.asModel(), a.B, a.P)
}

func (s *RPCServer) Netrc(args []byte, resp *[]byte) error {
	var a argumentsNetrc
	err := json.Unmarshal(args, &a)
	if err != nil {
		return err
	}
	netrc, err := s.Impl.Netrc(a.U.asModel(), a.R.asModel())
	if err != nil {
		return err
	}
	*resp, err = json.Marshal(netrc)
	return err
}

func (s *RPCServer) Activate(args []byte, resp *[]byte) error {
	var a argumentsActivateDeactivate
	err := json.Unmarshal(args, &a)
	if err != nil {
		return err
	}
	*resp = []byte{}
	return s.Impl.Activate(mkCtx(), a.U.asModel(), a.R.asModel(), a.Link)
}

func (s *RPCServer) Deactivate(args []byte, resp *[]byte) error {
	var a argumentsActivateDeactivate
	err := json.Unmarshal(args, &a)
	if err != nil {
		return err
	}
	*resp = []byte{}
	return s.Impl.Deactivate(mkCtx(), a.U.asModel(), a.R.asModel(), a.Link)
}

func (s *RPCServer) Branches(args []byte, resp *[]byte) error {
	var a argumentsBranchesPullRequests
	err := json.Unmarshal(args, &a)
	if err != nil {
		return err
	}
	branches, err := s.Impl.Branches(mkCtx(), a.U.asModel(), a.R.asModel(), a.P)
	if err != nil {
		return err
	}
	*resp, err = json.Marshal(branches)
	return err
}

func (s *RPCServer) BranchHead(args []byte, resp *[]byte) error {
	var a argumentsBranchHead
	err := json.Unmarshal(args, &a)
	if err != nil {
		return err
	}
	commit, err := s.Impl.BranchHead(mkCtx(), a.U.asModel(), a.R.asModel(), a.Branch)
	if err != nil {
		return err
	}
	*resp, err = json.Marshal(commit)
	return err
}

func (s *RPCServer) PullRequests(args []byte, resp *[]byte) error {
	var a argumentsBranchesPullRequests
	err := json.Unmarshal(args, &a)
	if err != nil {
		return err
	}
	prs, err := s.Impl.PullRequests(mkCtx(), a.U.asModel(), a.R.asModel(), a.P)
	if err != nil {
		return err
	}
	*resp, err = json.Marshal(prs)
	return err
}

func (s *RPCServer) OrgMembership(args []byte, resp *[]byte) error {
	var a argumentsOrgMembershipOrg
	err := json.Unmarshal(args, &a)
	if err != nil {
		return err
	}
	org, err := s.Impl.OrgMembership(mkCtx(), a.U.asModel(), a.Org)
	if err != nil {
		return err
	}
	*resp, err = json.Marshal(org)
	return err
}

func (s *RPCServer) Org(args []byte, resp *[]byte) error {
	var a argumentsOrgMembershipOrg
	err := json.Unmarshal(args, &a)
	if err != nil {
		return err
	}
	org, err := s.Impl.Org(mkCtx(), a.U.asModel(), a.Org)
	if err != nil {
		return err
	}
	*resp, err = json.Marshal(org)
	return err
}

func (s *RPCServer) Hook(args []byte, resp *[]byte) error {
	var a httpRequest
	err := json.Unmarshal(args, &a)
	if err != nil {
		return err
	}
	req, err := http.NewRequest(a.Method, a.URL, bytes.NewBuffer(a.Body))
	if err != nil {
		return err
	}
	req.Header = a.Header
	req.Form = a.Form
	repo, pipeline, err := s.Impl.Hook(mkCtx(), req)
	if err != nil {
		return err
	}
	*resp, err = json.Marshal(&responseHook{
		Repo:     modelRepoFromModel(repo),
		Pipeline: pipeline,
	})
	return err
}

func (s *RPCServer) Login(args []byte, resp *[]byte) error {
	var a *types.OAuthRequest
	err := json.Unmarshal(args, a)
	if err != nil {
		return err
	}
	user, red, err := s.Impl.Login(mkCtx(), a)
	if err != nil {
		return err
	}
	*resp, err = json.Marshal(&responseLogin{
		User:        modelUserFromModel(user),
		RedirectURL: red,
	})
	return err
}
