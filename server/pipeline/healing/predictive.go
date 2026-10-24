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
	"math"
	"sort"
	"sync"
	"time"
)

// StepMetricRecord captures peak memory and execution telemetry for a step.
type StepMetricRecord struct {
	RepoID      int64     `json:"repo_id"`
	StepName    string    `json:"step_name"`
	PeakMemoryMB int64     `json:"peak_memory_mb"`
	DurationSec float64   `json:"duration_sec"`
	ExitCode    int       `json:"exit_code"`
	RecordedAt  time.Time `json:"recorded_at"`
}

// MemoryPredictor provides pre-emptive memory rightsizing to prevent OOM before container start.
type MemoryPredictor interface {
	RecordStepMetric(ctx context.Context, metric StepMetricRecord) error
	PredictStepMemory(ctx context.Context, repoID int64, stepName string, defaultLimitMB int64) (int64, bool)
}

// TelemetryMemoryPredictor stores historical metrics and calculates P95 headroom.
type TelemetryMemoryPredictor struct {
	mu      sync.RWMutex
	metrics map[string][]StepMetricRecord // key: "repoID:stepName"
}

// NewTelemetryMemoryPredictor creates a new predictor.
func NewTelemetryMemoryPredictor() *TelemetryMemoryPredictor {
	return &TelemetryMemoryPredictor{
		metrics: make(map[string][]StepMetricRecord),
	}
}

func makeMetricKey(repoID int64, stepName string) string {
	return stepName // Can be scoped by repo or global for step
}

// RecordStepMetric records execution metrics for future predictions.
func (p *TelemetryMemoryPredictor) RecordStepMetric(ctx context.Context, metric StepMetricRecord) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	key := makeMetricKey(metric.RepoID, metric.StepName)
	if metric.RecordedAt.IsZero() {
		metric.RecordedAt = time.Now().UTC()
	}

	p.metrics[key] = append(p.metrics[key], metric)
	return nil
}

// PredictStepMemory calculates the P95 memory requirement with a 20% safety headroom.
// Returns (recommendedMB, wasRightsized).
func (p *TelemetryMemoryPredictor) PredictStepMemory(ctx context.Context, repoID int64, stepName string, defaultLimitMB int64) (int64, bool) {
	p.mu.RLock()
	defer p.mu.RUnlock()

	key := makeMetricKey(repoID, stepName)
	records, exists := p.metrics[key]
	if !exists || len(records) == 0 {
		return defaultLimitMB, false
	}

	// Extract peak memory values from successful or non-killed runs
	var memorySamples []int64
	for _, rec := range records {
		if rec.PeakMemoryMB > 0 {
			memorySamples = append(memorySamples, rec.PeakMemoryMB)
		}
	}

	if len(memorySamples) == 0 {
		return defaultLimitMB, false
	}

	sort.Slice(memorySamples, func(i, j int) bool {
		return memorySamples[i] < memorySamples[j]
	})

	// Calculate 95th percentile index
	p95Idx := int(math.Ceil(0.95*float64(len(memorySamples)))) - 1
	if p95Idx < 0 {
		p95Idx = 0
	}
	if p95Idx >= len(memorySamples) {
		p95Idx = len(memorySamples) - 1
	}

	p95Memory := memorySamples[p95Idx]

	// Add 20% safety margin
	recommended := int64(math.Ceil(float64(p95Memory) * 1.20))

	// If the recommended memory exceeds defaultLimitMB, pre-emptively scale up!
	if recommended > defaultLimitMB {
		return recommended, true
	}

	return defaultLimitMB, false
}
