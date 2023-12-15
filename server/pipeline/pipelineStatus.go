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
	"time"

	"go.woodpecker-ci.org/woodpecker/v2/pipeline/errors"
	"go.woodpecker-ci.org/woodpecker/v2/server/model"
)

func UpdateToStatusRunning(store model.UpdatePipelineStore, pipeline model.Pipeline, started int64) (*model.Pipeline, error) {
	pipeline.Status = model.StatusRunning
	pipeline.Started = started
	return &pipeline, store.UpdatePipeline(&pipeline)
}

func UpdateToStatusPending(store model.UpdatePipelineStore, pipeline model.Pipeline, reviewer string) (*model.Pipeline, error) {
	if reviewer != "" {
		pipeline.Reviewer = reviewer
		pipeline.Reviewed = time.Now().Unix()
	}
	pipeline.Status = model.StatusPending
	return &pipeline, store.UpdatePipeline(&pipeline)
}

func UpdateToStatusDeclined(store model.UpdatePipelineStore, pipeline model.Pipeline, reviewer string) (*model.Pipeline, error) {
	pipeline.Reviewer = reviewer
	pipeline.Status = model.StatusDeclined
	pipeline.Reviewed = time.Now().Unix()
	pipeline.AccessToken = ""
	return &pipeline, store.UpdatePipeline(&pipeline)
}

func UpdateStatusToDone(store model.UpdatePipelineStore, pipeline model.Pipeline, status model.StatusValue, stopped int64) (*model.Pipeline, error) {
	pipeline.Status = status
	pipeline.Finished = stopped
	pipeline.AccessToken = ""
	return &pipeline, store.UpdatePipeline(&pipeline)
}

func UpdateToStatusError(store model.UpdatePipelineStore, pipeline model.Pipeline, err error) (*model.Pipeline, error) {
	pipeline.Errors = errors.GetPipelineErrors(err)
	pipeline.Status = model.StatusError
	pipeline.Started = time.Now().Unix()
	pipeline.Finished = pipeline.Started
	pipeline.AccessToken = ""
	return &pipeline, store.UpdatePipeline(&pipeline)
}

func UpdateToStatusKilled(store model.UpdatePipelineStore, pipeline model.Pipeline) (*model.Pipeline, error) {
	pipeline.Status = model.StatusKilled
	pipeline.Finished = time.Now().Unix()
	pipeline.AccessToken = ""
	return &pipeline, store.UpdatePipeline(&pipeline)
}
