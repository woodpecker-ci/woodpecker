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
	"sync"

	"go.woodpecker-ci.org/woodpecker/v3/server/model"
	"go.woodpecker-ci.org/woodpecker/v3/shared/utils"
)

func NewAdmins(admins []string) *Admins {
	adminsLowercase := make([]string, len(admins))
	for i, a := range admins {
		adminsLowercase[i] = strings.ToLower(a)
	}
	return &Admins{admins: utils.SliceToBoolMap(adminsLowercase)}
}

type Admins struct {
	mu     sync.RWMutex
	admins map[string]bool
}

func (a *Admins) IsAdmin(user *model.User) bool {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return a.admins[strings.ToLower(user.Login)]
}

// Remove drops a login from the list until the server restarts, so a revoked
// admin does not get admin back at their next sign in.
func (a *Admins) Remove(login string) {
	a.mu.Lock()
	defer a.mu.Unlock()
	delete(a.admins, strings.ToLower(login))
}
