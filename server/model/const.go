// Copyright 2022 Woodpecker Authors
// Copyright 2018 Drone.IO Inc.
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

type WebhookEvent string

const (
	EventPush   WebhookEvent = "push"
	EventPull   WebhookEvent = "pull_request"
	EventTag    WebhookEvent = "tag"
	EventDeploy WebhookEvent = "deployment"
	EventCron   WebhookEvent = "cron"
	EventManual WebhookEvent = "manual"
)

type WebhookEventList []WebhookEvent

func (wel WebhookEventList) Len() int           { return len(wel) }
func (wel WebhookEventList) Swap(i, j int)      { wel[i], wel[j] = wel[j], wel[i] }
func (wel WebhookEventList) Less(i, j int) bool { return wel[i] < wel[j] }

func ValidateWebhookEvent(s WebhookEvent) bool {
	switch s {
	case EventPush, EventPull, EventTag, EventDeploy, EventCron, EventManual:
		return true
	default:
		return false
	}
}

// StatusValue represent pipeline states woodpecker know
type StatusValue string

const (
	StatusSkipped  StatusValue = "skipped"
	StatusPending  StatusValue = "pending"
	StatusRunning  StatusValue = "running"
	StatusSuccess  StatusValue = "success"
	StatusFailure  StatusValue = "failure"
	StatusKilled   StatusValue = "killed"
	StatusError    StatusValue = "error"
	StatusBlocked  StatusValue = "blocked"
	StatusDeclined StatusValue = "declined"
)

// SCMKind represent different version control systems
type SCMKind string

const (
	RepoGit      SCMKind = "git"
	RepoHg       SCMKind = "hg"
	RepoFossil   SCMKind = "fossil"
	RepoPerforce SCMKind = "perforce"
)

// RepoVisibility represent to wat state a repo in woodpecker is visible to others
type RepoVisibility string

const (
	VisibilityPublic   RepoVisibility = "public"
	VisibilityPrivate  RepoVisibility = "private"
	VisibilityInternal RepoVisibility = "internal"
)
