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
	"errors"
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
	"github.com/woodpecker-ci/woodpecker/server/pipeline"
	"github.com/woodpecker-ci/woodpecker/server/pubsub"
	"github.com/woodpecker-ci/woodpecker/server/queue"
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
		agent, err := s.getAgentFromContext(c)
		if err != nil {
			return nil, err
		} else if agent.NoSchedule {
			return nil, nil
		}

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

	currentPipeline, err := s.store.GetPipeline(pstep.PipelineID)
	if err != nil {
		log.Error().Msgf("error: cannot find pipeline with id %d: %s", pstep.PipelineID, err)
		return err
	}

	step, err := s.store.StepChild(currentPipeline, pstep.PID, state.Step)
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

	repo, err := s.store.GetRepo(currentPipeline.RepoID)
	if err != nil {
		log.Error().Msgf("error: cannot find repo with id %d: %s", currentPipeline.RepoID, err)
		return err
	}

	if _, err = pipeline.UpdateStepStatus(s.store, *step, state, currentPipeline.Started); err != nil {
		log.Error().Err(err).Msg("rpc.update: cannot update step")
	}

	// TODO get all
	if currentPipeline.Steps, err = s.store.StepList(currentPipeline, &model.PaginationData{Page: 1, PerPage: 50}); err != nil {
		log.Error().Err(err).Msg("can not get step list from store")
	}
	if currentPipeline.Steps, err = model.Tree(currentPipeline.Steps); err != nil {
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
		Pipeline: *currentPipeline,
	})
	if err := s.pubsub.Publish(c, "topic/events", message); err != nil {
		log.Error().Err(err).Msg("can not publish step list to")
	}

	return nil
}

// Upload implements the rpc.Upload function
func (s *RPC) Upload(_ context.Context, id string, file *rpc.File) error {
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

	currentPipeline, err := s.store.GetPipeline(step.PipelineID)
	if err != nil {
		log.Error().Msgf("error: cannot find pipeline with id %d: %s", step.PipelineID, err)
		return err
	}

	repo, err := s.store.GetRepo(currentPipeline.RepoID)
	if err != nil {
		log.Error().Msgf("error: cannot find repo with id %d: %s", currentPipeline.RepoID, err)
		return err
	}

	if currentPipeline.Status == model.StatusPending {
		if currentPipeline, err = pipeline.UpdateToStatusRunning(s.store, *currentPipeline, state.Started); err != nil {
			log.Error().Msgf("error: init: cannot update build_id %d state: %s", currentPipeline.ID, err)
		}
	}

	defer func() {
		// TODO get all
		currentPipeline.Steps, _ = s.store.StepList(currentPipeline, &model.PaginationData{Page: 1, PerPage: 50})
		message := pubsub.Message{
			Labels: map[string]string{
				"repo":    repo.FullName,
				"private": strconv.FormatBool(repo.IsSCMPrivate),
			},
		}
		message.Data, _ = json.Marshal(model.Event{
			Repo:     *repo,
			Pipeline: *currentPipeline,
		})
		if err := s.pubsub.Publish(c, "topic/events", message); err != nil {
			log.Error().Err(err).Msg("can not publish step list to")
		}
	}()

	_, err = pipeline.UpdateStepToStatusStarted(s.store, *step, state)
	return err
}

