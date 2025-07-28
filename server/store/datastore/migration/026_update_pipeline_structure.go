// Copyright 2025 Woodpecker Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package migration

import (
	"strings"

	"src.techknowlogick.com/xormigrate"
	"xorm.io/xorm"

	"go.woodpecker-ci.org/woodpecker/v3/pipeline/errors/types"
	"go.woodpecker-ci.org/woodpecker/v3/server/model"
)

var updatePipelineStructure = xormigrate.Migration{
	ID: "update-pipeline-structure",
	MigrateSession: func(sess *xorm.Session) error {
		// perPage024 set the size of the slice to read per page.
		var perPage024 = 100

		type commitAuthor struct {
			Author string `json:"author"`
			Email  string `json:"email"`
		}

		type commit struct {
			SHA      string       `json:"sha"`
			Message  string       `json:"message"`
			ForgeURL string       `json:"forge_url"`
			Author   commitAuthor `json:"author"`
		}

		type pullRequest struct {
			Index    model.ForgeRemoteID `json:"index"`
			Title    string              `json:"title"`
			Labels   []string            `json:"labels,omitempty"`
			FromFork bool                `json:"from_fork,omitempty"`
		}

		type deployment struct {
			Target      string `json:"target"`
			Task        string `json:"task"`
			Description string `json:"description"`
		}

		type release struct {
			IsPrerelease bool   `json:"is_prerelease,omitempty"`
			Title        string `json:"title,omitempty"`
			TagTitle     string `json:"tag_title,omitempty"`
		}

		type pipelines struct {
			ID                  int64                  `json:"id"                  xorm:"pk autoincr 'id'"`
			RepoID              int64                  `json:"-"                   xorm:"UNIQUE(s) INDEX 'repo_id'"`
			Number              int64                  `json:"number"              xorm:"UNIQUE(s) 'number'"`
			Parent              int64                  `json:"parent"              xorm:"parent"`
			Status              model.StatusValue      `json:"status"              xorm:"INDEX 'status'"`
			Errors              []*types.PipelineError `json:"errors"              xorm:"json 'errors'"`
			Created             int64                  `json:"created"             xorm:"'created' NOT NULL DEFAULT 0 created"`
			Updated             int64                  `json:"updated"             xorm:"'updated' NOT NULL DEFAULT 0 updated"`
			Started             int64                  `json:"started"             xorm:"started"`
			Finished            int64                  `json:"finished"            xorm:"finished"`
			Reviewer            string                 `json:"reviewed_by"         xorm:"reviewer"`
			Reviewed            int64                  `json:"reviewed"            xorm:"reviewed"`
			AdditionalVariables map[string]string      `json:"variables,omitempty" xorm:"json 'additional_variables'"`

			// event related

			Event model.WebhookEvent `json:"event"                       xorm:"event"`
			// TODO change json to 'commit' in next major
			Commit       *commit  `json:"commit_pipeline"             xorm:"json 'commit'"`
			Branch       string   `json:"branch"                      xorm:"branch"`
			Ref          string   `json:"ref"                         xorm:"ref"`
			Refspec      string   `json:"refspec"                     xorm:"refspec"`
			ForgeURL     string   `json:"forge_url"                   xorm:"forge_url"`
			Author       string   `json:"author"                      xorm:"author"` // The user sending the webhook data or triggering the pipeline event
			Avatar       string   `json:"author_avatar"               xorm:"varchar(500) 'avatar'"`
			ChangedFiles []string `json:"changed_files,omitempty"     xorm:"LONGTEXT 'changed_files'"`

			// new fields
			CommitNew   *commit      `xorm:"json 'commit_new'"`
			Deployment  *deployment  `xorm:"json 'deployment'"`
			PullRequest *pullRequest `xorm:"json 'pr'"`
			Cron        string       `xorm:"cron"`
			Release     *release     `xorm:"json 'release'"`
			TagTitle    string       `xorm:"tag_title"`

			// removed without replacement
			Timestamp int64  `xorm:"'timestamp'"`
			Email     string `xorm:"varchar(500) email"`
		}

		type oldPipelineI struct {
			*pipelines
		}

		if err := sess.Sync(new(pipelines)); err != nil {
			return err
		}

		page := 0
		oldPipelines := make([]*pipelines, 0, perPage024)

		for {
			oldPipelines = oldPipelines[:0]

			err := sess.Limit(perPage024, page*perPage024).Cols("id", "event", "author", "forge_url", "commit", "title", "message", "sender", "deploy", "deploy_task", "pr_labels", "from_fork", "is_prerelease", "email").Find(&oldPipelines)
			if err != nil {
				return err
			}

			for _, oldPipeline := range oldPipelines {
				var newPipeline pipelines
				newPipeline.ID = oldPipeline.ID
				newPipeline.CommitNew = &commit{
					SHA:      oldPipeline.Commit,
					Message:  oldPipeline.Message,
					ForgeURL: oldPipeline.ForgeURL,
					Author: commitAuthor{
						Author: oldPipeline.Author,
						Email:  oldPipeline.Email,
					},
				}

				switch oldPipeline.Event {
				case model.EventRelease:
					newPipeline.Release = &release{
						TagTitle:     strings.TrimPrefix(oldPipeline.Message, "created release "),
						IsPrerelease: oldPipeline.IsPrerelease,
					}
					newPipeline.TagTitle = strings.TrimPrefix(oldPipeline.Ref, "refs/tags/")
				case model.EventTag:
					newPipeline.TagTitle = strings.TrimPrefix(oldPipeline.Ref, "refs/tags/")
				case model.EventCron:
					newPipeline.Cron = oldPipeline.Sender
				case model.EventPull, model.EventPullClosed:
					newPipeline.PullRequest = &pullRequest{
						Title: oldPipeline.Title,
						Index: model.ForgeRemoteID(
							strings.TrimSuffix(
								strings.TrimSuffix(
									strings.TrimPrefix(
										strings.TrimPrefix(oldPipeline.Ref, "refs/pull/"),
										"refs/merge-requests/",
									),
									"/merge"),
								"/head",
							),
						),
						FromFork: oldPipeline.FromFork,
						Labels:   oldPipeline.PullRequestLabels,
					}
				case model.EventDeploy:
					newPipeline.Deployment = &deployment{
						Description: oldPipeline.Message,
						Target:      oldPipeline.DeployTo,
						Task:        oldPipeline.DeployTask,
					}
				}

				if _, err := sess.ID(oldPipeline.ID).Cols("commit_new", "deployment", "pr", "cron", "release", "tag_title").Update(newPipeline); err != nil {
					return err
				}
			}

			if len(oldPipelines) < perPage024 {
				break
			}

			page++
		}

		// if err := dropTableColumns(sess, "pipelines", "email", "timestamp", "sender", "commit", "title", "message", "deploy", "deploy_task", "pr_labels", "from_fork"); err != nil {
		// 	return err
		// }

		// return renameColumn(sess, "pipelines", "commit_new", "commit")
		return nil
	},
}
