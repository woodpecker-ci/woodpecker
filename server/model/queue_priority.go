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
	"fmt"
	"strconv"
	"strings"

	"github.com/bmatcuk/doublestar/v4"
)

// QueuePriorityRule adds or subtracts priority from a pipeline's queued tasks
// when all configured match fields match. Higher resulting task priority runs
// earlier in the queue.
type QueuePriorityRule struct {
	Priority      int
	Repo          string
	Event         WebhookEvent
	Branch        string
	Ref           string
	Author        string
	Sender        string
	PullLabel     string
	EventReason   string
	MinRerunCount int64
}

// ParseQueuePriorityRules parses a list of whitespace-separated key=value rule
// strings. Every rule must include priority=<int>.
func ParseQueuePriorityRules(values []string) ([]QueuePriorityRule, error) {
	rules := make([]QueuePriorityRule, 0, len(values))
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value == "" {
			continue
		}

		rule, err := parseQueuePriorityRule(value)
		if err != nil {
			return nil, err
		}
		rules = append(rules, rule)
	}
	return rules, nil
}

func parseQueuePriorityRule(value string) (QueuePriorityRule, error) {
	var rule QueuePriorityRule
	hasPriority := false

	for _, field := range strings.Fields(value) {
		key, raw, ok := strings.Cut(field, "=")
		if !ok || strings.TrimSpace(key) == "" {
			return rule, fmt.Errorf("invalid queue priority rule field %q in %q", field, value)
		}
		if raw == "" {
			return rule, fmt.Errorf("empty queue priority rule value for %q in %q", key, value)
		}

		switch strings.ReplaceAll(strings.ToLower(key), "-", "_") {
		case "priority":
			priority, err := strconv.Atoi(raw)
			if err != nil {
				return rule, fmt.Errorf("invalid queue priority %q in %q: %w", raw, value, err)
			}
			rule.Priority = priority
			hasPriority = true
		case "repo":
			if err := validateQueuePriorityGlob(key, raw); err != nil {
				return rule, err
			}
			rule.Repo = raw
		case "event":
			event := WebhookEvent(raw)
			if err := event.Validate(); err != nil {
				return rule, err
			}
			rule.Event = event
		case "branch":
			if err := validateQueuePriorityGlob(key, raw); err != nil {
				return rule, err
			}
			rule.Branch = raw
		case "ref":
			if err := validateQueuePriorityGlob(key, raw); err != nil {
				return rule, err
			}
			rule.Ref = raw
		case "author":
			if err := validateQueuePriorityGlob(key, raw); err != nil {
				return rule, err
			}
			rule.Author = raw
		case "sender":
			if err := validateQueuePriorityGlob(key, raw); err != nil {
				return rule, err
			}
			rule.Sender = raw
		case "pr_label", "pull_label", "pull_request_label":
			if err := validateQueuePriorityGlob(key, raw); err != nil {
				return rule, err
			}
			rule.PullLabel = raw
		case "event_reason":
			if err := validateQueuePriorityGlob(key, raw); err != nil {
				return rule, err
			}
			rule.EventReason = raw
		case "min_rerun_count":
			min, err := strconv.ParseInt(raw, 10, 64)
			if err != nil {
				return rule, fmt.Errorf("invalid queue priority min_rerun_count %q in %q: %w", raw, value, err)
			}
			if min < 0 {
				return rule, fmt.Errorf("invalid queue priority min_rerun_count %q in %q: must be greater than or equal to zero", raw, value)
			}
			rule.MinRerunCount = min
		default:
			return rule, fmt.Errorf("unknown queue priority rule field %q in %q", key, value)
		}
	}

	if !hasPriority {
		return rule, fmt.Errorf("queue priority rule %q is missing priority", value)
	}
	return rule, nil
}

func validateQueuePriorityGlob(key, pattern string) error {
	if _, err := doublestar.Match(pattern, ""); err != nil {
		return fmt.Errorf("invalid queue priority glob for %q: %w", key, err)
	}
	return nil
}

// QueuePriority returns the sum of all matching priority rules.
func QueuePriority(repo *Repo, pipeline *Pipeline, rules []QueuePriorityRule) int {
	priority := 0
	for _, rule := range rules {
		if rule.Match(repo, pipeline) {
			priority += rule.Priority
		}
	}
	return priority
}

// Match reports whether the rule matches the pipeline metadata.
func (r QueuePriorityRule) Match(repo *Repo, pipeline *Pipeline) bool {
	if repo == nil || pipeline == nil {
		return false
	}
	if r.Repo != "" && !globMatch(r.Repo, repo.FullName) {
		return false
	}
	if r.Event != "" && r.Event != pipeline.Event {
		return false
	}
	if r.Branch != "" && !globMatch(r.Branch, pipeline.Branch) {
		return false
	}
	if r.Ref != "" && !globMatch(r.Ref, pipeline.Ref) {
		return false
	}
	if r.Author != "" && !globMatch(r.Author, pipeline.Author) {
		return false
	}
	if r.Sender != "" && !globMatch(r.Sender, pipeline.Sender) {
		return false
	}
	if r.PullLabel != "" && !matchAny(r.PullLabel, pipeline.PullRequestLabels) {
		return false
	}
	if r.EventReason != "" && !matchAny(r.EventReason, pipeline.EventReason) {
		return false
	}
	if r.MinRerunCount > 0 && pipeline.RerunCount < r.MinRerunCount {
		return false
	}
	return true
}

func matchAny(pattern string, values []string) bool {
	for _, value := range values {
		if globMatch(pattern, value) {
			return true
		}
	}
	return false
}

func globMatch(pattern, value string) bool {
	matched, err := doublestar.Match(pattern, value)
	return err == nil && matched
}
