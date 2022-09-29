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

type mockUpdatePipelineStore struct{}

func (m *mockUpdatePipelineStore) UpdatePipeline(pipeline *model.Pipeline) error {
	return nil
}

func TestUpdateToStatusRunning(t *testing.T) {
	t.Parallel()

	build, _ := UpdateToStatusRunning(&mockUpdatePipelineStore{}, model.Pipeline{}, int64(1))

	if model.StatusRunning != build.Status {
		t.Errorf("Pipeline status not equals '%s' != '%s'", model.StatusRunning, build.Status)
	} else if int64(1) != build.Started {
		t.Errorf("Pipeline started not equals 1 != %d", build.Started)
	}
}

func TestUpdateToStatusPending(t *testing.T) {
	t.Parallel()

	now := time.Now().Unix()

	build, _ := UpdateToStatusPending(&mockUpdatePipelineStore{}, model.Pipeline{}, "Reviewer")

	if model.StatusPending != build.Status {
		t.Errorf("Pipeline status not equals '%s' != '%s'", model.StatusPending, build.Status)
	} else if build.Reviewer != "Reviewer" {
		t.Errorf("Reviewer not equals 'Reviewer' != '%s'", build.Reviewer)
	} else if now > build.Reviewed {
		t.Errorf("Reviewed not updated %d !< %d", now, build.Reviewed)
	}
}

func TestUpdateToStatusDeclined(t *testing.T) {
	t.Parallel()

	now := time.Now().Unix()

	build, _ := UpdateToStatusDeclined(&mockUpdatePipelineStore{}, model.Pipeline{}, "Reviewer")

	if model.StatusDeclined != build.Status {
		t.Errorf("Pipeline status not equals '%s' != '%s'", model.StatusDeclined, build.Status)
	} else if build.Reviewer != "Reviewer" {
		t.Errorf("Reviewer not equals 'Reviewer' != '%s'", build.Reviewer)
	} else if now > build.Reviewed {
		t.Errorf("Reviewed not updated %d !< %d", now, build.Reviewed)
	}
}

func TestUpdateToStatusToDone(t *testing.T) {
	t.Parallel()

	build, _ := UpdateStatusToDone(&mockUpdatePipelineStore{}, model.Pipeline{}, "status", int64(1))

	if build.Status != "status" {
		t.Errorf("Pipeline status not equals 'status' != '%s'", build.Status)
	} else if int64(1) != build.Finished {
		t.Errorf("Pipeline finished not equals 1 != %d", build.Finished)
	}
}

func TestUpdateToStatusError(t *testing.T) {
	t.Parallel()

	now := time.Now().Unix()

	build, _ := UpdateToStatusError(&mockUpdatePipelineStore{}, model.Pipeline{}, errors.New("error"))

	if build.Error != "error" {
		t.Errorf("Pipeline error not equals 'error' != '%s'", build.Error)
	} else if model.StatusError != build.Status {
		t.Errorf("Pipeline status not equals '%s' != '%s'", model.StatusError, build.Status)
	} else if now > build.Started {
		t.Errorf("Started not updated %d !< %d", now, build.Started)
	} else if build.Started != build.Finished {
		t.Errorf("Pipeline started and finished not equals %d != %d", build.Started, build.Finished)
	}
}

func TestUpdateToStatusKilled(t *testing.T) {
	t.Parallel()

	now := time.Now().Unix()

	build, _ := UpdateToStatusKilled(&mockUpdatePipelineStore{}, model.Pipeline{})

	if model.StatusKilled != build.Status {
		t.Errorf("Pipeline status not equals '%s' != '%s'", model.StatusKilled, build.Status)
	} else if now > build.Finished {
		t.Errorf("Finished not updated %d !< %d", now, build.Finished)
	}
}
