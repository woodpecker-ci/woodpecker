package environservice

import (
	"encoding/json"
	"net/rpc"

	"go.woodpecker-ci.org/woodpecker/v2/server/model"
)

type RPC struct{ client *rpc.Client }

func (g *RPC) EnvironList(repo *model.Repo) ([]*model.Environ, error) {
	args, err := json.Marshal(repo)
	if err != nil {
		return nil, err
	}
	var jsonResp []byte
	err = g.client.Call("Plugin.EnvironList", args, &jsonResp)
	if err != nil {
		return nil, err
	}

	var resp []*model.Environ
	return resp, json.Unmarshal(jsonResp, &resp)
}
