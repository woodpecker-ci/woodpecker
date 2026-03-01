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

func TestStatusValueMerge(t *testing.T) {
	tests := []struct {
		s model.StatusValue
		t model.StatusValue
		e model.StatusValue
	}{
		{
			s: model.StatusSuccess,
			t: model.StatusSkipped,
			e: model.StatusSuccess,
		},
		{
			s: model.StatusSuccess,
			t: model.StatusSuccess,
			e: model.StatusSuccess,
		},
		{
			s: model.StatusFailure,
			t: model.StatusSuccess,
			e: model.StatusFailure,
		},
		{
			s: model.StatusRunning,
			t: model.StatusSuccess,
			e: model.StatusRunning,
		},
		{
			s: model.StatusRunning,
			t: model.StatusFailure,
			e: model.StatusRunning,
		},
		{
			s: model.StatusFailure,
			t: model.StatusKilled,
			e: model.StatusKilled,
		},
		{
			s: model.StatusSkipped,
			t: model.StatusKilled,
			e: model.StatusKilled,
		},
		{
			s: model.StatusSkipped,
			t: model.StatusSkipped,
			e: model.StatusSkipped,
		},
		{
			s: model.StatusSkipped,
			t: model.StatusCancelled,
			e: model.StatusKilled,
		},
		{
			s: model.StatusSuccess,
			t: model.StatusCancelled,
			e: model.StatusKilled,
		},
		{
			s: model.StatusFailure,
			t: model.StatusCancelled,
			e: model.StatusKilled,
		},
	}
	for _, tt := range tests {
		assert.Equal(t, tt.e, MergeStatusValues(tt.s, tt.t))
		assert.Equal(t, tt.e, MergeStatusValues(tt.t, tt.s))
	}
}
