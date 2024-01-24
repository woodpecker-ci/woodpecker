package configservice

import (
	"encoding/json"
	"net/rpc"
	"time"

	forgetypes "go.woodpecker-ci.org/woodpecker/v2/server/forge/types"
	"go.woodpecker-ci.org/woodpecker/v2/server/model"
)

type ExtensionRPC struct{ client *rpc.Client }

func (g *ExtensionRPC) FetchConfig(repo *model.Repo, pipeline *model.Pipeline, currentFileMeta []*forgetypes.FileMeta, netrc *model.Netrc, timeout time.Duration) ([]*forgetypes.FileMeta, bool, error) {
	args, err := json.Marshal(&arguments{
		Repo:            repo,
		Pipeline:        pipeline,
		CurrentFileMeta: currentFileMeta,
		Netrc:           netrc,
		Timeout:         timeout,
	})
	if err != nil {
		return nil, false, err
	}
	var jsonResp []byte
	err = g.client.Call("Plugin.FetchConfig", args, &jsonResp)
	if err != nil {
		return nil, false, err
	}

	var resp response
	err = json.Unmarshal(jsonResp, &resp)
	if err != nil {
		return nil, false, err
	}
	return resp.ConfigData, resp.UseOld, err
}
