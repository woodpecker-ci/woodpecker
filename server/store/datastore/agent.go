// Copyright 2021 Woodpecker Authors
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

package datastore

import (
	"errors"

	"github.com/woodpecker-ci/woodpecker/server/model"
)

func (s storage) AgentList(p *model.ListOptions) ([]*model.Agent, error) {
	var agents []*model.Agent
	return agents, s.paginate(p).Find(&agents)
}

func (s storage) AgentFind(id int64) (*model.Agent, error) {
	agent := new(model.Agent)
	return agent, wrapGet(s.engine.ID(id).Get(agent))
}

func (s storage) AgentFindByToken(token string) (*model.Agent, error) {
	// Searching with an empty token would result in an empty where clause and therefore returning first item
	if token == "" {
		return nil, errors.New("Please provide a token")
	}
	agent := &model.Agent{
		Token: token,
	}
	return agent, wrapGet(s.engine.Get(agent))
}

func (s storage) AgentCreate(agent *model.Agent) error {
	// only Insert set auto created ID back to object
	_, err := s.engine.Insert(agent)
	return err
}

func (s storage) AgentUpdate(agent *model.Agent) error {
	_, err := s.engine.ID(agent.ID).AllCols().Update(agent)
	return err
}

func (s storage) AgentDelete(agent *model.Agent) error {
	return wrapDelete(s.engine.ID(agent.ID).Delete(new(model.Agent)))
}
