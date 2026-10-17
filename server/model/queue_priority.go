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

	"github.com/bmatcuk/doublestar/v4"
	"go.yaml.in/yaml/v4"
)

// QueuePriorityRule adds or subtracts priority from a pipeline's queued tasks
// when all configured match fields match. Higher resulting task priority runs
// earlier in the queue.
type QueuePriorityRule struct {
	Priority      int          `yaml:"priority"`
	Repo          string       `yaml:"repo"`
	Event         WebhookEvent `yaml:"event"`
	Branch        string       `yaml:"branch"`
	Ref           string       `yaml:"ref"`
	Author        string       `yaml:"author"`
	Sender        string       `yaml:"sender"`
	PullLabel     string       `yaml:"pr_label"`
	EventReason   string       `yaml:"event_reason"`
	MinRerunCount int64        `yaml:"min_rerun_count"`
}

type queuePriorityRuleFile struct {
	Rules []queuePriorityRuleConfig `yaml:"rules"`
}

type queuePriorityRuleConfig struct {
	Priority      *int         `yaml:"priority"`
	Repo          string       `yaml:"repo"`
	Event         WebhookEvent `yaml:"event"`
	Branch        string       `yaml:"branch"`
	Ref           string       `yaml:"ref"`
	Author        string       `yaml:"author"`
	Sender        string       `yaml:"sender"`
	PullLabel     string       `yaml:"pr_label"`
	EventReason   string       `yaml:"event_reason"`
	MinRerunCount int64        `yaml:"min_rerun_count"`
}

// ParseQueuePriorityRuleFile parses a YAML queue priority config.
func ParseQueuePriorityRuleFile(data []byte) ([]QueuePriorityRule, error) {
	var file queuePriorityRuleFile
	if err := yaml.Unmarshal(data, &file); err != nil {
		return nil, err
	}
	rules := make([]QueuePriorityRule, 0, len(file.Rules))
	for i := range file.Rules {
		rule, err := file.Rules[i].asRule()
		if err != nil {
			return nil, fmt.Errorf("invalid rule %d: %w", i+1, err)
		}
		rules = append(rules, rule)
	}
	return rules, nil
}

func (r queuePriorityRuleConfig) asRule() (QueuePriorityRule, error) {
	if r.Priority == nil {
		return QueuePriorityRule{}, fmt.Errorf("missing priority")
	}
	rule := QueuePriorityRule{
		Priority:      *r.Priority,
		Repo:          r.Repo,
		Event:         r.Event,
		Branch:        r.Branch,
		Ref:           r.Ref,
		Author:        r.Author,
		Sender:        r.Sender,
		PullLabel:     r.PullLabel,
		EventReason:   r.EventReason,
		MinRerunCount: r.MinRerunCount,
	}
	return rule, rule.Validate()
}

// Validate checks a queue priority rule loaded from config.
func (r QueuePriorityRule) Validate() error {
	if r.Event != "" {
		if err := r.Event.Validate(); err != nil {
			return err
		}
	}
	for key, pattern := range map[string]string{
		"repo":         r.Repo,
		"branch":       r.Branch,
		"ref":          r.Ref,
		"author":       r.Author,
		"sender":       r.Sender,
		"pr_label":     r.PullLabel,
		"event_reason": r.EventReason,
	} {
		if pattern != "" {
			if err := validateQueuePriorityGlob(key, pattern); err != nil {
				return err
			}
		}
	}
	if r.MinRerunCount < 0 {
		return fmt.Errorf("min_rerun_count must be greater than or equal to zero")
	}
	return nil
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
