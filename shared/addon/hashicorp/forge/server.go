package configservice

import (
	"context"
	"encoding/json"

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
