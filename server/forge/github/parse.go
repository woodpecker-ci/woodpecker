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

	"github.com/google/go-github/v56/github"

	"go.woodpecker-ci.org/woodpecker/server/forge/types"
	"go.woodpecker-ci.org/woodpecker/server/model"
	"go.woodpecker-ci.org/woodpecker/shared/utils"
)

const (
	hookField = "payload"

	actionOpen = "opened"
	actionSync = "synchronize"

	stateOpen = "open"
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
		repo, pipeline, err := parsePushHook(hook)
		return nil, repo, pipeline, err
	case *github.DeploymentEvent:
		repo, pipeline, err := parseDeployHook(hook)
		return nil, repo, pipeline, err
	case *github.PullRequestEvent:
		return parsePullHook(hook, merge)
	default:
		return nil, nil, nil, &types.ErrIgnoreEvent{Event: github.Stringify(hook)}
	}
}

// parsePushHook parses a push hook and returns the Repo and Pipeline details.
// If the commit type is unsupported nil values are returned.
func parsePushHook(hook *github.PushEvent) (*model.Repo, *model.Pipeline, error) {
	if hook.Deleted != nil && *hook.Deleted {
		return nil, nil, nil
	}

	pipeline := &model.Pipeline{
		Event:        model.EventPush,
		Commit:       hook.GetHeadCommit().GetID(),
		Ref:          hook.GetRef(),
		ForgeURL:     hook.GetHeadCommit().GetURL(),
		Branch:       strings.Replace(hook.GetRef(), "refs/heads/", "", -1),
		Message:      hook.GetHeadCommit().GetMessage(),
		Email:        hook.GetHeadCommit().GetAuthor().GetEmail(),
		Avatar:       hook.GetSender().GetAvatarURL(),
		Author:       hook.GetSender().GetLogin(),
		CloneURL:     hook.GetRepo().GetCloneURL(),
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
			pipeline.Branch = strings.Replace(hook.GetBaseRef(), "refs/heads/", "", -1)
		}
	}

	return convertRepoHook(hook.GetRepo()), pipeline, nil
}

// parseDeployHook parses a deployment and returns the Repo and Pipeline details.
// If the commit type is unsupported nil values are returned.
func parseDeployHook(hook *github.DeploymentEvent) (*model.Repo, *model.Pipeline, error) {
	pipeline := &model.Pipeline{
		Event:    model.EventDeploy,
		Commit:   hook.GetDeployment().GetSHA(),
		ForgeURL: hook.GetDeployment().GetURL(),
		Message:  hook.GetDeployment().GetDescription(),
		Ref:      hook.GetDeployment().GetRef(),
		Branch:   hook.GetDeployment().GetRef(),
		Deploy:   hook.GetDeployment().GetEnvironment(),
		Avatar:   hook.GetSender().GetAvatarURL(),
		Author:   hook.GetSender().GetLogin(),
		Sender:   hook.GetSender().GetLogin(),
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

	return convertRepo(hook.GetRepo()), pipeline, nil
}

// parsePullHook parses a pull request hook and returns the Repo and Pipeline
// details. If the pull request is closed nil values are returned.
func parsePullHook(hook *github.PullRequestEvent, merge bool) (*github.PullRequest, *model.Repo, *model.Pipeline, error) {
	// only listen to new merge-requests and pushes to open ones
	if hook.GetAction() != actionOpen && hook.GetAction() != actionSync {
		return nil, nil, nil, nil
	}
	if hook.GetPullRequest().GetState() != stateOpen {
		return nil, nil, nil, nil
	}

	pipeline := &model.Pipeline{
		Event:    model.EventPull,
		Commit:   hook.GetPullRequest().GetHead().GetSHA(),
		ForgeURL: hook.GetPullRequest().GetHTMLURL(),
		Ref:      fmt.Sprintf(headRefs, hook.GetPullRequest().GetNumber()),
		Branch:   hook.GetPullRequest().GetBase().GetRef(),
		Message:  hook.GetPullRequest().GetTitle(),
		Author:   hook.GetPullRequest().GetUser().GetLogin(),
		Avatar:   hook.GetPullRequest().GetUser().GetAvatarURL(),
		Title:    hook.GetPullRequest().GetTitle(),
		Sender:   hook.GetSender().GetLogin(),
		CloneURL: hook.GetPullRequest().GetHead().GetRepo().GetCloneURL(),
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

func getChangedFilesFromCommits(commits []*github.HeadCommit) []string {
	// assume a capacity of 4 changed files per commit
	files := make([]string, 0, len(commits)*4)
	for _, cm := range commits {
		files = append(files, cm.Added...)
		files = append(files, cm.Removed...)
		files = append(files, cm.Modified...)
	}
	return utils.DedupStrings(files)
}
