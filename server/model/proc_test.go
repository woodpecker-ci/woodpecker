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

package model

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTree(t *testing.T) {
	procs := []*Proc{{
		ID:      25,
		PID:     2,
		BuildID: 6,
		PPID:    1,
		PGID:    2,
		Name:    "clone",
		State:   StatusSuccess,
		Error:   "0",
	}, {
		ID:      24,
		BuildID: 6,
		PID:     1,
		PPID:    0,
		PGID:    1,
		Name:    "lint",
		State:   StatusFailure,
		Error:   "1",
	}, {
		ID:      26,
		BuildID: 6,
		PID:     3,
		PPID:    1,
		PGID:    3,
		Name:    "lint",
		State:   StatusFailure,
		Error:   "1",
	}}
	procs, err := Tree(procs)
	assert.NoError(t, err)
	assert.Len(t, procs, 1)
	assert.Len(t, procs[0].Children, 2)

	procs = []*Proc{{
		ID:      25,
		PID:     2,
		BuildID: 6,
		PPID:    1,
		PGID:    2,
		Name:    "clone",
		State:   StatusSuccess,
		Error:   "0",
	}}
	_, err = Tree(procs)
	assert.Error(t, err)
}
