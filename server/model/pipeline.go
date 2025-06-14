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
	"go.woodpecker-ci.org/woodpecker/v3/pipeline/errors/types"
)

type Pipeline struct {
	ID                  int64                  `json:"id"                  xorm:"pk autoincr 'id'"`
	RepoID              int64                  `json:"-"                   xorm:"UNIQUE(s) INDEX 'repo_id'"`
	Number              int64                  `json:"number"              xorm:"UNIQUE(s) 'number'"`
	Parent              int64                  `json:"parent"              xorm:"parent"`
	Status              StatusValue            `json:"status"              xorm:"INDEX 'status'"`
	Errors              []*types.PipelineError `json:"errors"              xorm:"json 'errors'"`
	Created             int64                  `json:"created"             xorm:"'created' NOT NULL DEFAULT 0 created"`
	Updated             int64                  `json:"updated"             xorm:"'updated' NOT NULL DEFAULT 0 updated"`
	Started             int64                  `json:"started"             xorm:"started"`
	Finished            int64                  `json:"finished"            xorm:"finished"`
	Reviewer            string                 `json:"reviewed_by"         xorm:"reviewer"`
	Reviewed            int64                  `json:"reviewed"            xorm:"reviewed"`
	Workflows           []*Workflow            `json:"workflows,omitempty" xorm:"-"`
	AdditionalVariables map[string]string      `json:"variables,omitempty" xorm:"json 'additional_variables'"`

	// event related

	Event WebhookEvent `json:"event"                       xorm:"event"`
	// TODO change json to 'commit' in next major
	Commit       *Commit      `json:"commit_pipeline"             xorm:"json 'commit'"`
	Branch       string       `json:"branch"                      xorm:"branch"`
	Ref          string       `json:"ref"                         xorm:"ref"`
	Refspec      string       `json:"refspec"                     xorm:"refspec"`
	ForgeURL     string       `json:"forge_url"                   xorm:"forge_url"`
	Author       string       `json:"author"                      xorm:"author"`
	Avatar       string       `json:"author_avatar"               xorm:"varchar(500) 'avatar'"`
	ChangedFiles []string     `json:"changed_files,omitempty"     xorm:"LONGTEXT 'changed_files'"`
	Deployment   *Deployment  `json:"deployment,omitempty"        xorm:"json 'deployment'"`
	PullRequest  *PullRequest `json:"pull_request,omitempty"      xorm:"json 'pr'"`
	Cron         string       `json:"cron,omitempty"              xorm:"cron"`
	Release      *Release     `json:"release,omitempty"      xorm:"json 'release'"`
}

// APIPipeline TODO remove in next major.
type APIPipeline struct {
	*Pipeline

	DeployTo   string `json:"deploy_to"`
	DeployTask string `json:"deploy_task"`
	Commit     string `json:"commit"`

	Title             string   `json:"title"`
	Message           string   `json:"message"`
	Timestamp         int64    `json:"timestamp"`
	Sender            string   `json:"sender"`
	Email             string   `json:"author_email"`
	PullRequestLabels []string `json:"pr_labels,omitempty"`
	FromFork          bool     `json:"from_fork,omitempty"`
	IsPrerelease      bool     `json:"is_prerelease,omitempty"`
} //	@name Pipeline

// TableName return database table name for xorm.
func (Pipeline) TableName() string {
	return "pipelines"
}

func (p *Pipeline) ToAPIModel() *APIPipeline {
	ap := &APIPipeline{
		Pipeline:  p,
		Commit:    p.Commit.SHA,
		Title:     p.Commit.Message,
		Message:   p.Commit.Message,
		Timestamp: p.Created,
		Sender:    p.Author,
		Email:     p.Commit.Author.Email,
	}

	if p.Deployment != nil {
		ap.DeployTo = p.Deployment.Target
		ap.DeployTask = p.Deployment.Task
	}
	if p.PullRequest != nil {
		ap.Message = p.PullRequest.Title
		ap.PullRequestLabels = p.PullRequest.Labels
		ap.FromFork = p.PullRequest.FromFork
	}
	if p.Release != nil {
		ap.IsPrerelease = p.Release.IsPrerelease
	}
	switch p.Event {
	case EventCron:
		ap.Message = p.Cron
	case EventTag:
		ap.Message = "created tag " + p.Release.TagTitle
	case EventRelease:
		ap.Message = p.Release.TagTitle
	}

	return ap
}

type PipelineFilter struct {
	Before      int64
	After       int64
	Branch      string
	Events      []WebhookEvent
	RefContains string
	Status      StatusValue
}

// IsMultiPipeline checks if step list contain more than one parent step.
func (p Pipeline) IsMultiPipeline() bool {
	return len(p.Workflows) > 1
}

type PipelineOptions struct {
	Branch    string            `json:"branch"`
	Variables map[string]string `json:"variables"`
} //	@name PipelineOptions

type Deployment struct {
	Target      string `json:"target"`
	Task        string `json:"task"`
	Description string `json:"description"`
}

type Release struct {
	IsPrerelease bool   `json:"is_prerelease,omitempty"`
	TagTitle     string `json:"release_tag_title,omitempty"`
}
