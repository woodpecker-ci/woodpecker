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

func NewOwnersAllowlist(owners []string) *OwnersAllowlist {
	ownersLowercase := make([]string, len(owners))
	for i, a := range owners {
		ownersLowercase[i] = strings.ToLower(a)
	}
	return &OwnersAllowlist{owners: utils.SliceToBoolMap(ownersLowercase)}
}

type OwnersAllowlist struct {
	owners map[string]bool
}

func (o *OwnersAllowlist) IsAllowed(repo *model.Repo) bool {
	return len(o.owners) < 1 || o.owners[strings.ToLower(repo.Owner)]
}
