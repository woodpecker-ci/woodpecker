package secretservice

import (
	"encoding/json"
	"net/rpc"

	"go.woodpecker-ci.org/woodpecker/v2/server/model"
)

type RPC struct{ client *rpc.Client }

func (g *RPC) SecretListPipeline(repo *model.Repo, pipeline *model.Pipeline, listOptions *model.ListOptions) ([]*model.Secret, error) {
	args, err := json.Marshal(&argumentsListPipeline{
		Repo:        repo,
		Pipeline:    pipeline,
		ListOptions: listOptions,
	})
	if err != nil {
		return nil, err
	}
	var jsonResp []byte
	err = g.client.Call("Plugin.SecretListPipeline", args, &jsonResp)
	if err != nil {
		return nil, err
	}

	var resp []*model.Secret
	return resp, json.Unmarshal(jsonResp, &resp)
}

func (g *RPC) SecretFind(repo *model.Repo, name string) (*model.Secret, error) {
	args, err := json.Marshal(&argumentsFindDelete{
		Repo: repo,
		Name: name,
	})
	if err != nil {
		return nil, err
	}
	var jsonResp []byte
	err = g.client.Call("Plugin.SecretFind", args, &jsonResp)
	if err != nil {
		return nil, err
	}

	var resp *model.Secret
	return resp, json.Unmarshal(jsonResp, &resp)
}

func (g *RPC) SecretList(repo *model.Repo, listOptions *model.ListOptions) ([]*model.Secret, error) {
	args, err := json.Marshal(&argumentsList{
		Repo:        repo,
		ListOptions: listOptions,
	})
	if err != nil {
		return nil, err
	}
	var jsonResp []byte
	err = g.client.Call("Plugin.SecretList", args, &jsonResp)
	if err != nil {
		return nil, err
	}

	var resp []*model.Secret
	return resp, json.Unmarshal(jsonResp, &resp)
}

func (g *RPC) SecretCreate(repo *model.Repo, secret *model.Secret) error {
	args, err := json.Marshal(&argumentsCreateUpdate{
		Repo:   repo,
		Secret: secret,
	})
	if err != nil {
		return err
	}
	var jsonResp []byte
	err = g.client.Call("Plugin.SecretCreate", args, &jsonResp)
	if err != nil {
		return err
	}

	var resp []byte
	return json.Unmarshal(jsonResp, &resp)
}

func (g *RPC) SecretUpdate(repo *model.Repo, secret *model.Secret) error {
	args, err := json.Marshal(&argumentsCreateUpdate{
		Repo:   repo,
		Secret: secret,
	})
	if err != nil {
		return err
	}
	var jsonResp []byte
	err = g.client.Call("Plugin.SecretUpdate", args, &jsonResp)
	if err != nil {
		return err
	}

	var resp []byte
	return json.Unmarshal(jsonResp, &resp)
}

func (g *RPC) SecretDelete(repo *model.Repo, name string) error {
	args, err := json.Marshal(&argumentsFindDelete{
		Repo: repo,
		Name: name,
	})
	if err != nil {
		return err
	}
	var jsonResp []byte
	err = g.client.Call("Plugin.SecretDelete", args, &jsonResp)
	if err != nil {
		return err
	}

	var resp []byte
	return json.Unmarshal(jsonResp, &resp)
}

func (g *RPC) OrgSecretFind(orgID int64, name string) (*model.Secret, error) {
	args, err := json.Marshal(&argumentsOrgFindDelete{
		OrgID: orgID,
		Name:  name,
	})
	if err != nil {
		return nil, err
	}
	var jsonResp []byte
	err = g.client.Call("Plugin.OrgSecretFind", args, &jsonResp)
	if err != nil {
		return nil, err
	}

	var resp *model.Secret
	return resp, json.Unmarshal(jsonResp, &resp)
}

func (g *RPC) OrgSecretList(orgID int64, listOptions *model.ListOptions) ([]*model.Secret, error) {
	args, err := json.Marshal(&argumentsOrgList{
		OrgID:       orgID,
		ListOptions: listOptions,
	})
	if err != nil {
		return nil, err
	}
	var jsonResp []byte
	err = g.client.Call("Plugin.OrgSecretList", args, &jsonResp)
	if err != nil {
		return nil, err
	}

	var resp []*model.Secret
	return resp, json.Unmarshal(jsonResp, &resp)
}

func (g *RPC) OrgSecretCreate(orgID int64, secret *model.Secret) error {
	args, err := json.Marshal(&argumentsOrgCreateUpdate{
		OrgID:  orgID,
		Secret: secret,
	})
	if err != nil {
		return err
	}
	var jsonResp []byte
	err = g.client.Call("Plugin.OrgSecretCreate", args, &jsonResp)
	if err != nil {
		return err
	}

	var resp []byte
	return json.Unmarshal(jsonResp, &resp)
}

func (g *RPC) OrgSecretUpdate(orgID int64, secret *model.Secret) error {
	args, err := json.Marshal(&argumentsOrgCreateUpdate{
		OrgID:  orgID,
		Secret: secret,
	})
	if err != nil {
		return err
	}
	var jsonResp []byte
	err = g.client.Call("Plugin.OrgSecretUpdate", args, &jsonResp)
	if err != nil {
		return err
	}

	var resp []byte
	return json.Unmarshal(jsonResp, &resp)
}

func (g *RPC) OrgSecretDelete(orgID int64, name string) error {
	args, err := json.Marshal(&argumentsOrgFindDelete{
		OrgID: orgID,
		Name:  name,
	})
	if err != nil {
		return err
	}
	var jsonResp []byte
	err = g.client.Call("Plugin.OrgSecretDelete", args, &jsonResp)
	if err != nil {
		return err
	}

	var resp []byte
	return json.Unmarshal(jsonResp, &resp)
}

func (g *RPC) GlobalSecretFind(name string) (*model.Secret, error) {
	var jsonResp []byte
	err := g.client.Call("Plugin.GlobalSecretFind", name, &jsonResp)
	if err != nil {
		return nil, err
	}

	var resp *model.Secret
	return resp, json.Unmarshal(jsonResp, &resp)
}

func (g *RPC) GlobalSecretList(options *model.ListOptions) ([]*model.Secret, error) {
	args, err := json.Marshal(options)
	if err != nil {
		return nil, err
	}
	var jsonResp []byte
	err = g.client.Call("Plugin.GlobalSecretList", args, &jsonResp)
	if err != nil {
		return nil, err
	}

	var resp []*model.Secret
	return resp, json.Unmarshal(jsonResp, &resp)
}

func (g *RPC) GlobalSecretCreate(secret *model.Secret) error {
	args, err := json.Marshal(secret)
	if err != nil {
		return err
	}
	var jsonResp []byte
	err = g.client.Call("Plugin.GlobalSecretCreate", args, &jsonResp)
	if err != nil {
		return err
	}

	var resp []byte
	return json.Unmarshal(jsonResp, &resp)
}

func (g *RPC) GlobalSecretUpdate(secret *model.Secret) error {
	args, err := json.Marshal(secret)
	if err != nil {
		return err
	}
	var jsonResp []byte
	err = g.client.Call("Plugin.GlobalSecretUpdate", args, &jsonResp)
	if err != nil {
		return err
	}

	var resp []byte
	return json.Unmarshal(jsonResp, &resp)
}

func (g *RPC) GlobalSecretDelete(name string) error {
	var jsonResp []byte
	err := g.client.Call("Plugin.GlobalSecretDelete", name, &jsonResp)
	if err != nil {
		return err
	}

	var resp []byte
	return json.Unmarshal(jsonResp, &resp)
}
