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
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/tevino/abool/v2"
	"google.golang.org/grpc/metadata"

	"go.woodpecker-ci.org/woodpecker/v2/pipeline"
	backend "go.woodpecker-ci.org/woodpecker/v2/pipeline/backend/types"
	"go.woodpecker-ci.org/woodpecker/v2/pipeline/rpc"
	"go.woodpecker-ci.org/woodpecker/v2/shared/logger/errorattr"
	"go.woodpecker-ci.org/woodpecker/v2/shared/utils"
)

type Runner struct {
	client   rpc.Peer
	filter   rpc.Filter
	hostname string
	counter  *State
	backend  *backend.Backend
}

func NewRunner(workEngine rpc.Peer, f rpc.Filter, h string, state *State, backend *backend.Backend) Runner {
	return Runner{
		client:   workEngine,
		filter:   f,
		hostname: h,
		counter:  state,
		backend:  backend,
	}
}

func (r *Runner) Run(runnerCtx context.Context) error {
	slog.Debug("request next execution")

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

	logger := slog.With(
		slog.String("repo", repoName),
		slog.String("pipeline", pipelineNumber),
		slog.String("id", work.ID))

	logger.Debug("received execution")

	workflowCtx, cancel := context.WithTimeout(ctxmeta, timeout)
	defer cancel()

	// Add sigterm support for internal context.
	// Required when the pipeline is terminated by external signals
	// like kubernetes.
	workflowCtx = utils.WithContextSigtermCallback(workflowCtx, func() {
		logger.Error("Received sigterm termination signal")
	})

	canceled := abool.New()
	go func() {
		logger.Debug("listen for cancel signal")

		if werr := r.client.Wait(workflowCtx, work.ID); werr != nil {
			canceled.SetTo(true)
			slog.Warn("cancel signal received", errorattr.Default(werr))

			cancel()
		} else {
			slog.Debug("stop listening for cancel signal")
		}
	}()

	go func() {
		for {
			select {
			case <-workflowCtx.Done():
				logger.Debug("pipeline done")

				return
			case <-time.After(time.Minute):
				logger.Debug("pipeline lease renewed")

				if err := r.client.Extend(workflowCtx, work.ID); err != nil {
					slog.Error("extending pipeline deadline failed", errorattr.Default(err))
				}
			}
		}
	}()

	state := rpc.State{}
	state.Started = time.Now().Unix()

	err = r.client.Init(runnerCtx, work.ID, state)
	if err != nil {
		slog.Error("pipeline initialization failed", errorattr.Default(err))
	}

	var uploads sync.WaitGroup
	//nolint:contextcheck
	err = pipeline.New(work.Config,
		pipeline.WithContext(workflowCtx),
		pipeline.WithTaskUUID(fmt.Sprint(work.ID)),
		pipeline.WithLogger(r.createLogger(logger, &uploads, work)),
		pipeline.WithTracer(r.createTracer(ctxmeta, logger, work)),
		pipeline.WithBackend(*r.backend),
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
	} else if err != nil {
		pExitError := &pipeline.ExitError{}
		switch {
		case errors.As(err, &pExitError):
			state.ExitCode = pExitError.Code
		case errors.Is(err, pipeline.ErrCancel):
			state.Error = ""
			state.ExitCode = 137
			canceled.SetTo(true)
		default:
			state.ExitCode = 1
			state.Error = err.Error()
		}
	}

	slog.Debug("pipeline complete", slog.String("error", state.Error), slog.Int("exit_code", state.ExitCode), slog.Bool("canceled", canceled.IsSet()))

	slog.Debug("uploading logs")
	uploads.Wait()
	slog.Debug("uploading logs complete")

	slog.Debug("updating pipeline status", slog.String("error", state.Error), slog.Int("exit_code", state.ExitCode))

	if err := r.client.Done(runnerCtx, work.ID, state); err != nil {
		slog.Error("updating pipeline status failed", errorattr.Default(err))
	} else {
		slog.Debug("updating pipeline status complete")
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
