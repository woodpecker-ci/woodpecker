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

package pipeline

// SetDroneEnviron set dedicated to DroneCI environment vars as compatibility
// layer. Main purpose is to be compatible with drone plugins.
func SetDroneEnviron(env map[string]string) {
	// webhook
	env["DRONE_BRANCH"] = env["CI_COMMIT_BRANCH"]
	env["DRONE_PULL_REQUEST"] = env["CI_COMMIT_PULL_REQUEST"]
	env["DRONE_TAG"] = env["CI_COMMIT_TAG"]
	env["DRONE_SOURCE_BRANCH"] = env["CI_COMMIT_SOURCE_BRANCH"]
	env["DRONE_TARGET_BRANCH"] = env["CI_COMMIT_TARGET_BRANCH"]
	// pipeline
	env["DRONE_BUILD_NUMBER"] = env["CI_PIPELINE_NUMBER"]
	env["DRONE_BUILD_PARENT"] = env["CI_PIPELINE_PARENT"]
	env["DRONE_BUILD_EVENT"] = env["CI_PIPELINE_EVENT"]
	env["DRONE_BUILD_STATUS"] = env["CI_PIPELINE_STATUS"]
	env["DRONE_BUILD_LINK"] = env["CI_PIPELINE_LINK"]
	env["DRONE_BUILD_CREATED"] = env["CI_PIPELINE_CREATED"]
	env["DRONE_BUILD_STARTED"] = env["CI_PIPELINE_STARTED"]
	env["DRONE_BUILD_FINISHED"] = env["CI_PIPELINE_FINISHED"]
	// commit
	env["DRONE_COMMIT"] = env["CI_COMMIT_SHA"]
	env["DRONE_COMMIT_SHA"] = env["CI_COMMIT_SHA"]
	env["DRONE_COMMIT_BEFORE"] = env["CI_PREV_COMMIT_SHA"]
	env["DRONE_COMMIT_REF"] = env["CI_COMMIT_REF"]
	env["DRONE_COMMIT_BRANCH"] = env["CI_COMMIT_BRANCH"]
	env["DRONE_COMMIT_LINK"] = env["CI_COMMIT_LINK"]
	env["DRONE_COMMIT_MESSAGE"] = env["CI_COMMIT_MESSAGE"]
	env["DRONE_COMMIT_AUTHOR"] = env["CI_COMMIT_AUTHOR"]
	env["DRONE_COMMIT_AUTHOR_NAME"] = env["CI_COMMIT_AUTHOR"]
	env["DRONE_COMMIT_AUTHOR_EMAIL"] = env["CI_COMMIT_AUTHOR_EMAIL"]
	env["DRONE_COMMIT_AUTHOR_AVATAR"] = env["CI_COMMIT_AUTHOR_AVATAR"]
	// repo
	env["DRONE_REPO"] = env["CI_REPO"]
	env["DRONE_REPO_SCM"] = env["CI_REPO_SCM"]
	env["DRONE_REPO_OWNER"] = env["CI_REPO_OWNER"]
	env["DRONE_REPO_NAME"] = env["CI_REPO_NAME"]
	env["DRONE_REPO_LINK"] = env["CI_REPO_LINK"]
	env["DRONE_REPO_BRANCH"] = env["CI_REPO_DEFAULT_BRANCH"]
	env["DRONE_REPO_PRIVATE"] = env["CI_REPO_PRIVATE"]
	// clone
	env["DRONE_REMOTE_URL"] = env["CI_REPO_CLONE_URL"]
	env["DRONE_GIT_HTTP_URL"] = env["CI_REPO_CLONE_URL"]
	// misc
	env["DRONE_SYSTEM_HOST"] = env["CI_SYSTEM_HOST"]
	env["DRONE_STEP_NUMBER"] = env["CI_STEP_NUMBER"]
}
