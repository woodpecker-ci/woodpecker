// Copyright 2023 Woodpecker Authors
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

package metadata

import (
	"fmt"
	"path"
	"regexp"
	"strconv"
	"strings"
)

var pullRegexp = regexp.MustCompile(`\d+`)

// Environ returns the metadata as a map of environment variables.
func (m *Metadata) Environ() map[string]string {
	var (
		sourceBranch string
		targetBranch string
	)

	branchParts := strings.Split(m.Curr.Commit.Refspec, ":")
	if len(branchParts) == 2 {
		sourceBranch = branchParts[0]
		targetBranch = branchParts[1]
	}

	params := map[string]string{
		"CI":                     m.Sys.Name,
		"CI_REPO":                path.Join(m.Repo.Owner, m.Repo.Name),
		"CI_REPO_NAME":           m.Repo.Name,
		"CI_REPO_OWNER":          m.Repo.Owner,
		"CI_REPO_REMOTE_ID":      m.Repo.RemoteID,
		"CI_REPO_SCM":            "git",
		"CI_REPO_URL":            m.Repo.Link,
		"CI_REPO_CLONE_URL":      m.Repo.CloneURL,
		"CI_REPO_CLONE_SSH_URL":  m.Repo.CloneSSHURL,
		"CI_REPO_DEFAULT_BRANCH": m.Repo.Branch,
		"CI_REPO_PRIVATE":        strconv.FormatBool(m.Repo.Private),
		"CI_REPO_TRUSTED":        strconv.FormatBool(m.Repo.Trusted),

		"CI_COMMIT_SHA":                 m.Curr.Commit.Sha,
		"CI_COMMIT_REF":                 m.Curr.Commit.Ref,
		"CI_COMMIT_REFSPEC":             m.Curr.Commit.Refspec,
		"CI_COMMIT_BRANCH":              m.Curr.Commit.Branch,
		"CI_COMMIT_SOURCE_BRANCH":       sourceBranch,
		"CI_COMMIT_TARGET_BRANCH":       targetBranch,
		"CI_COMMIT_URL":                 m.Curr.Link,
		"CI_COMMIT_MESSAGE":             m.Curr.Commit.Message,
		"CI_COMMIT_AUTHOR":              m.Curr.Commit.Author.Name,
		"CI_COMMIT_AUTHOR_EMAIL":        m.Curr.Commit.Author.Email,
		"CI_COMMIT_AUTHOR_AVATAR":       m.Curr.Commit.Author.Avatar,
		"CI_COMMIT_TAG":                 "", // will be set if event is tag
		"CI_COMMIT_PULL_REQUEST":        "", // will be set if event is pr
		"CI_COMMIT_PULL_REQUEST_LABELS": "", // will be set if event is pr

		"CI_PIPELINE_NUMBER":        strconv.FormatInt(m.Curr.Number, 10),
		"CI_PIPELINE_PARENT":        strconv.FormatInt(m.Curr.Parent, 10),
		"CI_PIPELINE_EVENT":         m.Curr.Event,
		"CI_PIPELINE_URL":           m.getPipelineStatusLink(m.Curr, 0),
		"CI_PIPELINE_FORGE_URL":     m.Curr.Link,
		"CI_PIPELINE_DEPLOY_TARGET": m.Curr.Target,
		"CI_PIPELINE_STATUS":        m.Curr.Status,
		"CI_PIPELINE_CREATED":       strconv.FormatInt(m.Curr.Created, 10),
		"CI_PIPELINE_STARTED":       strconv.FormatInt(m.Curr.Started, 10),
		"CI_PIPELINE_FINISHED":      strconv.FormatInt(m.Curr.Finished, 10),

		"CI_WORKFLOW_NAME":   m.Workflow.Name,
		"CI_WORKFLOW_NUMBER": strconv.Itoa(m.Workflow.Number),

		"CI_STEP_NAME":     m.Step.Name,
		"CI_STEP_NUMBER":   strconv.Itoa(m.Step.Number),
		"CI_STEP_STATUS":   "", // will be set by agent
		"CI_STEP_STARTED":  "", // will be set by agent
		"CI_STEP_FINISHED": "", // will be set by agent
		"CI_STEP_URL":      m.getPipelineStatusLink(m.Curr, m.Step.Number),

		"CI_PREV_COMMIT_SHA":           m.Prev.Commit.Sha,
		"CI_PREV_COMMIT_REF":           m.Prev.Commit.Ref,
		"CI_PREV_COMMIT_REFSPEC":       m.Prev.Commit.Refspec,
		"CI_PREV_COMMIT_BRANCH":        m.Prev.Commit.Branch,
		"CI_PREV_COMMIT_URL":           m.Prev.Link,
		"CI_PREV_COMMIT_MESSAGE":       m.Prev.Commit.Message,
		"CI_PREV_COMMIT_AUTHOR":        m.Prev.Commit.Author.Name,
		"CI_PREV_COMMIT_AUTHOR_EMAIL":  m.Prev.Commit.Author.Email,
		"CI_PREV_COMMIT_AUTHOR_AVATAR": m.Prev.Commit.Author.Avatar,

		"CI_PREV_PIPELINE_NUMBER":        strconv.FormatInt(m.Prev.Number, 10),
		"CI_PREV_PIPELINE_PARENT":        strconv.FormatInt(m.Prev.Parent, 10),
		"CI_PREV_PIPELINE_EVENT":         m.Prev.Event,
		"CI_PREV_PIPELINE_URL":           m.getPipelineStatusLink(m.Prev, 0),
		"CI_PREV_PIPELINE_FORGE_URL":     m.Curr.Link,
		"CI_PREV_PIPELINE_DEPLOY_TARGET": m.Prev.Target,
		"CI_PREV_PIPELINE_STATUS":        m.Prev.Status,
		"CI_PREV_PIPELINE_CREATED":       strconv.FormatInt(m.Prev.Created, 10),
		"CI_PREV_PIPELINE_STARTED":       strconv.FormatInt(m.Prev.Started, 10),
		"CI_PREV_PIPELINE_FINISHED":      strconv.FormatInt(m.Prev.Finished, 10),

		"CI_SYSTEM_NAME":     m.Sys.Name,
		"CI_SYSTEM_URL":      m.Sys.Link,
		"CI_SYSTEM_HOST":     m.Sys.Host,
		"CI_SYSTEM_PLATFORM": m.Sys.Platform, // will be set by pipeline platform option or by agent
		"CI_SYSTEM_VERSION":  m.Sys.Version,

		"CI_FORGE_TYPE": m.Forge.Type,
		"CI_FORGE_URL":  m.Forge.URL,
	}
	if m.Curr.Event == EventTag {
		params["CI_COMMIT_TAG"] = strings.TrimPrefix(m.Curr.Commit.Ref, "refs/tags/")
	}
	if m.Curr.Event == EventPull {
		params["CI_COMMIT_PULL_REQUEST"] = pullRegexp.FindString(m.Curr.Commit.Ref)
		params["CI_COMMIT_PULL_REQUEST_LABELS"] = strings.Join(m.Curr.Commit.PullRequestLabels, ",")
	}

	return params
}

func (m *Metadata) getPipelineStatusLink(pipeline Pipeline, stepNumber int) string {
	if stepNumber == 0 {
		return fmt.Sprintf("%s/repos/%d/pipeline/%d", m.Sys.Link, m.Repo.ID, pipeline.Number)
	}

	return fmt.Sprintf("%s/repos/%d/pipeline/%d/%d", m.Sys.Link, m.Repo.ID, pipeline.Number, stepNumber)
}
