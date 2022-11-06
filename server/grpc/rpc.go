// Copyright 2022 Woodpecker Authors
// Copyright 2021 Informatyka Boguslawski sp. z o.o. sp.k., http://www.ib.pl/
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
//
// This file has been modified by Informatyka Boguslawski sp. z o.o. sp.k.

package grpc

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/rs/zerolog/log"

	"github.com/prometheus/client_golang/prometheus"
	"google.golang.org/grpc/metadata"
	grpcMetadata "google.golang.org/grpc/metadata"

	"github.com/woodpecker-ci/woodpecker/pipeline/rpc"
	"github.com/woodpecker-ci/woodpecker/server"
	"github.com/woodpecker-ci/woodpecker/server/forge"
	"github.com/woodpecker-ci/woodpecker/server/logging"
	"github.com/woodpecker-ci/woodpecker/server/model"
	"github.com/woodpecker-ci/woodpecker/server/pubsub"
	"github.com/woodpecker-ci/woodpecker/server/queue"
	"github.com/woodpecker-ci/woodpecker/server/shared"
	"github.com/woodpecker-ci/woodpecker/server/store"
)

type RPC struct {
	forge         forge.Forge
	queue         queue.Queue
	pubsub        pubsub.Publisher
	logger        logging.Log
	store         store.Store
	host          string
	pipelineTime  *prometheus.GaugeVec
	pipelineCount *prometheus.CounterVec
}

// Next implements the rpc.Next function
func (s *RPC) Next(c context.Context, agentFilter rpc.Filter) (*rpc.Pipeline, error) {
	metadata, ok := grpcMetadata.FromIncomingContext(c)
	if ok {
		hostname, ok := metadata["hostname"]
		if ok && len(hostname) != 0 {
			log.Debug().Msgf("agent connected: %s: polling", hostname[0])
		}
	}

	fn, err := createFilterFunc(agentFilter)
	if err != nil {
		return nil, err
	}
	for {
		task, err := s.queue.Poll(c, fn)
		if err != nil {
			return nil, err
		} else if task == nil {
			return nil, nil
		}

		if task.ShouldRun() {
			pipeline := new(rpc.Pipeline)
			err = json.Unmarshal(task.Data, pipeline)
			return pipeline, err
		}
		if err := s.Done(c, task.ID, rpc.State{}); err != nil {
			log.Error().Err(err).Msgf("mark task '%s' done failed", task.ID)
		}
	}
}

// Wait implements the rpc.Wait function
func (s *RPC) Wait(c context.Context, id string) error {
	return s.queue.Wait(c, id)
}

// Extend implements the rpc.Extend function
func (s *RPC) Extend(c context.Context, id string) error {
	return s.queue.Extend(c, id)
}

// Update implements the rpc.Update function
func (s *RPC) Update(c context.Context, id string, state rpc.State) error {
	stepID, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		return err
	}

	pstep, err := s.store.StepLoad(stepID)
	if err != nil {
		log.Error().Msgf("error: rpc.update: cannot find step with id %d: %s", stepID, err)
		return err
	}

	pipeline, err := s.store.GetPipeline(pstep.PipelineID)
	if err != nil {
		log.Error().Msgf("error: cannot find pipeline with id %d: %s", pstep.PipelineID, err)
		return err
	}

	step, err := s.store.StepChild(pipeline, pstep.PID, state.Step)
	if err != nil {
		log.Error().Msgf("error: cannot find step with name %s: %s", state.Step, err)
		return err
	}

	metadata, ok := grpcMetadata.FromIncomingContext(c)
	if ok {
		hostname, ok := metadata["hostname"]
		if ok && len(hostname) != 0 {
			step.Machine = hostname[0]
		}
	}

	repo, err := s.store.GetRepo(pipeline.RepoID)
	if err != nil {
		log.Error().Msgf("error: cannot find repo with id %d: %s", pipeline.RepoID, err)
		return err
	}

	if _, err = shared.UpdateStepStatus(s.store, *step, state, pipeline.Started); err != nil {
		log.Error().Err(err).Msg("rpc.update: cannot update step")
	}

	if pipeline.Steps, err = s.store.StepList(pipeline); err != nil {
		log.Error().Err(err).Msg("can not get step list from store")
	}
	if pipeline.Steps, err = model.Tree(pipeline.Steps); err != nil {
		log.Error().Err(err).Msg("can not build tree from step list")
		return err
	}
	message := pubsub.Message{
		Labels: map[string]string{
			"repo":    repo.FullName,
			"private": strconv.FormatBool(repo.IsSCMPrivate),
		},
	}
	message.Data, _ = json.Marshal(model.Event{
		Repo:     *repo,
		Pipeline: *pipeline,
	})
	if err := s.pubsub.Publish(c, "topic/events", message); err != nil {
		log.Error().Err(err).Msg("can not publish step list to")
	}

	return nil
}

