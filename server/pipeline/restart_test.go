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

package pipeline

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"go.woodpecker-ci.org/woodpecker/v3/server/model"
)

func TestCreateNewOutOfOldMergesRestartVariables(t *testing.T) {
	old := &model.Pipeline{
		AdditionalVariables: map[string]string{
			"KEEP":     "old",
			"OVERRIDE": "old",
		},
	}

	newPipeline := createNewOutOfOld(old, map[string]string{
		"OVERRIDE": "restart",
		"EXTRA":    "restart",
	})

	// restart-supplied variables are persisted with the new pipeline so a
	// later (re-)compilation sees them; explicit restart input wins over
	// inherited variables.
	assert.Equal(t, map[string]string{
		"KEEP":     "old",
		"OVERRIDE": "restart",
		"EXTRA":    "restart",
	}, newPipeline.AdditionalVariables)

	// the old pipeline's map must not be mutated
	assert.Equal(t, map[string]string{
		"KEEP":     "old",
		"OVERRIDE": "old",
	}, old.AdditionalVariables)
}

func TestCreateNewOutOfOldNilMaps(t *testing.T) {
	newPipeline := createNewOutOfOld(&model.Pipeline{}, nil)
	assert.Empty(t, newPipeline.AdditionalVariables)

	newPipeline = createNewOutOfOld(&model.Pipeline{}, map[string]string{"A": "b"})
	assert.Equal(t, map[string]string{"A": "b"}, newPipeline.AdditionalVariables)
}
