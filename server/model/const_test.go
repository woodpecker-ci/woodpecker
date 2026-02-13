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

package model

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestStatusValueMerge(t *testing.T) {
	tests := []struct {
		s   StatusValue
		t StatusValue
		e StatusValue
	}{
		{
			s: StatusSuccess,
			t: StatusSkipped,
			e: StatusSkipped,
		},
		{
			s: StatusSuccess,
			t: StatusSuccess,
			e: StatusSuccess,
		},
		{
			s: StatusFailure,
			t: StatusSuccess,
			e: StatusFailure,
		},
		{
			s: StatusRunning,
			t: StatusSuccess,
			e: StatusRunning,
		},
		{
			s: StatusRunning,
			t: StatusFailure,
			e: StatusRunning,
		},
		{
			s: StatusFailure,
			t: StatusKilled,
			e: StatusKilled,
		},
		{
			s: StatusSkipped,
			t: StatusKilled,
			e: StatusKilled,
		},
		{
			s: StatusSkipped,
			t: StatusSkipped,
			e: StatusSkipped,
		},
	}
	for _, tt := range tests {
		assert.Equal(t, tt.e, tt.s.Merge(tt.t))
		assert.Equal(t, tt.e, tt.t.Merge(tt.s))
	}
}
