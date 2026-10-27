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
	"fmt"
	"math/rand/v2"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"go.woodpecker-ci.org/woodpecker/v3/server/model"
	service_log "go.woodpecker-ci.org/woodpecker/v3/server/services/log"
	"go.woodpecker-ci.org/woodpecker/v3/server/store"
)

var (
	// Finalizing one step takes at most stepTimeout, so a hung log store cannot stall the finalizer.
	stepTimeout = time.Minute

	// A claim on a step log counts as left behind by a crash after recoverDelay plus
	// up to recoverJitter (so recoveries do not all happen at once): finalizing or deleting takes a minute.
	recoverDelay  = 30 * time.Minute
	recoverJitter = 10 * time.Minute

	// Run accepts no interval shorter than minInterval.
	minInterval = time.Minute

	// Prometheus gauge of the steps whose log is complete but not yet finalized.
	unfinalizedSteps = promauto.NewGauge(prometheus.GaugeOpts{
		Namespace: "woodpecker",
		Name:      "log_steps_unfinalized",
		Help:      "Number of steps whose log is complete but not yet finalized.",
	})

	// Prometheus counter of the steps ticks finalized (result done) or failed on (result failed).
	finalizedSteps = promauto.NewCounterVec(prometheus.CounterOpts{
		Namespace: "woodpecker",
		Name:      "log_steps_finalized_total",
		Help:      "Number of steps whose log was finalized, by result (done or failed).",
	}, []string{"result"})
)

// Finalizer is the only caller of StepFinished.
// It remembers which steps were finalized, what the log store does with the log is its own business.
// Several finalizers may run concurrently: they claim a step log before they work on it.
type Finalizer struct {
	Store    store.Store
	Logs     service_log.Service
	Interval time.Duration
	// Batch is the maximum number of pipelines handled per tick.
	Batch  int
	Paused bool
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

// Tick finalizes the steps of the oldest Batch pipelines due. A step that fails stays claimed
// until RecoverClaims hands it back, so it cannot block newer pipelines and is retried then.
func (f *Finalizer) Tick(ctx context.Context, now time.Time) (Result, error) {
	var res Result

	due := &model.StepLogFilter{States: []model.LogState{model.LogStateOpen}, CompleteBefore: before(now)}
	count, err := f.Store.StepCountByLogState(ctx, due)
	if err != nil {
		return res, err
	}
	unfinalizedSteps.Set(float64(count))
	if f.Paused {
		return res, nil
	}

	due.Pipelines = f.Batch
	steps, err := f.Store.StepListByLogState(ctx, due)
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
			finalizedSteps.WithLabelValues("failed").Inc()
		case finalized:
			res.Done++
			finalizedSteps.WithLabelValues("done").Inc()
		}
	}

	if res.Done > 0 || res.Failed > 0 {
		level := zerolog.DebugLevel
		if res.Failed > 0 {
			level = zerolog.WarnLevel
		}
		log.WithLevel(level).Int("finalized", res.Done).Int("failed", res.Failed).Msg("finalized step logs")
	}
	return res, nil
}

// finalize claims a step log, hands it to the log store and marks it finalized.
// It reports false for a log a concurrent finalize or delete claimed first.
func (f *Finalizer) finalize(ctx context.Context, step *model.Step) (bool, error) {
	ctx, cancel := context.WithTimeout(ctx, stepTimeout)
	defer cancel()

	if claimed, err := f.Store.StepLogStateSwap(ctx, step, model.LogStateFinalizing); !claimed {
		return false, err
	}

	// A step that never started has no log, it is only marked.
	if step.Started != 0 {
		if err := f.Logs.StepFinished(ctx, step); err != nil {
			// stays claimed, RecoverClaims hands it back later
			return false, err
		}
	}
	_, err := f.Store.StepLogStateSwap(ctx, step, model.LogStateFinalized)
	return err == nil, err
}

// RecoverClaims hands back the step logs left claimed (finalizing or deleting) by a failed finalize or a crash:
// it lists the claims, waits 30 to 40 minutes and reopens those still in the same state, until ctx is done.
func RecoverClaims(ctx context.Context, s store.Store) {
	for {
		claimed, err := s.StepListByLogState(ctx, &model.StepLogFilter{
			States: []model.LogState{model.LogStateFinalizing, model.LogStateDeleting},
		})
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
	if f.Interval < minInterval {
		return fmt.Errorf("log finalizer interval must be at least %s, got %s", minInterval, f.Interval)
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
