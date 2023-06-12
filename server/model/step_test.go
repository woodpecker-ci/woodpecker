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
	steps := []*Step{{
		ID:         25,
		UUID:       "f80df0bb-77a7-4964-9412-2e1049872d57",
		PID:        2,
		PipelineID: 6,
		PPID:       1,
		Name:       "clone",
		State:      StatusSuccess,
		Error:      "0",
	}, {
		ID:         24,
		UUID:       "c19b49c5-990d-4722-ba9c-1c4fe9db1f91",
		PipelineID: 6,
		PID:        1,
		PPID:       0,
		Name:       "lint",
		State:      StatusFailure,
		Error:      "1",
	}, {
		ID:         26,
		UUID:       "4380146f-c0ff-4482-8107-c90937d1faba",
		PipelineID: 6,
		PID:        3,
		PPID:       1,
		Name:       "lint",
		State:      StatusFailure,
		Error:      "1",
	}}
	steps, err := Tree(steps)
	assert.NoError(t, err)
	assert.Len(t, steps, 1)
	assert.Len(t, steps[0].Children, 2)

	steps = []*Step{{
		ID:         25,
		UUID:       "f80df0bb-77a7-4964-9412-2e1049872d57",
		PID:        2,
		PipelineID: 6,
		PPID:       1,
		Name:       "clone",
		State:      StatusSuccess,
		Error:      "0",
	}}
	_, err = Tree(steps)
	assert.Error(t, err)
}
