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

package woodpecker

// The types the API of the server uses are generated from its models.
//go:generate go run ../typegen -model ../../server/model User Repo RepoPatch APIPipeline PipelineOptions Registry Secret APIFeed QueueInfo LogEntry Cron Agent Task Org OrgPatch

type ApprovalMode string

var (
	RequireApprovalNone         ApprovalMode = "none"          // require approval for no events
	RequireApprovalForks        ApprovalMode = "forks"         // require approval for PRs from forks
	RequireApprovalPullRequests ApprovalMode = "pull_requests" // require approval for all PRs (default)
	RequireApprovalAllEvents    ApprovalMode = "all_events"    // require approval for all events
)

func (mode ApprovalMode) Valid() bool {
	switch mode {
	case RequireApprovalNone,
		RequireApprovalForks,
		RequireApprovalPullRequests,
		RequireApprovalAllEvents:
		return true
	default:
		return false
	}
}

type (
	PipelineError struct {
		Type      string `json:"type"`
		Message   string `json:"message"`
		IsWarning bool   `json:"is_warning"`
		Data      any    `json:"data"`
	}

	// Version provides system version details.
	Version struct {
		Source  string `json:"source,omitempty"`
		Version string `json:"version,omitempty"`
		Commit  string `json:"commit,omitempty"`
	}

	// LogLevel is for checking/setting logging level.
	LogLevel struct {
		Level string `json:"log-level"`
	}
)
