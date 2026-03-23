// Copyright 2023 Woodpecker Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package permissions

import (
	"strings"

	"go.woodpecker-ci.org/woodpecker/v3/server/model"
	"go.woodpecker-ci.org/woodpecker/v3/shared/utils"
)

func NewOrgs(orgs []string) *Orgs {
	orgsLowercase := make([]string, len(orgs))
	for i, a := range orgs {
		orgsLowercase[i] = strings.ToLower(a)
	}
	return &Orgs{
		IsConfigured: len(orgs) > 0,
		orgs:         utils.SliceToBoolMap(orgsLowercase),
	}
}

type Orgs struct {
	IsConfigured bool
	orgs         map[string]bool
}

func (o *Orgs) IsMember(teams []*model.Team) bool {
	for _, team := range teams {
		if o.orgs[strings.ToLower(team.Login)] {
			return true
		}
	}
	return false
}
