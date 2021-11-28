// Copyright 2018 Drone.IO Inc.
// Copyright 2021 Informatyka Boguslawski sp. z o.o. sp.k., http://www.ib.pl/
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

	"github.com/rs/zerolog/log"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	grpcMetadata "google.golang.org/grpc/metadata"

	"github.com/woodpecker-ci/expr"
	"github.com/woodpecker-ci/woodpecker/pipeline/rpc"
	"github.com/woodpecker-ci/woodpecker/pipeline/rpc/proto"
	"github.com/woodpecker-ci/woodpecker/server"
	"github.com/woodpecker-ci/woodpecker/server/logging"
	"github.com/woodpecker-ci/woodpecker/server/model"
	"github.com/woodpecker-ci/woodpecker/server/pubsub"
	"github.com/woodpecker-ci/woodpecker/server/queue"
	"github.com/woodpecker-ci/woodpecker/server/remote"
	"github.com/woodpecker-ci/woodpecker/server/shared"
	"github.com/woodpecker-ci/woodpecker/server/store"
)

type RPC struct {
	remote     remote.Remote
	queue      queue.Queue
	pubsub     pubsub.Publisher
	logger     logging.Log
	store      store.Store
	host       string
	buildTime  *prometheus.GaugeVec
	buildCount *prometheus.CounterVec
}

// Next implements the rpc.Next function
func (s *RPC) Next(c context.Context, filter rpc.Filter) (*rpc.Pipeline, error) {
	metadata, ok := grpcMetadata.FromIncomingContext(c)
	if ok {
		hostname, ok := metadata["hostname"]
		if ok && len(hostname) != 0 {
			log.Debug().Msgf("agent connected: %s: polling", hostname[0])
		}
	}

	fn, err := createFilterFunc(filter)
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
	procID, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		return err
	}

	pproc, err := s.store.ProcLoad(procID)
	if err != nil {
		log.Error().Msgf("error: rpc.update: cannot find pproc with id %d: %s", procID, err)
		return err
	}

	build, err := s.store.GetBuild(pproc.BuildID)
	if err != nil {
		log.Error().Msgf("error: cannot find build with id %d: %s", pproc.BuildID, err)
		return err
	}

	proc, err := s.store.ProcChild(build, pproc.PID, state.Proc)
	if err != nil {
		log.Error().Msgf("error: cannot find proc with name %s: %s", state.Proc, err)
		return err
	}

	metadata, ok := grpcMetadata.FromIncomingContext(c)
	if ok {
		hostname, ok := metadata["hostname"]
		if ok && len(hostname) != 0 {
			proc.Machine = hostname[0]
		}
	}

	repo, err := s.store.GetRepo(build.RepoID)
	if err != nil {
		log.Error().Msgf("error: cannot find repo with id %d: %s", build.RepoID, err)
		return err
	}

	if _, err = shared.UpdateProcStatus(s.store, *proc, state, build.Started); err != nil {
		log.Error().Msgf("error: rpc.update: cannot update proc: %s", err)
	}

	build.Procs, _ = s.store.ProcList(build)
	build.Procs = model.Tree(build.Procs)
	message := pubsub.Message{
		Labels: map[string]string{
			"repo":    repo.FullName,
			"private": strconv.FormatBool(repo.IsSCMPrivate),
		},
	}
	message.Data, _ = json.Marshal(model.Event{
		Repo:  *repo,
		Build: *build,
	})
	if err := s.pubsub.Publish(c, "topic/events", message); err != nil {
		log.Error().Err(err).Msg("can not publish proc list to")
	}

	return nil
}

