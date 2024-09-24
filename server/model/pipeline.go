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
	"go.woodpecker-ci.org/woodpecker/v2/pipeline/errors/types"
)

type Pipeline struct {
	ID                  int64                  `json:"id"                      xorm:"pk autoincr 'id'"`
	RepoID              int64                  `json:"-"                       xorm:"UNIQUE(s) INDEX 'repo_id'"`
	Number              int64                  `json:"number"                  xorm:"UNIQUE(s) 'number'"`
	Author              string                 `json:"author"                  xorm:"INDEX 'author'"`
	Parent              int64                  `json:"parent"                  xorm:"parent"`
	Event               WebhookEvent           `json:"event"                   xorm:"event"`
	Status              StatusValue            `json:"status"                  xorm:"INDEX 'status'"`
	Errors              []*types.PipelineError `json:"errors"                  xorm:"json 'errors'"`
	Created             int64                  `json:"created"                 xorm:"'created' NOT NULL DEFAULT 0 created"`
	Updated             int64                  `json:"updated"                 xorm:"'updated' NOT NULL DEFAULT 0 updated"`
	Started             int64                  `json:"started"                 xorm:"started"`
	Finished            int64                  `json:"finished"                xorm:"finished"`
	DeployTo            string                 `json:"deploy_to"               xorm:"deploy"`
	DeployTask          string                 `json:"deploy_task"             xorm:"deploy_task"`
	Commit              string                 `json:"commit"                  xorm:"commit"`
	Branch              string                 `json:"branch"                  xorm:"branch"`
	Ref                 string                 `json:"ref"                     xorm:"ref"`
	Refspec             string                 `json:"refspec"                 xorm:"refspec"`
	Title               string                 `json:"title"                   xorm:"title"`
	Message             string                 `json:"message"                 xorm:"TEXT 'message'"`
	Timestamp           int64                  `json:"timestamp"               xorm:"'timestamp'"`
	Sender              string                 `json:"sender"                  xorm:"sender"` // uses reported user for webhooks and name of cron for cron pipelines
	Avatar              string                 `json:"author_avatar"           xorm:"avatar"`
	Email               string                 `json:"author_email"            xorm:"email"`
	ForgeURL            string                 `json:"forge_url"               xorm:"forge_url"`
	Reviewer            string                 `json:"reviewed_by"             xorm:"reviewer"`
	Reviewed            int64                  `json:"reviewed"                xorm:"reviewed"`
	Workflows           []*Workflow            `json:"workflows,omitempty"     xorm:"-"`
	ChangedFiles        []string               `json:"changed_files,omitempty" xorm:"LONGTEXT 'changed_files'"`
	AdditionalVariables map[string]string      `json:"variables,omitempty"     xorm:"json 'additional_variables'"`
	PullRequestLabels   []string               `json:"pr_labels,omitempty"     xorm:"json 'pr_labels'"`
	IsPrerelease        bool                   `json:"is_prerelease,omitempty" xorm:"is_prerelease"`
} //	@name Pipeline

// TableName return database table name for xorm.
func (Pipeline) TableName() string {
	return "pipelines"
}

type PipelineFilter struct {
	Before int64
	After  int64
}

// IsMultiPipeline checks if step list contain more than one parent step.
func (p Pipeline) IsMultiPipeline() bool {
	return len(p.Workflows) > 1
}

type PipelineOptions struct {
	Branch    string            `json:"branch"`
	Variables map[string]string `json:"variables"`
} //	@name PipelineOptions
