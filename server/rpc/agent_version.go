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

const maxAgentVersionLen = 64

// sanitizeAgentVersion returns v if it looks like a plausible version string
// (alphanumerics with `.`, `_`, `-`, `+`, length-bounded), or "" otherwise.
// The agent-reported version is stored on the agent record and rendered in
// the UI/logs, so we never trust it raw.
func sanitizeAgentVersion(v string) string {
	if v == "" || len(v) > maxAgentVersionLen {
		return ""
	}
	for _, r := range v {
		switch {
		case r >= 'a' && r <= 'z',
			r >= 'A' && r <= 'Z',
			r >= '0' && r <= '9',
			r == '.', r == '-', r == '_', r == '+':
		default:
			return ""
		}
	}
	return v
}
