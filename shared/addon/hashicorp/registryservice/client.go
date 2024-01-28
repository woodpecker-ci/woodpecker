package registryservice

import (
	"encoding/json"
	"net/rpc"

	"go.woodpecker-ci.org/woodpecker/v2/server/model"
)

type RPC struct{ client *rpc.Client }

func (g *RPC) RegistryFind(repo *model.Repo, name string) (*model.Registry, error) {
	args, err := json.Marshal(&argumentsFindDelete{
		Repo: repo,
		Name: name,
	})
	if err != nil {
		return nil, err
	}
	var jsonResp []byte
	err = g.client.Call("Plugin.RegistryFind", args, &jsonResp)
	if err != nil {
		return nil, err
	}

	var resp *model.Registry
	return resp, json.Unmarshal(jsonResp, resp)
}

func (g *RPC) RegistryList(repo *model.Repo, listOptions *model.ListOptions) ([]*model.Registry, error) {
	args, err := json.Marshal(&argumentsList{
		Repo:        repo,
		ListOptions: listOptions,
	})
	if err != nil {
		return nil, err
	}
	var jsonResp []byte
	err = g.client.Call("Plugin.RegistryList", args, &jsonResp)
	if err != nil {
		return nil, err
	}

	var resp []*model.Registry
	return resp, json.Unmarshal(jsonResp, &resp)
}

func (g *RPC) RegistryDelete(repo *model.Repo, name string) error {
	args, err := json.Marshal(&argumentsFindDelete{
		Repo: repo,
		Name: name,
	})
	if err != nil {
		return err
	}
	var jsonResp []byte
	return g.client.Call("Plugin.RegistryDelete", args, &jsonResp)
}

func (g *RPC) RegistryCreate(repo *model.Repo, registry *model.Registry) error {
	args, err := json.Marshal(&argumentsCreateUpdate{
		Repo:     repo,
		Registry: registry,
	})
	if err != nil {
		return err
	}
	var jsonResp []byte
	return g.client.Call("Plugin.RegistryCreate", args, &jsonResp)
}

func (g *RPC) RegistryUpdate(repo *model.Repo, registry *model.Registry) error {
	args, err := json.Marshal(&argumentsCreateUpdate{
		Repo:     repo,
		Registry: registry,
	})
	if err != nil {
		return err
	}
	var jsonResp []byte
	return g.client.Call("Plugin.RegistryUpdate", args, &jsonResp)
}
