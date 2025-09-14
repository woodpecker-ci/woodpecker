// Copyright 2025 Woodpecker Authors
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
	"encoding/json"
	"net/rpc"
	"os/exec"

	"github.com/hashicorp/go-plugin"
	"github.com/rs/zerolog/log"

	"go.woodpecker-ci.org/woodpecker/v3/server/model"
	logService "go.woodpecker-ci.org/woodpecker/v3/server/services/log"
	"go.woodpecker-ci.org/woodpecker/v3/shared/logger"
)

// make sure RPC implements logService.Service.
var _ logService.Service = new(RPC)

func Load(file string) (logService.Service, error) {
	client := plugin.NewClient(&plugin.ClientConfig{
		HandshakeConfig: HandshakeConfig,
		Plugins: map[string]plugin.Plugin{
			pluginKey: &Plugin{},
		},
		Cmd: exec.Command(file),
		Logger: &logger.AddonClientLogger{
			Logger: log.With().Str("addon", file).Logger(),
		},
	})
	// TODO: defer client.Kill()

	rpcClient, err := client.Client()
	if err != nil {
		return nil, err
	}

	raw, err := rpcClient.Dispense(pluginKey)
	if err != nil {
		return nil, err
	}

	extension, _ := raw.(logService.Service)
	return extension, nil
}

type RPC struct {
	client *rpc.Client
}

func (g *RPC) LogFind(step *model.Step) ([]*model.LogEntry, error) {
	args, err := json.Marshal(step)
	if err != nil {
		return nil, err
	}
	var jsonResp []byte
	err = g.client.Call("Plugin.LogFind", args, &jsonResp)
	if err != nil {
		return nil, err
	}

	var resp []*model.LogEntry
	err = json.Unmarshal(jsonResp, &resp)
	if err != nil {
		return nil, err
	}

	return resp, nil
}

func (g *RPC) LogAppend(step *model.Step, logEntries []*model.LogEntry) error {
	args, err := json.Marshal(&argumentsAppend{
		Step:       step,
		LogEntries: logEntries,
	})
	if err != nil {
		return err
	}
	var jsonResp []byte
	return g.client.Call("Plugin.LogAppend", args, &jsonResp)
}

func (g *RPC) LogDelete(step *model.Step) error {
	args, err := json.Marshal(step)
	if err != nil {
		return err
	}
	var jsonResp []byte
	return g.client.Call("Plugin.LogDelete", args, &jsonResp)
}
