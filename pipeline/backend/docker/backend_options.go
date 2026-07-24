// Copyright 2024 Woodpecker Authors
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

package docker

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/go-viper/mapstructure/v2"

	backend_types "go.woodpecker-ci.org/woodpecker/v3/pipeline/backend/types"
)

const (
	oomScoreAdjMin = 0
	oomScoreAdjMax = 1000
)

// BackendOptions defines all the advanced options for the docker backend.
type BackendOptions struct {
	User        string    `mapstructure:"user"`
	Resources   Resources `mapstructure:"resources"`
	OomScoreAdj int       `mapstructure:"oom_score_adj"`
}

// Resources holds resource limit configuration for a step.
type Resources struct {
	Limits ResourceList `mapstructure:"limits"`
}

// ResourceList holds the specific resource limits.
type ResourceList struct {
	Memory string  `mapstructure:"memory"` // e.g. "512m", "1g", "536870912"
	CPUs   float64 `mapstructure:"cpus"`   // e.g. 0.5, 1.0, 2.0
}

func parseBackendOptions(step *backend_types.Step) (BackendOptions, error) {
	var result BackendOptions
	if step == nil || step.BackendOptions == nil {
		return result, nil
	}
	if err := mapstructure.WeakDecode(step.BackendOptions[EngineName], &result); err != nil {
		return result, err
	}

	if result.OomScoreAdj < oomScoreAdjMin || result.OomScoreAdj > oomScoreAdjMax {
		return BackendOptions{}, fmt.Errorf("oom_score_adj must be in [%d, %d], got %d", oomScoreAdjMin, oomScoreAdjMax, result.OomScoreAdj)
	}

	if result.Resources.Limits.Memory != "" {
		if _, err := parseMemory(result.Resources.Limits.Memory); err != nil {
			return BackendOptions{}, fmt.Errorf("invalid memory limit %q: %w", result.Resources.Limits.Memory, err)
		}
	}

	return result, nil
}

// parseMemory converts a human-readable memory string to bytes.
// Supported suffixes: k/K (kibibytes), m/M (mebibytes), g/G (gibibytes).
// A bare integer is treated as bytes.
func parseMemory(s string) (int64, error) {
	if s == "" {
		return 0, nil
	}

	const (
		kibibyte = int64(1024)
		mebibyte = kibibyte * 1024
		gibibyte = mebibyte * 1024
	)

	lower := strings.ToLower(s)
	var multiplier int64
	var numStr string

	switch {
	case strings.HasSuffix(lower, "g"):
		multiplier = gibibyte
		numStr = s[:len(s)-1]
	case strings.HasSuffix(lower, "m"):
		multiplier = mebibyte
		numStr = s[:len(s)-1]
	case strings.HasSuffix(lower, "k"):
		multiplier = kibibyte
		numStr = s[:len(s)-1]
	default:
		multiplier = 1
		numStr = s
	}

	val, err := strconv.ParseInt(numStr, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("cannot parse memory value %q: %w", s, err)
	}
	return val * multiplier, nil
}
