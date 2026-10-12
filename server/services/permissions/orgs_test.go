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
	"testing"

	"github.com/stretchr/testify/assert"

	"go.woodpecker-ci.org/woodpecker/v3/server/model"
)

func TestOrgs(t *testing.T) {
	o := NewOrgs([]string{"woodpecker-ci"})
	assert.True(t, o.IsConfigured)
	assert.True(t, o.IsMember([]*model.Team{{Login: "woodpecker-ci"}}))
	assert.False(t, o.IsMember([]*model.Team{{Login: "not-woodpecker-ci"}}))
	empty := NewOrgs([]string{})
	assert.False(t, empty.IsConfigured)
	assert.False(t, empty.IsMember([]*model.Team{{Login: "woodpecker-ci"}}))
	assert.False(t, empty.IsMember([]*model.Team{{Login: "not-woodpecker-ci"}}))
}

func TestOrgsRejectsSameNamedGroup(t *testing.T) {
	org := NewOrgs([]string{"woodpecker"})

	// the real top-level group
	assert.True(t, org.IsMember([]*model.Team{{Login: "woodpecker"}}))

	// anyone can create a group named "woodpecker" below their own namespace,
	// membership there must never satisfy an allowed org of "woodpecker"
	assert.False(t, org.IsMember([]*model.Team{{Login: "eve/woodpecker"}}))
	assert.False(t, org.IsMember([]*model.Team{{Login: "eve/sub/woodpecker"}}))

	// subgroups are not matched either, only exact membership counts
	assert.False(t, org.IsMember([]*model.Team{{Login: "woodpecker/infra"}}))
}

func TestOrgsWith(t *testing.T) {
	global := NewOrgs([]string{"woodpecker-ci"})
	merged := global.With([]string{"my-group"})

	// both the global and the added orgs are allowed
	assert.True(t, merged.IsConfigured)
	assert.True(t, merged.IsMember([]*model.Team{{Login: "woodpecker-ci"}}))
	assert.True(t, merged.IsMember([]*model.Team{{Login: "my-group"}}))
	assert.False(t, merged.IsMember([]*model.Team{{Login: "other-group"}}))

	// the original list is not modified
	assert.False(t, global.IsMember([]*model.Team{{Login: "my-group"}}))

	// adding to an unconfigured list only checks the added orgs
	fromEmpty := NewOrgs(nil).With([]string{"my-group"})
	assert.True(t, fromEmpty.IsConfigured)
	assert.True(t, fromEmpty.IsMember([]*model.Team{{Login: "my-group"}}))
	assert.False(t, fromEmpty.IsMember([]*model.Team{{Login: "woodpecker-ci"}}))
}

func TestOrgsIgnoresCase(t *testing.T) {
	org := NewOrgs([]string{"Woodpecker-CI"})

	// the configured name and the one reported by the forge may differ in case
	assert.True(t, org.IsMember([]*model.Team{{Login: "woodpecker-ci"}}))
	assert.True(t, org.IsMember([]*model.Team{{Login: "WOODPECKER-CI"}}))

	// this also holds for a GitLab full path
	group := NewOrgs([]string{"my-group/My-Subgroup"})
	assert.True(t, group.IsMember([]*model.Team{{Login: "My-Group/my-subgroup"}}))
}
