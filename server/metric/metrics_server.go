// Copyright 2024 Woodpecker Authors
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

package metric

import (
	"context"
	"errors"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/rs/zerolog/log"
	"github.com/urfave/cli/v3"

	"go.woodpecker-ci.org/woodpecker/v3/server"
	"go.woodpecker-ci.org/woodpecker/v3/server/store"
)

const (
	queueInfoRefreshInterval = 500 * time.Millisecond
	storeInfoRefreshInterval = 10 * time.Second
)

var (
	FailurePipelineStepInfoCount *prometheus.CounterVec   = nil
	StepDurationRecord           *prometheus.HistogramVec = nil
)

func StartMetricsCollector(ctx context.Context, c *cli.Command, _store store.Store) {
	detailedMetricsEnabled := c.Bool("step-level-metrics")
	pendingSteps := promauto.NewGauge(prometheus.GaugeOpts{
		Namespace: "woodpecker",
		Name:      "pending_steps",
		Help:      "Total number of pending pipeline steps.",
	})
	waitingSteps := promauto.NewGauge(prometheus.GaugeOpts{
		Namespace: "woodpecker",
		Name:      "waiting_steps",
		Help:      "Total number of pipeline waiting on deps.",
	})
	runningSteps := promauto.NewGauge(prometheus.GaugeOpts{
		Namespace: "woodpecker",
		Name:      "running_steps",
		Help:      "Total number of running pipeline steps.",
	})
	workers := promauto.NewGauge(prometheus.GaugeOpts{
		Namespace: "woodpecker",
		Name:      "worker_count",
		Help:      "Total number of workers.",
	})
	pipelines := promauto.NewGauge(prometheus.GaugeOpts{
		Namespace: "woodpecker",
		Name:      "pipeline_total_count",
		Help:      "Total number of pipelines.",
	})
	users := promauto.NewGauge(prometheus.GaugeOpts{
		Namespace: "woodpecker",
		Name:      "user_count",
		Help:      "Total number of users.",
	})
	repos := promauto.NewGauge(prometheus.GaugeOpts{
		Namespace: "woodpecker",
		Name:      "repo_count",
		Help:      "Total number of repos.",
	})

	if detailedMetricsEnabled {
		FailurePipelineStepInfoCount = promauto.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: "woodpecker",
				Name:      "step_failures_total",
				Help:      "Total number of pipeline step failures.",
			},
			[]string{"workflow", "repo", "step"},
		)

		StepDurationRecord = promauto.NewHistogramVec(
			prometheus.HistogramOpts{
				Namespace: "woodpecker",
				Name:      "step_duration_seconds",
				Help:      "Step duration in seconds.",
				Buckets:   []float64{1, 5, 10, 30, 60, 300, 600, 1800, 3600},
			},
			[]string{"workflow", "repo", "step"},
		)
	}
	go func() {
		log.Info().Msg("queue metric collector started")

		for {
			stats := server.Config.Services.Scheduler.Info(ctx)
			pendingSteps.Set(float64(stats.Stats.Pending))
			waitingSteps.Set(float64(stats.Stats.WaitingOnDeps))
			runningSteps.Set(float64(stats.Stats.Running))
			workers.Set(float64(stats.Stats.Workers))

			select {
			case <-ctx.Done():
				log.Info().Msg("queue metric collector stopped")
				return
			case <-time.After(queueInfoRefreshInterval):
			}
		}
	}()
	go func() {
		log.Info().Msg("store metric collector started")

		for {
			repoCount, repoErr := _store.GetRepoCount()
			userCount, userErr := _store.GetUserCount()
			pipelineCount, pipelineErr := _store.GetPipelineCount()
			pipelines.Set(float64(pipelineCount))
			users.Set(float64(userCount))
			repos.Set(float64(repoCount))

			if err := errors.Join(repoErr, userErr, pipelineErr); err != nil {
				log.Error().Err(err).Msg("could not update store information for metrics")
			}

			select {
			case <-ctx.Done():
				log.Info().Msg("store metric collector stopped")
				return
			case <-time.After(storeInfoRefreshInterval):
			}
		}
	}()
}
