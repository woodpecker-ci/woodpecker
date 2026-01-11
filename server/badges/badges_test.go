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

package badges

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"go.woodpecker-ci.org/woodpecker/v3/server/model"
)

// Generate an SVG badge based on a pipeline.
func TestGenerate(t *testing.T) {
	assert.Equal(t, badgeNone, Generate("pipeline", model.StatusDeclined))
	assert.Equal(t, badgeSuccess, Generate("pipeline", model.StatusSuccess))
	assert.Equal(t, badgeFailure, Generate("pipeline", model.StatusFailure))
	assert.Equal(t, badgeError, Generate("pipeline", model.StatusError))
	assert.Equal(t, badgeError, Generate("pipeline", model.StatusKilled))
	assert.Equal(t, badgeStarted, Generate("pipeline", model.StatusPending))
	assert.Equal(t, badgeStarted, Generate("pipeline", model.StatusRunning))
}
