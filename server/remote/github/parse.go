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
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/google/go-github/v39/github"
	"github.com/rs/zerolog/log"

	"github.com/woodpecker-ci/woodpecker/server/model"
	"github.com/woodpecker-ci/woodpecker/shared/utils"
)

const (
	hookField = "payload"

	actionOpen = "opened"
	actionSync = "synchronize"

	stateOpen = "open"
)

// parseHook parses a GitHub hook from an http.Request request and returns
// Repo and Build detail. If a hook type is unsupported nil values are returned.
func parseHook(r *http.Request, c *client) (*github.PullRequest, *model.Repo, *model.Build, error) {
	var reader io.Reader = r.Body

	if payload := r.FormValue(hookField); payload != "" {
		reader = bytes.NewBufferString(payload)
	}

	raw, err := ioutil.ReadAll(reader)
	if err != nil {
		return nil, nil, nil, err
	}

	var webhookType string = github.WebHookType(r)
	payload, err := github.ParseWebHook(webhookType, raw)
	if err != nil {
		return nil, nil, nil, err
	}

	switch hook := payload.(type) {
	case *github.PushEvent:
		repo, build, err := parsePushHook(hook)
		return nil, repo, build, err
	case *github.DeploymentEvent:
		repo, build, err := parseDeployHook(hook)
		return nil, repo, build, err
	case *github.ReleaseEvent:
		hasAction := false
		println(*hook.Action)
		for _, actionType := range c.ReleaseActions {
			if actionType == *hook.Action {
				hasAction = true
				break
			}
		}
		if !hasAction {
			log.Debug().
				Str("Hook", webhookType).
				Str("Action", *hook.Action).
				Msg(
					"Github release action ignored. " +
						"See WOODPECKER_GITHUB_RELEASE_ACTIONS")
			return nil, nil, nil, nil
		}
		repo, build, err := parseReleaseHook(hook)
		return nil, repo, build, err
	case *github.PullRequestEvent:
		return parsePullHook(hook, c.MergeRef)
	default:
		log.Debug().
			Str("Hook", webhookType).
			Msg("Github event ignored, and will not be parsed")
		return nil, nil, nil, nil
	}
}

// parsePushHook parses a push hook and returns the Repo and Build details.
// If the commit type is unsupported nil values are returned.
func parsePushHook(hook *github.PushEvent) (*model.Repo, *model.Build, error) {
	if hook.Deleted != nil && *hook.Deleted {
		return nil, nil, nil
	}

	build := &model.Build{
		Event:        model.EventPush,
		Commit:       hook.GetHeadCommit().GetID(),
		Ref:          hook.GetRef(),
		Link:         hook.GetHeadCommit().GetURL(),
		Branch:       strings.Replace(hook.GetRef(), "refs/heads/", "", -1),
		Message:      hook.GetHeadCommit().GetMessage(),
		Email:        hook.GetHeadCommit().GetAuthor().GetEmail(),
		Avatar:       hook.GetSender().GetAvatarURL(),
		Author:       hook.GetSender().GetLogin(),
		Remote:       hook.GetRepo().GetCloneURL(),
		Sender:       hook.GetSender().GetLogin(),
		ChangedFiles: getChangedFilesFromCommits(hook.Commits),
	}

	if len(build.Author) == 0 {
		build.Author = hook.GetHeadCommit().GetAuthor().GetLogin()
	}
	// if len(build.Email) == 0 {
	// TODO: default to gravatar?
	// }
	if strings.HasPrefix(build.Ref, "refs/tags/") {
		// just kidding, this is actually a tag event. Why did this come as a push
		// event we'll never know!
		build.Event = model.EventTag
		build.ChangedFiles = nil
		// For tags, if the base_ref (tag's base branch) is set, we're using it
		// as build's branch so that we can filter events base on it
		if strings.HasPrefix(hook.GetBaseRef(), "refs/heads/") {
			build.Branch = strings.Replace(hook.GetBaseRef(), "refs/heads/", "", -1)
		}
	}

	return convertRepoHook(hook.GetRepo()), build, nil
}

