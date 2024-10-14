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
	"go.woodpecker-ci.org/woodpecker/v2/server"
	"go.woodpecker-ci.org/woodpecker/v2/server/forge"
	"go.woodpecker-ci.org/woodpecker/v2/server/logging"
	"go.woodpecker-ci.org/woodpecker/v2/server/model"
	"go.woodpecker-ci.org/woodpecker/v2/server/pipeline"
	"go.woodpecker-ci.org/woodpecker/v2/server/pubsub"
	"go.woodpecker-ci.org/woodpecker/v2/server/queue"
	"go.woodpecker-ci.org/woodpecker/v2/server/store"
)

// updateAgentLastWorkDelay the delay before the LastWork info should be updated.
const updateAgentLastWorkDelay = time.Minute

type RPC struct {
	queue         queue.Queue
	pubsub        *pubsub.Publisher
	logger        logging.Log
	store         store.Store
	pipelineTime  *prometheus.GaugeVec
	pipelineCount *prometheus.CounterVec
}

// Next blocks until it provides the next workflow to execute.
func (s *RPC) Next(c context.Context, agentFilter rpc.Filter) (*rpc.Workflow, error) {
	if hostname, err := s.getHostnameFromContext(c); err == nil {
		log.Debug().Msgf("agent connected: %s: polling", hostname)
	}

	agent, err := s.getAgentFromContext(c)
	if err != nil {
		return nil, err
	}

	if agent.NoSchedule {
		time.Sleep(1 * time.Second)
		return nil, nil
	}

	agentServerLabels, err := agent.GetServerLabels()
	if err != nil {
		return nil, err
	}

	// enforce labels from server by overwriting agent labels
	for k, v := range agentServerLabels {
		agentFilter.Labels[k] = v
	}

	log.Trace().Msgf("Agent %s[%d] tries to pull task with labels: %v", agent.Name, agent.ID, agentFilter.Labels)

	filterFn := createFilterFunc(agentFilter)

	for {
		// poll blocks until a task is available or the context is canceled / worker is kicked
		task, err := s.queue.Poll(c, agent.ID, filterFn)
		if err != nil || task == nil {
			return nil, err
		}

		if task.ShouldRun() {
			workflow := new(rpc.Workflow)
			err = json.Unmarshal(task.Data, workflow)
			return workflow, err
		}

		// task should not run, so mark it as done
		if err := s.Done(c, task.ID, rpc.WorkflowState{}); err != nil {
			log.Error().Err(err).Msgf("marking workflow task '%s' as done failed", task.ID)
		}
	}
}

// Wait blocks until the workflow with the given ID is done.
func (s *RPC) Wait(c context.Context, workflowID string) error {
	agent, err := s.getAgentFromContext(c)
	if err != nil {
		return err
	}

	if err := s.checkAgentPermissionByWorkflow(c, agent, workflowID, nil, nil); err != nil {
		return err
	}

	return s.queue.Wait(c, workflowID)
}

// Extend extends the lease for the workflow with the given ID.
func (s *RPC) Extend(c context.Context, workflowID string) error {
	agent, err := s.getAgentFromContext(c)
	if err != nil {
		return err
	}

	err = s.updateAgentLastWork(agent)
	if err != nil {
		return err
	}

	if err := s.checkAgentPermissionByWorkflow(c, agent, workflowID, nil, nil); err != nil {
		return err
	}

	return s.queue.Extend(c, agent.ID, workflowID)
}

