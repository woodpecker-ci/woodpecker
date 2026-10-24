// Copyright 2026 Woodpecker Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package healing

import (
	"context"
	"fmt"
	"strings"
)

// ConfigIssue describes a detected flaw or policy violation in a pipeline configuration.
type ConfigIssue struct {
	RuleID      string `json:"rule_id"`
	Description string `json:"description"`
	Severity    string `json:"severity"` // "WARNING", "ERROR", "CRITICAL"
}

// ConfigHealingSuggestion contains the analyzed YAML, proposed remediation, and unified diff.
type ConfigHealingSuggestion struct {
	OriginalYAML    string        `json:"original_yaml"`
	RemediatedYAML  string        `json:"remediated_yaml"`
	UnifiedDiff     string        `json:"unified_diff"`
	IssuesDetected  []ConfigIssue `json:"issues_detected"`
	RequiresCSOReview bool        `json:"requires_cso_review"`
}

// ConfigRepairer analyzes and repairs Woodpecker YAML configs.
type ConfigRepairer interface {
	AnalyzeAndRepair(ctx context.Context, yamlContent string) (*ConfigHealingSuggestion, error)
}

// DefaultConfigRepairer handles automated config healing for clinical CI pipelines.
type DefaultConfigRepairer struct{}

// NewDefaultConfigRepairer creates a new config repairer instance.
func NewDefaultConfigRepairer() *DefaultConfigRepairer {
	return &DefaultConfigRepairer{}
}

// AnalyzeAndRepair inspects pipeline YAML for security and compliance issues and generates a remediation diff.
func (r *DefaultConfigRepairer) AnalyzeAndRepair(ctx context.Context, yamlContent string) (*ConfigHealingSuggestion, error) {
	suggestion := &ConfigHealingSuggestion{
		OriginalYAML: yamlContent,
	}

	remediated := yamlContent
	var issues []ConfigIssue

	// 1. Check for Zero-Mock L-CAID-10 violations (e.g. mock test flags, ?demo=true, x-test-bypass)
	if strings.Contains(remediated, "--mock") || strings.Contains(remediated, "jest.mock") || strings.Contains(remediated, "?demo=true") {
		issues = append(issues, ConfigIssue{
			RuleID:      "L-CAID-10",
			Description: "Prohibited synthetic mock bypass detected in clinical pipeline step",
			Severity:    "CRITICAL",
		})
		remediated = strings.ReplaceAll(remediated, "--mock", "--real-services")
		remediated = strings.ReplaceAll(remediated, "?demo=true", "")
		suggestion.RequiresCSOReview = true
	}

	// 2. Check for missing Tink AEAD KMS keyset in secrets or environment
	if !strings.Contains(remediated, "WOODPECKER_ENCRYPTION_TINK_KEYSET_FILE") && strings.Contains(remediated, "secret") {
		issues = append(issues, ConfigIssue{
			RuleID:      "SEC-TINK-01",
			Description: "Plaintext secrets used without mandatory Google Tink AEAD at-rest encryption",
			Severity:    "ERROR",
		})
		// Inject Tink keyset env into secrets configuration
		if strings.Contains(remediated, "environment:") {
			remediated = strings.Replace(remediated, "environment:", "environment:\n      WOODPECKER_ENCRYPTION_TINK_KEYSET_FILE: /etc/woodpecker/keyset.json", 1)
		}
	}

	// 3. Check for deprecated or privileged container execution
	if strings.Contains(remediated, "privileged: true") {
		issues = append(issues, ConfigIssue{
			RuleID:      "SEC-SANDBOX-01",
			Description: "Privileged container execution violates clinical air-gap sandbox policy",
			Severity:    "CRITICAL",
		})
		remediated = strings.ReplaceAll(remediated, "privileged: true", "privileged: false")
		suggestion.RequiresCSOReview = true
	}

	suggestion.RemediatedYAML = remediated
	suggestion.IssuesDetected = issues
	suggestion.UnifiedDiff = generateSimpleDiff(yamlContent, remediated)

	return suggestion, nil
}

// generateSimpleDiff creates a standard diff representation between original and remediated YAML.
func generateSimpleDiff(original, remediated string) string {
	if original == remediated {
		return ""
	}

	var sb strings.Builder
	sb.WriteString("--- .woodpecker.yaml (original)\n")
	sb.WriteString("+++ .woodpecker.yaml (healed)\n")
	sb.WriteString("@@ -1 +1 @@\n")

	origLines := strings.Split(original, "\n")
	remLines := strings.Split(remediated, "\n")

	for _, l := range origLines {
		if !containsLine(remLines, l) {
			sb.WriteString(fmt.Sprintf("-%s\n", l))
		}
	}
	for _, l := range remLines {
		if !containsLine(origLines, l) {
			sb.WriteString(fmt.Sprintf("+%s\n", l))
		}
	}

	return sb.String()
}

func containsLine(lines []string, target string) bool {
	for _, l := range lines {
		if l == target {
			return true
		}
	}
	return false
}
