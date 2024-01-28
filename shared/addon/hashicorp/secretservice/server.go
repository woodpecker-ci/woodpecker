package secretservice

import (
	"encoding/json"

	"go.woodpecker-ci.org/woodpecker/v2/server/model"
)

type RPCServer struct {
	Impl model.SecretService
}

func (s *RPCServer) SecretListPipeline(args []byte, resp *[]byte) error {
	var a argumentsListPipeline
	err := json.Unmarshal(args, &a)
	if err != nil {
		return err
	}
	secrets, err := s.Impl.SecretListPipeline(a.Repo, a.Pipeline, a.ListOptions)
	if err != nil {
		return err
	}
	*resp, err = json.Marshal(secrets)
	return err
}

func (s *RPCServer) SecretFind(args []byte, resp *[]byte) error {
	var a argumentsFindDelete
	err := json.Unmarshal(args, &a)
	if err != nil {
		return err
	}
	secret, err := s.Impl.SecretFind(a.Repo, a.Name)
	if err != nil {
		return err
	}
	*resp, err = json.Marshal(secret)
	return err
}

func (s *RPCServer) SecretList(args []byte, resp *[]byte) error {
	var a argumentsList
	err := json.Unmarshal(args, &a)
	if err != nil {
		return err
	}
	secrets, err := s.Impl.SecretList(a.Repo, a.ListOptions)
	if err != nil {
		return err
	}
	*resp, err = json.Marshal(secrets)
	return err
}

func (s *RPCServer) SecretDelete(args []byte, resp *[]byte) error {
	var a argumentsFindDelete
	err := json.Unmarshal(args, &a)
	if err != nil {
		return err
	}
	*resp = []byte{}
	return s.Impl.SecretDelete(a.Repo, a.Name)
}

func (s *RPCServer) SecretCreate(args []byte, resp *[]byte) error {
	var a argumentsCreateUpdate
	err := json.Unmarshal(args, &a)
	if err != nil {
		return err
	}
	*resp = []byte{}
	return s.Impl.SecretCreate(a.Repo, a.Secret)
}

func (s *RPCServer) SecretUpdate(args []byte, resp *[]byte) error {
	var a argumentsCreateUpdate
	err := json.Unmarshal(args, &a)
	if err != nil {
		return err
	}
	*resp = []byte{}
	return s.Impl.SecretUpdate(a.Repo, a.Secret)
}

func (s *RPCServer) OrgSecretFind(args []byte, resp *[]byte) error {
	var a argumentsOrgFindDelete
	err := json.Unmarshal(args, &a)
	if err != nil {
		return err
	}
	secret, err := s.Impl.OrgSecretFind(a.OrgID, a.Name)
	if err != nil {
		return err
	}
	*resp, err = json.Marshal(secret)
	return err
}

func (s *RPCServer) OrgSecretList(args []byte, resp *[]byte) error {
	var a argumentsOrgList
	err := json.Unmarshal(args, &a)
	if err != nil {
		return err
	}
	secrets, err := s.Impl.OrgSecretList(a.OrgID, a.ListOptions)
	if err != nil {
		return err
	}
	*resp, err = json.Marshal(secrets)
	return err
}

func (s *RPCServer) OrgSecretDelete(args []byte, resp *[]byte) error {
	var a argumentsOrgFindDelete
	err := json.Unmarshal(args, &a)
	if err != nil {
		return err
	}
	*resp = []byte{}
	return s.Impl.OrgSecretDelete(a.OrgID, a.Name)
}

func (s *RPCServer) OrgSecretCreate(args []byte, resp *[]byte) error {
	var a argumentsOrgCreateUpdate
	err := json.Unmarshal(args, &a)
	if err != nil {
		return err
	}
	*resp = []byte{}
	return s.Impl.OrgSecretCreate(a.OrgID, a.Secret)
}

func (s *RPCServer) OrgSecretUpdate(args []byte, resp *[]byte) error {
	var a argumentsOrgCreateUpdate
	err := json.Unmarshal(args, &a)
	if err != nil {
		return err
	}
	*resp = []byte{}
	return s.Impl.OrgSecretUpdate(a.OrgID, a.Secret)
}

func (s *RPCServer) GlobalSecretFind(args string, resp *[]byte) error {
	secret, err := s.Impl.GlobalSecretFind(args)
	if err != nil {
		return err
	}
	*resp, err = json.Marshal(secret)
	return err
}

func (s *RPCServer) GlobalSecretList(args []byte, resp *[]byte) error {
	var opts *model.ListOptions
	err := json.Unmarshal(args, opts)
	if err != nil {
		return err
	}
	secrets, err := s.Impl.GlobalSecretList(opts)
	if err != nil {
		return err
	}
	*resp, err = json.Marshal(secrets)
	return err
}

func (s *RPCServer) GlobalSecretDelete(args string, resp *[]byte) error {
	*resp = []byte{}
	return s.Impl.GlobalSecretDelete(args)
}

func (s *RPCServer) GlobalSecretCreate(args []byte, resp *[]byte) error {
	var secret *model.Secret
	err := json.Unmarshal(args, secret)
	if err != nil {
		return err
	}
	*resp = []byte{}
	return s.Impl.GlobalSecretCreate(secret)
}

func (s *RPCServer) GlobalSecretUpdate(args []byte, resp *[]byte) error {
	var secret *model.Secret
	err := json.Unmarshal(args, secret)
	if err != nil {
		return err
	}
	*resp = []byte{}
	return s.Impl.GlobalSecretUpdate(secret)
}
