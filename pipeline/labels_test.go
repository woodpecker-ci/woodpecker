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

import "testing"

func TestMissingRequiredLabels(t *testing.T) {
	t.Parallel()

	testdata := []struct {
		taskLabels     map[string]string
		requiredLabels map[string]string
		want           bool
	}{
		// Required label present and matches
		{
			taskLabels:     map[string]string{"os": "linux"},
			requiredLabels: map[string]string{"!os": "linux", "platform": "arm64"},
			want:           false,
		},
		// Required label present but does not match
		{
			taskLabels:     map[string]string{"os": "windows"},
			requiredLabels: map[string]string{"!os": "linux", "platform": "amd64"},
			want:           true,
		},
		// Required label missing
		{
			taskLabels:     map[string]string{"arch": "amd64"},
			requiredLabels: map[string]string{"!os": "linux"},
			want:           true,
		},
		// No agent labels
		{
			taskLabels:     map[string]string{"os": "linux"},
			requiredLabels: map[string]string{},
			want:           false,
		},
	}

	for _, tt := range testdata {
		if got := requiredLabelsMissing(tt.taskLabels, tt.requiredLabels); got != tt.want {
			t.Errorf("requiredLabelsMissing(%v, %v) = %v, want %v", tt.taskLabels, tt.requiredLabels, got, tt.want)
		}
	}
}