// Upload implements the rpc.Upload function
func (s *RPC) Upload(c context.Context, id string, file *rpc.File) error {
	stepID, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		return err
	}

	pstep, err := s.store.StepLoad(stepID)
	if err != nil {
		log.Error().Msgf("error: cannot find parent step with id %d: %s", stepID, err)
		return err
	}

	pipeline, err := s.store.GetPipeline(pstep.PipelineID)
	if err != nil {
		log.Error().Msgf("error: cannot find pipeline with id %d: %s", pstep.PipelineID, err)
		return err
	}

	step, err := s.store.StepChild(pipeline, pstep.PID, file.Step)
	if err != nil {
		log.Error().Msgf("error: cannot find child step with name %s: %s", file.Step, err)
		return err
	}

	if file.Mime == "application/json+logs" {
		return s.store.LogSave(
			step,
			bytes.NewBuffer(file.Data),
		)
	}

	report := &model.File{
		PipelineID: step.PipelineID,
		StepID:     step.ID,
		PID:        step.PID,
		Mime:       file.Mime,
		Name:       file.Name,
		Size:       file.Size,
		Time:       file.Time,
	}
	if d, ok := file.Meta["X-Tests-Passed"]; ok {
		report.Passed, _ = strconv.Atoi(d)
	}
	if d, ok := file.Meta["X-Tests-Failed"]; ok {
		report.Failed, _ = strconv.Atoi(d)
	}
	if d, ok := file.Meta["X-Tests-Skipped"]; ok {
		report.Skipped, _ = strconv.Atoi(d)
	}

	if d, ok := file.Meta["X-Checks-Passed"]; ok {
		report.Passed, _ = strconv.Atoi(d)
	}
	if d, ok := file.Meta["X-Checks-Failed"]; ok {
		report.Failed, _ = strconv.Atoi(d)
	}

	if d, ok := file.Meta["X-Coverage-Lines"]; ok {
		report.Passed, _ = strconv.Atoi(d)
	}
	if d, ok := file.Meta["X-Coverage-Total"]; ok {
		if total, _ := strconv.Atoi(d); total != 0 {
			report.Failed = total - report.Passed
		}
	}

	return server.Config.Storage.Files.FileCreate(
		report,
		bytes.NewBuffer(file.Data),
	)
}

