// Copyright 2019 mhmxs.
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

package shared

import (
	"errors"
	"testing"
	"time"

	"github.com/woodpecker-ci/woodpecker/server/model"
)

// TODO(974) move to server/pipeline/*

type mockUpdateBuildStore struct{}

func (m *mockUpdateBuildStore) UpdateBuild(build *model.Build) error {
	return nil
}

func TestUpdateToStatusRunning(t *testing.T) {
	t.Parallel()

	build, _ := UpdateToStatusRunning(&mockUpdateBuildStore{}, model.Build{}, int64(1))

	if model.StatusRunning != build.Status {
		t.Errorf("Build status not equals '%s' != '%s'", model.StatusRunning, build.Status)
	} else if int64(1) != build.Started {
		t.Errorf("Build started not equals 1 != %d", build.Started)
	}
}

func TestUpdateToStatusPending(t *testing.T) {
	t.Parallel()

	now := time.Now().Unix()

	build, _ := UpdateToStatusPending(&mockUpdateBuildStore{}, model.Build{}, "Reviewer")

	if model.StatusPending != build.Status {
		t.Errorf("Build status not equals '%s' != '%s'", model.StatusPending, build.Status)
	} else if build.Reviewer != "Reviewer" {
		t.Errorf("Reviewer not equals 'Reviewer' != '%s'", build.Reviewer)
	} else if now > build.Reviewed {
		t.Errorf("Reviewed not updated %d !< %d", now, build.Reviewed)
	}
}

func TestUpdateToStatusDeclined(t *testing.T) {
	t.Parallel()

	now := time.Now().Unix()

	build, _ := UpdateToStatusDeclined(&mockUpdateBuildStore{}, model.Build{}, "Reviewer")

	if model.StatusDeclined != build.Status {
		t.Errorf("Build status not equals '%s' != '%s'", model.StatusDeclined, build.Status)
	} else if build.Reviewer != "Reviewer" {
		t.Errorf("Reviewer not equals 'Reviewer' != '%s'", build.Reviewer)
	} else if now > build.Reviewed {
		t.Errorf("Reviewed not updated %d !< %d", now, build.Reviewed)
	}
}

func TestUpdateToStatusToDone(t *testing.T) {
	t.Parallel()

	build, _ := UpdateStatusToDone(&mockUpdateBuildStore{}, model.Build{}, "status", int64(1))

	if build.Status != "status" {
		t.Errorf("Build status not equals 'status' != '%s'", build.Status)
	} else if int64(1) != build.Finished {
		t.Errorf("Build finished not equals 1 != %d", build.Finished)
	}
}

func TestUpdateToStatusError(t *testing.T) {
	t.Parallel()

	now := time.Now().Unix()

	build, _ := UpdateToStatusError(&mockUpdateBuildStore{}, model.Build{}, errors.New("error"))

	if build.Error != "error" {
		t.Errorf("Build error not equals 'error' != '%s'", build.Error)
	} else if model.StatusError != build.Status {
		t.Errorf("Build status not equals '%s' != '%s'", model.StatusError, build.Status)
	} else if now > build.Started {
		t.Errorf("Started not updated %d !< %d", now, build.Started)
	} else if build.Started != build.Finished {
		t.Errorf("Build started and finished not equals %d != %d", build.Started, build.Finished)
	}
}

func TestUpdateToStatusKilled(t *testing.T) {
	t.Parallel()

	now := time.Now().Unix()

	build, _ := UpdateToStatusKilled(&mockUpdateBuildStore{}, model.Build{})

	if model.StatusKilled != build.Status {
		t.Errorf("Build status not equals '%s' != '%s'", model.StatusKilled, build.Status)
	} else if now > build.Finished {
		t.Errorf("Finished not updated %d !< %d", now, build.Finished)
	}
}
