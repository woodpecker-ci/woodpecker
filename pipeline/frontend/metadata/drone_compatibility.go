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

package metadata

// SetDroneEnviron set dedicated to DroneCI environment vars as compatibility
// layer. Main purpose is to be compatible with drone plugins.
func SetDroneEnviron(env map[string]string) {
	// webhook
	copyEnv("CI_COMMIT_BRANCH", "DRONE_BRANCH", env)
	copyEnv("CI_COMMIT_PULL_REQUEST", "DRONE_PULL_REQUEST", env)
	copyEnv("CI_COMMIT_PULL_REQUEST", "PULLREQUEST_DRONE_PULL_REQUEST", env)
	copyEnv("CI_COMMIT_TAG", "DRONE_TAG", env)
	copyEnv("CI_COMMIT_SOURCE_BRANCH", "DRONE_SOURCE_BRANCH", env)
	copyEnv("CI_COMMIT_TARGET_BRANCH", "DRONE_TARGET_BRANCH", env)
	// pipeline
	copyEnv("CI_PIPELINE_NUMBER", "DRONE_BUILD_NUMBER", env)
	copyEnv("CI_PIPELINE_PARENT", "DRONE_BUILD_PARENT", env)
	copyEnv("CI_PIPELINE_EVENT", "DRONE_BUILD_EVENT", env)
	copyEnv("CI_PIPELINE_URL", "DRONE_BUILD_LINK", env)
	copyEnv("CI_PIPELINE_CREATED", "DRONE_BUILD_CREATED", env)
	copyEnv("CI_PIPELINE_STARTED", "DRONE_BUILD_STARTED", env)
	// commit
	copyEnv("CI_COMMIT_SHA", "DRONE_COMMIT", env)
	copyEnv("CI_COMMIT_SHA", "DRONE_COMMIT_SHA", env)
	copyEnv("CI_PREV_COMMIT_SHA", "DRONE_COMMIT_BEFORE", env)
	copyEnv("CI_COMMIT_REF", "DRONE_COMMIT_REF", env)
	copyEnv("CI_COMMIT_BRANCH", "DRONE_COMMIT_BRANCH", env)
	copyEnv("CI_PIPELINE_FORGE_URL", "DRONE_COMMIT_LINK", env)
	copyEnv("CI_COMMIT_MESSAGE", "DRONE_COMMIT_MESSAGE", env)
	copyEnv("CI_COMMIT_AUTHOR", "DRONE_COMMIT_AUTHOR", env)
	copyEnv("CI_COMMIT_AUTHOR", "DRONE_COMMIT_AUTHOR_NAME", env)
	copyEnv("CI_COMMIT_AUTHOR_EMAIL", "DRONE_COMMIT_AUTHOR_EMAIL", env)
	copyEnv("CI_COMMIT_AUTHOR_AVATAR", "DRONE_COMMIT_AUTHOR_AVATAR", env)
	// repo
	copyEnv("CI_REPO", "DRONE_REPO", env)
	copyEnv("CI_REPO_SCM", "DRONE_REPO_SCM", env)
	copyEnv("CI_REPO_OWNER", "DRONE_REPO_OWNER", env)
	copyEnv("CI_REPO_NAME", "DRONE_REPO_NAME", env)
	copyEnv("CI_REPO_URL", "DRONE_REPO_LINK", env)
	copyEnv("CI_REPO_DEFAULT_BRANCH", "DRONE_REPO_BRANCH", env)
	copyEnv("CI_REPO_PRIVATE", "DRONE_REPO_PRIVATE", env)
	// clone
	copyEnv("CI_REPO_CLONE_URL", "DRONE_REMOTE_URL", env)
	copyEnv("CI_REPO_CLONE_URL", "DRONE_GIT_HTTP_URL", env)
	// misc
	copyEnv("CI_SYSTEM_HOST", "DRONE_SYSTEM_HOST", env)
	copyEnv("CI_STEP_NUMBER", "DRONE_STEP_NUMBER", env)

	env["DRONE_BUILD_STATUS"] = "success"

	// some quirks

	// Legacy env var to prevent the plugin from throwing an error
	// when converting an empty string to a number
	//
	// plugins affected: "plugins/manifest"
	if env["CI_COMMIT_PULL_REQUEST"] == "" {
		env["PULLREQUEST_DRONE_PULL_REQUEST"] = "0"
	}
}

func copyEnv(woodpecker, drone string, env map[string]string) {
	var present bool
	var value string

	value, present = env[woodpecker]
	if present {
		env[drone] = value
	}
}
