// Copyright 2026 Woodpecker Authors
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

	"go.woodpecker-ci.org/woodpecker/v3/server/model"
)

var updatePipelineStructure = xormigrate.Migration{
	ID: "update-pipeline-structure",
	MigrateSession: func(sess *xorm.Session) error {
		// perPage024 set the size of the slice to read per page.
		perPage024 := 100

		type commitAuthor struct {
			Name  string `json:"name"`
			Email string `json:"email"`
		}

		type commit struct {
			SHA       string       `json:"sha"`
			Message   string       `json:"message"`
			ForgeURL  string       `json:"forge_url"`
			Author    commitAuthor `json:"author"`
			Timestamp int64        `json:"timestamp"`
		}

		type pullRequest struct {
			Index     model.ForgeRemoteID `json:"index"`
			Title     string              `json:"title"`
			Labels    []string            `json:"labels,omitempty"`
			Milestone string              `json:"milestone,omitempty"`
			FromFork  bool                `json:"from_fork,omitempty"`
		}

		type deployment struct {
			Target      string `json:"target"`
			Task        string `json:"task"`
			Description string `json:"description"`
		}

		type release struct {
			IsPrerelease bool   `json:"is_prerelease,omitempty"`
			Title        string `json:"title,omitempty"`
		}

		type pipelines struct {
			ID       int64              `xorm:"pk autoincr 'id'"`
			Event    model.WebhookEvent `xorm:"event"`
			Author   string             `xorm:"INDEX 'author'"`
			ForgeURL string             `xorm:"forge_url"`
			Ref      string             `xorm:"ref"`

			Commit               string   `xorm:"commit"`
			Title                string   `xorm:"title"`
			Message              string   `xorm:"TEXT 'message'"`
			Sender               string   `xorm:"sender"` // uses reported user for webhooks and name of cron for cron pipelines
			DeployTo             string   `xorm:"deploy"`
			DeployTask           string   `xorm:"deploy_task"`
			PullRequestLabels    []string `xorm:"json 'pr_labels'"`
			PullRequestMilestone string   `xorm:"pr_milestone"`
			FromFork             bool     `xorm:"from_fork"`
			IsPrerelease         bool     `xorm:"is_prerelease"`
			Timestamp            int64    `xorm:"'timestamp'"`

			// new fields
			CommitNew   *commit      `xorm:"json 'commit_new'"`
			Deployment  *deployment  `xorm:"json 'deployment'"`
			PullRequest *pullRequest `xorm:"json 'pull_request'"`
			Release     *release     `xorm:"json 'release'"`
			TagTitle    string       `xorm:"tag_title"`

			// removed without replacement
			Email string `xorm:"varchar(500) email"`
		}

		if err := sess.Sync(new(pipelines)); err != nil {
			return err
		}

		page := 0
		oldPipelines := make([]*pipelines, 0, perPage024)

		for {
			oldPipelines = oldPipelines[:0]

			err := sess.Limit(perPage024, page*perPage024).Find(&oldPipelines)
			if err != nil {
				return err
			}

			// fill new fields with old values
			for _, p := range oldPipelines {
				p.CommitNew = &commit{
					SHA:      p.Commit,
					Message:  p.Message,
					ForgeURL: p.ForgeURL,
					Author: commitAuthor{
						Name:  p.Author,
						Email: p.Email,
					},
					Timestamp: p.Timestamp,
				}

				switch p.Event {
				case model.EventRelease:
					p.Release = &release{
						Title:        strings.TrimPrefix(p.Message, "created release "),
						IsPrerelease: p.IsPrerelease,
					}
					p.TagTitle = strings.TrimPrefix(p.Ref, "refs/tags/")
				case model.EventTag:
					p.TagTitle = strings.TrimPrefix(p.Ref, "refs/tags/")
				case model.EventPull, model.EventPullClosed:
					p.PullRequest = &pullRequest{
						Title: p.Title,
						Index: model.ForgeRemoteID(
							strings.TrimSuffix(
								strings.TrimSuffix(
									strings.TrimPrefix(
										strings.TrimPrefix(p.Ref, "refs/pull/"),
										"refs/merge-requests/",
									),
									"/merge",
								),
								"/head",
							),
						),
						FromFork:  p.FromFork,
						Labels:    p.PullRequestLabels,
						Milestone: p.PullRequestMilestone,
					}
				case model.EventDeploy:
					p.Deployment = &deployment{
						Description: p.Message,
						Target:      p.DeployTo,
						Task:        p.DeployTask,
					}
				}

				if _, err := sess.ID(p.ID).Cols("commit_new", "deployment", "pull_request", "release", "tag_title").Update(p); err != nil {
					return err
				}
			}

			if len(oldPipelines) < perPage024 {
				break
			}

			page++
		}

		if err := dropTableColumns(sess, "pipelines", "email", "timestamp", "sender", "commit", "title", "message", "deploy", "deploy_task", "pr_labels", "pr_milestone", "is_prerelease", "from_fork"); err != nil {
			return err
		}

		return renameColumn(sess, "pipelines", "commit_new", "commit")
	},
}
