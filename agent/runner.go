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
	"encoding/json"
	"io"
	"runtime"
	"strconv"
	"sync"
	"time"

	"github.com/rs/zerolog/log"
	"github.com/tevino/abool"
	"google.golang.org/grpc/metadata"

	"github.com/woodpecker-ci/woodpecker/pipeline"
	backend "github.com/woodpecker-ci/woodpecker/pipeline/backend/types"
	"github.com/woodpecker-ci/woodpecker/pipeline/multipart"
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

func (r *Runner) Run(ctx context.Context) error {
	log.Debug().Msg("request next execution")

	meta, _ := metadata.FromOutgoingContext(ctx)
	ctxmeta := metadata.NewOutgoingContext(context.Background(), meta)

	// get the next workflow from the queue
	work, err := r.client.Next(ctx, r.filter)
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

	ctx, cancel := context.WithTimeout(ctxmeta, timeout)
	defer cancel()

	// Add sigterm support for internal context.
	// Required when the pipeline is terminated by external signals
	// like kubernetes.
	ctx = utils.WithContextSigtermCallback(ctx, func() {
		logger.Error().Msg("Received sigterm termination signal")
	})

	canceled := abool.New()
	go func() {
		logger.Debug().Msg("listen for cancel signal")

		if werr := r.client.Wait(ctx, work.ID); werr != nil {
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
			case <-ctx.Done():
				logger.Debug().Msg("pipeline done")

				return
			case <-time.After(time.Minute):
				logger.Debug().Msg("pipeline lease renewed")

				if err := r.client.Extend(ctx, work.ID); err != nil {
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
	defaultLogger := pipeline.LogFunc(func(step *backend.Step, rc multipart.Reader) error {
		loglogger := logger.With().
			Str("image", step.Image).
			Str("stage", step.Alias).
			Logger()

		part, rerr := rc.NextPart()
		if rerr != nil {
			return rerr
		}
		uploads.Add(1)

		var secrets []string
		for _, secret := range work.Config.Secrets {
			if secret.Mask {
				secrets = append(secrets, secret.Value)
			}
		}

		loglogger.Debug().Msg("log stream opened")

		limitedPart := io.LimitReader(part, maxLogsUpload)
		logStream := rpc.NewLineWriter(r.client, work.ID, step.Alias, secrets...)
		if _, err := io.Copy(logStream, limitedPart); err != nil {
			log.Error().Err(err).Msg("copy limited logStream part")
		}

		loglogger.Debug().Msg("log stream copied")

		data, err := json.Marshal(logStream.Lines())
		if err != nil {
			loglogger.Err(err).Msg("could not marshal logstream")
		}

		file := &rpc.File{
			Mime: "application/json+logs",
			Step: step.Alias,
			Name: "logs.json",
			Data: data,
			Size: len(data),
			Time: time.Now().Unix(),
		}

		loglogger.Debug().Msg("log stream uploading")
		if serr := r.client.Upload(ctxmeta, work.ID, file); serr != nil {
			loglogger.Error().Err(serr).Msg("log stream upload error")
		} else {
			loglogger.Debug().Msg("log stream upload complete")
		}

		defer func() {
			loglogger.Debug().Msg("log stream closed")
			uploads.Done()
		}()

		part, rerr = rc.NextPart()
		if rerr != nil {
			return nil
		}
		// TODO should be configurable
		limitedPart = io.LimitReader(part, maxFileUpload)
		data, err = io.ReadAll(limitedPart)
		if err != nil {
			loglogger.Err(err).Msg("could not read limited part")
		}

		file = &rpc.File{
			Mime: part.Header().Get("Content-Type"),
			Step: step.Alias,
			Name: part.FileName(),
			Data: data,
			Size: len(data),
			Time: time.Now().Unix(),
			Meta: make(map[string]string),
		}
		for key, value := range part.Header() {
			file.Meta[key] = value[0]
		}

		loglogger.Debug().
			Str("file", file.Name).
			Str("mime", file.Mime).
			Msg("file stream uploading")

		if serr := r.client.Upload(ctxmeta, work.ID, file); serr != nil {
			loglogger.Error().
				Err(serr).
				Str("file", file.Name).
				Str("mime", file.Mime).
				Msg("file stream upload error")
		}

		loglogger.Debug().
			Str("file", file.Name).
			Str("mime", file.Mime).
			Msg("file stream upload complete")
		return nil
	})

	defaultTracer := pipeline.TraceFunc(func(state *pipeline.State) error {
		steplogger := logger.With().
			Str("image", state.Pipeline.Step.Image).
			Str("stage", state.Pipeline.Step.Alias).
			Err(state.Process.Error).
			Int("exit_code", state.Process.ExitCode).
			Bool("exited", state.Process.Exited).
			Logger()

		stepState := rpc.State{
			Step:     state.Pipeline.Step.Alias,
			Exited:   state.Process.Exited,
			ExitCode: state.Process.ExitCode,
			Started:  time.Now().Unix(), // TODO do not do this
			Finished: time.Now().Unix(),
		}
		if state.Process.Error != nil {
			stepState.Error = state.Process.Error.Error()
		}

		defer func() {
			steplogger.Debug().Msg("update step status")

			if uerr := r.client.Update(ctxmeta, work.ID, stepState); uerr != nil {
				steplogger.Debug().
					Err(uerr).
					Msg("update step status error")
			}

			steplogger.Debug().Msg("update step status complete")
		}()
		if state.Process.Exited {
			return nil
		}
		if state.Pipeline.Step.Environment == nil {
			state.Pipeline.Step.Environment = map[string]string{}
		}

		// TODO: find better way to update this state and move it to pipeline to have the same env in cli-exec
		state.Pipeline.Step.Environment["CI_MACHINE"] = r.hostname

		state.Pipeline.Step.Environment["CI_PIPELINE_STATUS"] = "success"
		state.Pipeline.Step.Environment["CI_PIPELINE_STARTED"] = strconv.FormatInt(state.Pipeline.Time, 10)
		state.Pipeline.Step.Environment["CI_PIPELINE_FINISHED"] = strconv.FormatInt(time.Now().Unix(), 10)

		state.Pipeline.Step.Environment["CI_STEP_STATUS"] = "success"
		state.Pipeline.Step.Environment["CI_STEP_STARTED"] = strconv.FormatInt(state.Pipeline.Time, 10)
		state.Pipeline.Step.Environment["CI_STEP_FINISHED"] = strconv.FormatInt(time.Now().Unix(), 10)

		state.Pipeline.Step.Environment["CI_SYSTEM_ARCH"] = runtime.GOOS + "/" + runtime.GOARCH

		// DEPRECATED
		state.Pipeline.Step.Environment["CI_BUILD_STATUS"] = "success"
		state.Pipeline.Step.Environment["CI_BUILD_STARTED"] = strconv.FormatInt(state.Pipeline.Time, 10)
		state.Pipeline.Step.Environment["CI_BUILD_FINISHED"] = strconv.FormatInt(time.Now().Unix(), 10)
		state.Pipeline.Step.Environment["CI_JOB_STATUS"] = "success"
		state.Pipeline.Step.Environment["CI_JOB_STARTED"] = strconv.FormatInt(state.Pipeline.Time, 10)
		state.Pipeline.Step.Environment["CI_JOB_FINISHED"] = strconv.FormatInt(time.Now().Unix(), 10)

		if state.Pipeline.Error != nil {
			state.Pipeline.Step.Environment["CI_PIPELINE_STATUS"] = "failure"
			state.Pipeline.Step.Environment["CI_STEP_STATUS"] = "failure"

			// DEPRECATED
			state.Pipeline.Step.Environment["CI_BUILD_STATUS"] = "failure"
			state.Pipeline.Step.Environment["CI_JOB_STATUS"] = "failure"
		}

		return nil
	})

	// crate func to kill still running worflow if r.client.Info tell us it got cancled
	err = pipeline.New(work.Config,
		pipeline.WithContext(ctx),
		pipeline.WithLogger(defaultLogger),
		pipeline.WithTracer(defaultTracer),
		pipeline.WithEngine(*r.engine),
		pipeline.WithDescription(map[string]string{
			"ID":       work.ID,
			"Repo":     repoName,
			"Pipeline": pipelineNumber,
		}),
	).Run()

	state.Finished = time.Now().Unix()
	state.Exited = true
	if err != nil {
		switch xerr := err.(type) {
		case *pipeline.ExitError:
			state.ExitCode = xerr.Code
		default:
			state.ExitCode = 1
			state.Error = err.Error()
		}
		if canceled.IsSet() {
			state.ExitCode = 137
		}
	}

	logger.Debug().
		Str("error", state.Error).
		Int("exit_code", state.ExitCode).
		Msg("pipeline complete")

	logger.Debug().Msg("uploading logs")

	uploads.Wait()

	logger.Debug().Msg("uploading logs complete")

	logger.Debug().
		Str("error", state.Error).
		Int("exit_code", state.ExitCode).
		Msg("updating pipeline status")

	err = r.client.Done(ctxmeta, work.ID, state)
	if err != nil {
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
