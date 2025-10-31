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
	"slices"
	"sync"
	"time"

	"github.com/rs/zerolog/log"
	"google.golang.org/grpc/metadata"

	"go.woodpecker-ci.org/woodpecker/v3/pipeline"
	backend "go.woodpecker-ci.org/woodpecker/v3/pipeline/backend/types"
	"go.woodpecker-ci.org/woodpecker/v3/pipeline/rpc"
	"go.woodpecker-ci.org/woodpecker/v3/shared/constant"
	"go.woodpecker-ci.org/woodpecker/v3/shared/utils"
)

type Runner struct {
	client   rpc.Peer
	filter   rpc.Filter
	hostname string
	counter  *State
	backend  *backend.Backend
	trusted  TrustedRepos
}

type TrustedRepos struct {
	Volumes  []int64
	Network  []int64
	Security []int64
}

func NewRunner(workEngine rpc.Peer, f rpc.Filter, h string, state *State, backend *backend.Backend, trusted TrustedRepos) Runner {
	return Runner{
		client:   workEngine,
		filter:   f,
		hostname: h,
		counter:  state,
		backend:  backend,
		trusted:  trusted,
	}
}

func (r *Runner) Run(runnerCtx, shutdownCtx context.Context) error { //nolint:contextcheck
	log.Debug().Msg("request next execution")

	meta, _ := metadata.FromOutgoingContext(runnerCtx)
	ctxMeta := metadata.NewOutgoingContext(context.Background(), meta)

	// get the next workflow from the queue
	workflow, err := r.client.Next(runnerCtx, r.filter)
	if err != nil {
		return err
	}
	if workflow == nil {
		return nil
	}

	timeout := time.Hour
	if minutes := workflow.Timeout; minutes != 0 {
		timeout = time.Duration(minutes) * time.Minute
	}

	r.counter.Add(
		workflow.ID,
		timeout,
		workflow.RepoName,
		workflow.PipelineNumber,
	)
	defer r.counter.Done(workflow.ID)

	logger := log.With().
		Str("repo", workflow.RepoName).
		Str("pipeline", workflow.PipelineNumber).
		Str("workflow_id", workflow.ID).
		Logger()

	logger.Debug().Msg("received execution")

	workflowCtx, cancel := context.WithTimeout(ctxMeta, timeout)
	defer cancel()

	// Add sigterm support for internal context.
	// Required when the pipeline is terminated by external signals
	// like kubernetes.
	workflowCtx = utils.WithContextSigtermCallback(workflowCtx, func() {
		logger.Error().Msg("Received sigterm termination signal")
	})

	canceled := false
	go func() {
		logger.Debug().Msg("listen for cancel signal")

		if err := r.client.Wait(workflowCtx, workflow.ID); err != nil {
			canceled = true
			logger.Warn().Err(err).Msg("cancel signal received")
			cancel()
		} else {
			logger.Debug().Msg("done listening for cancel signal")
		}
	}()

	go func() {
		for {
			select {
			case <-workflowCtx.Done():
				logger.Debug().Msg("pipeline done")
				return

			case <-time.After(constant.TaskTimeout / 3):
				logger.Debug().Msg("pipeline lease renewed")
				if err := r.client.Extend(workflowCtx, workflow.ID); err != nil {
					log.Error().Err(err).Msg("extending pipeline deadline failed")
				}
			}
		}
	}()

	state := rpc.WorkflowState{}
	state.Started = time.Now().Unix()

	err = r.client.Init(runnerCtx, workflow.ID, state)
	if err != nil {
		logger.Error().Err(err).Msg("workflow initialization failed")
		// TODO: should we return here?
	}

	trusted := backend.TrustedConfiguration{
		Network:  slices.Contains(r.trusted.Network, workflow.RepoID),
		Volumes:  slices.Contains(r.trusted.Volumes, workflow.RepoID),
		Security: slices.Contains(r.trusted.Security, workflow.RepoID),
	}

	var uploads sync.WaitGroup
	//nolint:contextcheck
	err = pipeline.New(workflow.Config,
		pipeline.WithContext(workflowCtx),
		pipeline.WithTaskUUID(fmt.Sprint(workflow.ID)),
		pipeline.WithLogger(r.createLogger(logger, &uploads, workflow)),
		pipeline.WithTracer(r.createTracer(ctxMeta, &uploads, logger, workflow, trusted)),
		pipeline.WithBackend(*r.backend),
		pipeline.WithDescription(map[string]string{
			"workflow_id":     workflow.ID,
			"repo":            workflow.RepoName,
			"pipeline_number": workflow.PipelineNumber,
		}),
		pipeline.WithTrustedConfiguration(trusted),
	).Run(runnerCtx)

	state.Finished = time.Now().Unix()

	if errors.Is(err, pipeline.ErrCancel) {
		canceled = true
	} else if canceled {
		err = errors.Join(err, pipeline.ErrCancel)
	}

	if err != nil {
		state.Error = err.Error()
	}

	logger.Debug().
		Str("error", state.Error).
		Bool("canceled", canceled).
		Msg("workflow finished")

	logger.Debug().Msg("uploading logs and traces / states ...")
	uploads.Wait()
	logger.Debug().Msg("uploaded logs and traces / states")

	logger.Debug().
		Str("error", state.Error).
		Msg("updating workflow status")

	doneCtx := runnerCtx
	if doneCtx.Err() != nil {
		doneCtx = shutdownCtx
	}
	if err := r.client.Done(doneCtx, workflow.ID, state); err != nil {
		logger.Error().Err(err).Msg("updating workflow status failed")
	} else {
		logger.Debug().Msg("updating workflow status complete")
	}

	return nil
}
