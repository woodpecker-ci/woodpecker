// Copyright 2022 Woodpecker Authors
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

package frontend

import (
	"strconv"
	"strings"
)

// setDroneEnviron set dedicated to DroneCI environment vars as compatibility layer
func (m *Metadata) setDroneEnviron(env map[string]string,
	scmType, repoOwner, repoName, sourceBranch, targetBranch string,
) {
	// webhook
	env["DRONE_BRANCH"] = m.Curr.Commit.Branch
	env["DRONE_PULL_REQUEST"] = ""
	if m.Curr.Event == EventPull {
		env["DRONE_PULL_REQUEST"] = pullRegexp.FindString(m.Curr.Commit.Ref)
	}
	env["DRONE_TAG"] = ""
	if m.Curr.Event == EventTag {
		env["DRONE_TAG"] = strings.TrimPrefix(m.Curr.Commit.Ref, "refs/tags/")
	}
	env["DRONE_SOURCE_BRANCH"] = sourceBranch
	env["DRONE_TARGET_BRANCH"] = targetBranch
	// pipeline
	env["DRONE_BUILD_NUMBER"] = strconv.FormatInt(m.Curr.Number, 10)
	env["DRONE_BUILD_PARENT"] = strconv.FormatInt(m.Curr.Parent, 10)
	env["DRONE_BUILD_EVENT"] = m.Curr.Event
	env["DRONE_BUILD_STATUS"] = m.Curr.Status
	env["DRONE_BUILD_LINK"] = m.Curr.Link
	env["DRONE_BUILD_CREATED"] = strconv.FormatInt(m.Curr.Created, 10)
	env["DRONE_BUILD_STARTED"] = strconv.FormatInt(m.Curr.Started, 10)
	env["DRONE_BUILD_FINISHED"] = strconv.FormatInt(m.Curr.Finished, 10)
	// commit
	env["DRONE_COMMIT"] = m.Curr.Commit.Sha
	env["DRONE_COMMIT_BEFORE"] = m.Prev.Commit.Sha
	env["DRONE_COMMIT_REF"] = m.Curr.Commit.Ref
	env["DRONE_COMMIT_BRANCH"] = m.Curr.Commit.Branch
	env["DRONE_COMMIT_LINK"] = m.Curr.Link
	env["DRONE_COMMIT_MESSAGE"] = m.Curr.Commit.Message
	env["DRONE_COMMIT_AUTHOR"] = m.Curr.Commit.Author.Name
	env["DRONE_COMMIT_AUTHOR_NAME"] = m.Curr.Commit.Author.Name
	env["DRONE_COMMIT_AUTHOR_EMAIL"] = m.Curr.Commit.Author.Email
	env["DRONE_COMMIT_AUTHOR_AVATAR"] = m.Curr.Commit.Author.Avatar
	// repo
	env["DRONE_REPO"] = m.Repo.Name
	env["DRONE_REPO_SCM"] = scmType
	env["DRONE_REPO_OWNER"] = repoOwner
	env["DRONE_REPO_NAME"] = repoName
	env["DRONE_REPO_LINK"] = m.Repo.Link
	env["DRONE_REPO_BRANCH"] = m.Repo.Branch
	env["DRONE_REPO_PRIVATE"] = strconv.FormatBool(m.Repo.Private)
	// clone
	env["DRONE_REMOTE_URL"] = m.Repo.Remote
	if scmType == "git" {
		env["DRONE_GIT_HTTP_URL"] = m.Repo.Remote
	}
	// misc
	env["DRONE_SYSTEM_HOST"] = m.Sys.Host
	env["DRONE_STEP_NUMBER"] = strconv.Itoa(m.Job.Number)
}
