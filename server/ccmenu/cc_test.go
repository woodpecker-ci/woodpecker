// Copyright 2022 Woodpecker Authors
// Copyright 2018 Drone.IO Inc.
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

package ccmenu

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"go.woodpecker-ci.org/woodpecker/v3/server/model"
)

func TestCC(t *testing.T) {
	t.Run("create a project", func(t *testing.T) {
		now := time.Now().Unix()
		nowFmt := time.Unix(now, 0).Format(time.RFC3339)
		r := &model.Repo{
			FullName: "foo/bar",
		}
		b := &model.Pipeline{
			Status:  model.StatusSuccess,
			Number:  1,
			Started: now,
		}
		cc := New(r, b, "http://localhost/foo/bar/1")

		assert.Equal(t, "foo/bar", cc.Project.Name)
		assert.Equal(t, "Sleeping", cc.Project.Activity)
		assert.Equal(t, "Success", cc.Project.LastBuildStatus)
		assert.Equal(t, "1", cc.Project.LastBuildLabel)
		assert.Equal(t, nowFmt, cc.Project.LastBuildTime)
		assert.Equal(t, "http://localhost/foo/bar/1", cc.Project.WebURL)
	})

	t.Run("properly label exceptions", func(t *testing.T) {
		r := &model.Repo{FullName: "foo/bar"}
		b := &model.Pipeline{
			Status:  model.StatusError,
			Number:  1,
			Started: 1257894000,
		}
		cc := New(r, b, "http://localhost/foo/bar/1")
		assert.Equal(t, "Exception", cc.Project.LastBuildStatus)
		assert.Equal(t, "Sleeping", cc.Project.Activity)
	})

	t.Run("properly label success", func(t *testing.T) {
		r := &model.Repo{FullName: "foo/bar"}
		b := &model.Pipeline{
			Status:  model.StatusSuccess,
			Number:  1,
			Started: 1257894000,
		}
		cc := New(r, b, "http://localhost/foo/bar/1")
		assert.Equal(t, "Success", cc.Project.LastBuildStatus)
		assert.Equal(t, "Sleeping", cc.Project.Activity)
	})

	t.Run("properly label failure", func(t *testing.T) {
		r := &model.Repo{FullName: "foo/bar"}
		b := &model.Pipeline{
			Status:  model.StatusFailure,
			Number:  1,
			Started: 1257894000,
		}
		cc := New(r, b, "http://localhost/foo/bar/1")
		assert.Equal(t, "Failure", cc.Project.LastBuildStatus)
		assert.Equal(t, "Sleeping", cc.Project.Activity)
	})

	t.Run("properly label running", func(t *testing.T) {
		r := &model.Repo{FullName: "foo/bar"}
		b := &model.Pipeline{
			Status:  model.StatusRunning,
			Number:  1,
			Started: 1257894000,
		}
		cc := New(r, b, "http://localhost/foo/bar/1")
		assert.Equal(t, "Building", cc.Project.Activity)
		assert.Equal(t, "Unknown", cc.Project.LastBuildStatus)
		assert.Equal(t, "Unknown", cc.Project.LastBuildLabel)
	})
}
