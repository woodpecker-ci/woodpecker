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
	"encoding/json"
	"fmt"
	"reflect"
	"strings"

	"src.techknowlogick.com/xormigrate"
	"xorm.io/builder"
	"xorm.io/xorm"
	"xorm.io/xorm/schemas"

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
			Draft     bool                `json:"draft,omitempty"`
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
			PullRequestDraft     bool     `xorm:"pr_draft"`
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

		dialect := sess.Engine().Dialect().URI().DBType

		// jsonThen is the THEN placeholder used for JSON columns.
		// Postgres maps `json` column tag to a native json type,
		// so untyped bind parameters inside a CASE need an explicit cast.
		jsonThen := "?"
		if dialect == schemas.POSTGRES {
			jsonThen = "CAST(? AS json)"
		}

		// marshalJSON encodes v as a JSON string, returning a nil interface for
		// nil pointers so the column is written as SQL NULL rather than the
		// literal "null".
		marshalJSON := func(v any) (any, error) {
			if rv := reflect.ValueOf(v); rv.Kind() == reflect.Pointer && rv.IsNil() {
				return nil, nil
			}
			b, err := json.Marshal(v)
			if err != nil {
				return nil, err
			}
			return string(b), nil
		}

		page := 0
		oldPipelines := make([]*pipelines, 0, perPage024)

		for {
			oldPipelines = oldPipelines[:0]

			err := sess.Limit(perPage024, page*perPage024).
				OrderBy("id ASC").
				Where(builder.In("event", model.EventPull, model.EventPullClosed, model.EventPullMetadata, model.EventDeploy)).
				Find(&oldPipelines)
			if err != nil {
				return err
			}
			if len(oldPipelines) == 0 {
				break
			}

			// Build a single bulk UPDATE for the whole page instead of one
			// statement per row. Every target column becomes a CASE keyed by
			// id, collapsing perPage024 round-trips into one statement.
			var (
				commitCase  strings.Builder
				deployCase  strings.Builder
				prCase      strings.Builder
				releaseCase strings.Builder
				tagCase     strings.Builder

				commitArgs  []any
				deployArgs  []any
				prArgs      []any
				releaseArgs []any
				tagArgs     []any
				ids         []any
			)

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
				case model.EventPull, model.EventPullClosed, model.EventPullMetadata:
					// derive the pull request index from the ref, covering every forge
					// ref layout: GitHub/Gitea/Forgejo (refs/pull/N/head), GitLab
					// (refs/merge-requests/N/head|merge) and Bitbucket cloud + datacenter
					// (refs/pull-requests/N/from).
					prIndex := p.Ref
					for _, prefix := range []string{"refs/pull/", "refs/merge-requests/", "refs/pull-requests/"} {
						prIndex = strings.TrimPrefix(prIndex, prefix)
					}
					for _, suffix := range []string{"/merge", "/head", "/from"} {
						prIndex = strings.TrimSuffix(prIndex, suffix)
					}
					p.PullRequest = &pullRequest{
						Title:     p.Title,
						Index:     model.ForgeRemoteID(prIndex),
						FromFork:  p.FromFork,
						Labels:    p.PullRequestLabels,
						Milestone: p.PullRequestMilestone,
						Draft:     p.PullRequestDraft,
					}
				case model.EventDeploy:
					p.Deployment = &deployment{
						Description: p.Message,
						Target:      p.DeployTo,
						Task:        p.DeployTask,
					}
				}

				commitJSON, err := marshalJSON(p.CommitNew)
				if err != nil {
					return err
				}
				deployJSON, err := marshalJSON(p.Deployment)
				if err != nil {
					return err
				}
				prJSON, err := marshalJSON(p.PullRequest)
				if err != nil {
					return err
				}
				releaseJSON, err := marshalJSON(p.Release)
				if err != nil {
					return err
				}

				ids = append(ids, p.ID)

				commitCase.WriteString(" WHEN ? THEN " + jsonThen)
				commitArgs = append(commitArgs, p.ID, commitJSON)
				deployCase.WriteString(" WHEN ? THEN " + jsonThen)
				deployArgs = append(deployArgs, p.ID, deployJSON)
				prCase.WriteString(" WHEN ? THEN " + jsonThen)
				prArgs = append(prArgs, p.ID, prJSON)
				releaseCase.WriteString(" WHEN ? THEN " + jsonThen)
				releaseArgs = append(releaseArgs, p.ID, releaseJSON)
				tagCase.WriteString(" WHEN ? THEN ?")
				tagArgs = append(tagArgs, p.ID, p.TagTitle)
			}

			placeholders := strings.TrimSuffix(strings.Repeat("?,", len(ids)), ",")
			query := fmt.Sprintf(
				"UPDATE `pipelines` SET "+
					"`commit_new` = CASE `id`%s END, "+
					"`deployment` = CASE `id`%s END, "+
					"`pull_request` = CASE `id`%s END, "+
					"`release` = CASE `id`%s END, "+
					"`tag_title` = CASE `id`%s END "+
					"WHERE `id` IN (%s)",
				commitCase.String(), deployCase.String(), prCase.String(),
				releaseCase.String(), tagCase.String(), placeholders,
			)

			// Argument order must match the textual order of the placeholders
			// above: each column's CASE, then the WHERE IN list.
			execArgs := make([]any, 0, 1+len(commitArgs)+len(deployArgs)+len(prArgs)+len(releaseArgs)+len(tagArgs)+len(ids))
			execArgs = append(execArgs, query)
			execArgs = append(execArgs, commitArgs...)
			execArgs = append(execArgs, deployArgs...)
			execArgs = append(execArgs, prArgs...)
			execArgs = append(execArgs, releaseArgs...)
			execArgs = append(execArgs, tagArgs...)
			execArgs = append(execArgs, ids...)

			if _, err := sess.Exec(execArgs...); err != nil {
				return err
			}

			if len(oldPipelines) < perPage024 {
				break
			}

			page++
		}

		if err := dropTableColumns(sess, "pipelines", "email", "timestamp", "sender", "commit", "title", "message", "deploy", "deploy_task", "pr_labels", "pr_milestone", "pr_draft", "is_prerelease", "from_fork"); err != nil {
			return err
		}

		if err := renameColumn(sess, "pipelines", "avatar", "author_avatar"); err != nil {
			return err
		}

		return renameColumn(sess, "pipelines", "commit_new", "commit")
	},
}
