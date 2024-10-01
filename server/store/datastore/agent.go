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

	"go.woodpecker-ci.org/woodpecker/v2/server/model"
)

var ErrNoTokenProvided = errors.New("please provide a token")

func (s storage) AgentList(p *model.ListOptions) (agents []*model.Agent, _ error) {
	return agents, s.paginate(p).OrderBy("id").Find(&agents)
}

func (s storage) AgentFind(id int64) (*model.Agent, error) {
	agent := new(model.Agent)
	return agent, wrapGet(s.engine.ID(id).Get(agent))
}

func (s storage) AgentFindByToken(token string) (*model.Agent, error) {
	// Searching with an empty token would result in an empty where clause and therefore returning first item
	if token == "" {
		return nil, ErrNoTokenProvided
	}
	agent := new(model.Agent)
	return agent, wrapGet(s.engine.Where("token = ?", token).Get(agent))
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

func (s storage) AgentListForOrg(orgID int64, p *model.ListOptions) (agents []*model.Agent, _ error) {
	return agents, s.paginate(p).Where("org_id = ?", orgID).OrderBy("id").Find(&agents)
}
