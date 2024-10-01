// Copyright 2024 Woodpecker Authors
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

func TestTask_GetLabels(t *testing.T) {
	t.Run("Nil Repo", func(t *testing.T) {
		task := &Task{}
		err := task.ApplyLabelsFromRepo(nil)

		assert.Error(t, err)
		assert.Nil(t, task.Labels)
		assert.EqualError(t, err, "repo is nil but needed to get task labels")
	})

	t.Run("Empty Repo", func(t *testing.T) {
		task := &Task{}
		repo := &Repo{}

		err := task.ApplyLabelsFromRepo(repo)

		assert.NoError(t, err)
		assert.NotNil(t, task.Labels)
		assert.Equal(t, map[string]string{
			"repo":           "",
			agentFilterOrgID: "0",
		}, task.Labels)
	})

	t.Run("Empty Labels", func(t *testing.T) {
		task := &Task{}
		repo := &Repo{
			FullName: "test/repo",
			ID:       123,
			OrgID:    456,
		}

		err := task.ApplyLabelsFromRepo(repo)

		assert.NoError(t, err)
		assert.NotNil(t, task.Labels)
		assert.Equal(t, map[string]string{
			"repo":           "test/repo",
			agentFilterOrgID: "456",
		}, task.Labels)
	})

	t.Run("Existing Labels", func(t *testing.T) {
		task := &Task{
			Labels: map[string]string{
				"existing": "label",
			},
		}
		repo := &Repo{
			FullName: "test/repo",
			ID:       123,
			OrgID:    456,
		}

		err := task.ApplyLabelsFromRepo(repo)

		assert.NoError(t, err)
		assert.NotNil(t, task.Labels)
		assert.Equal(t, map[string]string{
			"existing":       "label",
			"repo":           "test/repo",
			agentFilterOrgID: "456",
		}, task.Labels)
	})
}
