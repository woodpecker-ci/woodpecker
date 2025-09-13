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
	ID                  int64                  `json:"id"                      xorm:"pk autoincr 'id'"`
	RepoID              int64                  `json:"-"                       xorm:"UNIQUE(s) INDEX 'repo_id'"`
	Number              int64                  `json:"number"                  xorm:"UNIQUE(s) 'number'"`
	Author              string                 `json:"author"                  xorm:"author"` // The user sending the webhook data or triggering the pipeline event
	Parent              int64                  `json:"parent"                  xorm:"parent"`
	Event               WebhookEvent           `json:"event"                   xorm:"event"`
	Status              StatusValue            `json:"status"                  xorm:"INDEX 'status'"`
	Errors              []*types.PipelineError `json:"errors"                  xorm:"json 'errors'"`
	Created             int64                  `json:"created"                 xorm:"'created' NOT NULL DEFAULT 0 created"`
	Updated             int64                  `json:"updated"                 xorm:"'updated' NOT NULL DEFAULT 0 updated"`
	Started             int64                  `json:"started"                 xorm:"started"`
	Finished            int64                  `json:"finished"                xorm:"finished"`
	Commit              *Commit                `json:"commit_pipeline"         xorm:"json 'commit'"` // TODO change json to 'commit' in next major
	Branch              string                 `json:"branch"                  xorm:"branch"`
	Ref                 string                 `json:"ref"                     xorm:"ref"`
	Refspec             string                 `json:"refspec"                 xorm:"refspec"`
	AuthorAvatar        string                 `json:"author_avatar"           xorm:"varchar(500) 'avatar'"` // Avatar URL of the author of the commit
	ForgeURL            string                 `json:"forge_url"               xorm:"forge_url"`
	Reviewer            string                 `json:"reviewed_by"             xorm:"reviewer"`
	Reviewed            int64                  `json:"reviewed"                xorm:"reviewed"` // timestamp of the review
	Workflows           []*Workflow            `json:"workflows,omitempty"     xorm:"-"`
	ChangedFiles        []string               `json:"changed_files,omitempty" xorm:"LONGTEXT 'changed_files'"`
	AdditionalVariables map[string]string      `json:"variables,omitempty"     xorm:"json 'additional_variables'"`
	Deployment          *Deployment            `json:"deployment,omitempty"    xorm:"json 'deployment'"`
	PullRequest         *PullRequest           `json:"pull_request,omitempty"  xorm:"json 'pull_request'"`
	Cron                string                 `json:"cron,omitempty"          xorm:"cron"` // name of the cron job
	Release             *Release               `json:"release,omitempty"       xorm:"json 'release'"`
	TagTitle            string                 `json:"tag_title,omitempty"     xorm:"tag_title"`
}

// APIPipeline TODO remove deprecated properties in next major.
type APIPipeline struct {
	*Pipeline

	DeployTo          string   `json:"deploy_to"`               // deprecated, use deployment.target instead
	DeployTask        string   `json:"deploy_task"`             // deprecated, use deployment.task instead
	Commit            string   `json:"commit"`                  // deprecated, use commit_pipeline.sha instead
	Title             string   `json:"title"`                   // deprecated, use pull_request.title (pull_request & pull_request_closed) or deployment.description
	Message           string   `json:"message"`                 // deprecated, use commit.message (pull_request, pull_request_closed & push), deployment.description, cron (cron) or tag_title (tag) instead
	Timestamp         int64    `json:"timestamp"`               // deprecated, use created instead
	Sender            string   `json:"sender"`                  // deprecated, use author instead
	Email             string   `json:"author_email"`            // deprecated, use commit.author.email instead
	PullRequestLabels []string `json:"pr_labels,omitempty"`     // deprecated, use pull_request.labels instead
	FromFork          bool     `json:"from_fork,omitempty"`     // deprecated, use pull_request.from_fork instead
	IsPrerelease      bool     `json:"is_prerelease,omitempty"` // deprecated, use release.is_prerelease instead
	Avatar            string   `json:"avatar"`                  // deprecated, use author_avatar instead
} //	@name	Pipeline

// TableName return database table name for xorm.
func (Pipeline) TableName() string {
	return "pipelines"
}

func (p *Pipeline) ToAPIModel() *APIPipeline {
	ap := &APIPipeline{
		Pipeline:  p,
		Commit:    p.Commit.SHA,
		Message:   p.Commit.Message,
		Timestamp: p.Created,
		Sender:    p.Author,
		Email:     p.Commit.Author.Email,
	}

	ap.Author = p.Commit.Author.Name
	ap.Avatar = p.AuthorAvatar

	switch p.Event {
	case EventCron:
		ap.Message = p.Cron
	case EventTag:
		ap.Message = p.TagTitle
	case EventRelease:
		if p.Release != nil {
			ap.IsPrerelease = p.Release.IsPrerelease
			ap.Title = p.Release.Title
		}
		ap.Message = "created release " + p.TagTitle
	case EventManual:
		ap.Message = "MANUAL PIPELINE" + p.Branch
	case EventDeploy:
		if p.Deployment != nil {
			ap.DeployTo = p.Deployment.Target
			ap.DeployTask = p.Deployment.Task
			ap.Message = p.Deployment.Description
			ap.Title = p.Deployment.Description
		}
	case EventPull, EventPullClosed:
		if p.PullRequest != nil {
			ap.Title = p.PullRequest.Title
			ap.PullRequestLabels = p.PullRequest.Labels
			ap.FromFork = p.PullRequest.FromFork
		}
		ap.Message = p.Commit.Message
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
} //	@name	PipelineOptions

type Deployment struct {
	Target      string `json:"target"`
	Task        string `json:"task"`
	Description string `json:"description"`
}

type Release struct {
	IsPrerelease bool   `json:"is_prerelease,omitempty"`
	Title        string `json:"title,omitempty"`
}
