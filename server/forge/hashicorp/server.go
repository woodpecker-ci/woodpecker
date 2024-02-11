package hashicorp

import (
	"context"
	"encoding/json"
	"net/http"

	"go.woodpecker-ci.org/woodpecker/v2/server/forge"
	"go.woodpecker-ci.org/woodpecker/v2/server/model"
)

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
	var a *model.User
	err := json.Unmarshal(args, a)
	if err != nil {
		return err
	}
	teams, err := s.Impl.Teams(mkCtx(), a)
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
	repos, err := s.Impl.Repo(mkCtx(), a.U, a.RemoteID, a.Owner, a.Name)
	if err != nil {
		return err
	}
	*resp, err = json.Marshal(repos)
	return err
}

func (s *RPCServer) Repos(args []byte, resp *[]byte) error {
	var a *model.User
	err := json.Unmarshal(args, a)
	if err != nil {
		return err
	}
	repos, err := s.Impl.Repos(mkCtx(), a)
	if err != nil {
		return err
	}
	*resp, err = json.Marshal(repos)
	return err
}

func (s *RPCServer) File(args []byte, resp *[]byte) error {
	var a argumentsFileDir
	err := json.Unmarshal(args, &a)
	if err != nil {
		return err
	}
	*resp, err = s.Impl.File(mkCtx(), a.U, a.R, a.B, a.F)
	return err
}

func (s *RPCServer) Dir(args []byte, resp *[]byte) error {
	var a argumentsFileDir
	err := json.Unmarshal(args, &a)
	if err != nil {
		return err
	}
	meta, err := s.Impl.Dir(mkCtx(), a.U, a.R, a.B, a.F)
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
	return s.Impl.Status(mkCtx(), a.U, a.R, a.B, a.P)
}

func (s *RPCServer) Netrc(args []byte, resp *[]byte) error {
	var a argumentsNetrc
	err := json.Unmarshal(args, &a)
	if err != nil {
		return err
	}
	netrc, err := s.Impl.Netrc(a.U, a.R)
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
	return s.Impl.Activate(mkCtx(), a.U, a.R, a.Link)
}

func (s *RPCServer) Deactivate(args []byte, resp *[]byte) error {
	var a argumentsActivateDeactivate
	err := json.Unmarshal(args, &a)
	if err != nil {
		return err
	}
	*resp = []byte{}
	return s.Impl.Deactivate(mkCtx(), a.U, a.R, a.Link)
}

func (s *RPCServer) Branches(args []byte, resp *[]byte) error {
	var a argumentsBranchesPullRequests
	err := json.Unmarshal(args, &a)
	if err != nil {
		return err
	}
	branches, err := s.Impl.Branches(mkCtx(), a.U, a.R, a.P)
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
	commit, err := s.Impl.BranchHead(mkCtx(), a.U, a.R, a.Branch)
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
	prs, err := s.Impl.PullRequests(mkCtx(), a.U, a.R, a.P)
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
	org, err := s.Impl.OrgMembership(mkCtx(), a.U, a.Org)
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
	org, err := s.Impl.Org(mkCtx(), a.U, a.Org)
	if err != nil {
		return err
	}
	*resp, err = json.Marshal(org)
	return err
}

func (s *RPCServer) Hook(args []byte, resp *[]byte) error {
	// TODO http.request json
	var a *http.Request
	err := json.Unmarshal(args, &a)
	if err != nil {
		return err
	}
	repo, pipeline, err := s.Impl.Hook(mkCtx(), a)
	if err != nil {
		return err
	}
	*resp, err = json.Marshal(&responseHook{
		Repo:     repo,
		Pipeline: pipeline,
	})
	return err
}

func (s *RPCServer) Login(args []byte, resp *[]byte) error {
	// TODO http.request and iowriter json
	var a *http.Request
	err := json.Unmarshal(args, &a)
	if err != nil {
		return err
	}
	user, err := s.Impl.Login(mkCtx(), nil, a)
	if err != nil {
		return err
	}
	*resp, err = json.Marshal(user)
	return err
}