// Upload implements the rpc.Upload function
func (s *RPC) Upload(c context.Context, id string, file *rpc.File) error {
	procID, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		return err
	}

	pproc, err := s.store.ProcLoad(procID)
	if err != nil {
		log.Error().Msgf("error: cannot find parent proc with id %d: %s", procID, err)
		return err
	}

	build, err := s.store.GetBuild(pproc.BuildID)
	if err != nil {
		log.Error().Msgf("error: cannot find build with id %d: %s", pproc.BuildID, err)
		return err
	}

	proc, err := s.store.ProcChild(build, pproc.PID, file.Proc)
	if err != nil {
		log.Error().Msgf("error: cannot find child proc with name %s: %s", file.Proc, err)
		return err
	}

	if file.Mime == "application/json+logs" {
		return s.store.LogSave(
			proc,
			bytes.NewBuffer(file.Data),
		)
	}

	report := &model.File{
		BuildID: proc.BuildID,
		ProcID:  proc.ID,
		PID:     proc.PID,
		Mime:    file.Mime,
		Name:    file.Name,
		Size:    file.Size,
		Time:    file.Time,
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
	procID, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		return err
	}

	proc, err := s.store.ProcLoad(procID)
	if err != nil {
		log.Error().Msgf("error: cannot find proc with id %d: %s", procID, err)
		return err
	}
	metadata, ok := grpcMetadata.FromIncomingContext(c)
	if ok {
		hostname, ok := metadata["hostname"]
		if ok && len(hostname) != 0 {
			proc.Machine = hostname[0]
		}
	}

	build, err := s.store.GetBuild(proc.BuildID)
	if err != nil {
		log.Error().Msgf("error: cannot find build with id %d: %s", proc.BuildID, err)
		return err
	}

	repo, err := s.store.GetRepo(build.RepoID)
	if err != nil {
		log.Error().Msgf("error: cannot find repo with id %d: %s", build.RepoID, err)
		return err
	}

	if build.Status == model.StatusPending {
		if build, err = shared.UpdateToStatusRunning(s.store, *build, state.Started); err != nil {
			log.Error().Msgf("error: init: cannot update build_id %d state: %s", build.ID, err)
		}
	}

	defer func() {
		build.Procs, _ = s.store.ProcList(build)
		message := pubsub.Message{
			Labels: map[string]string{
				"repo":    repo.FullName,
				"private": strconv.FormatBool(repo.IsSCMPrivate),
			},
		}
		message.Data, _ = json.Marshal(model.Event{
			Repo:  *repo,
			Build: *build,
		})
		if err := s.pubsub.Publish(c, "topic/events", message); err != nil {
			log.Error().Err(err).Msg("can not publish proc list to")
		}
	}()

	_, err = shared.UpdateProcToStatusStarted(s.store, *proc, state)
	return err
}

// Done implements the rpc.Done function
func (s *RPC) Done(c context.Context, id string, state rpc.State) error {
	procID, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		return err
	}

	proc, err := s.store.ProcLoad(procID)
	if err != nil {
		log.Error().Msgf("error: cannot find proc with id %d: %s", procID, err)
		return err
	}

	build, err := s.store.GetBuild(proc.BuildID)
	if err != nil {
		log.Error().Msgf("error: cannot find build with id %d: %s", proc.BuildID, err)
		return err
	}

	repo, err := s.store.GetRepo(build.RepoID)
	if err != nil {
		log.Error().Msgf("error: cannot find repo with id %d: %s", build.RepoID, err)
		return err
	}

	if proc, err = shared.UpdateProcStatusToDone(s.store, *proc, state); err != nil {
		log.Error().Msgf("error: done: cannot update proc_id %d state: %s", proc.ID, err)
	}

	var queueErr error
	if proc.Failing() {
		queueErr = s.queue.Error(c, id, fmt.Errorf("Proc finished with exitcode %d, %s", state.ExitCode, state.Error))
	} else {
		queueErr = s.queue.Done(c, id, proc.State)
	}
	if queueErr != nil {
		log.Error().Msgf("error: done: cannot ack proc_id %d: %s", procID, err)
	}

	procs, _ := s.store.ProcList(build)
	s.completeChildrenIfParentCompleted(procs, proc)

	if !isThereRunningStage(procs) {
		if build, err = shared.UpdateStatusToDone(s.store, *build, buildStatus(procs), proc.Stopped); err != nil {
			log.Error().Msgf("error: done: cannot update build_id %d final state: %s", build.ID, err)
		}

		if !isMultiPipeline(procs) {
			s.updateRemoteStatus(c, repo, build, nil)
		}
	}

	if isMultiPipeline(procs) {
		s.updateRemoteStatus(c, repo, build, proc)
	}

	if err := s.logger.Close(c, id); err != nil {
		log.Error().Msgf("error: done: cannot close build_id %d logger: %s", proc.ID, err)
	}

	s.notify(c, repo, build, procs)

	if build.Status == model.StatusSuccess || build.Status == model.StatusFailure {
		s.buildCount.WithLabelValues(repo.FullName, build.Branch, string(build.Status), "total").Inc()
		s.buildTime.WithLabelValues(repo.FullName, build.Branch, string(build.Status), "total").Set(float64(build.Finished - build.Started))
	}
	if isMultiPipeline(procs) {
		s.buildTime.WithLabelValues(repo.FullName, build.Branch, string(proc.State), proc.Name).Set(float64(proc.Stopped - proc.Started))
	}

	return nil
}

