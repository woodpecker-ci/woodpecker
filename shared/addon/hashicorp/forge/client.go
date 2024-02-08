package configservice

import (
	"context"
	"encoding/json"
	"net/http"
	"net/rpc"

	"go.woodpecker-ci.org/woodpecker/v2/server/forge"
	"go.woodpecker-ci.org/woodpecker/v2/server/forge/types"
	"go.woodpecker-ci.org/woodpecker/v2/server/model"
)

var f forge.Forge = new(RPC)

// TODO issue: user models are not sent with token/secret (token/secret is json:"-")
// possible solution: two-way-communication with two funcs: 1. token/secret for user 2. token/secret for repo
// however, that's an issue in both directions: the addon can't return tokens/secrets
type RPC struct{ client *rpc.Client }

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

func (g *RPC) Login(ctx context.Context, w http.ResponseWriter, r *http.Request) (*model.User, error) {
	/* TODO args, err := json.Marshal(&arguments{
		Repo:            repo,
		Pipeline:        pipeline,
		CurrentFileMeta: currentFileMeta,
		Netrc:           netrc,
		Timeout:         timeout,
	})
	if err != nil {
		return nil, err
	}*/
	var jsonResp []byte
	/*err = g.client.Call("Plugin.Login", args, &jsonResp)
	if err != nil {
		return nil, err
	}*/

	var resp *model.User
	return resp, json.Unmarshal(jsonResp, resp)
}

func (g *RPC) Auth(ctx context.Context, token, secret string) (string, error) {
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

func (g *RPC) Teams(ctx context.Context, u *model.User) ([]*model.Team, error) {
	args, err := json.Marshal(u)
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

func (g *RPC) Repo(ctx context.Context, u *model.User, remoteID model.ForgeRemoteID, owner, name string) (*model.Repo, error) {
	args, err := json.Marshal(&argumentsRepo{
		U:        u,
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

	var resp *model.Repo
	return resp, json.Unmarshal(jsonResp, resp)
}

func (g *RPC) Repos(ctx context.Context, u *model.User) ([]*model.Repo, error) {
	args, err := json.Marshal(u)
	if err != nil {
		return nil, err
	}
	var jsonResp []byte
	err = g.client.Call("Plugin.Repos", args, &jsonResp)
	if err != nil {
		return nil, err
	}

	var resp []*model.Repo
	return resp, json.Unmarshal(jsonResp, &resp)
}

func (g *RPC) File(ctx context.Context, u *model.User, r *model.Repo, b *model.Pipeline, f string) ([]byte, error) {
	args, err := json.Marshal(&argumentsFileDir{
		U: u,
		R: r,
		B: b,
		F: f,
	})
	if err != nil {
		return nil, err
	}
	var resp []byte
	return resp, g.client.Call("Plugin.File", args, &resp)
}

func (g *RPC) Dir(ctx context.Context, u *model.User, r *model.Repo, b *model.Pipeline, f string) ([]*types.FileMeta, error) {
	args, err := json.Marshal(&argumentsFileDir{
		U: u,
		R: r,
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

func (g *RPC) Status(ctx context.Context, u *model.User, r *model.Repo, b *model.Pipeline, p *model.Workflow) error {
	args, err := json.Marshal(&argumentsStatus{
		U: u,
		R: r,
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
		U: u,
		R: r,
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

func (g *RPC) Activate(ctx context.Context, u *model.User, r *model.Repo, link string) error {
	//TODO implement me
	panic("implement me")
}

func (g *RPC) Deactivate(ctx context.Context, u *model.User, r *model.Repo, link string) error {
	//TODO implement me
	panic("implement me")
}

func (g *RPC) Branches(ctx context.Context, u *model.User, r *model.Repo, p *model.ListOptions) ([]string, error) {
	//TODO implement me
	panic("implement me")
}

func (g *RPC) BranchHead(ctx context.Context, u *model.User, r *model.Repo, branch string) (string, error) {
	//TODO implement me
	panic("implement me")
}

func (g *RPC) PullRequests(ctx context.Context, u *model.User, r *model.Repo, p *model.ListOptions) ([]*model.PullRequest, error) {
	//TODO implement me
	panic("implement me")
}

func (g *RPC) Hook(ctx context.Context, r *http.Request) (repo *model.Repo, pipeline *model.Pipeline, err error) {
	//TODO implement me
	panic("implement me")
}

func (g *RPC) OrgMembership(ctx context.Context, u *model.User, org string) (*model.OrgPerm, error) {
	//TODO implement me
	panic("implement me")
}

func (g *RPC) Org(ctx context.Context, u *model.User, org string) (*model.Org, error) {
	//TODO implement me
	panic("implement me")
}
