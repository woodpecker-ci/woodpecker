package registryservice

import (
	"encoding/json"

	"go.woodpecker-ci.org/woodpecker/v2/server/model"
)

type RPCServer struct {
	Impl model.RegistryService
}

func (s *RPCServer) RegistryFind(args []byte, resp *[]byte) error {
	var a argumentsFindDelete
	err := json.Unmarshal(args, &a)
	if err != nil {
		return err
	}
	registry, err := s.Impl.RegistryFind(a.Repo, a.Name)
	if err != nil {
		return err
	}
	*resp, err = json.Marshal(registry)
	return err
}

func (s *RPCServer) RegistryDelete(args []byte, resp *[]byte) error {
	var a argumentsFindDelete
	err := json.Unmarshal(args, &a)
	if err != nil {
		return err
	}
	*resp = []byte{}
	return s.Impl.RegistryDelete(a.Repo, a.Name)
}

func (s *RPCServer) RegistryCreate(args []byte, resp *[]byte) error {
	var a argumentsCreateUpdate
	err := json.Unmarshal(args, &a)
	if err != nil {
		return err
	}
	*resp = []byte{}
	return s.Impl.RegistryCreate(a.Repo, a.Registry)
}

func (s *RPCServer) RegistryUpdate(args []byte, resp *[]byte) error {
	var a argumentsCreateUpdate
	err := json.Unmarshal(args, &a)
	if err != nil {
		return err
	}
	*resp = []byte{}
	return s.Impl.RegistryUpdate(a.Repo, a.Registry)
}

func (s *RPCServer) RegistryList(args []byte, resp *[]byte) error {
	var a argumentsList
	err := json.Unmarshal(args, &a)
	if err != nil {
		return err
	}
	registries, err := s.Impl.RegistryList(a.Repo, a.ListOptions)
	if err != nil {
		return err
	}
	*resp, err = json.Marshal(registries)
	return err
}