func isMultiPipeline(procs []*model.Proc) bool {
	countPPIDZero := 0
	for _, proc := range procs {
		if proc.PPID == 0 {
			countPPIDZero++
		}
	}
	return countPPIDZero > 1
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

func (s *RPC) completeChildrenIfParentCompleted(procs []*model.Proc, completedProc *model.Proc) {
	for _, p := range procs {
		if p.Running() && p.PPID == completedProc.PID {
			if _, err := shared.UpdateProcToStatusSkipped(s.store, *p, completedProc.Stopped); err != nil {
				log.Error().Msgf("error: done: cannot update proc_id %d child state: %s", p.ID, err)
			}
		}
	}
}

func isThereRunningStage(procs []*model.Proc) bool {
	for _, p := range procs {
		if p.PPID == 0 {
			if p.Running() {
				return true
			}
		}
	}
	return false
}

func buildStatus(procs []*model.Proc) model.StatusValue {
	status := model.StatusSuccess

	for _, p := range procs {
		if p.PPID == 0 {
			if p.Failing() {
				status = p.State
			}
		}
	}

	return status
}

func (s *RPC) updateRemoteStatus(ctx context.Context, repo *model.Repo, build *model.Build, proc *model.Proc) {
	user, err := s.store.GetUser(repo.UserID)
	if err == nil {
		if refresher, ok := s.remote.(remote.Refresher); ok {
			ok, err := refresher.Refresh(ctx, user)
			if err != nil {
				log.Error().Err(err).Msgf("grpc: refresh oauth token of user '%s' failed", user.Login)
			} else if ok {
				if err := s.store.UpdateUser(user); err != nil {
					log.Error().Err(err).Msg("fail to save user to store after refresh oauth token")
				}
			}
		}
		uri := fmt.Sprintf("%s/%s/%d", server.Config.Server.Host, repo.FullName, build.Number)
		err = s.remote.Status(ctx, user, repo, build, uri, proc)
		if err != nil {
			log.Error().Msgf("error setting commit status for %s/%d: %v", repo.FullName, build.Number, err)
		}
	}
}

func (s *RPC) notify(c context.Context, repo *model.Repo, build *model.Build, procs []*model.Proc) {
	build.Procs = model.Tree(procs)
	message := pubsub.Message{
		Labels: map[string]string{
			"repo":    repo.FullName,
			"private": strconv.FormatBool(repo.IsSCMPrivate),
		},
	}
	message.Data, _ = json.Marshal(model.Event{
		Repo:  *repo,
		Build: *build,
	})
	if err := s.pubsub.Publish(c, "topic/events", message); err != nil {
		log.Error().Err(err).Msgf("grpc could not notify event: '%v'", message)
	}
}

func createFilterFunc(filter rpc.Filter) (queue.Filter, error) {
	var st *expr.Selector
	var err error

	if filter.Expr != "" {
		st, err = expr.ParseString(filter.Expr)
		if err != nil {
			return nil, err
		}
	}

	return func(task *queue.Task) bool {
		if st != nil {
			match, _ := st.Eval(expr.NewRow(task.Labels))
			return match
		}

		for k, v := range filter.Labels {
			if task.Labels[k] != v {
				return false
			}
		}
		return true
	}, nil
}

//
//
//

// WoodpeckerServer is a grpc server implementation.
type WoodpeckerServer struct {
	proto.UnimplementedWoodpeckerServer
	peer RPC
}

func NewWoodpeckerServer(remote remote.Remote, queue queue.Queue, logger logging.Log, pubsub pubsub.Publisher, store store.Store, host string) *WoodpeckerServer {
	buildTime := promauto.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: "woodpecker",
		Name:      "build_time",
		Help:      "Build time.",
	}, []string{"repo", "branch", "status", "pipeline"})
	buildCount := promauto.NewCounterVec(prometheus.CounterOpts{
		Namespace: "woodpecker",
		Name:      "build_count",
		Help:      "Build count.",
	}, []string{"repo", "branch", "status", "pipeline"})
	peer := RPC{
		remote:     remote,
		store:      store,
		queue:      queue,
		pubsub:     pubsub,
		logger:     logger,
		host:       host,
		buildTime:  buildTime,
		buildCount: buildCount,
	}
	return &WoodpeckerServer{peer: peer}
}

