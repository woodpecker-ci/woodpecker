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

package rpc

import (
	"strings"
	"testing"
)

func TestSanitizeAgentVersion(t *testing.T) {
	cases := []struct {
		name string
		in   string
		out  string
	}{
		{"empty", "", ""},
		{"semver", "v3.14.0", "v3.14.0"},
		{"dev", "dev", "dev"},
		{"prerelease", "v3.14.0-rc1+build.7", "v3.14.0-rc1+build.7"},
		{"next snapshot", "next-abc123", "next-abc123"},
		{"underscored", "snapshot_2026_05", "snapshot_2026_05"},
		{"too long", strings.Repeat("a", maxAgentVersionLen+1), ""},
		{"newline injected", "v3.14\nrogue", ""},
		{"shell metachar", "v3.14;rm -rf /", ""},
		{"unicode", "v3.14-α", ""},
		{"space", "v3.14 rc1", ""},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := sanitizeAgentVersion(tc.in)
			if got != tc.out {
				t.Fatalf("sanitizeAgentVersion(%q) = %q, want %q", tc.in, got, tc.out)
			}
		})
	}
}
