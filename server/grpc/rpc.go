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
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/rs/zerolog/log"
	grpcMetadata "google.golang.org/grpc/metadata"

	"go.woodpecker-ci.org/woodpecker/v2/pipeline/rpc"
	"go.woodpecker-ci.org/woodpecker/v2/server/forge"
	"go.woodpecker-ci.org/woodpecker/v2/server/logging"
	"go.woodpecker-ci.org/woodpecker/v2/server/model"
	"go.woodpecker-ci.org/woodpecker/v2/server/pipeline"
	"go.woodpecker-ci.org/woodpecker/v2/server/pubsub"
	"go.woodpecker-ci.org/woodpecker/v2/server/queue"
	"go.woodpecker-ci.org/woodpecker/v2/server/store"
)

type RPC struct {
	forge         forge.Forge
	queue         queue.Queue
	pubsub        *pubsub.Publisher
	logger        logging.Log
	store         store.Store
	pipelineTime  *prometheus.GaugeVec
	pipelineCount *prometheus.CounterVec
}

// Next implements the rpc.Next function
func (s *RPC) Next(c context.Context, agentFilter rpc.Filter) (*rpc.Workflow, error) {
	if hostname, err := s.getHostnameFromContext(c); err == nil {
		log.Debug().Msgf("agent connected: %s: polling", hostname)
	}

	fn := createFilterFunc(agentFilter)
	for {
		agent, err := s.getAgentFromContext(c)
		if err != nil {
			return nil, err
		} else if agent.NoSchedule {
			return nil, nil
		}

		task, err := s.queue.Poll(c, agent.ID, fn)
		if err != nil {
			return nil, err
		} else if task == nil {
			return nil, nil
		}

		if task.ShouldRun() {
			workflow := new(rpc.Workflow)
			err = json.Unmarshal(task.Data, workflow)
			return workflow, err
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
func (s *RPC) Update(_ context.Context, id string, state rpc.State) error {
	workflowID, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		return err
	}

	workflow, err := s.store.WorkflowLoad(workflowID)
	if err != nil {
		log.Error().Msgf("error: rpc.update: cannot find workflow with id %d: %s", workflowID, err)
		return err
	}

	currentPipeline, err := s.store.GetPipeline(workflow.PipelineID)
	if err != nil {
		log.Error().Msgf("error: cannot find pipeline with id %d: %s", workflow.PipelineID, err)
		return err
	}

	step, err := s.store.StepChild(currentPipeline, workflow.PID, state.Step)
	if err != nil {
		log.Error().Msgf("error: cannot find step with name %s: %s", state.Step, err)
		return err
	}

	repo, err := s.store.GetRepo(currentPipeline.RepoID)
	if err != nil {
		log.Error().Msgf("error: cannot find repo with id %d: %s", currentPipeline.RepoID, err)
		return err
	}

	if err := pipeline.UpdateStepStatus(s.store, step, state, currentPipeline.Started); err != nil {
		log.Error().Err(err).Msg("rpc.update: cannot update step")
	}

	if currentPipeline.Workflows, err = s.store.WorkflowGetTree(currentPipeline); err != nil {
		log.Error().Err(err).Msg("can not build tree from step list")
		return err
	}
	message := pubsub.Message{
		Labels: map[string]string{
			"repo":    repo.FullName,
			"private": strconv.FormatBool(repo.IsSCMPrivate),
		},
	}
	message.Data, err = json.Marshal(model.Event{
		Repo:     *repo,
		Pipeline: *currentPipeline,
	})
	if err != nil {
		return err
	}
	s.pubsub.Publish(message)

	return nil
}

// Init implements the rpc.Init function
func (s *RPC) Init(c context.Context, id string, state rpc.State) error {
	stepID, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		return err
	}

	workflow, err := s.store.WorkflowLoad(stepID)
	if err != nil {
		log.Error().Msgf("error: cannot find step with id %d: %s", stepID, err)
		return err
	}

	agent, err := s.getAgentFromContext(c)
	if err != nil {
		return err
	}
	workflow.AgentID = agent.ID

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

	if currentPipeline.Status == model.StatusPending {
		if currentPipeline, err = pipeline.UpdateToStatusRunning(s.store, *currentPipeline, state.Started); err != nil {
			log.Error().Msgf("error: init: cannot update build_id %d state: %s", currentPipeline.ID, err)
		}
	}

	s.updateForgeStatus(c, repo, currentPipeline, workflow)

	defer func() {
		currentPipeline.Workflows, _ = s.store.WorkflowGetTree(currentPipeline)
		message := pubsub.Message{
			Labels: map[string]string{
				"repo":    repo.FullName,
				"private": strconv.FormatBool(repo.IsSCMPrivate),
			},
		}
		message.Data, err = json.Marshal(model.Event{
			Repo:     *repo,
			Pipeline: *currentPipeline,
		})
		if err != nil {
			log.Error().Err(err).Msgf("could not marshal JSON")
			return
		}
		s.pubsub.Publish(message)
	}()

	workflow, err = pipeline.UpdateWorkflowToStatusStarted(s.store, *workflow, state)
	if err != nil {
		return err
	}
	s.updateForgeStatus(c, repo, currentPipeline, workflow)
	return nil
}

// Done implements the rpc.Done function
func (s *RPC) Done(c context.Context, id string, state rpc.State) error {
	workflowID, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		return err
	}

	workflow, err := s.store.WorkflowLoad(workflowID)
	if err != nil {
		log.Error().Err(err).Msgf("cannot find step with id %d", workflowID)
		return err
	}

	workflow.Children, err = s.store.StepListFromWorkflowFind(workflow)
	if err != nil {
		return err
	}

	currentPipeline, err := s.store.GetPipeline(workflow.PipelineID)
	if err != nil {
		log.Error().Err(err).Msgf("cannot find pipeline with id %d", workflow.PipelineID)
		return err
	}

	repo, err := s.store.GetRepo(currentPipeline.RepoID)
	if err != nil {
		log.Error().Err(err).Msgf("cannot find repo with id %d", currentPipeline.RepoID)
		return err
	}

	logger := log.With().
		Str("repo_id", fmt.Sprint(repo.ID)).
		Str("pipeline_id", fmt.Sprint(currentPipeline.ID)).
		Str("workflow_id", id).Logger()

	logger.Trace().Msgf("gRPC Done with state: %#v", state)

	if workflow, err = pipeline.UpdateWorkflowStatusToDone(s.store, *workflow, state); err != nil {
		logger.Error().Err(err).Msgf("pipeline.UpdateStepStatusToDone: cannot update workflow state: %s", err)
	}

	var queueErr error
	if workflow.Failing() {
		queueErr = s.queue.Error(c, id, fmt.Errorf("Step finished with exit code %d, %s", state.ExitCode, state.Error))
	} else {
		queueErr = s.queue.Done(c, id, workflow.State)
	}
	if queueErr != nil {
		logger.Error().Err(queueErr).Msg("queue.Done: cannot ack workflow")
	}

	currentPipeline.Workflows, err = s.store.WorkflowGetTree(currentPipeline)
	if err != nil {
		return err
	}
	s.completeChildrenIfParentCompleted(workflow)

	if !model.IsThereRunningStage(currentPipeline.Workflows) {
		if currentPipeline, err = pipeline.UpdateStatusToDone(s.store, *currentPipeline, model.PipelineStatus(currentPipeline.Workflows), workflow.Stopped); err != nil {
			logger.Error().Err(err).Msgf("pipeline.UpdateStatusToDone: cannot update workflow final state")
		}
	}

	s.updateForgeStatus(c, repo, currentPipeline, workflow)

	// make sure writes to pubsub are non blocking (https://github.com/woodpecker-ci/woodpecker/blob/c919f32e0b6432a95e1a6d3d0ad662f591adf73f/server/logging/log.go#L9)
	go func() {
		for _, step := range workflow.Children {
			if err := s.logger.Close(c, step.ID); err != nil {
				logger.Error().Err(err).Msgf("done: cannot close log stream for step %d", step.ID)
			}
		}
	}()

	if err := s.notify(repo, currentPipeline); err != nil {
		return err
	}

	if currentPipeline.Status == model.StatusSuccess || currentPipeline.Status == model.StatusFailure {
		s.pipelineCount.WithLabelValues(repo.FullName, currentPipeline.Branch, string(currentPipeline.Status), "total").Inc()
		s.pipelineTime.WithLabelValues(repo.FullName, currentPipeline.Branch, string(currentPipeline.Status), "total").Set(float64(currentPipeline.Finished - currentPipeline.Started))
	}
	if currentPipeline.IsMultiPipeline() {
		s.pipelineTime.WithLabelValues(repo.FullName, currentPipeline.Branch, string(workflow.State), workflow.Name).Set(float64(workflow.Stopped - workflow.Started))
	}

	return nil
}

