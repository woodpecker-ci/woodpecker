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
