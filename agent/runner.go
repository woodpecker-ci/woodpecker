// Copyright 2022 Woodpecker Authors
// Copyright 2018 Drone.IO Inc.
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

package agent

import (
	"context"
	"errors"
	"sync"
	"time"

	"github.com/rs/zerolog/log"
	"github.com/tevino/abool"
	"google.golang.org/grpc/metadata"

	"github.com/woodpecker-ci/woodpecker/pipeline"
	backend "github.com/woodpecker-ci/woodpecker/pipeline/backend/types"
	"github.com/woodpecker-ci/woodpecker/pipeline/rpc"
	"github.com/woodpecker-ci/woodpecker/shared/utils"
)

// TODO: Implement log streaming.
// Until now we need to limit the size of the logs and files that we upload.
// The maximum grpc payload size is 4194304. So we need to set these limits below the maximum.
const (
	maxLogsUpload = 2000000 // this is per step
	maxFileUpload = 1000000
)

type Runner struct {
	client   rpc.Peer
	filter   rpc.Filter
	hostname string
	counter  *State
	engine   *backend.Engine
}

func NewRunner(workEngine rpc.Peer, f rpc.Filter, h string, state *State, backend *backend.Engine) Runner {
	return Runner{
		client:   workEngine,
		filter:   f,
		hostname: h,
		counter:  state,
		engine:   backend,
	}
}

func (r *Runner) Run(runnerCtx context.Context) error {
	log.Debug().Msg("request next execution")

	meta, _ := metadata.FromOutgoingContext(runnerCtx)
	ctxmeta := metadata.NewOutgoingContext(context.Background(), meta)

	// get the next workflow from the queue
	work, err := r.client.Next(runnerCtx, r.filter)
	if err != nil {
		return err
	}
	if work == nil {
		return nil
	}

	timeout := time.Hour
	if minutes := work.Timeout; minutes != 0 {
		timeout = time.Duration(minutes) * time.Minute
	}

	repoName := extractRepositoryName(work.Config)       // hack
	pipelineNumber := extractPipelineNumber(work.Config) // hack

	r.counter.Add(
		work.ID,
		timeout,
		repoName,
		pipelineNumber,
	)
	defer r.counter.Done(work.ID)

	logger := log.With().
		Str("repo", repoName).
		Str("pipeline", pipelineNumber).
		Str("id", work.ID).
		Logger()

	logger.Debug().Msg("received execution")

	workflowCtx, cancel := context.WithTimeout(ctxmeta, timeout)
	defer cancel()

	// Add sigterm support for internal context.
	// Required when the pipeline is terminated by external signals
	// like kubernetes.
	workflowCtx = utils.WithContextSigtermCallback(workflowCtx, func() {
		logger.Error().Msg("Received sigterm termination signal")
	})

	canceled := abool.New()
	go func() {
		logger.Debug().Msg("listen for cancel signal")

		if werr := r.client.Wait(workflowCtx, work.ID); werr != nil {
			canceled.SetTo(true)
			logger.Warn().Err(werr).Msg("cancel signal received")

			cancel()
		} else {
			logger.Debug().Msg("stop listening for cancel signal")
		}
	}()

	go func() {
		for {
			select {
			case <-workflowCtx.Done():
				logger.Debug().Msg("pipeline done")

				return
			case <-time.After(time.Minute):
				logger.Debug().Msg("pipeline lease renewed")

				if err := r.client.Extend(workflowCtx, work.ID); err != nil {
					log.Error().Err(err).Msg("extending pipeline deadline failed")
				}
			}
		}
	}()

	state := rpc.State{}
	state.Started = time.Now().Unix()

	err = r.client.Init(ctxmeta, work.ID, state)
	if err != nil {
		logger.Error().Err(err).Msg("pipeline initialization failed")
	}

	var uploads sync.WaitGroup
	err = pipeline.New(work.Config,
		pipeline.WithContext(workflowCtx),
		pipeline.WithLogger(r.createLogger(ctxmeta, logger, &uploads, work)),
		pipeline.WithTracer(r.createTracer(ctxmeta, logger, work)),
		pipeline.WithEngine(*r.engine),
		pipeline.WithDescription(map[string]string{
			"ID":       work.ID,
			"Repo":     repoName,
			"Pipeline": pipelineNumber,
		}),
	).Run(runnerCtx)

	state.Finished = time.Now().Unix()
	state.Exited = true

	if canceled.IsSet() {
		state.Error = ""
		state.ExitCode = 137
	} else {
		if err != nil {
			pExitError := &pipeline.ExitError{}
			if errors.As(err, &pExitError) {
				state.ExitCode = pExitError.Code
			} else if errors.Is(err, pipeline.ErrCancel) {
				state.Error = ""
				state.ExitCode = 137
				canceled.SetTo(true)
			} else {
				state.ExitCode = 1
				state.Error = err.Error()
			}
		}
	}

	logger.Debug().
		Str("error", state.Error).
		Int("exit_code", state.ExitCode).
		Bool("canceled", canceled.IsSet()).
		Msg("pipeline complete")

	logger.Debug().Msg("uploading logs")
	uploads.Wait()
	logger.Debug().Msg("uploading logs complete")

	logger.Debug().
		Str("error", state.Error).
		Int("exit_code", state.ExitCode).
		Msg("updating pipeline status")

	if err := r.client.Done(ctxmeta, work.ID, state); err != nil {
		logger.Error().Err(err).Msg("updating pipeline status failed")
	} else {
		logger.Debug().Msg("updating pipeline status complete")
	}

	return nil
}

// extract repository name from the configuration
func extractRepositoryName(config *backend.Config) string {
	return config.Stages[0].Steps[0].Environment["CI_REPO"]
}

// extract pipeline number from the configuration
func extractPipelineNumber(config *backend.Config) string {
	return config.Stages[0].Steps[0].Environment["CI_PIPELINE_NUMBER"]
}