func (s *WoodpeckerServer) Next(c context.Context, req *proto.NextRequest) (*proto.NextReply, error) {
	filter := rpc.Filter{
		Labels: req.GetFilter().GetLabels(),
		Expr:   req.GetFilter().GetExpr(),
	}

	res := new(proto.NextReply)
	pipeline, err := s.peer.Next(c, filter)
	if err != nil {
		return res, err
	}
	if pipeline == nil {
		return res, err
	}

	res.Pipeline = new(proto.Pipeline)
	res.Pipeline.Id = pipeline.ID
	res.Pipeline.Timeout = pipeline.Timeout
	res.Pipeline.Payload, _ = json.Marshal(pipeline.Config)

	return res, err
}

func (s *WoodpeckerServer) Init(c context.Context, req *proto.InitRequest) (*proto.Empty, error) {
	state := rpc.State{
		Error:    req.GetState().GetError(),
		ExitCode: int(req.GetState().GetExitCode()),
		Finished: req.GetState().GetFinished(),
		Started:  req.GetState().GetStarted(),
		Proc:     req.GetState().GetName(),
		Exited:   req.GetState().GetExited(),
	}
	res := new(proto.Empty)
	err := s.peer.Init(c, req.GetId(), state)
	return res, err
}

func (s *WoodpeckerServer) Update(c context.Context, req *proto.UpdateRequest) (*proto.Empty, error) {
	state := rpc.State{
		Error:    req.GetState().GetError(),
		ExitCode: int(req.GetState().GetExitCode()),
		Finished: req.GetState().GetFinished(),
		Started:  req.GetState().GetStarted(),
		Proc:     req.GetState().GetName(),
		Exited:   req.GetState().GetExited(),
	}
	res := new(proto.Empty)
	err := s.peer.Update(c, req.GetId(), state)
	return res, err
}

func (s *WoodpeckerServer) Upload(c context.Context, req *proto.UploadRequest) (*proto.Empty, error) {
	file := &rpc.File{
		Data: req.GetFile().GetData(),
		Mime: req.GetFile().GetMime(),
		Name: req.GetFile().GetName(),
		Proc: req.GetFile().GetProc(),
		Size: int(req.GetFile().GetSize()),
		Time: req.GetFile().GetTime(),
		Meta: req.GetFile().GetMeta(),
	}

	res := new(proto.Empty)
	err := s.peer.Upload(c, req.GetId(), file)
	return res, err
}

func (s *WoodpeckerServer) Done(c context.Context, req *proto.DoneRequest) (*proto.Empty, error) {
	state := rpc.State{
		Error:    req.GetState().GetError(),
		ExitCode: int(req.GetState().GetExitCode()),
		Finished: req.GetState().GetFinished(),
		Started:  req.GetState().GetStarted(),
		Proc:     req.GetState().GetName(),
		Exited:   req.GetState().GetExited(),
	}
	res := new(proto.Empty)
	err := s.peer.Done(c, req.GetId(), state)
	return res, err
}

func (s *WoodpeckerServer) Wait(c context.Context, req *proto.WaitRequest) (*proto.Empty, error) {
	res := new(proto.Empty)
	err := s.peer.Wait(c, req.GetId())
	return res, err
}

func (s *WoodpeckerServer) Extend(c context.Context, req *proto.ExtendRequest) (*proto.Empty, error) {
	res := new(proto.Empty)
	err := s.peer.Extend(c, req.GetId())
	return res, err
}

func (s *WoodpeckerServer) Log(c context.Context, req *proto.LogRequest) (*proto.Empty, error) {
	line := &rpc.Line{
		Out:  req.GetLine().GetOut(),
		Pos:  int(req.GetLine().GetPos()),
		Time: req.GetLine().GetTime(),
		Proc: req.GetLine().GetProc(),
	}
	res := new(proto.Empty)
	err := s.peer.Log(c, req.GetId(), line)
	return res, err
}
