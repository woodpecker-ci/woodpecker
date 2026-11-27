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

import "strings"

// MatchLabels reports if an agent with the given labels may execute a
// workflow with the given labels, and how well it fits: the more labels match
// exactly instead of by wildcard, the higher the score.
func MatchLabels(workflowLabels, agentLabels map[string]string) (matches bool, score int) {
	if requiredLabelsMissing(workflowLabels, agentLabels) {
		return false, 0
	}

	for label, value := range workflowLabels {
		// internal labels are not for filtering, and empty ones are ignored
		if strings.HasPrefix(label, InternalLabelPrefix) || value == "" {
			continue
		}

		// all workflow labels are required to be present for an agent to match
		agentValue, ok := agentLabels[label]
		if !ok {
			// Check for required label
			agentValue, ok = agentLabels["!"+label]
			if !ok {
				return false, 0
			}
		}

		switch agentValue {
		// if agent label has a wildcard
		case "*":
			score++
		// if agent label has an exact match
		case value:
			score += 10
		// agent doesn't match
		default:
			return false, 0
		}
	}
	return true, score
}

// requiredLabelsMissing reports if the workflow lacks a label the agent
// requires, which are the agent labels starting with "!".
func requiredLabelsMissing(workflowLabels, agentLabels map[string]string) bool {
	for label, value := range agentLabels {
		if len(label) > 0 && label[0] == '!' {
			val, ok := workflowLabels[label[1:]]
			if !ok || val != value {
				return true
			}
		}
	}
	return false
}
