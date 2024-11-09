// Copyright 2022 Woodpecker Authors
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

package github

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/google/go-github/v66/github"

	"go.woodpecker-ci.org/woodpecker/v2/server/forge/types"
	"go.woodpecker-ci.org/woodpecker/v2/server/model"
	"go.woodpecker-ci.org/woodpecker/v2/shared/utils"
)

const (
	hookField = "payload"

	actionOpen     = "opened"
	actionClose    = "closed"
	actionSync     = "synchronize"
	actionReleased = "released"

	stateOpen  = "open"
	stateClose = "closed"
)

// parseHook parses a GitHub hook from an http.Request request and returns
// Repo and Pipeline detail. If a hook type is unsupported nil values are returned.
func parseHook(r *http.Request, merge bool) (*github.PullRequest, *model.Repo, *model.Pipeline, error) {
	var reader io.Reader = r.Body

	if payload := r.FormValue(hookField); payload != "" {
		reader = bytes.NewBufferString(payload)
	}

	raw, err := io.ReadAll(reader)
	if err != nil {
		return nil, nil, nil, err
	}

	payload, err := github.ParseWebHook(github.WebHookType(r), raw)
	if err != nil {
		return nil, nil, nil, err
	}

	switch hook := payload.(type) {
	case *github.PushEvent:
		repo, pipeline := parsePushHook(hook)
		return nil, repo, pipeline, nil
	case *github.DeploymentEvent:
		repo, pipeline := parseDeployHook(hook)
		return nil, repo, pipeline, nil
	case *github.PullRequestEvent:
		return parsePullHook(hook, merge)
	case *github.ReleaseEvent:
		repo, pipeline := parseReleaseHook(hook)
		return nil, repo, pipeline, nil
	default:
		return nil, nil, nil, &types.ErrIgnoreEvent{Event: github.Stringify(hook)}
	}
}

// parsePushHook parses a push hook and returns the Repo and Pipeline details.
// If the commit type is unsupported nil values are returned.
func parsePushHook(hook *github.PushEvent) (*model.Repo, *model.Pipeline) {
	if hook.Deleted != nil && *hook.Deleted {
		return nil, nil
	}

	pipeline := &model.Pipeline{
		Event:        model.EventPush,
		Commit:       hook.GetHeadCommit().GetID(),
		Ref:          hook.GetRef(),
		ForgeURL:     hook.GetHeadCommit().GetURL(),
		Branch:       strings.ReplaceAll(hook.GetRef(), "refs/heads/", ""),
		Message:      hook.GetHeadCommit().GetMessage(),
		Email:        hook.GetHeadCommit().GetAuthor().GetEmail(),
		Avatar:       hook.GetSender().GetAvatarURL(),
		Author:       hook.GetSender().GetLogin(),
		Sender:       hook.GetSender().GetLogin(),
		ChangedFiles: getChangedFilesFromCommits(hook.Commits),
	}

	if len(pipeline.Author) == 0 {
		pipeline.Author = hook.GetHeadCommit().GetAuthor().GetLogin()
	}
	if strings.HasPrefix(pipeline.Ref, "refs/tags/") {
		// just kidding, this is actually a tag event. Why did this come as a push
		// event we'll never know!
		pipeline.Event = model.EventTag
		pipeline.ChangedFiles = nil
		// For tags, if the base_ref (tag's base branch) is set, we're using it
		// as pipeline's branch so that we can filter events base on it
		if strings.HasPrefix(hook.GetBaseRef(), "refs/heads/") {
			pipeline.Branch = strings.ReplaceAll(hook.GetBaseRef(), "refs/heads/", "")
		}
	}

	return convertRepoHook(hook.GetRepo()), pipeline
}

