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
	"context"
	"encoding/json"
	"net/rpc"
	"os/exec"

	"github.com/hashicorp/go-plugin"
	"github.com/rs/zerolog/log"

	"go.woodpecker-ci.org/woodpecker/v3/server/model"
	service_log "go.woodpecker-ci.org/woodpecker/v3/server/services/log"
	"go.woodpecker-ci.org/woodpecker/v3/shared/logger"
)

// make sure RPC implements service_log.Service.
var _ service_log.Service = new(RPC)

func Load(file string) (service_log.Service, error) {
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

	extension, _ := raw.(service_log.Service)
	return extension, nil
}

type RPC struct {
	client *rpc.Client
}

func (g *RPC) LogFind(ctx context.Context, step *model.Step) ([]*model.LogEntry, error) {
	args, err := json.Marshal(step)
	if err != nil {
		return nil, err
	}
	var jsonResp []byte
	err = g.call(ctx, "Plugin.LogFind", args, &jsonResp)
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

func (g *RPC) LogAppend(ctx context.Context, step *model.Step, logEntries []*model.LogEntry) error {
	args, err := json.Marshal(&argumentsAppend{
		Step:       step,
		LogEntries: logEntries,
	})
	if err != nil {
		return err
	}
	var jsonResp []byte
	return g.call(ctx, "Plugin.LogAppend", args, &jsonResp)
}

func (g *RPC) LogDelete(ctx context.Context, step *model.Step) error {
	args, err := json.Marshal(step)
	if err != nil {
		return err
	}
	var jsonResp []byte
	return g.call(ctx, "Plugin.LogDelete", args, &jsonResp)
}

func (g *RPC) StepFinished(ctx context.Context, step *model.Step) error {
	args, err := json.Marshal(step)
	if err != nil {
		return err
	}
	var jsonResp []byte
	return g.call(ctx, "Plugin.StepFinished", args, &jsonResp)
}

func (g *RPC) call(ctx context.Context, method string, args []byte, resp *[]byte) error {
	c := g.client.Go(method, args, resp, make(chan *rpc.Call, 1))
	select {
	case <-c.Done:
		return c.Error
	case <-ctx.Done():
		return context.Cause(ctx)
	}
}
