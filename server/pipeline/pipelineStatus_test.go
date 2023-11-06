// Copyright 2022 Woodpecker Authors
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

package pipeline

import (
	"errors"
	"testing"
	"time"

	"go.woodpecker-ci.org/woodpecker/server/model"
)

type mockUpdatePipelineStore struct{}

func (m *mockUpdatePipelineStore) UpdatePipeline(_ *model.Pipeline) error {
	return nil
}

func TestUpdateToStatusRunning(t *testing.T) {
	t.Parallel()

	pipeline, _ := UpdateToStatusRunning(&mockUpdatePipelineStore{}, model.Pipeline{}, int64(1))

	if model.StatusRunning != pipeline.Status {
		t.Errorf("Pipeline status not equals '%s' != '%s'", model.StatusRunning, pipeline.Status)
	} else if int64(1) != pipeline.Started {
		t.Errorf("Pipeline started not equals 1 != %d", pipeline.Started)
	}
}

func TestUpdateToStatusPending(t *testing.T) {
	t.Parallel()

	now := time.Now().Unix()

	pipeline, _ := UpdateToStatusPending(&mockUpdatePipelineStore{}, model.Pipeline{}, "Reviewer")

	if model.StatusPending != pipeline.Status {
		t.Errorf("Pipeline status not equals '%s' != '%s'", model.StatusPending, pipeline.Status)
	} else if pipeline.Reviewer != "Reviewer" {
		t.Errorf("Reviewer not equals 'Reviewer' != '%s'", pipeline.Reviewer)
	} else if now > pipeline.Reviewed {
		t.Errorf("Reviewed not updated %d !< %d", now, pipeline.Reviewed)
	}
}

func TestUpdateToStatusDeclined(t *testing.T) {
	t.Parallel()

	now := time.Now().Unix()

	pipeline, _ := UpdateToStatusDeclined(&mockUpdatePipelineStore{}, model.Pipeline{}, "Reviewer")

	if model.StatusDeclined != pipeline.Status {
		t.Errorf("Pipeline status not equals '%s' != '%s'", model.StatusDeclined, pipeline.Status)
	} else if pipeline.Reviewer != "Reviewer" {
		t.Errorf("Reviewer not equals 'Reviewer' != '%s'", pipeline.Reviewer)
	} else if now > pipeline.Reviewed {
		t.Errorf("Reviewed not updated %d !< %d", now, pipeline.Reviewed)
	}
}

func TestUpdateToStatusToDone(t *testing.T) {
	t.Parallel()

	pipeline, _ := UpdateStatusToDone(&mockUpdatePipelineStore{}, model.Pipeline{}, "status", int64(1))

	if pipeline.Status != "status" {
		t.Errorf("Pipeline status not equals 'status' != '%s'", pipeline.Status)
	} else if int64(1) != pipeline.Finished {
		t.Errorf("Pipeline finished not equals 1 != %d", pipeline.Finished)
	}
}

func TestUpdateToStatusError(t *testing.T) {
	t.Parallel()

	now := time.Now().Unix()

	pipeline, _ := UpdateToStatusError(&mockUpdatePipelineStore{}, model.Pipeline{}, errors.New("this is an error"))

	if len(pipeline.Errors) != 1 {
		t.Errorf("Expected one error, got %d", len(pipeline.Errors))
	} else if pipeline.Errors[0].Error() != "[generic] this is an error" {
		t.Errorf("Pipeline error not equals '[generic] this is an error' != '%s'", pipeline.Errors[0].Error())
	} else if model.StatusError != pipeline.Status {
		t.Errorf("Pipeline status not equals '%s' != '%s'", model.StatusError, pipeline.Status)
	} else if now > pipeline.Started {
		t.Errorf("Started not updated %d !< %d", now, pipeline.Started)
	} else if pipeline.Started != pipeline.Finished {
		t.Errorf("Pipeline started and finished not equals %d != %d", pipeline.Started, pipeline.Finished)
	}
}

func TestUpdateToStatusKilled(t *testing.T) {
	t.Parallel()

	now := time.Now().Unix()

	pipeline, _ := UpdateToStatusKilled(&mockUpdatePipelineStore{}, model.Pipeline{})

	if model.StatusKilled != pipeline.Status {
		t.Errorf("Pipeline status not equals '%s' != '%s'", model.StatusKilled, pipeline.Status)
	} else if now > pipeline.Finished {
		t.Errorf("Finished not updated %d !< %d", now, pipeline.Finished)
	}
}