// Update updates the state of a step.
func (s *RPC) Update(c context.Context, strWorkflowID string, state rpc.StepState) error {
	workflowID, err := strconv.ParseInt(strWorkflowID, 10, 64)
	if err != nil {
		return err
	}

	workflow, err := s.store.WorkflowLoad(workflowID)
	if err != nil {
		log.Error().Err(err).Msgf("rpc.update: cannot find workflow with id %d", workflowID)
		return err
	}

	currentPipeline, err := s.store.GetPipeline(workflow.PipelineID)
	if err != nil {
		log.Error().Err(err).Msgf("cannot find pipeline with id %d", workflow.PipelineID)
		return err
	}

	agent, err := s.getAgentFromContext(c)
	if err != nil {
		return err
	}

	step, err := s.store.StepByUUID(state.StepUUID)
	if err != nil {
		log.Error().Err(err).Msgf("cannot find step with uuid %s", state.StepUUID)
		return err
	}

	if step.PipelineID != currentPipeline.ID {
		msg := fmt.Sprintf("agent returned status with step uuid '%s' which does not belong to current pipeline", state.StepUUID)
		log.Error().
			Int64("stepPipelineID", step.PipelineID).
			Int64("currentPipelineID", currentPipeline.ID).
			Msg(msg)
		return errors.New(msg)
	}

	repo, err := s.store.GetRepo(currentPipeline.RepoID)
	if err != nil {
		log.Error().Err(err).Msgf("cannot find repo with id %d", currentPipeline.RepoID)
		return err
	}

	// check before agent can alter some state
	if err := s.checkAgentPermissionByWorkflow(c, agent, strWorkflowID, currentPipeline, repo); err != nil {
		return err
	}

	if err := pipeline.UpdateStepStatus(s.store, step, state); err != nil {
		log.Error().Err(err).Msg("rpc.update: cannot update step")
	}

	if currentPipeline.Workflows, err = s.store.WorkflowGetTree(currentPipeline); err != nil {
		log.Error().Err(err).Msg("cannot build tree from step list")
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

// Init implements the rpc.Init function.
func (s *RPC) Init(c context.Context, strWorkflowID string, state rpc.WorkflowState) error {
	workflowID, err := strconv.ParseInt(strWorkflowID, 10, 64)
	if err != nil {
		return err
	}

	workflow, err := s.store.WorkflowLoad(workflowID)
	if err != nil {
		log.Error().Err(err).Msgf("cannot find workflow with id %d", workflowID)
		return err
	}

	agent, err := s.getAgentFromContext(c)
	if err != nil {
		return err
	}

	workflow.AgentID = agent.ID

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

	// check before agent can alter some state
	if err := s.checkAgentPermissionByWorkflow(c, agent, strWorkflowID, currentPipeline, repo); err != nil {
		return err
	}

	if currentPipeline.Status == model.StatusPending {
		if currentPipeline, err = pipeline.UpdateToStatusRunning(s.store, *currentPipeline, state.Started); err != nil {
			log.Error().Err(err).Msgf("init: cannot update pipeline %d state", currentPipeline.ID)
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
			log.Error().Err(err).Msg("could not marshal JSON")
			return
		}
		s.pubsub.Publish(message)
	}()

	workflow, err = pipeline.UpdateWorkflowStatusToRunning(s.store, *workflow, state)
	if err != nil {
		return err
	}
	s.updateForgeStatus(c, repo, currentPipeline, workflow)

	return s.updateAgentLastWork(agent)
}

// Done marks the workflow with the given ID as done.
func (s *RPC) Done(c context.Context, strWorkflowID string, state rpc.WorkflowState) error {
	workflowID, err := strconv.ParseInt(strWorkflowID, 10, 64)
	if err != nil {
		return err
	}

	workflow, err := s.store.WorkflowLoad(workflowID)
	if err != nil {
		log.Error().Err(err).Msgf("cannot find workflow with id %d", workflowID)
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

	agent, err := s.getAgentFromContext(c)
	if err != nil {
		return err
	}

	// check before agent can alter some state
	if err := s.checkAgentPermissionByWorkflow(c, agent, strWorkflowID, currentPipeline, repo); err != nil {
		return err
	}

	logger := log.With().
		Str("repo_id", fmt.Sprint(repo.ID)).
		Str("pipeline_id", fmt.Sprint(currentPipeline.ID)).
		Str("workflow_id", strWorkflowID).Logger()

	logger.Trace().Msgf("gRPC Done with state: %#v", state)

	if workflow, err = pipeline.UpdateWorkflowStatusToDone(s.store, *workflow, state); err != nil {
		logger.Error().Err(err).Msgf("pipeline.UpdateWorkflowStatusToDone: cannot update workflow state: %s", err)
	}

	var queueErr error
	if workflow.Failing() {
		queueErr = s.queue.Error(c, strWorkflowID, fmt.Errorf("workflow finished with error %s", state.Error))
	} else {
		queueErr = s.queue.Done(c, strWorkflowID, workflow.State)
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
		if currentPipeline, err = pipeline.UpdateStatusToDone(s.store, *currentPipeline, model.PipelineStatus(currentPipeline.Workflows), workflow.Finished); err != nil {
			logger.Error().Err(err).Msgf("pipeline.UpdateStatusToDone: cannot update workflows final state")
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
		s.pipelineTime.WithLabelValues(repo.FullName, currentPipeline.Branch, string(workflow.State), workflow.Name).Set(float64(workflow.Finished - workflow.Started))
	}

	return s.updateAgentLastWork(agent)
}

// Log writes a log entry to the database and publishes it to the pubsub.
// An explicit stepUUID makes it obvious that all entries must come from the same step.
func (s *RPC) Log(c context.Context, stepUUID string, rpcLogEntries []*rpc.LogEntry) error {
	step, err := s.store.StepByUUID(stepUUID)
	if err != nil {
		return fmt.Errorf("could not find step with uuid %s in store: %w", stepUUID, err)
	}

	agent, err := s.getAgentFromContext(c)
	if err != nil {
		return err
	}

	currentPipeline, err := s.store.GetPipeline(step.PipelineID)
	if err != nil {
		log.Error().Err(err).Msgf("cannot find pipeline with id %d", step.PipelineID)
		return err
	}

	// check before agent can alter some state
	if err := s.checkAgentPermissionByWorkflow(c, agent, "", currentPipeline, nil); err != nil {
		return err
	}

	err = s.updateAgentLastWork(agent)
	if err != nil {
		return err
	}

	var logEntries []*model.LogEntry

	for _, rpcLogEntry := range rpcLogEntries {
		if rpcLogEntry.StepUUID != stepUUID {
			return fmt.Errorf("expected step UUID %s, got %s", stepUUID, rpcLogEntry.StepUUID)
		}
		logEntries = append(logEntries, &model.LogEntry{
			StepID: step.ID,
			Time:   rpcLogEntry.Time,
			Line:   rpcLogEntry.Line,
			Data:   rpcLogEntry.Data,
			Type:   model.LogEntryType(rpcLogEntry.Type),
		})
	}

	// make sure writes to pubsub are non blocking (https://github.com/woodpecker-ci/woodpecker/blob/c919f32e0b6432a95e1a6d3d0ad662f591adf73f/server/logging/log.go#L9)
	go func() {
		// write line to listening web clients
		if err := s.logger.Write(c, step.ID, logEntries); err != nil {
			log.Error().Err(err).Msgf("rpc server could not write to logger")
		}
	}()

	if err = server.Config.Services.LogStore.LogAppend(step, logEntries); err != nil {
		log.Error().Err(err).Msg("could not store log entries")
	}

	return nil
}

func (s *RPC) RegisterAgent(ctx context.Context, info rpc.AgentInfo) (int64, error) {
	agent, err := s.getAgentFromContext(ctx)
	if err != nil {
		return -1, err
	}

	if agent.Name == "" {
		if hostname, err := s.getHostnameFromContext(ctx); err == nil {
			agent.Name = hostname
		}
	}

	agent.Backend = info.Backend
	agent.Platform = info.Platform
	agent.Capacity = int32(info.Capacity)
	agent.Version = info.Version
	agent.CustomLabels = info.CustomLabels

	err = s.store.AgentUpdate(agent)
	if err != nil {
		return -1, err
	}

	return agent.ID, nil
}

// UnregisterAgent removes the agent from the database.
func (s *RPC) UnregisterAgent(ctx context.Context) error {
	agent, err := s.getAgentFromContext(ctx)
	if !agent.IsSystemAgent() {
		// registered with individual agent token -> do not unregister
		return nil
	}
	log.Debug().Msgf("un-registering agent with ID %d", agent.ID)
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
		//nolint:stylecheck
		return errors.New("Are you alive?")
	}

	agent.LastContact = time.Now().Unix()

	return s.store.AgentUpdate(agent)
}

func (s *RPC) checkAgentPermissionByWorkflow(_ context.Context, agent *model.Agent, strWorkflowID string, pipeline *model.Pipeline, repo *model.Repo) error {
	var err error
	if repo == nil && pipeline == nil {
		workflowID, err := strconv.ParseInt(strWorkflowID, 10, 64)
		if err != nil {
			return err
		}

		workflow, err := s.store.WorkflowLoad(workflowID)
		if err != nil {
			log.Error().Err(err).Msgf("cannot find workflow with id %d", workflowID)
			return err
		}

		pipeline, err = s.store.GetPipeline(workflow.PipelineID)
		if err != nil {
			log.Error().Err(err).Msgf("cannot find pipeline with id %d", workflow.PipelineID)
			return err
		}
	}

	if repo == nil {
		repo, err = s.store.GetRepo(pipeline.RepoID)
		if err != nil {
			log.Error().Err(err).Msgf("cannot find repo with id %d", pipeline.RepoID)
			return err
		}
	}

	if agent.CanAccessRepo(repo) {
		return nil
	}

	msg := fmt.Sprintf("agent '%d' is not allowed to interact with repo[%d] '%s'", agent.ID, repo.ID, repo.FullName)
	log.Error().Int64("repoId", repo.ID).Msg(msg)
	return errors.New(msg)
}

func (s *RPC) completeChildrenIfParentCompleted(completedWorkflow *model.Workflow) {
	for _, c := range completedWorkflow.Children {
		if c.Running() {
			if _, err := pipeline.UpdateStepToStatusSkipped(s.store, *c, completedWorkflow.Finished); err != nil {
				log.Error().Err(err).Msgf("done: cannot update step_id %d child state", c.ID)
			}
		}
	}
}

func (s *RPC) updateForgeStatus(ctx context.Context, repo *model.Repo, pipeline *model.Pipeline, workflow *model.Workflow) {
	user, err := s.store.GetUser(repo.UserID)
	if err != nil {
		log.Error().Err(err).Msgf("cannot get user with id '%d'", repo.UserID)
		return
	}

	_forge, err := server.Config.Services.Manager.ForgeFromRepo(repo)
	if err != nil {
		log.Error().Err(err).Msgf("can not get forge for repo '%s'", repo.FullName)
		return
	}

	forge.Refresh(ctx, _forge, s.store, user)

	// only do status updates for parent steps
	if workflow != nil {
		err = _forge.Status(ctx, user, repo, pipeline, workflow)
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

func (s *RPC) updateAgentLastWork(agent *model.Agent) error {
	// only update agent.LastWork if not recently updated
	if time.Unix(agent.LastWork, 0).Add(updateAgentLastWorkDelay).After(time.Now()) {
		return nil
	}

	agent.LastWork = time.Now().Unix()
	if err := s.store.AgentUpdate(agent); err != nil {
		return err
	}

	return nil
}