// Log implements the rpc.Log function
func (s *RPC) Log(c context.Context, _logEntry *rpc.LogEntry) error {
	// convert rpc log_entry to model.log_entry
	step, err := s.store.StepByUUID(_logEntry.StepUUID)
	if err != nil {
		return fmt.Errorf("could not find step with uuid %s in store: %w", _logEntry.StepUUID, err)
	}
	logEntry := &model.LogEntry{
		StepID: step.ID,
		Time:   _logEntry.Time,
		Line:   _logEntry.Line,
		Data:   []byte(_logEntry.Data),
		Type:   model.LogEntryType(_logEntry.Type),
	}
	// make sure writes to pubsub are non blocking (https://github.com/woodpecker-ci/woodpecker/blob/c919f32e0b6432a95e1a6d3d0ad662f591adf73f/server/logging/log.go#L9)
	go func() {
		// write line to listening web clients
		if err := s.logger.Write(c, logEntry.StepID, logEntry); err != nil {
			log.Error().Err(err).Msgf("rpc server could not write to logger")
		}
	}()
	// make line persistent in database
	return s.store.LogAppend(logEntry)
}

func (s *RPC) RegisterAgent(ctx context.Context, platform, backend, version string, capacity int32) (int64, error) {
	agent, err := s.getAgentFromContext(ctx)
	if err != nil {
		return -1, err
	}

	if agent.Name == "" {
		if hostname, err := s.getHostnameFromContext(ctx); err == nil {
			agent.Name = hostname
		}
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

func (s *RPC) UnregisterAgent(ctx context.Context) error {
	agent, err := s.getAgentFromContext(ctx)
	if !agent.IsSystemAgent() {
		// registered with individual agent token -> do not unregister
		return nil
	}
	log.Debug().Msgf("unregistering agent with ID %d", agent.ID)
	if err != nil {
		return err
	}

	err = s.store.AgentDelete(agent)

	return err
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

func (s *RPC) completeChildrenIfParentCompleted(completedWorkflow *model.Workflow) {
	for _, c := range completedWorkflow.Children {
		if c.Running() {
			if _, err := pipeline.UpdateStepToStatusSkipped(s.store, *c, completedWorkflow.Stopped); err != nil {
				log.Error().Msgf("error: done: cannot update step_id %d child state: %s", c.ID, err)
			}
		}
	}
}

func (s *RPC) updateForgeStatus(ctx context.Context, repo *model.Repo, pipeline *model.Pipeline, workflow *model.Workflow) {
	user, err := s.store.GetUser(repo.UserID)
	if err != nil {
		log.Error().Err(err).Msgf("can not get user with id '%d'", repo.UserID)
		return
	}

	forge.Refresh(ctx, s.forge, s.store, user)

	// only do status updates for parent steps
	if workflow != nil {
		err = s.forge.Status(ctx, user, repo, pipeline, workflow)
		if err != nil {
			log.Error().Err(err).Msgf("error setting commit status for %s/%d", repo.FullName, pipeline.Number)
		}
	}
}

func (s *RPC) notify(repo *model.Repo, pipeline *model.Pipeline) (err error) {
	message := pubsub.Message{
		Labels: map[string]string{
			"repo":    repo.FullName,
			"private": strconv.FormatBool(repo.IsSCMPrivate),
		},
	}
	message.Data, err = json.Marshal(model.Event{
		Repo:     *repo,
		Pipeline: *pipeline,
	})
	if err != nil {
		return err
	}
	s.pubsub.Publish(message)
	return nil
}

func (s *RPC) getAgentFromContext(ctx context.Context) (*model.Agent, error) {
	md, ok := grpcMetadata.FromIncomingContext(ctx)
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

func (s *RPC) getHostnameFromContext(ctx context.Context) (string, error) {
	metadata, ok := grpcMetadata.FromIncomingContext(ctx)
	if ok {
		hostname, ok := metadata["hostname"]
		if ok && len(hostname) != 0 {
			return hostname[0], nil
		}
	}
	return "", errors.New("no hostname in metadata")
}
