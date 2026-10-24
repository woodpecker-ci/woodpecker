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

package runtime

import (
	"fmt"
	"math"
	"strconv"
	"strings"

	backend_types "go.woodpecker-ci.org/woodpecker/v3/pipeline/backend/types"
)

const (
	defaultOOMMaxRetries   = 2
	defaultFallbackMemLimit = "2Gi"
)

// isOOMSelfHealingEnabled checks whether Layer 1 OOM self-healing is active for the step.
// Enabled by default unless explicitly disabled via CI_SELF_HEALING_OOM=false.
func (r *Runtime) isOOMSelfHealingEnabled(step *backend_types.Step) (bool, int) {
	if step.Environment != nil {
		if val, ok := step.Environment["CI_SELF_HEALING_OOM"]; ok {
			if val == "false" || val == "0" {
				return false, 0
			}
		}
		if val, ok := step.Environment["CI_SELF_HEALING_OOM_MAX_RETRIES"]; ok {
			if retries, err := strconv.Atoi(val); err == nil && retries >= 0 {
				return retries > 0, retries
			}
		}
	}
	return true, defaultOOMMaxRetries
}

// escalateStepMemory dynamically increases memory limits and requests for Kubernetes/Docker backends.
func (r *Runtime) escalateStepMemory(step *backend_types.Step) {
	logger := r.makeLogger()

	if step.BackendOptions == nil {
		step.BackendOptions = make(map[string]any)
	}

	// 1. Kubernetes resources escalation
	resRaw, exists := step.BackendOptions["resources"]
	if exists {
		if resMap, ok := resRaw.(map[string]any); ok {
			if limitsRaw, hasLimits := resMap["limits"]; hasLimits {
				if limitsMap, ok := limitsRaw.(map[string]string); ok {
					if mem, hasMem := limitsMap["memory"]; hasMem {
						newMem := scaleMemoryString(mem)
						limitsMap["memory"] = newMem
						logger.Info().Str("step", step.Name).Str("old_memory", mem).Str("new_memory", newMem).Msg("Escalated Kubernetes memory limit")
					}
				}
			}
			if reqRaw, hasReq := resMap["requests"]; hasReq {
				if reqMap, ok := reqRaw.(map[string]string); ok {
					if mem, hasMem := reqMap["memory"]; hasMem {
						newMem := scaleMemoryString(mem)
						reqMap["memory"] = newMem
					}
				}
			}
		}
	} else {
		// If no resources were set, provision a safe fallback
		step.BackendOptions["resources"] = map[string]any{
			"limits": map[string]string{
				"memory": defaultFallbackMemLimit,
			},
			"requests": map[string]string{
				"memory": "1Gi",
			},
		}
		logger.Info().Str("step", step.Name).Str("allocated", defaultFallbackMemLimit).Msg("Provisioned initial Kubernetes memory limit after OOM")
	}

	// 2. Mark environment so child processes and step telemetry know an escalation occurred
	if step.Environment == nil {
		step.Environment = make(map[string]string)
	}
	step.Environment["CI_HEALED_FROM_OOM"] = "true"
}

// scaleMemoryString takes a Kubernetes resource string (e.g. 512Mi, 2Gi) and scales it up.
func scaleMemoryString(mem string) string {
	mem = strings.TrimSpace(mem)
	if mem == "" {
		return defaultFallbackMemLimit
	}

	suffix := ""
	for _, s := range []string{"Gi", "Mi", "Ki", "G", "M", "K"} {
		if strings.HasSuffix(mem, s) {
			suffix = s
			break
		}
	}

	if suffix == "" {
		// Attempt numeric parse as raw bytes
		if bytesVal, err := strconv.ParseInt(mem, 10, 64); err == nil && bytesVal > 0 {
			return strconv.FormatInt(bytesVal*2, 10)
		}
		return defaultFallbackMemLimit
	}

	numPart := strings.TrimSuffix(mem, suffix)
	val, err := strconv.ParseFloat(numPart, 64)
	if err != nil || val <= 0 {
		return defaultFallbackMemLimit
	}

	// Scale by 1.5x (or 2x for <= 1Gi)
	var multiplier float64 = 1.5
	if val <= 1.0 {
		multiplier = 2.0
	}

	newVal := math.Ceil(val * multiplier)
	return fmt.Sprintf("%d%s", int64(newVal), suffix)
}