// Init implements the rpc.Init function
func (s *RPC) Init(c context.Context, id string, state rpc.State) error {
	stepID, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		return err
	}

	step, err := s.store.StepLoad(stepID)
	if err != nil {
		log.Error().Msgf("error: cannot find step with id %d: %s", stepID, err)
		return err
	}
	metadata, ok := grpcMetadata.FromIncomingContext(c)
	if ok {
		hostname, ok := metadata["hostname"]
		if ok && len(hostname) != 0 {
			step.Machine = hostname[0]
		}
	}

	pipeline, err := s.store.GetPipeline(step.PipelineID)
	if err != nil {
		log.Error().Msgf("error: cannot find pipeline with id %d: %s", step.PipelineID, err)
		return err
	}

	repo, err := s.store.GetRepo(pipeline.RepoID)
	if err != nil {
		log.Error().Msgf("error: cannot find repo with id %d: %s", pipeline.RepoID, err)
		return err
	}

	if pipeline.Status == model.StatusPending {
		if pipeline, err = shared.UpdateToStatusRunning(s.store, *pipeline, state.Started); err != nil {
			log.Error().Msgf("error: init: cannot update build_id %d state: %s", pipeline.ID, err)
		}
	}

	defer func() {
		pipeline.Steps, _ = s.store.StepList(pipeline)
		message := pubsub.Message{
			Labels: map[string]string{
				"repo":    repo.FullName,
				"private": strconv.FormatBool(repo.IsSCMPrivate),
			},
		}
		message.Data, _ = json.Marshal(model.Event{
			Repo:     *repo,
			Pipeline: *pipeline,
		})
		if err := s.pubsub.Publish(c, "topic/events", message); err != nil {
			log.Error().Err(err).Msg("can not publish step list to")
		}
	}()

	_, err = shared.UpdateStepToStatusStarted(s.store, *step, state)
	return err
}

// Done implements the rpc.Done function
func (s *RPC) Done(c context.Context, id string, state rpc.State) error {
	stepID, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		return err
	}

	step, err := s.store.StepLoad(stepID)
	if err != nil {
		log.Error().Msgf("error: cannot find step with id %d: %s", stepID, err)
		return err
	}

	pipeline, err := s.store.GetPipeline(step.PipelineID)
	if err != nil {
		log.Error().Msgf("error: cannot find pipeline with id %d: %s", step.PipelineID, err)
		return err
	}

	repo, err := s.store.GetRepo(pipeline.RepoID)
	if err != nil {
		log.Error().Msgf("error: cannot find repo with id %d: %s", pipeline.RepoID, err)
		return err
	}

	log.Trace().
		Str("repo_id", fmt.Sprint(repo.ID)).
		Str("build_id", fmt.Sprint(pipeline.ID)).
		Str("step_id", id).
		Msgf("gRPC Done with state: %#v", state)

	if step, err = shared.UpdateStepStatusToDone(s.store, *step, state); err != nil {
		log.Error().Msgf("error: done: cannot update step_id %d state: %s", step.ID, err)
	}

	var queueErr error
	if step.Failing() {
		queueErr = s.queue.Error(c, id, fmt.Errorf("Step finished with exitcode %d, %s", state.ExitCode, state.Error))
	} else {
		queueErr = s.queue.Done(c, id, step.State)
	}
	if queueErr != nil {
		log.Error().Msgf("error: done: cannot ack step_id %d: %s", stepID, err)
	}

	steps, err := s.store.StepList(pipeline)
	if err != nil {
		return err
	}
	s.completeChildrenIfParentCompleted(steps, step)

	if !model.IsThereRunningStage(steps) {
		if pipeline, err = shared.UpdateStatusToDone(s.store, *pipeline, model.PipelineStatus(steps), step.Stopped); err != nil {
			log.Error().Err(err).Msgf("error: done: cannot update build_id %d final state", pipeline.ID)
		}
	}

	s.updateForgeStatus(c, repo, pipeline, step)

	if err := s.logger.Close(c, id); err != nil {
		log.Error().Err(err).Msgf("done: cannot close build_id %d logger", step.ID)
	}

	if err := s.notify(c, repo, pipeline, steps); err != nil {
		return err
	}

	if pipeline.Status == model.StatusSuccess || pipeline.Status == model.StatusFailure {
		s.pipelineCount.WithLabelValues(repo.FullName, pipeline.Branch, string(pipeline.Status), "total").Inc()
		s.pipelineTime.WithLabelValues(repo.FullName, pipeline.Branch, string(pipeline.Status), "total").Set(float64(pipeline.Finished - pipeline.Started))
	}
	if model.IsMultiPipeline(steps) {
		s.pipelineTime.WithLabelValues(repo.FullName, pipeline.Branch, string(step.State), step.Name).Set(float64(step.Stopped - step.Started))
	}

	return nil
}