// parseDeployHook parses a deployment and returns the Repo and Pipeline details.
// If the commit type is unsupported nil values are returned.
func parseDeployHook(hook *github.DeploymentEvent) (*model.Repo, *model.Pipeline) {
	pipeline := &model.Pipeline{
		Event:      model.EventDeploy,
		Commit:     hook.GetDeployment().GetSHA(),
		ForgeURL:   hook.GetDeployment().GetURL(),
		Message:    hook.GetDeployment().GetDescription(),
		Ref:        hook.GetDeployment().GetRef(),
		Branch:     hook.GetDeployment().GetRef(),
		Avatar:     hook.GetSender().GetAvatarURL(),
		Author:     hook.GetSender().GetLogin(),
		Sender:     hook.GetSender().GetLogin(),
		DeployTo:   hook.GetDeployment().GetEnvironment(),
		DeployTask: hook.GetDeployment().GetTask(),
	}
	// if the ref is a sha or short sha we need to manually construct the ref.
	if strings.HasPrefix(pipeline.Commit, pipeline.Ref) || pipeline.Commit == pipeline.Ref {
		pipeline.Branch = hook.GetRepo().GetDefaultBranch()
		pipeline.Ref = fmt.Sprintf("refs/heads/%s", pipeline.Branch)
	}
	// if the ref is a branch we should make sure it has refs/heads prefix
	if !strings.HasPrefix(pipeline.Ref, "refs/") { // branch or tag
		pipeline.Ref = fmt.Sprintf("refs/heads/%s", pipeline.Branch)
	}

	return convertRepo(hook.GetRepo()), pipeline
}

// parsePullHook parses a pull request hook and returns the Repo and Pipeline
// details.
func parsePullHook(hook *github.PullRequestEvent, merge bool) (*github.PullRequest, *model.Repo, *model.Pipeline, error) {
	if hook.GetAction() != actionOpen && hook.GetAction() != actionSync && hook.GetAction() != actionClose {
		return nil, nil, nil, nil
	}

	event := model.EventPull
	if hook.GetPullRequest().GetState() == stateClose {
		event = model.EventPullClosed
	}

	pipeline := &model.Pipeline{
		Event:    event,
		Commit:   hook.GetPullRequest().GetHead().GetSHA(),
		ForgeURL: hook.GetPullRequest().GetHTMLURL(),
		Ref:      fmt.Sprintf(headRefs, hook.GetPullRequest().GetNumber()),
		Branch:   hook.GetPullRequest().GetBase().GetRef(),
		Message:  hook.GetPullRequest().GetTitle(),
		Author:   hook.GetPullRequest().GetUser().GetLogin(),
		Avatar:   hook.GetPullRequest().GetUser().GetAvatarURL(),
		Title:    hook.GetPullRequest().GetTitle(),
		Sender:   hook.GetSender().GetLogin(),
		Refspec: fmt.Sprintf(refSpec,
			hook.GetPullRequest().GetHead().GetRef(),
			hook.GetPullRequest().GetBase().GetRef(),
		),
		PullRequestLabels: convertLabels(hook.GetPullRequest().Labels),
	}
	if merge {
		pipeline.Ref = fmt.Sprintf(mergeRefs, hook.GetPullRequest().GetNumber())
	}

	return hook.GetPullRequest(), convertRepo(hook.GetRepo()), pipeline, nil
}

// parseReleaseHook parses a release hook and returns the Repo and Pipeline
// details.
func parseReleaseHook(hook *github.ReleaseEvent) (*model.Repo, *model.Pipeline) {
	if hook.GetAction() != actionReleased {
		return nil, nil
	}

	name := hook.GetRelease().GetName()
	if name == "" {
		name = hook.GetRelease().GetTagName()
	}

	pipeline := &model.Pipeline{
		Event:        model.EventRelease,
		ForgeURL:     hook.GetRelease().GetHTMLURL(),
		Ref:          fmt.Sprintf("refs/tags/%s", hook.GetRelease().GetTagName()),
		Branch:       hook.GetRelease().GetTargetCommitish(), // cspell:disable-line
		Message:      fmt.Sprintf("created release %s", name),
		Author:       hook.GetRelease().GetAuthor().GetLogin(),
		Avatar:       hook.GetRelease().GetAuthor().GetAvatarURL(),
		Sender:       hook.GetSender().GetLogin(),
		IsPrerelease: hook.GetRelease().GetPrerelease(),
	}

	return convertRepo(hook.GetRepo()), pipeline
}

func getChangedFilesFromCommits(commits []*github.HeadCommit) []string {
	// assume a capacity of 4 changed files per commit
	files := make([]string, 0, len(commits)*4)
	for _, cm := range commits {
		files = append(files, cm.Added...)
		files = append(files, cm.Removed...)
		files = append(files, cm.Modified...)
	}
	return utils.DeduplicateStrings(files)
}