// Done implements the rpc.Done function
func (s *RPC) Done(c context.Context, id string, state rpc.State) error {
	workflowID, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		return err
	}

	workflow, err := s.store.StepLoad(workflowID)
	if err != nil {
		log.Error().Msgf("error: cannot find step with id %d: %s", workflowID, err)
		return err
	}

	currentPipeline, err := s.store.GetPipeline(workflow.PipelineID)
	if err != nil {
		log.Error().Msgf("error: cannot find pipeline with id %d: %s", workflow.PipelineID, err)
		return err
	}

	repo, err := s.store.GetRepo(currentPipeline.RepoID)
	if err != nil {
		log.Error().Msgf("error: cannot find repo with id %d: %s", currentPipeline.RepoID, err)
		return err
	}

	log.Trace().
		Str("repo_id", fmt.Sprint(repo.ID)).
		Str("build_id", fmt.Sprint(currentPipeline.ID)).
		Str("step_id", id).
		Msgf("gRPC Done with state: %#v", state)

	if workflow, err = pipeline.UpdateStepStatusToDone(s.store, *workflow, state); err != nil {
		log.Error().Msgf("error: done: cannot update step_id %d state: %s", workflow.ID, err)
	}

	var queueErr error
	if workflow.Failing() {
		queueErr = s.queue.Error(c, id, fmt.Errorf("Step finished with exitcode %d, %s", state.ExitCode, state.Error))
	} else {
		queueErr = s.queue.Done(c, id, workflow.State)
	}
	if queueErr != nil {
		log.Error().Msgf("error: done: cannot ack step_id %d: %s", workflowID, err)
	}

	// TODO get all
	steps, err := s.store.StepList(currentPipeline, &model.PaginationData{Page: 1, PerPage: 50})
	if err != nil {
		return err
	}
	s.completeChildrenIfParentCompleted(steps, workflow)

	if !model.IsThereRunningStage(steps) {
		if currentPipeline, err = pipeline.UpdateStatusToDone(s.store, *currentPipeline, model.PipelineStatus(steps), workflow.Stopped); err != nil {
			log.Error().Err(err).Msgf("error: done: cannot update build_id %d final state", currentPipeline.ID)
		}
	}

	s.updateForgeStatus(c, repo, currentPipeline, workflow)

	if err := s.logger.Close(c, id); err != nil {
		log.Error().Err(err).Msgf("done: cannot close build_id %d logger", workflow.ID)
	}

	if err := s.notify(c, repo, currentPipeline, steps); err != nil {
		return err
	}

	if currentPipeline.Status == model.StatusSuccess || currentPipeline.Status == model.StatusFailure {
		s.pipelineCount.WithLabelValues(repo.FullName, currentPipeline.Branch, string(currentPipeline.Status), "total").Inc()
		s.pipelineTime.WithLabelValues(repo.FullName, currentPipeline.Branch, string(currentPipeline.Status), "total").Set(float64(currentPipeline.Finished - currentPipeline.Started))
	}
	if model.IsMultiPipeline(steps) {
		s.pipelineTime.WithLabelValues(repo.FullName, currentPipeline.Branch, string(workflow.State), workflow.Name).Set(float64(workflow.Stopped - workflow.Started))
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

func (s *RPC) RegisterAgent(ctx context.Context, platform, backend, version string, capacity int32) (int64, error) {
	agent, err := s.getAgentFromContext(ctx)
	if err != nil {
		return -1, err
	}

	agent.Backend = backend
	agent.Platform = platform
	agent.Capacity = capacity
	agent.Version = version

	err = s.store.AgentUpdate(agent)
	if err != nil {
		return -1, err
	}

	return agent.ID, nil
}

func (s *RPC) ReportHealth(ctx context.Context, status string) error {
	agent, err := s.getAgentFromContext(ctx)
	if err != nil {
		return err
	}

	if status != "I am alive!" {
		return errors.New("Are you alive?")
	}

	agent.LastContact = time.Now().Unix()

	return s.store.AgentUpdate(agent)
}

func (s *RPC) completeChildrenIfParentCompleted(steps []*model.Step, completedWorkflow *model.Step) {
	for _, p := range steps {
		if p.Running() && p.PPID == completedWorkflow.PID {
			if _, err := pipeline.UpdateStepToStatusSkipped(s.store, *p, completedWorkflow.Stopped); err != nil {
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
		return nil, errors.New("metadata is not provided")
	}

	values := md["agent_id"]
	if len(values) == 0 {
		return nil, errors.New("agent_id is not provided")
	}

	_agentID := values[0]
	agentID, err := strconv.ParseInt(_agentID, 10, 64)
	if err != nil {
		return nil, errors.New("agent_id is not a valid integer")
	}

	return s.store.AgentFind(agentID)
}
