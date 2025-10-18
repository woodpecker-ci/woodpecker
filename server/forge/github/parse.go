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

	"github.com/google/go-github/v76/github"

	"go.woodpecker-ci.org/woodpecker/v3/server/forge/common"
	"go.woodpecker-ci.org/woodpecker/v3/server/forge/types"
	"go.woodpecker-ci.org/woodpecker/v3/server/model"
)

const (
	hookField = "payload"

	actionOpen             = "opened"
	actionReopen           = "reopened"
	actionClose            = "closed"
	actionSync             = "synchronize"
	actionReleased         = "released"
	actionAssigned         = "assigned"
	actionConvertedToDraft = "converted_to_draft"
	actionDemilestoned     = "demilestoned"
	actionEdited           = "edited"
	actionLabeled          = "labeled"
	actionLocked           = "locked"
	actionMilestoned       = "milestoned"
	actionReadyForReview   = "ready_for_review"
	actionUnassigned       = "unassigned"
	actionUnlabeled        = "unlabeled"
	actionUnlocked         = "unlocked"

	labelCleared = "label_cleared"
	labelUpdated = "label_updated"
)

// parseHook parses a GitHub hook from an http.Request request and returns
// Repo and Pipeline detail. If a hook type is unsupported nil values are returned.
func parseHook(r *http.Request, merge bool) (_ *github.PullRequest, _ *model.Repo, _ *model.Pipeline, currCommit, prevCommit string, _ error) {
	var reader io.Reader = r.Body

	if payload := r.FormValue(hookField); payload != "" {
		reader = bytes.NewBufferString(payload)
	}

	raw, err := io.ReadAll(reader)
	if err != nil {
		return nil, nil, nil, "", "", err
	}

	payload, err := github.ParseWebHook(github.WebHookType(r), raw)
	if err != nil {
		return nil, nil, nil, "", "", err
	}

	switch hook := payload.(type) {
	case *github.PushEvent:
		repo, pipeline, curr, prev := parsePushHook(hook)
		return nil, repo, pipeline, curr, prev, nil
	case *github.DeploymentEvent:
		repo, pipeline := parseDeployHook(hook)
		return nil, repo, pipeline, "", "", nil
	case *github.PullRequestEvent:
		pr, repo, pipeline, err := parsePullHook(hook, merge)
		return pr, repo, pipeline, "", "", err
	case *github.ReleaseEvent:
		repo, pipeline := parseReleaseHook(hook)
		return nil, repo, pipeline, "", "", nil
	default:
		return nil, nil, nil, "", "", &types.ErrIgnoreEvent{Event: github.Stringify(hook)}
	}
}

// parsePushHook parses a push hook and returns the Repo and Pipeline details.
// If the commit type is unsupported nil values are returned.
func parsePushHook(hook *github.PushEvent) (_ *model.Repo, _ *model.Pipeline, curr, prev string) {
	if hook.Deleted != nil && *hook.Deleted {
		return nil, nil, "", ""
	}

	pipeline := &model.Pipeline{
		Event: model.EventPush,
		Commit: &model.Commit{
			SHA:      hook.GetHeadCommit().GetID(),
			Author:   convertCommitAuthor(hook.GetHeadCommit().GetAuthor()),
			Message:  hook.GetHeadCommit().GetMessage(),
			ForgeURL: hook.GetHeadCommit().GetURL(),
		},
		Ref:          hook.GetRef(),
		ForgeURL:     hook.GetHeadCommit().GetURL(),
		Branch:       strings.ReplaceAll(hook.GetRef(), "refs/heads/", ""),
		AuthorAvatar: hook.GetSender().GetAvatarURL(),
		Author:       hook.GetSender().GetLogin(),
	}
	repo := convertRepoHook(hook.GetRepo())

	if strings.HasPrefix(pipeline.Ref, "refs/tags/") {
		// just kidding, this is actually a tag event. Why did this come as a push
		// event we'll never know!
		pipeline.Event = model.EventTag
		// For tags, if the base_ref (tag's base branch) is set, we're using it
		// as pipeline's branch so that we can filter events base on it
		if strings.HasPrefix(hook.GetBaseRef(), "refs/heads/") {
			pipeline.Branch = strings.ReplaceAll(hook.GetBaseRef(), "refs/heads/", "")
		}
		return repo, pipeline, "", ""
	}

	return repo, pipeline, hook.GetHeadCommit().GetID(), hook.GetBefore()
}

