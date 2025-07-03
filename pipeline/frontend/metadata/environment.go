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
	"encoding/json"
	"fmt"
	"path"
	"regexp"
	"strconv"
	"strings"

	"github.com/rs/zerolog/log"
)

const (
	initialEnvMapSize = 100
	maxChangedFiles   = 500
)

var pullRegexp = regexp.MustCompile(`\d+`)

// Environ returns the metadata as a map of environment variables.
func (m *Metadata) Environ() map[string]string {
	params := make(map[string]string, initialEnvMapSize)

	system := m.Sys
	setNonEmptyEnvVar(params, "CI", system.Name)
	setNonEmptyEnvVar(params, "CI_SYSTEM_NAME", system.Name)
	setNonEmptyEnvVar(params, "CI_SYSTEM_URL", system.URL)
	setNonEmptyEnvVar(params, "CI_SYSTEM_HOST", system.Host)
	setNonEmptyEnvVar(params, "CI_SYSTEM_PLATFORM", system.Platform) // will be set by pipeline platform option or by agent
	setNonEmptyEnvVar(params, "CI_SYSTEM_VERSION", system.Version)

	forge := m.Forge
	setNonEmptyEnvVar(params, "CI_FORGE_TYPE", forge.Type)
	setNonEmptyEnvVar(params, "CI_FORGE_URL", forge.URL)

	repo := m.Repo
	setNonEmptyEnvVar(params, "CI_REPO", path.Join(repo.Owner, repo.Name))
	setNonEmptyEnvVar(params, "CI_REPO_NAME", repo.Name)
	setNonEmptyEnvVar(params, "CI_REPO_OWNER", repo.Owner)
	setNonEmptyEnvVar(params, "CI_REPO_REMOTE_ID", repo.RemoteID)
	setNonEmptyEnvVar(params, "CI_REPO_URL", repo.ForgeURL)
	setNonEmptyEnvVar(params, "CI_REPO_CLONE_URL", repo.CloneURL)
	setNonEmptyEnvVar(params, "CI_REPO_CLONE_SSH_URL", repo.CloneSSHURL)
	setNonEmptyEnvVar(params, "CI_REPO_DEFAULT_BRANCH", repo.Branch)
	setNonEmptyEnvVar(params, "CI_REPO_PRIVATE", strconv.FormatBool(repo.Private))

	pipeline := m.Curr
	setNonEmptyEnvVar(params, "CI_PIPELINE_NUMBER", strconv.FormatInt(pipeline.Number, 10))
	setNonEmptyEnvVar(params, "CI_PIPELINE_PARENT", strconv.FormatInt(pipeline.Parent, 10))
	setNonEmptyEnvVar(params, "CI_PIPELINE_EVENT", pipeline.Event)
	setNonEmptyEnvVar(params, "CI_PIPELINE_URL", m.getPipelineWebURL(pipeline, 0))
	setNonEmptyEnvVar(params, "CI_PIPELINE_FORGE_URL", pipeline.ForgeURL)
	setNonEmptyEnvVar(params, "CI_PIPELINE_DEPLOY_TARGET", pipeline.DeployTo)
	setNonEmptyEnvVar(params, "CI_PIPELINE_DEPLOY_TASK", pipeline.DeployTask)
	setNonEmptyEnvVar(params, "CI_PIPELINE_CREATED", strconv.FormatInt(pipeline.Created, 10))
	setNonEmptyEnvVar(params, "CI_PIPELINE_STARTED", strconv.FormatInt(pipeline.Started, 10))
	setNonEmptyEnvVar(params, "CI_PIPELINE_AUTHOR", pipeline.Author)
	setNonEmptyEnvVar(params, "CI_PIPELINE_AVATAR", pipeline.Avatar)

	workflow := m.Workflow
	setNonEmptyEnvVar(params, "CI_WORKFLOW_NAME", workflow.Name)
	setNonEmptyEnvVar(params, "CI_WORKFLOW_NUMBER", strconv.Itoa(workflow.Number))

	step := m.Step
	setNonEmptyEnvVar(params, "CI_STEP_NAME", step.Name)
	setNonEmptyEnvVar(params, "CI_STEP_NUMBER", strconv.Itoa(step.Number))
	setNonEmptyEnvVar(params, "CI_STEP_URL", m.getPipelineWebURL(pipeline, step.Number))
	// CI_STEP_STARTED will be set by agent

	commit := pipeline.Commit
	setNonEmptyEnvVar(params, "CI_COMMIT_SHA", commit.Sha)
	setNonEmptyEnvVar(params, "CI_COMMIT_REF", commit.Ref)
	setNonEmptyEnvVar(params, "CI_COMMIT_REFSPEC", commit.Refspec)
	setNonEmptyEnvVar(params, "CI_COMMIT_MESSAGE", commit.Message)
	setNonEmptyEnvVar(params, "CI_COMMIT_BRANCH", commit.Branch)
	setNonEmptyEnvVar(params, "CI_COMMIT_AUTHOR", commit.Author.Name)
	setNonEmptyEnvVar(params, "CI_COMMIT_AUTHOR_EMAIL", commit.Author.Email)
	setNonEmptyEnvVar(params, "CI_COMMIT_AUTHOR_AVATAR", commit.Author.Avatar)
	if pipeline.Event == EventTag || pipeline.Event == EventRelease || strings.HasPrefix(pipeline.Commit.Ref, "refs/tags/") {
		setNonEmptyEnvVar(params, "CI_COMMIT_TAG", strings.TrimPrefix(pipeline.Commit.Ref, "refs/tags/"))
	}
	if pipeline.Event == EventRelease {
		setNonEmptyEnvVar(params, "CI_COMMIT_PRERELEASE", strconv.FormatBool(pipeline.Commit.IsPrerelease))
	}
	if pipeline.Event == EventPull || pipeline.Event == EventPullClosed {
		sourceBranch, targetBranch := getSourceTargetBranches(commit.Refspec)
		setNonEmptyEnvVar(params, "CI_COMMIT_SOURCE_BRANCH", sourceBranch)
		setNonEmptyEnvVar(params, "CI_COMMIT_TARGET_BRANCH", targetBranch)
		setNonEmptyEnvVar(params, "CI_COMMIT_PULL_REQUEST", pullRegexp.FindString(pipeline.Commit.Ref))
		setNonEmptyEnvVar(params, "CI_COMMIT_PULL_REQUEST_LABELS", strings.Join(pipeline.Commit.PullRequestLabels, ","))
	}

	// Only export changed files if maxChangedFiles is not exceeded
	changedFiles := commit.ChangedFiles
	if len(changedFiles) == 0 {
		params["CI_PIPELINE_FILES"] = "[]"
	} else if len(changedFiles) <= maxChangedFiles {
		// we have to use json, as other separators like ;, or space are valid filename chars
		changedFiles, err := json.Marshal(changedFiles)
		if err != nil {
			log.Error().Err(err).Msg("marshal changed files")
		}
		params["CI_PIPELINE_FILES"] = string(changedFiles)
	}

	prevPipeline := m.Prev
	setNonEmptyEnvVar(params, "CI_PREV_PIPELINE_NUMBER", strconv.FormatInt(prevPipeline.Number, 10))
	setNonEmptyEnvVar(params, "CI_PREV_PIPELINE_PARENT", strconv.FormatInt(prevPipeline.Parent, 10))
	setNonEmptyEnvVar(params, "CI_PREV_PIPELINE_EVENT", prevPipeline.Event)
	setNonEmptyEnvVar(params, "CI_PREV_PIPELINE_URL", m.getPipelineWebURL(prevPipeline, 0))
	setNonEmptyEnvVar(params, "CI_PREV_PIPELINE_FORGE_URL", prevPipeline.ForgeURL)
	setNonEmptyEnvVar(params, "CI_PREV_COMMIT_URL", prevPipeline.ForgeURL) // why commit url?
	setNonEmptyEnvVar(params, "CI_PREV_PIPELINE_DEPLOY_TARGET", prevPipeline.DeployTo)
	setNonEmptyEnvVar(params, "CI_PREV_PIPELINE_DEPLOY_TASK", prevPipeline.DeployTask)
	setNonEmptyEnvVar(params, "CI_PREV_PIPELINE_STATUS", prevPipeline.Status)
	setNonEmptyEnvVar(params, "CI_PREV_PIPELINE_CREATED", strconv.FormatInt(prevPipeline.Created, 10))
	setNonEmptyEnvVar(params, "CI_PREV_PIPELINE_STARTED", strconv.FormatInt(prevPipeline.Started, 10))
	setNonEmptyEnvVar(params, "CI_PREV_PIPELINE_FINISHED", strconv.FormatInt(prevPipeline.Finished, 10))
	setNonEmptyEnvVar(params, "CI_PREV_PIPELINE_AUTHOR", prevPipeline.Author)
	setNonEmptyEnvVar(params, "CI_PREV_PIPELINE_AVATAR", prevPipeline.Avatar)

	prevCommit := prevPipeline.Commit
	setNonEmptyEnvVar(params, "CI_PREV_COMMIT_SHA", prevCommit.Sha)
	setNonEmptyEnvVar(params, "CI_PREV_COMMIT_REF", prevCommit.Ref)
	setNonEmptyEnvVar(params, "CI_PREV_COMMIT_REFSPEC", prevCommit.Refspec)
	setNonEmptyEnvVar(params, "CI_PREV_COMMIT_MESSAGE", prevCommit.Message)
	setNonEmptyEnvVar(params, "CI_PREV_COMMIT_BRANCH", prevCommit.Branch)
	setNonEmptyEnvVar(params, "CI_PREV_COMMIT_AUTHOR", prevCommit.Author.Name)
	setNonEmptyEnvVar(params, "CI_PREV_COMMIT_AUTHOR_EMAIL", prevCommit.Author.Email)
	setNonEmptyEnvVar(params, "CI_PREV_COMMIT_AUTHOR_AVATAR", prevCommit.Author.Avatar)
	if prevPipeline.Event == EventPull || prevPipeline.Event == EventPullClosed {
		prevSourceBranch, prevTargetBranch := getSourceTargetBranches(prevCommit.Refspec)
		setNonEmptyEnvVar(params, "CI_PREV_COMMIT_SOURCE_BRANCH", prevSourceBranch)
		setNonEmptyEnvVar(params, "CI_PREV_COMMIT_TARGET_BRANCH", prevTargetBranch)
	}

	return params
}

func (m *Metadata) getPipelineWebURL(pipeline Pipeline, stepNumber int) string {
	if stepNumber == 0 {
		return fmt.Sprintf("%s/repos/%d/pipeline/%d", m.Sys.URL, m.Repo.ID, pipeline.Number)
	}

	return fmt.Sprintf("%s/repos/%d/pipeline/%d/%d", m.Sys.URL, m.Repo.ID, pipeline.Number, stepNumber)
}

func getSourceTargetBranches(refspec string) (string, string) {
	var (
		sourceBranch string
		targetBranch string
	)

	branchParts := strings.Split(refspec, ":")
	if len(branchParts) == 2 { //nolint:mnd
		sourceBranch = branchParts[0]
		targetBranch = branchParts[1]
	}

	return sourceBranch, targetBranch
}

func setNonEmptyEnvVar(env map[string]string, key, value string) {
	if len(value) > 0 {
		env[key] = value
	} else {
		log.Trace().Str("variable", key).Msg("env var is filtered as it's empty")
	}
}
