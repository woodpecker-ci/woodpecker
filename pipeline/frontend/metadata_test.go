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

package frontend_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/woodpecker-ci/woodpecker/pipeline/frontend"
	"github.com/woodpecker-ci/woodpecker/pipeline/frontend/metadata"
	"github.com/woodpecker-ci/woodpecker/server/forge/mocks"
	"github.com/woodpecker-ci/woodpecker/server/model"
)

func TestEnvVarSubst(t *testing.T) {
	testCases := []struct {
		name    string
		yaml    string
		environ map[string]string
		want    string
	}{{
		name: "simple substitution",
		yaml: `steps:
		step1:
			image: ${HELLO_IMAGE}`,
		environ: map[string]string{"HELLO_IMAGE": "hello-world"},
		want: `steps:
		step1:
			image: hello-world`,
	}}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			result, err := frontend.EnvVarSubst(testCase.yaml, testCase.environ)
			assert.NoError(t, err)
			assert.EqualValues(t, testCase.want, result)
		})
	}
}

func TestMetadataFromStruct(t *testing.T) {
	forge := mocks.NewForge(t)
	forge.On("Name").Return("gitea")
	forge.On("URL").Return("https://gitea.com")

	testCases := []struct {
		name             string
		forge            metadata.ServerForge
		repo             *model.Repo
		pipeline, last   *model.Pipeline
		workflow         *model.Workflow
		link             string
		expectedMetadata metadata.Metadata
		expectedEnviron  map[string]string
	}{
		{
			name:             "Test with empty info",
			expectedMetadata: metadata.Metadata{Sys: metadata.System{Name: "woodpecker"}},
			expectedEnviron: map[string]string{
				"CI": "woodpecker", "CI_BUILD_CREATED": "0", "CI_BUILD_DEPLOY_TARGET": "", "CI_BUILD_EVENT": "", "CI_BUILD_FINISHED": "0", "CI_BUILD_LINK": "", "CI_BUILD_NUMBER": "0", "CI_BUILD_PARENT": "0",
				"CI_BUILD_STARTED": "0", "CI_BUILD_STATUS": "", "CI_COMMIT_AUTHOR": "", "CI_COMMIT_AUTHOR_AVATAR": "", "CI_COMMIT_AUTHOR_EMAIL": "", "CI_COMMIT_BRANCH": "", "CI_COMMIT_LINK": "",
				"CI_COMMIT_MESSAGE": "", "CI_COMMIT_PULL_REQUEST": "", "CI_COMMIT_PULL_REQUEST_LABELS": "", "CI_COMMIT_REF": "", "CI_COMMIT_REFSPEC": "", "CI_COMMIT_SHA": "", "CI_COMMIT_SOURCE_BRANCH": "",
				"CI_COMMIT_TAG": "", "CI_COMMIT_TARGET_BRANCH": "", "CI_COMMIT_URL": "", "CI_FORGE_TYPE": "", "CI_FORGE_URL": "", "CI_JOB_FINISHED": "", "CI_JOB_NUMBER": "0", "CI_JOB_STARTED": "",
				"CI_JOB_STATUS": "", "CI_PIPELINE_CREATED": "0", "CI_PIPELINE_DEPLOY_TARGET": "", "CI_PIPELINE_EVENT": "", "CI_PIPELINE_FINISHED": "0", "CI_PIPELINE_LINK": "", "CI_PIPELINE_NUMBER": "0",
				"CI_PIPELINE_PARENT": "0", "CI_PIPELINE_STARTED": "0", "CI_PIPELINE_STATUS": "", "CI_PIPELINE_URL": "", "CI_PREV_BUILD_CREATED": "0", "CI_PREV_BUILD_DEPLOY_TARGET": "",
				"CI_PREV_BUILD_EVENT": "", "CI_PREV_BUILD_FINISHED": "0", "CI_PREV_BUILD_LINK": "", "CI_PREV_BUILD_NUMBER": "0", "CI_PREV_BUILD_PARENT": "0", "CI_PREV_BUILD_STARTED": "0",
				"CI_PREV_BUILD_STATUS": "", "CI_PREV_COMMIT_AUTHOR": "", "CI_PREV_COMMIT_AUTHOR_AVATAR": "", "CI_PREV_COMMIT_AUTHOR_EMAIL": "", "CI_PREV_COMMIT_BRANCH": "", "CI_PREV_COMMIT_LINK": "",
				"CI_PREV_COMMIT_MESSAGE": "", "CI_PREV_COMMIT_REF": "", "CI_PREV_COMMIT_REFSPEC": "", "CI_PREV_COMMIT_SHA": "", "CI_PREV_COMMIT_URL": "", "CI_PREV_PIPELINE_CREATED": "0",
				"CI_PREV_PIPELINE_DEPLOY_TARGET": "", "CI_PREV_PIPELINE_EVENT": "", "CI_PREV_PIPELINE_FINISHED": "0", "CI_PREV_PIPELINE_LINK": "", "CI_PREV_PIPELINE_NUMBER": "0", "CI_PREV_PIPELINE_PARENT": "0",
				"CI_PREV_PIPELINE_STARTED": "0", "CI_PREV_PIPELINE_STATUS": "", "CI_PREV_PIPELINE_URL": "", "CI_REPO": "", "CI_REPO_CLONE_URL": "", "CI_REPO_DEFAULT_BRANCH": "", "CI_REPO_LINK": "", "CI_REPO_REMOTE_ID": "",
				"CI_REPO_NAME": "", "CI_REPO_OWNER": "", "CI_REPO_PRIVATE": "false", "CI_REPO_REMOTE": "", "CI_REPO_SCM": "git", "CI_REPO_TRUSTED": "false", "CI_REPO_URL": "", "CI_STEP_FINISHED": "",
				"CI_STEP_NAME": "", "CI_STEP_NUMBER": "0", "CI_STEP_STARTED": "", "CI_STEP_STATUS": "", "CI_SYSTEM_ARCH": "", "CI_SYSTEM_HOST": "", "CI_SYSTEM_LINK": "", "CI_SYSTEM_NAME": "woodpecker",
				"CI_SYSTEM_PLATFORM": "", "CI_SYSTEM_URL": "", "CI_SYSTEM_VERSION": "", "CI_WORKFLOW_NAME": "", "CI_WORKFLOW_NUMBER": "0",
			},
		},
		{
			name:     "Test with forge",
			forge:    forge,
			repo:     &model.Repo{FullName: "testUser/testRepo", Link: "https://gitea.com/testUser/testRepo", Clone: "https://gitea.com/testUser/testRepo.git", Branch: "main", IsSCMPrivate: true},
			pipeline: &model.Pipeline{Number: 3},
			last:     &model.Pipeline{Number: 2},
			workflow: &model.Workflow{Name: "hello"},
			link:     "https://example.com",
			expectedMetadata: metadata.Metadata{
				Forge:    metadata.Forge{Type: "gitea", URL: "https://gitea.com"},
				Sys:      metadata.System{Name: "woodpecker", Host: "example.com", Link: "https://example.com"},
				Repo:     metadata.Repo{Owner: "testUser", Name: "testRepo", Link: "https://gitea.com/testUser/testRepo", CloneURL: "https://gitea.com/testUser/testRepo.git", Branch: "main", Private: true},
				Curr:     metadata.Pipeline{Number: 3},
				Prev:     metadata.Pipeline{Number: 2},
				Workflow: metadata.Workflow{Name: "hello"},
			},
			expectedEnviron: map[string]string{
				"CI": "woodpecker", "CI_BUILD_CREATED": "0", "CI_BUILD_DEPLOY_TARGET": "", "CI_BUILD_EVENT": "", "CI_BUILD_FINISHED": "0", "CI_BUILD_LINK": "", "CI_BUILD_NUMBER": "3", "CI_BUILD_PARENT": "0",
				"CI_BUILD_STARTED": "0", "CI_BUILD_STATUS": "", "CI_COMMIT_AUTHOR": "", "CI_COMMIT_AUTHOR_AVATAR": "", "CI_COMMIT_AUTHOR_EMAIL": "", "CI_COMMIT_BRANCH": "", "CI_COMMIT_LINK": "",
				"CI_COMMIT_MESSAGE": "", "CI_COMMIT_PULL_REQUEST": "", "CI_COMMIT_PULL_REQUEST_LABELS": "", "CI_COMMIT_REF": "", "CI_COMMIT_REFSPEC": "", "CI_COMMIT_SHA": "", "CI_COMMIT_SOURCE_BRANCH": "",
				"CI_COMMIT_TAG": "", "CI_COMMIT_TARGET_BRANCH": "", "CI_COMMIT_URL": "", "CI_FORGE_TYPE": "gitea", "CI_FORGE_URL": "https://gitea.com", "CI_JOB_FINISHED": "", "CI_JOB_NUMBER": "0",
				"CI_JOB_STARTED": "", "CI_JOB_STATUS": "", "CI_PIPELINE_CREATED": "0", "CI_PIPELINE_DEPLOY_TARGET": "", "CI_PIPELINE_EVENT": "", "CI_PIPELINE_FINISHED": "0", "CI_PIPELINE_LINK": "",
				"CI_PIPELINE_NUMBER": "3", "CI_PIPELINE_PARENT": "0", "CI_PIPELINE_STARTED": "0", "CI_PIPELINE_STATUS": "", "CI_PIPELINE_URL": "", "CI_PREV_BUILD_CREATED": "0", "CI_PREV_BUILD_DEPLOY_TARGET": "",
				"CI_PREV_BUILD_EVENT": "", "CI_PREV_BUILD_FINISHED": "0", "CI_PREV_BUILD_LINK": "", "CI_PREV_BUILD_NUMBER": "2", "CI_PREV_BUILD_PARENT": "0", "CI_PREV_BUILD_STARTED": "0",
				"CI_PREV_BUILD_STATUS": "", "CI_PREV_COMMIT_AUTHOR": "", "CI_PREV_COMMIT_AUTHOR_AVATAR": "", "CI_PREV_COMMIT_AUTHOR_EMAIL": "", "CI_PREV_COMMIT_BRANCH": "", "CI_PREV_COMMIT_LINK": "",
				"CI_PREV_COMMIT_MESSAGE": "", "CI_PREV_COMMIT_REF": "", "CI_PREV_COMMIT_REFSPEC": "", "CI_PREV_COMMIT_SHA": "", "CI_PREV_COMMIT_URL": "", "CI_PREV_PIPELINE_CREATED": "0",
				"CI_PREV_PIPELINE_DEPLOY_TARGET": "", "CI_PREV_PIPELINE_EVENT": "", "CI_PREV_PIPELINE_FINISHED": "0", "CI_PREV_PIPELINE_LINK": "", "CI_PREV_PIPELINE_NUMBER": "2", "CI_PREV_PIPELINE_PARENT": "0",
				"CI_PREV_PIPELINE_STARTED": "0", "CI_PREV_PIPELINE_STATUS": "", "CI_PREV_PIPELINE_URL": "", "CI_REPO": "testUser/testRepo", "CI_REPO_CLONE_URL": "https://gitea.com/testUser/testRepo.git",
				"CI_REPO_DEFAULT_BRANCH": "main", "CI_REPO_LINK": "https://gitea.com/testUser/testRepo", "CI_REPO_NAME": "testRepo", "CI_REPO_OWNER": "testUser", "CI_REPO_PRIVATE": "true", "CI_REPO_REMOTE_ID": "",
				"CI_REPO_REMOTE": "https://gitea.com/testUser/testRepo.git", "CI_REPO_SCM": "git", "CI_REPO_TRUSTED": "false", "CI_REPO_URL": "https://gitea.com/testUser/testRepo", "CI_STEP_FINISHED": "",
				"CI_STEP_NAME": "", "CI_STEP_NUMBER": "0", "CI_STEP_STARTED": "", "CI_STEP_STATUS": "", "CI_SYSTEM_ARCH": "", "CI_SYSTEM_HOST": "example.com", "CI_SYSTEM_LINK": "https://example.com",
				"CI_SYSTEM_NAME": "woodpecker", "CI_SYSTEM_PLATFORM": "", "CI_SYSTEM_URL": "https://example.com", "CI_SYSTEM_VERSION": "", "CI_WORKFLOW_NAME": "hello", "CI_WORKFLOW_NUMBER": "0",
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			result := frontend.MetadataFromStruct(testCase.forge, testCase.repo, testCase.pipeline, testCase.last, testCase.workflow, testCase.link)
			assert.EqualValues(t, testCase.expectedMetadata, result)
			assert.EqualValues(t, testCase.expectedEnviron, result.Environ())
		})
	}
}
