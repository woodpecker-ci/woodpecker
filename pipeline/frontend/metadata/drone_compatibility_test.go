// Copyright 2023 Woodpecker Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package metadata_test

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"go.woodpecker-ci.org/woodpecker/v2/pipeline/frontend/metadata"
)

func TestSetDroneEnvironOnPull(t *testing.T) {
	woodpeckerVars := `CI=woodpecker
CI_COMMIT_AUTHOR=6543
CI_COMMIT_AUTHOR_AVATAR=https://codeberg.org/avatars/09a234c768cb9bca78f6b2f82d6af173
CI_COMMIT_BRANCH=main
CI_COMMIT_MESSAGE=fix testscript
CI_COMMIT_PULL_REQUEST=9
CI_COMMIT_PULL_REQUEST_LABELS=tests,bugfix
CI_COMMIT_REF=refs/pull/9/head
CI_COMMIT_REFSPEC=fix_fail-on-err:main
CI_COMMIT_SHA=a778b069d9f5992786d2db9be493b43868cfce76
CI_COMMIT_SOURCE_BRANCH=fix_fail-on-err
CI_COMMIT_TARGET_BRANCH=main
CI_MACHINE=7939910e431b
CI_PIPELINE_CREATED=1685749339
CI_PIPELINE_EVENT=pull_request
CI_PIPELINE_FINISHED=1685749350
CI_PIPELINE_NUMBER=41
CI_PIPELINE_STARTED=1685749339
CI_PIPELINE_STATUS=success
CI_PREV_COMMIT_AUTHOR=6543
CI_PREV_COMMIT_AUTHOR_AVATAR=https://codeberg.org/avatars/09a234c768cb9bca78f6b2f82d6af173
CI_PREV_COMMIT_BRANCH=main
CI_PREV_COMMIT_MESSAGE=Print filename and linenuber on fail
CI_PREV_COMMIT_REF=refs/pull/13/head
CI_PREV_COMMIT_REFSPEC=print_file_and_line:main
CI_PREV_COMMIT_SHA=e246aff5a9466df2e522efc9007823a7496d9d41
CI_PREV_PIPELINE_CREATED=1685748680
CI_PREV_PIPELINE_EVENT=pull_request
CI_PREV_PIPELINE_FINISHED=1685748704
CI_PREV_PIPELINE_NUMBER=40
CI_PREV_PIPELINE_STARTED=1685748680
CI_PREV_PIPELINE_STATUS=success
CI_REPO=Epsilon_02/todo-checker
CI_REPO_CLONE_URL=https://codeberg.org/Epsilon_02/todo-checker.git
CI_REPO_DEFAULT_BRANCH=main
CI_REPO_NAME=todo-checker
CI_REPO_OWNER=Epsilon_02
CI_REPO_SCM=git
CI_STEP_FINISHED=1685749350
CI_STEP_NAME=wp_01h1z7v5d1tskaqjexw0ng6w7d_0_step_3
CI_STEP_STARTED=1685749339
CI_STEP_STATUS=success
CI_SYSTEM_PLATFORM=linux/amd64
CI_SYSTEM_HOST=ci.codeberg.org
CI_SYSTEM_NAME=woodpecker
CI_SYSTEM_VERSION=next-dd644da3
CI_WORKFLOW_NAME=woodpecker
CI_WORKFLOW_NUMBER=1
CI_WORKSPACE=/woodpecker/src/codeberg.org/Epsilon_02/todo-checker`

	droneVars := `DRONE_BRANCH=main
DRONE_BUILD_CREATED=1685749339
DRONE_BUILD_EVENT=pull_request
DRONE_BUILD_FINISHED=1685749350
DRONE_BUILD_NUMBER=41
DRONE_BUILD_STARTED=1685749339
DRONE_BUILD_STATUS=success
DRONE_COMMIT=a778b069d9f5992786d2db9be493b43868cfce76
DRONE_COMMIT_AUTHOR=6543
DRONE_COMMIT_AUTHOR_AVATAR=https://codeberg.org/avatars/09a234c768cb9bca78f6b2f82d6af173
DRONE_COMMIT_AUTHOR_NAME=6543
DRONE_COMMIT_BEFORE=e246aff5a9466df2e522efc9007823a7496d9d41
DRONE_COMMIT_BRANCH=main
DRONE_COMMIT_MESSAGE=fix testscript
DRONE_COMMIT_REF=refs/pull/9/head
DRONE_COMMIT_SHA=a778b069d9f5992786d2db9be493b43868cfce76
DRONE_GIT_HTTP_URL=https://codeberg.org/Epsilon_02/todo-checker.git
DRONE_PULL_REQUEST=9
DRONE_REMOTE_URL=https://codeberg.org/Epsilon_02/todo-checker.git
DRONE_REPO=Epsilon_02/todo-checker
DRONE_REPO_BRANCH=main
DRONE_REPO_NAME=todo-checker
DRONE_REPO_OWNER=Epsilon_02
DRONE_REPO_SCM=git
DRONE_SOURCE_BRANCH=fix_fail-on-err
DRONE_SYSTEM_HOST=ci.codeberg.org
DRONE_TARGET_BRANCH=main
PULLREQUEST_DRONE_PULL_REQUEST=9`

	env := convertListToEnvMap(t, woodpeckerVars)
	metadata.SetDroneEnviron(env)
	// filter only new added env vars
	for k := range convertListToEnvMap(t, woodpeckerVars) {
		delete(env, k)
	}
	assert.EqualValues(t, convertListToEnvMap(t, droneVars), env)
}