// parseDeployHook parses a deployment and returns the Repo and Build details.
// If the commit type is unsupported nil values are returned.
func parseDeployHook(hook *github.DeploymentEvent) (*model.Repo, *model.Build, error) {
	build := &model.Build{
		Event:   model.EventDeploy,
		Commit:  hook.GetDeployment().GetSHA(),
		Link:    hook.GetDeployment().GetURL(),
		Message: hook.GetDeployment().GetDescription(),
		Ref:     hook.GetDeployment().GetRef(),
		Branch:  hook.GetDeployment().GetRef(),
		Deploy:  hook.GetDeployment().GetEnvironment(),
		Avatar:  hook.GetSender().GetAvatarURL(),
		Author:  hook.GetSender().GetLogin(),
		Sender:  hook.GetSender().GetLogin(),
	}
	// if the ref is a sha or short sha we need to manually construct the ref.
	if strings.HasPrefix(build.Commit, build.Ref) || build.Commit == build.Ref {
		build.Branch = hook.GetRepo().GetDefaultBranch()
		if build.Branch == "" {
			build.Branch = defaultBranch
		}
		build.Ref = fmt.Sprintf("refs/heads/%s", build.Branch)
	}
	// if the ref is a branch we should make sure it has refs/heads prefix
	if !strings.HasPrefix(build.Ref, "refs/") { // branch or tag
		build.Ref = fmt.Sprintf("refs/heads/%s", build.Branch)
	}

	return convertRepo(hook.GetRepo()), build, nil
}

// parseDeployHook parses a deployment and returns the Repo and Build details.
// If the commit type is unsupported nil values are returned.
func parseReleaseHook(hook *github.ReleaseEvent) (*model.Repo, *model.Build, error) {
	release := hook.GetRelease()

	build := &model.Build{

		Event: model.EventRelease,
		// Commit: "",
		// Cannot retrieve the commit since
		// it is hidden in the tag. It seems that github dose
		// not provide the commit SHA with a release.
		Created: release.CreatedAt.UTC().Unix(),
		Link:    release.GetURL(),
		Message: "Release (" + hook.GetAction() + "):" + release.GetName(), // Use the body of the release. There is no message.
		Title:   release.GetName(),
		// Tag name here is the ref. We should add the refs/tags so
		// it is known its a tag (git-plugin looks for it)
		Ref: "refs/tags/" + release.GetTagName(),
		// Branch:  *release.TagName, Dose not exist here. Github releases
		// is always a tag

		Avatar: hook.GetSender().GetAvatarURL(),
		Author: hook.GetSender().GetLogin(),
		Sender: hook.GetSender().GetLogin(),
		Remote: hook.GetRepo().GetCloneURL(),
	}

	return convertRepo(hook.GetRepo()), build, nil
}

// parsePullHook parses a pull request hook and returns the Repo and Build
// details. If the pull request is closed nil values are returned.
func parsePullHook(hook *github.PullRequestEvent, merge bool) (*github.PullRequest, *model.Repo, *model.Build, error) {
	// only listen to new merge-requests and pushes to open ones
	if hook.GetAction() != actionOpen && hook.GetAction() != actionSync {
		return nil, nil, nil, nil
	}
	if hook.GetPullRequest().GetState() != stateOpen {
		return nil, nil, nil, nil
	}

	build := &model.Build{
		Event:   model.EventPull,
		Commit:  hook.GetPullRequest().GetHead().GetSHA(),
		Link:    hook.GetPullRequest().GetHTMLURL(),
		Ref:     fmt.Sprintf(headRefs, hook.GetPullRequest().GetNumber()),
		Branch:  hook.GetPullRequest().GetBase().GetRef(),
		Message: hook.GetPullRequest().GetTitle(),
		Author:  hook.GetPullRequest().GetUser().GetLogin(),
		Avatar:  hook.GetPullRequest().GetUser().GetAvatarURL(),
		Title:   hook.GetPullRequest().GetTitle(),
		Sender:  hook.GetSender().GetLogin(),
		Remote:  hook.GetPullRequest().GetHead().GetRepo().GetCloneURL(),
		Refspec: fmt.Sprintf(refSpec,
			hook.GetPullRequest().GetHead().GetRef(),
			hook.GetPullRequest().GetBase().GetRef(),
		),
	}
	if merge {
		build.Ref = fmt.Sprintf(mergeRefs, hook.GetPullRequest().GetNumber())
	}

	return hook.GetPullRequest(), convertRepo(hook.GetRepo()), build, nil
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
