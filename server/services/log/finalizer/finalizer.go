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

// Package finalizer calls StepFinished of the log store for every step whose log is complete.
package finalizer

import (
	"context"
	"errors"
	"fmt"
	"math/rand/v2"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/rs/zerolog/log"

	"go.woodpecker-ci.org/woodpecker/v3/server/model"
	service_log "go.woodpecker-ci.org/woodpecker/v3/server/services/log"
	"go.woodpecker-ci.org/woodpecker/v3/server/store"
)

// stepTimeout bounds finalizing one step, so a hung log store cannot stall the finalizer.
var stepTimeout = time.Minute

// A claim on a step log counts as left behind by a server that crashed after recoverDelay plus
// up to recoverJitter (so servers do not all recover at once): finalizing or deleting takes a minute.
var (
	recoverDelay  = 30 * time.Minute
	recoverJitter = 10 * time.Minute
)

var unfinalizedSteps = promauto.NewGauge(prometheus.GaugeOpts{
	Namespace: "woodpecker",
	Name:      "log_steps_unfinalized",
	Help:      "Number of steps whose log is complete but not yet finalized.",
})

// Finalizer is the only caller of StepFinished.
// It remembers which steps were finalized, what the log store does with the log is its own business.
// Several servers may each run one: they claim a step log before they work on it.
// It is not safe for concurrent use.
type Finalizer struct {
	Store    store.Store
	Logs     service_log.Service
	Interval time.Duration
	// Batch is the maximum number of pipelines handled per tick.
	Batch  int
	Paused bool

	// A pass goes through the pipelines due when it started, a batch per tick:
	// after is the pipeline ID the next tick continues after, until the last one of the pass.
	after, until int64
}

// Result counts the steps a tick finalized or failed on.
type Result struct {
	Done   int
	Failed int
}

// before returns the finish time a pipeline and step must be older than for their log to be complete:
// the late-log window has passed and so has the time an append admitted just before may take.
func before(now time.Time) int64 {
	return now.Add(-service_log.StreamDelayAllowed - service_log.AppendTimeout).Unix()
}

// Tick finalizes one batch of pipelines. Passes keep steps that fail from blocking newer pipelines
// and newer pipelines from keeping failed steps from being retried.
func (f *Finalizer) Tick(ctx context.Context, now time.Time) (Result, error) {
	var res Result

	finishedBefore := before(now)
	count, maxPipelineID, err := f.Store.StepLogUnfinalizedStats(ctx, finishedBefore)
	if err != nil {
		return res, err
	}
	unfinalizedSteps.Set(float64(count))
	if f.Paused {
		return res, nil
	}

	if f.after == 0 {
		f.until = maxPipelineID
	}
	steps, err := f.Store.StepListLogUnfinalized(ctx, finishedBefore, f.after, f.until, f.Batch)
	if err != nil {
		return res, err
	}

	for _, step := range steps {
		if err := ctx.Err(); err != nil {
			return res, err
		}
		finalized, err := f.finalize(ctx, step)
		switch {
		case err != nil:
			log.Error().Err(err).Int64("step", step.ID).Msg("could not finalize step log")
			res.Failed++
		case finalized:
			res.Done++
		}
	}

	// the next tick continues the pass after the last pipeline, or starts a new one
	f.after = 0
	if n := len(steps); n > 0 && steps[n-1].PipelineID < f.until {
		f.after = steps[n-1].PipelineID
	}

	if res.Done > 0 || res.Failed > 0 {
		log.Info().Int("finalized", res.Done).Int("failed", res.Failed).Msg("finalized step logs")
	}
	return res, nil
}

// finalize claims a step log, hands it to the log store and marks it finalized.
// It reports false for a log another server or a delete claimed first.
func (f *Finalizer) finalize(ctx context.Context, step *model.Step) (bool, error) {
	ctx, cancel := context.WithTimeout(ctx, stepTimeout)
	defer cancel()

	if claimed, err := f.Store.StepLogStateSwap(ctx, step, model.LogStateFinalizing); !claimed {
		return false, err
	}

	// A step that never started has no log, it is only marked.
	if step.Started != 0 {
		if err := f.Logs.StepFinished(ctx, step); err != nil {
			// hand it back to the next pass, else RecoverClaims does
			releaseCtx, cancel := service_log.ReleaseContext(ctx)
			defer cancel()
			_, reopenErr := f.Store.StepLogStateSwap(releaseCtx, step, model.LogStateOpen)
			return false, errors.Join(err, reopenErr)
		}
	}
	_, err := f.Store.StepLogStateSwap(ctx, step, model.LogStateFinalized)
	return err == nil, err
}

// RecoverClaims hands back the step logs a crashed server left claimed (finalizing or deleting):
// it lists the claims, waits 30 to 40 minutes and reopens those still in the same state, until ctx is done.
func RecoverClaims(ctx context.Context, s store.Store) {
	for {
		claimed, err := s.StepListLogClaimed(ctx)
		if err != nil && ctx.Err() == nil {
			log.Error().Err(err).Msg("could not list claimed step logs")
		}

		select {
		case <-ctx.Done():
			return
		case <-time.After(recoverDelay + rand.N(recoverJitter)): //nolint:gosec // jitter, not security
		}

		for _, step := range claimed {
			// only swaps if the state is still the same
			if _, err := s.StepLogStateSwap(ctx, step, model.LogStateOpen); err != nil && ctx.Err() == nil {
				log.Error().Err(err).Int64("step", step.ID).Msg("could not reopen claimed step log")
			}
		}
	}
}

// Run checks the configuration and ticks every Interval until the context is done.
func (f *Finalizer) Run(ctx context.Context) error {
	if f.Interval <= 0 {
		return fmt.Errorf("log finalizer interval must be positive, got %s", f.Interval)
	}
	if f.Batch < 1 {
		return fmt.Errorf("log finalizer batch must be at least 1, got %d", f.Batch)
	}

	for {
		select {
		case <-ctx.Done():
			return nil
		case <-time.After(f.Interval):
			if _, err := f.Tick(ctx, time.Now()); err != nil && ctx.Err() == nil {
				log.Error().Err(err).Msg("log finalizer tick failed")
			}
		}
	}
}
