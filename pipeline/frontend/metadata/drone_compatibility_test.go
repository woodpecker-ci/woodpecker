package metadata_test

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/woodpecker-ci/woodpecker/pipeline/frontend/metadata"
)

func TestSetDroneEnviron(t *testing.T) {
	woodpeckerVars := `CI=woodpecker
CI_BUILD_CREATED=1685749339
CI_BUILD_FINISHED=1685749350
CI_BUILD_LINK=https://codeberg.org/Epsilon_02/todo-checker/pulls/9
CI_BUILD_STARTED=1685749339
CI_BUILD_STATUS=success
CI_COMMIT_AUTHOR=6543
CI_COMMIT_AUTHOR_AVATAR=https://codeberg.org/avatars/09a234c768cb9bca78f6b2f82d6af173
CI_COMMIT_BRANCH=main
CI_COMMIT_LINK=https://codeberg.org/Epsilon_02/todo-checker/pulls/9
CI_COMMIT_MESSAGE=fix testscript
CI_COMMIT_PULL_REQUEST=9
CI_COMMIT_REF=refs/pull/9/head
CI_COMMIT_REFSPEC=fix_fail-on-err:main
CI_COMMIT_SHA=a778b069d9f5992786d2db9be493b43868cfce76
CI_COMMIT_SOURCE_BRANCH=fix_fail-on-err
CI_COMMIT_TARGET_BRANCH=main
CI_JOB_FINISHED=1685749350
CI_JOB_STARTED=1685749339
CI_JOB_STATUS=success
CI_MACHINE=7939910e431b
CI_PIPELINE_CREATED=1685749339
CI_PIPELINE_EVENT=pull_request
CI_PIPELINE_FINISHED=1685749350
CI_PIPELINE_LINK=https://codeberg.org/Epsilon_02/todo-checker/pulls/9
CI_PIPELINE_NUMBER=41
CI_PIPELINE_STARTED=1685749339
CI_PIPELINE_STATUS=success
CI_PREV_BUILD_CREATED=1685748680
CI_PREV_BUILD_EVENT=pull_request
CI_PREV_BUILD_FINISHED=1685748704
CI_PREV_BUILD_LINK=https://codeberg.org/Epsilon_02/todo-checker/pulls/13
CI_PREV_BUILD_NUMBER=40
CI_PREV_BUILD_STARTED=1685748680
CI_PREV_BUILD_STATUS=success
CI_PREV_COMMIT_AUTHOR=6543
CI_PREV_COMMIT_AUTHOR_AVATAR=https://codeberg.org/avatars/09a234c768cb9bca78f6b2f82d6af173
CI_PREV_COMMIT_BRANCH=main
CI_PREV_COMMIT_LINK=https://codeberg.org/Epsilon_02/todo-checker/pulls/13
CI_PREV_COMMIT_MESSAGE=Print filename and linenuber on fail
CI_PREV_COMMIT_REF=refs/pull/13/head
CI_PREV_COMMIT_REFSPEC=print_file_and_line:main
CI_PREV_COMMIT_SHA=e246aff5a9466df2e522efc9007823a7496d9d41
CI_PREV_PIPELINE_CREATED=1685748680
CI_PREV_PIPELINE_EVENT=pull_request
CI_PREV_PIPELINE_FINISHED=1685748704
CI_PREV_PIPELINE_LINK=https://codeberg.org/Epsilon_02/todo-checker/pulls/13
CI_PREV_PIPELINE_NUMBER=40
CI_PREV_PIPELINE_STARTED=1685748680
CI_PREV_PIPELINE_STATUS=success
CI_REPO=Epsilon_02/todo-checker
CI_REPO_CLONE_URL=https://codeberg.org/Epsilon_02/todo-checker.git
CI_REPO_DEFAULT_BRANCH=main
CI_REPO_LINK=https://codeberg.org/Epsilon_02/todo-checker
CI_REPO_NAME=todo-checker
CI_REPO_OWNER=Epsilon_02
CI_REPO_REMOTE=https://codeberg.org/Epsilon_02/todo-checker.git
CI_REPO_SCM=git
CI_STEP_FINISHED=1685749350
CI_STEP_NAME=wp_01h1z7v5d1tskaqjexw0ng6w7d_0_step_3
CI_STEP_STARTED=1685749339
CI_STEP_STATUS=success
CI_SYSTEM_ARCH=linux/amd64
CI_SYSTEM_HOST=ci.codeberg.org
CI_SYSTEM_LINK=https://ci.codeberg.org
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
DRONE_TARGET_BRANCH=main`

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
		ss := strings.SplitN(strings.TrimSpace(s), "=", 2)
		if len(ss) != 2 {
			t.Fatal("helper function got invalid test data")
		}
		result[ss[0]] = ss[1]
	}
	return result
}
