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

	"github.com/hashicorp/go-plugin"

	"go.woodpecker-ci.org/woodpecker/v3/server/model"
	"go.woodpecker-ci.org/woodpecker/v3/server/services/log"
)

func Serve(impl log.Service) {
	plugin.Serve(&plugin.ServeConfig{
		HandshakeConfig: HandshakeConfig,
		Plugins: map[string]plugin.Plugin{
			pluginKey: &Plugin{Impl: impl},
		},
	})
}

type RPCServer struct {
	Impl log.Service
}

type argumentsAppend struct {
	Step       *model.Step       `json:"step"`
	LogEntries []*model.LogEntry `json:"log_entries"`
}

func (s *RPCServer) LogFind(args []byte, resp *[]byte) error {
	var a model.Step
	err := json.Unmarshal(args, &a)
	if err != nil {
		return err
	}
	log, err := s.Impl.LogFind(&a)
	if err != nil {
		return err
	}
	*resp, err = json.Marshal(log)
	return err
}

func (s *RPCServer) LogAppend(args []byte, resp *[]byte) error {
	var a argumentsAppend
	err := json.Unmarshal(args, &a)
	if err != nil {
		return err
	}
	*resp = []byte{}
	return s.Impl.LogAppend(a.Step, a.LogEntries)
}

func (s *RPCServer) LogDelete(args []byte, resp *[]byte) error {
	var a model.Step
	err := json.Unmarshal(args, &a)
	if err != nil {
		return err
	}
	*resp = []byte{}
	return s.Impl.LogDelete(&a)
}

func (s *RPCServer) StepFinished(args []byte, resp *[]byte) error {
	var a model.Step
	err := json.Unmarshal(args, &a)
	if err != nil {
		return err
	}
	*resp = []byte{}
	s.Impl.StepFinished(&a)
	return nil
}