func TestSetDroneEnvironOnPush(t *testing.T) {
	woodpeckerVars := `CI_COMMIT_AUTHOR=test
CI_COMMIT_AUTHOR_AVATAR=http://1.2.3.4:3000/avatars/dd46a756faad4727fb679320751f6dea
CI_COMMIT_AUTHOR_EMAIL=test@noreply.localhost
CI_COMMIT_BRANCH=main
CI_COMMIT_MESSAGE=revert 9b2aed1392fc097ef7b027712977722fb004d463
CI_COMMIT_PULL_REQUEST=
CI_COMMIT_PULL_REQUEST_LABELS=
CI_COMMIT_REF=refs/heads/main
CI_COMMIT_REFSPEC=
CI_COMMIT_SHA=8826c98181353075bbeee8f99b400496488e3523
CI_COMMIT_SOURCE_BRANCH=
CI_COMMIT_TAG=
CI_COMMIT_TARGET_BRANCH=
CI_COMMIT_URL=http://1.2.3.4:3000/test/woodpecker-test/commit/8826c98181353075bbeee8f99b400496488e3523
CI_FORGE_TYPE=gitea
CI_FORGE_URL=http://1.2.3.4:3000
CI_MACHINE=hagalaz
CI_PIPELINE_CREATED=1721328737
CI_PIPELINE_DEPLOY_TARGET=
CI_PIPELINE_DEPLOY_TASK=
CI_PIPELINE_EVENT=push
CI_PIPELINE_FILES=[".woodpecker.yaml"]
CI_PIPELINE_FINISHED=1721328738
CI_PIPELINE_FORGE_URL=http://1.2.3.4:3000/test/woodpecker-test/commit/8826c98181353075bbeee8f99b400496488e3523
CI_PIPELINE_NUMBER=24
CI_PIPELINE_PARENT=23
CI_PIPELINE_STARTED=1721328737
CI_PIPELINE_STATUS=success
CI_PIPELINE_URL=http://1.2.3.4:8000/repos/2/pipeline/24
CI_PREV_COMMIT_AUTHOR=test
CI_PREV_COMMIT_AUTHOR_AVATAR=http://1.2.3.4:3000/avatars/dd46a756faad4727fb679320751f6dea
CI_PREV_COMMIT_AUTHOR_EMAIL=test@noreply.localhost
CI_PREV_COMMIT_BRANCH=main
CI_PREV_COMMIT_MESSAGE=revert 9b2aed1392fc097ef7b027712977722fb004d463
CI_PREV_COMMIT_REF=refs/heads/main
CI_PREV_COMMIT_REFSPEC=
CI_PREV_COMMIT_SHA=8826c98181353075bbeee8f99b400496488e3523
CI_PREV_COMMIT_URL=http://1.2.3.4:3000/test/woodpecker-test/commit/8826c98181353075bbeee8f99b400496488e3523
CI_PREV_COMMIT_SOURCE_BRANCH=
CI_PREV_COMMIT_TARGET_BRANCH=
CI_PREV_PIPELINE_CREATED=1721086039
CI_PREV_PIPELINE_DEPLOY_TARGET=
CI_PREV_PIPELINE_DEPLOY_TASK=
CI_PREV_PIPELINE_EVENT=push
CI_PREV_PIPELINE_FINISHED=1721086056
CI_PREV_PIPELINE_FORGE_URL=http://1.2.3.4:3000/test/woodpecker-test/commit/8826c98181353075bbeee8f99b400496488e3523
CI_PREV_PIPELINE_NUMBER=23
CI_PREV_PIPELINE_PARENT=0
CI_PREV_PIPELINE_STARTED=1721086039
CI_PREV_PIPELINE_STATUS=failure
CI_PREV_PIPELINE_URL=http://1.2.3.4:8000/repos/2/pipeline/23
CI_REPO=test/woodpecker-test
CI_REPO_CLONE_SSH_URL=user@1.2.3.4:test/woodpecker-test.git
CI_REPO_CLONE_URL=http://1.2.3.4:3000/test/woodpecker-test.git
CI_REPO_DEFAULT_BRANCH=main
CI_REPO_NAME=woodpecker-test
CI_REPO_OWNER=test
CI_REPO_PRIVATE=false
CI_REPO_REMOTE_ID=4
CI_REPO_SCM=git
CI_REPO_TRUSTED=false
CI_REPO_URL=http://1.2.3.4:3000/test/woodpecker-test
CI_STEP_FINISHED=1721328738
CI_STEP_NAME=
CI_STEP_NUMBER=0
CI_STEP_STARTED=1721328737
CI_STEP_STATUS=success
CI_STEP_URL=http://1.2.3.4:8000/repos/2/pipeline/24
CI_SYSTEM_HOST=1.2.3.4:8000
CI_SYSTEM_NAME=woodpecker
CI_SYSTEM_PLATFORM=linux/amd64
CI_SYSTEM_URL=http://1.2.3.4:8000
CI_SYSTEM_VERSION=2.7.0
CI_WORKFLOW_NAME=woodpecker
CI_WORKFLOW_NUMBER=1
CI_WORKSPACE=/usr/local/src/1.2.3.4/test/woodpecker-test`

	droneVars := `DRONE_BRANCH=main
DRONE_BUILD_CREATED=1721328737
DRONE_BUILD_EVENT=push
DRONE_BUILD_FINISHED=1721328738
DRONE_BUILD_LINK=http://1.2.3.4:8000/repos/2/pipeline/24
DRONE_BUILD_NUMBER=24
DRONE_BUILD_PARENT=23
DRONE_BUILD_STARTED=1721328737
DRONE_BUILD_STATUS=success
DRONE_COMMIT=8826c98181353075bbeee8f99b400496488e3523
DRONE_COMMIT_AUTHOR=test
DRONE_COMMIT_AUTHOR_AVATAR=http://1.2.3.4:3000/avatars/dd46a756faad4727fb679320751f6dea
DRONE_COMMIT_AUTHOR_EMAIL=test@noreply.localhost
DRONE_COMMIT_AUTHOR_NAME=test
DRONE_COMMIT_BEFORE=8826c98181353075bbeee8f99b400496488e3523
DRONE_COMMIT_BRANCH=main
DRONE_COMMIT_LINK=http://1.2.3.4:3000/test/woodpecker-test/commit/8826c98181353075bbeee8f99b400496488e3523
DRONE_COMMIT_MESSAGE=revert 9b2aed1392fc097ef7b027712977722fb004d463
DRONE_COMMIT_REF=refs/heads/main
DRONE_COMMIT_SHA=8826c98181353075bbeee8f99b400496488e3523
DRONE_GIT_HTTP_URL=http://1.2.3.4:3000/test/woodpecker-test.git
DRONE_PULL_REQUEST=
DRONE_REMOTE_URL=http://1.2.3.4:3000/test/woodpecker-test.git
DRONE_REPO=test/woodpecker-test
DRONE_REPO_BRANCH=main
DRONE_REPO_LINK=http://1.2.3.4:3000/test/woodpecker-test
DRONE_REPO_NAME=woodpecker-test
DRONE_REPO_OWNER=test
DRONE_REPO_PRIVATE=false
DRONE_REPO_SCM=git
DRONE_SOURCE_BRANCH=
DRONE_STEP_NUMBER=0
DRONE_SYSTEM_HOST=1.2.3.4:8000
DRONE_TAG=
DRONE_TARGET_BRANCH=
PULLREQUEST_DRONE_PULL_REQUEST=0`

	env := convertListToEnvMap(t, woodpeckerVars)
	metadata.SetDroneEnviron(env)
	// filter only new added env vars
	for k := range convertListToEnvMap(t, woodpeckerVars) {
		delete(env, k)
	}
	assert.EqualValues(t, convertListToEnvMap(t, droneVars), env)
}

func convertListToEnvMap(t *testing.T, list string) map[string]string {
	result := make(map[string]string)
	for _, s := range strings.Split(list, "\n") {
		before, after, _ := strings.Cut(strings.TrimSpace(s), "=")
		if before == "" {
			t.Fatal("helper function got invalid test data")
		}
		result[before] = after
	}
	return result
}
