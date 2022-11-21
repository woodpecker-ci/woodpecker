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
	copy("CI_COMMIT_BRANCH", "DRONE_BRANCH", env)
	copy("CI_COMMIT_PULL_REQUEST", "DRONE_PULL_REQUEST", env)
	copy("CI_COMMIT_TAG", "DRONE_TAG", env)
	copy("CI_COMMIT_SOURCE_BRANCH", "DRONE_SOURCE_BRANCH", env)
	copy("CI_COMMIT_TARGET_BRANCH", "DRONE_TARGET_BRANCH", env)
	// pipeline
	copy("CI_PIPELINE_NUMBER", "DRONE_BUILD_NUMBER", env)
	copy("CI_PIPELINE_PARENT", "DRONE_BUILD_PARENT", env)
	copy("CI_PIPELINE_EVENT", "DRONE_BUILD_EVENT", env)
	copy("CI_PIPELINE_STATUS", "DRONE_BUILD_STATUS", env)
	copy("CI_PIPELINE_LINK", "DRONE_BUILD_LINK", env)
	copy("CI_PIPELINE_CREATED", "DRONE_BUILD_CREATED", env)
	copy("CI_PIPELINE_STARTED", "DRONE_BUILD_STARTED", env)
	copy("CI_PIPELINE_FINISHED", "DRONE_BUILD_FINISHED", env)
	// commit
	copy("CI_COMMIT_SHA", "DRONE_COMMIT", env)
	copy("CI_COMMIT_SHA", "DRONE_COMMIT_SHA", env)
	copy("CI_PREV_COMMIT_SHA", "DRONE_COMMIT_BEFORE", env)
	copy("CI_COMMIT_REF", "DRONE_COMMIT_REF", env)
	copy("CI_COMMIT_BRANCH", "DRONE_COMMIT_BRANCH", env)
	copy("CI_COMMIT_LINK", "DRONE_COMMIT_LINK", env)
	copy("CI_COMMIT_MESSAGE", "DRONE_COMMIT_MESSAGE", env)
	copy("CI_COMMIT_AUTHOR", "DRONE_COMMIT_AUTHOR", env)
	copy("CI_COMMIT_AUTHOR", "DRONE_COMMIT_AUTHOR_NAME", env)
	copy("CI_COMMIT_AUTHOR_EMAIL", "DRONE_COMMIT_AUTHOR_EMAIL", env)
	copy("CI_COMMIT_AUTHOR_AVATAR", "DRONE_COMMIT_AUTHOR_AVATAR", env)
	// repo
	copy("CI_REPO", "DRONE_REPO", env)
	copy("CI_REPO_SCM", "DRONE_REPO_SCM", env)
	copy("CI_REPO_OWNER", "DRONE_REPO_OWNER", env)
	copy("CI_REPO_NAME", "DRONE_REPO_NAME", env)
	copy("CI_REPO_LINK", "DRONE_REPO_LINK", env)
	copy("CI_REPO_DEFAULT_BRANCH", "DRONE_REPO_BRANCH", env)
	copy("CI_REPO_PRIVATE", "DRONE_REPO_PRIVATE", env)
	// clone
	copy("CI_REPO_CLONE_URL", "DRONE_REMOTE_URL", env)
	copy("CI_REPO_CLONE_URL", "DRONE_GIT_HTTP_URL", env)
	// misc
	copy("CI_SYSTEM_HOST", "DRONE_SYSTEM_HOST", env)
	copy("CI_STEP_NUMBER", "DRONE_STEP_NUMBER", env)
}

func copy(woodpecker, drone string, env map[string]string) {
	var present bool
	var value string

	value, present = env[woodpecker]
	if present {
		env[drone] = value
	}
}
