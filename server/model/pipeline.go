// Copyright 2021 Woodpecker Authors
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

package model

import (
	"go.woodpecker-ci.org/woodpecker/v2/pipeline/errors"
)

type Pipeline struct {
	ID                  int64                   `json:"id"                      xorm:"pk autoincr 'pipeline_id'"`
	RepoID              int64                   `json:"-"                       xorm:"UNIQUE(s) INDEX 'pipeline_repo_id'"`
	Number              int64                   `json:"number"                  xorm:"UNIQUE(s) 'pipeline_number'"`
	Author              string                  `json:"author"                  xorm:"INDEX 'pipeline_author'"`
	Parent              int64                   `json:"parent"                  xorm:"pipeline_parent"`
	Event               WebhookEvent            `json:"event"                   xorm:"pipeline_event"`
	Status              StatusValue             `json:"status"                  xorm:"INDEX 'pipeline_status'"`
	Errors              []*errors.PipelineError `json:"errors"                  xorm:"json 'pipeline_errors'"`
	Created             int64                   `json:"created_at"              xorm:"pipeline_created"`
	Updated             int64                   `json:"updated_at"              xorm:"updated NOT NULL DEFAULT 0 'updated'"`
	Started             int64                   `json:"started_at"              xorm:"pipeline_started"`
	Finished            int64                   `json:"finished_at"             xorm:"pipeline_finished"`
	Deploy              string                  `json:"deploy_to"               xorm:"pipeline_deploy"`
	Commit              string                  `json:"commit"                  xorm:"pipeline_commit"`
	Branch              string                  `json:"branch"                  xorm:"pipeline_branch"`
	Ref                 string                  `json:"ref"                     xorm:"pipeline_ref"`
	Refspec             string                  `json:"refspec"                 xorm:"pipeline_refspec"`
	Title               string                  `json:"title"                   xorm:"pipeline_title"`
	Message             string                  `json:"message"                 xorm:"TEXT 'pipeline_message'"`
	Timestamp           int64                   `json:"timestamp"               xorm:"pipeline_timestamp"`
	Sender              string                  `json:"sender"                  xorm:"pipeline_sender"` // uses reported user for webhooks and name of cron for cron pipelines
	Avatar              string                  `json:"author_avatar"           xorm:"pipeline_avatar"`
	Email               string                  `json:"author_email"            xorm:"pipeline_email"`
	ForgeURL            string                  `json:"forge_url"                xorm:"pipeline_forge_url"`
	Reviewer            string                  `json:"reviewed_by"             xorm:"pipeline_reviewer"`
	Reviewed            int64                   `json:"reviewed_at"             xorm:"pipeline_reviewed"`
	Workflows           []*Workflow             `json:"workflows,omitempty"     xorm:"-"`
	ChangedFiles        []string                `json:"changed_files,omitempty" xorm:"LONGTEXT 'changed_files'"`
	AdditionalVariables map[string]string       `json:"variables,omitempty"     xorm:"json 'additional_variables'"`
	PullRequestLabels   []string                `json:"pr_labels,omitempty"     xorm:"json 'pr_labels'"`
	IsPrerelease        bool                    `json:"is_prerelease,omitempty"     xorm:"is_prerelease"`
} //	@name Pipeline

// TableName return database table name for xorm
func (Pipeline) TableName() string {
	return "pipelines"
}

// IsMultiPipeline checks if step list contain more than one parent step
func (p Pipeline) IsMultiPipeline() bool {
	return len(p.Workflows) > 1
}

type PipelineOptions struct {
	Branch    string            `json:"branch"`
	Variables map[string]string `json:"variables"`
} //	@name PipelineOptions