// parseDeployHook parses a deployment and returns the Repo and Pipeline details.
// If the commit type is unsupported nil values are returned.
func parseDeployHook(hook *github.DeploymentEvent) (*model.Repo, *model.Pipeline) {
	pipeline := &model.Pipeline{
		Event: model.EventDeploy,
		Commit: &model.Commit{
			SHA: hook.GetDeployment().GetSHA(),
		},
		ForgeURL:     hook.GetDeployment().GetURL(),
		Ref:          hook.GetDeployment().GetRef(),
		Branch:       hook.GetDeployment().GetRef(),
		AuthorAvatar: hook.GetSender().GetAvatarURL(),
		Author:       hook.GetSender().GetLogin(),
		Deployment: &model.Deployment{
			Target:      hook.GetDeployment().GetEnvironment(),
			Task:        hook.GetDeployment().GetTask(),
			Description: hook.GetDeployment().GetDescription(),
		},
	}
	// if the ref is a sha or short sha we need to manually construct the ref.
	if strings.HasPrefix(pipeline.Commit.SHA, pipeline.Ref) || pipeline.Commit.SHA == pipeline.Ref {
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
	event := model.EventPull
	eventAction := ""

	switch hook.GetAction() {
	case actionOpen, actionReopen, actionSync:
		// default case nothing to do
	case actionClose:
		event = model.EventPullClosed
	case actionAssigned,
		actionConvertedToDraft,
		actionDemilestoned,
		actionEdited,
		actionLabeled,
		actionLocked,
		actionMilestoned,
		actionReadyForReview,
		actionUnassigned,
		actionUnlabeled,
		actionUnlocked:
		// metadata pull events
		event = model.EventPullMetadata
		eventAction = common.NormalizeEventReason(hook.GetAction())
	default:
		return nil, nil, nil, &types.ErrIgnoreEvent{
			Event:  string(model.EventPullMetadata),
			Reason: fmt.Sprintf("action %s is not supported", hook.GetAction()),
		}
	}

	pipeline := &model.Pipeline{
		Event:       event,
		EventReason: []string{eventAction},
		Commit: &model.Commit{
			SHA: hook.GetPullRequest().GetHead().GetSHA(),
		},
		ForgeURL:     hook.GetPullRequest().GetHTMLURL(),
		Ref:          fmt.Sprintf(headRefs, hook.GetPullRequest().GetNumber()),
		Branch:       hook.GetPullRequest().GetBase().GetRef(),
		AuthorAvatar: hook.GetSender().GetAvatarURL(),
		Author:       hook.GetSender().GetLogin(),
		Refspec: fmt.Sprintf(refSpec,
			hook.GetPullRequest().GetHead().GetRef(),
			hook.GetPullRequest().GetBase().GetRef(),
		),
		PullRequest: convertPullRequest(hook.GetPullRequest()),
	}
	if merge {
		pipeline.Ref = fmt.Sprintf(mergeRefs, hook.GetPullRequest().GetNumber())
	}

	// normalize label events to match other forges
	if eventAction == actionLabeled || eventAction == actionUnlabeled {
		if len(pipeline.PullRequest.Labels) == 0 {
			pipeline.EventReason = []string{labelCleared}
		} else {
			pipeline.EventReason = []string{labelUpdated}
		}
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
		Event:    model.EventRelease,
		ForgeURL: hook.GetRelease().GetHTMLURL(),
		Ref:      fmt.Sprintf("refs/tags/%s", hook.GetRelease().GetTagName()),
		Branch:   hook.GetRelease().GetTargetCommitish(), // cspell:disable-line
		Release: &model.Release{
			Title:        name,
			IsPrerelease: hook.GetRelease().GetPrerelease(),
		},
		TagTitle:     hook.GetRelease().GetTagName(),
		AuthorAvatar: hook.GetSender().GetAvatarURL(),
		Author:       hook.GetSender().GetLogin(),
	}

	return convertRepo(hook.GetRepo()), pipeline
}

func convertCommitAuthor(u *github.CommitAuthor) model.CommitAuthor {
	return model.CommitAuthor{
		Name:  u.GetName(),
		Email: u.GetEmail(),
	}
}