// Log implements the rpc.Log function
func (s *RPC) Log(c context.Context, id string, line *rpc.Line) error {
	entry := new(logging.Entry)
	entry.Data, _ = json.Marshal(line)
	if err := s.logger.Write(c, id, entry); err != nil {
		log.Error().Err(err).Msgf("rpc server could not write to logger")
	}
	return nil
}

func (s *RPC) RegisterAgent(ctx context.Context, platform, backend string, capacity int32) error {
	agent, err := s.getAgentFromContext(ctx)
	if err != nil {
		return err
	}

	agent.Backend = backend
	agent.Platform = platform
	agent.Capacity = capacity

	return s.store.AgentUpdate(agent)
}

func (s *RPC) ReportHealth(ctx context.Context, status string) error {
	agent, err := s.getAgentFromContext(ctx)
	if err != nil {
		return err
	}

	if status != "I am alive!" {
		return fmt.Errorf("Are you alive?")
	}

	agent.LastContact = time.Now().Unix()

	return s.store.AgentUpdate(agent)
}

func (s *RPC) completeChildrenIfParentCompleted(steps []*model.Step, completedStep *model.Step) {
	for _, p := range steps {
		if p.Running() && p.PPID == completedStep.PID {
			if _, err := shared.UpdateStepToStatusSkipped(s.store, *p, completedStep.Stopped); err != nil {
				log.Error().Msgf("error: done: cannot update step_id %d child state: %s", p.ID, err)
			}
		}
	}
}

func (s *RPC) updateForgeStatus(ctx context.Context, repo *model.Repo, pipeline *model.Pipeline, step *model.Step) {
	user, err := s.store.GetUser(repo.UserID)
	if err != nil {
		log.Error().Err(err).Msgf("can not get user with id '%d'", repo.UserID)
		return
	}

	if refresher, ok := s.forge.(forge.Refresher); ok {
		ok, err := refresher.Refresh(ctx, user)
		if err != nil {
			log.Error().Err(err).Msgf("grpc: refresh oauth token of user '%s' failed", user.Login)
		} else if ok {
			if err := s.store.UpdateUser(user); err != nil {
				log.Error().Err(err).Msg("fail to save user to store after refresh oauth token")
			}
		}
	}

	// only do status updates for parent steps
	if step != nil && step.IsParent() {
		err = s.forge.Status(ctx, user, repo, pipeline, step)
		if err != nil {
			log.Error().Err(err).Msgf("error setting commit status for %s/%d", repo.FullName, pipeline.Number)
		}
	}
}

func (s *RPC) notify(c context.Context, repo *model.Repo, pipeline *model.Pipeline, steps []*model.Step) (err error) {
	if pipeline.Steps, err = model.Tree(steps); err != nil {
		return err
	}
	message := pubsub.Message{
		Labels: map[string]string{
			"repo":    repo.FullName,
			"private": strconv.FormatBool(repo.IsSCMPrivate),
		},
	}
	message.Data, _ = json.Marshal(model.Event{
		Repo:     *repo,
		Pipeline: *pipeline,
	})
	if err := s.pubsub.Publish(c, "topic/events", message); err != nil {
		log.Error().Err(err).Msgf("grpc could not notify event: '%v'", message)
	}
	return nil
}

func (s *RPC) getAgentFromContext(ctx context.Context) (*model.Agent, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return nil, fmt.Errorf("metadata is not provided")
	}

	values := md["agent_id"]
	if len(values) == 0 {
		return nil, fmt.Errorf("agent_id is not provided")
	}

	_agentID := values[0]
	agentID, err := strconv.ParseInt(_agentID, 10, 64)
	if err != nil {
		return nil, fmt.Errorf("agent_id is not a valid integer")
	}

	return s.store.AgentFind(agentID)
}
