// Copyright 2026 Woodpecker Authors
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

package runner

import (
	backend_types "go.woodpecker-ci.org/woodpecker/v3/pipeline/backend/types"
)

// extractRepositoryName returns the CI_REPO value embedded in the first step's
// environment. This is a temporary workaround until the workflow payload carries
// repository metadata explicitly.
func extractRepositoryName(config *backend_types.Config) string {
	return config.Stages[0].Steps[0].Environment["CI_REPO"]
}

// extractPipelineNumber returns the CI_PIPELINE_NUMBER value embedded in the
// first step's environment. Same caveat as extractRepositoryName.
func extractPipelineNumber(config *backend_types.Config) string {
	return config.Stages[0].Steps[0].Environment["CI_PIPELINE_NUMBER"]
}
