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

func NewAdmins(admins []string, primaryForgeOnly bool) *Admins {
	adminsLowercase := make([]string, len(admins))
	for i, a := range admins {
		adminsLowercase[i] = strings.ToLower(a)
	}
	return &Admins{
		admins:           utils.SliceToBoolMap(adminsLowercase),
		revoked:          make(map[string]bool),
		primaryForgeOnly: primaryForgeOnly,
	}
}

type Admins struct {
	mu sync.RWMutex
	// admins is the list read at server start, it never changes
	admins           map[string]bool
	revoked          map[string]bool
	primaryForgeOnly bool
}

// IsAdmin reports whether the user gets admin at sign in: listed and not revoked.
func (a *Admins) IsAdmin(user *model.User) bool {
	a.mu.RLock()
	defer a.mu.RUnlock()

	if a.primaryForgeOnly && user.ForgeID != model.PrimaryForgeID {
		return false
	}

	login := strings.ToLower(user.Login)
	return a.admins[login] && !a.revoked[login]
}

// IsRevoked reports whether the user is listed, but admin was revoked until the
// server restarts.
func (a *Admins) IsRevoked(user *model.User) bool {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return a.revoked[strings.ToLower(user.Login)]
}

// Revoke stops granting admin to a listed login until the server restarts, so a
// revoked admin does not get admin back at their next sign in.
func (a *Admins) Revoke(login string) {
	a.mu.Lock()
	defer a.mu.Unlock()
	login = strings.ToLower(login)
	if a.admins[login] {
		a.revoked[login] = true
	}
}
