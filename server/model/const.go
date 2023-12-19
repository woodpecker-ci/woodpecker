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

import (
	"errors"
	"fmt"
)

type WebhookEvent string //	@name WebhookEvent

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

var ErrInvalidWebhookEvent = errors.New("invalid webhook event")

func ValidateWebhookEvent(s WebhookEvent) error {
	switch s {
	case EventPush, EventPull, EventTag, EventDeploy, EventCron, EventManual:
		return nil
	default:
		return fmt.Errorf("%w: %s", ErrInvalidWebhookEvent, s)
	}
}

// StatusValue represent pipeline states woodpecker know
type StatusValue string //	@name StatusValue

const (
	StatusSkipped  StatusValue = "skipped"  // skipped as another step failed
	StatusPending  StatusValue = "pending"  // pending to be executed
	StatusRunning  StatusValue = "running"  // currently running
	StatusSuccess  StatusValue = "success"  // successfully finished
	StatusFailure  StatusValue = "failure"  // failed to finish (exit code != 0)
	StatusKilled   StatusValue = "killed"   // killed by user
	StatusError    StatusValue = "error"    // error with the config / while parsing / some other system problem
	StatusBlocked  StatusValue = "blocked"  // waiting for approval
	StatusDeclined StatusValue = "declined" // blocked and declined
	StatusCreated  StatusValue = "created"  // created / internal use only
)

// SCMKind represent different version control systems
type SCMKind string //	@name SCMKind

const (
	RepoGit      SCMKind = "git"
	RepoHg       SCMKind = "hg"
	RepoFossil   SCMKind = "fossil"
	RepoPerforce SCMKind = "perforce"
)

// RepoVisibility represent to wat state a repo in woodpecker is visible to others
type RepoVisibility string //	@name RepoVisibility

const (
	VisibilityPublic   RepoVisibility = "public"
	VisibilityPrivate  RepoVisibility = "private"
	VisibilityInternal RepoVisibility = "internal"
)
