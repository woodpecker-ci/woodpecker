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
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"

	"go.woodpecker-ci.org/woodpecker/v3/server/model"
)

func TestAdmins(t *testing.T) {
	a := NewAdmins([]string{"woodpecker-ci"})
	assert.True(t, a.IsAdmin(&model.User{Login: "woodpecker-ci"}))
	assert.False(t, a.IsAdmin(&model.User{Login: "not-woodpecker-ci"}))
	empty := NewAdmins([]string{})
	assert.False(t, empty.IsAdmin(&model.User{Login: "woodpecker-ci"}))
	assert.False(t, empty.IsAdmin(&model.User{Login: "not-woodpecker-ci"}))
}

func TestAdminsRevoke(t *testing.T) {
	a := NewAdmins([]string{"Alice", "bob"})

	a.Revoke("ALICE")
	alice := &model.User{Login: "alice"}
	assert.False(t, a.IsAdmin(alice), "revoke must ignore case like NewAdmins")
	assert.True(t, a.IsRevoked(alice), "the UI must see the revoke")
	bob := &model.User{Login: "bob"}
	assert.True(t, a.IsAdmin(bob), "other admins must stay")
	assert.False(t, a.IsRevoked(bob))

	a.Revoke("unknown")
	unknown := &model.User{Login: "unknown"}
	assert.False(t, a.IsRevoked(unknown), "only listed logins can be revoked")
	assert.False(t, a.IsAdmin(unknown), "revoking must not add a login to the list")
}

// A revoke in the admin settings can run while other users sign in.
func TestAdminsRevokeConcurrentWithIsAdmin(t *testing.T) {
	a := NewAdmins([]string{"alice", "bob"})
	user := &model.User{Login: "bob"}

	var wg sync.WaitGroup
	for range 10 {
		wg.Go(func() { a.IsAdmin(user) })
		wg.Go(func() { a.IsRevoked(user) })
	}
	wg.Go(func() { a.Revoke("alice") })
	wg.Wait()

	assert.False(t, a.IsAdmin(&model.User{Login: "alice"}))
}
