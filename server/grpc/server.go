// Copyright 2022 Woodpecker Authors
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

	"github.com/prometheus/client_golang/prometheus"
	prometheus_auto "github.com/prometheus/client_golang/prometheus/promauto"

	"go.woodpecker-ci.org/woodpecker/v2/pipeline/rpc"
	"go.woodpecker-ci.org/woodpecker/v2/pipeline/rpc/proto"
	"go.woodpecker-ci.org/woodpecker/v2/server/logging"
	"go.woodpecker-ci.org/woodpecker/v2/server/pubsub"
	"go.woodpecker-ci.org/woodpecker/v2/server/queue"
	"go.woodpecker-ci.org/woodpecker/v2/server/store"
	"go.woodpecker-ci.org/woodpecker/v2/version"
)

// WoodpeckerServer is a grpc server implementation.
type WoodpeckerServer struct {
	proto.UnimplementedWoodpeckerServer
	peer RPC
}

func NewWoodpeckerServer(queue queue.Queue, logger logging.Log, pubsub *pubsub.Publisher, store store.Store) proto.WoodpeckerServer {
	pipelineTime := prometheus_auto.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: "woodpecker",
		Name:      "pipeline_time",
		Help:      "Pipeline time.",
	}, []string{"repo", "branch", "status", "pipeline"})
	pipelineCount := prometheus_auto.NewCounterVec(prometheus.CounterOpts{
		Namespace: "woodpecker",
		Name:      "pipeline_count",
		Help:      "Pipeline count.",
	}, []string{"repo", "branch", "status", "pipeline"})
	peer := RPC{
		store:         store,
		queue:         queue,
		pubsub:        pubsub,
		logger:        logger,
		pipelineTime:  pipelineTime,
		pipelineCount: pipelineCount,
	}
	return &WoodpeckerServer{peer: peer}
}

func (s *WoodpeckerServer) Version(_ context.Context, _ *proto.Empty) (*proto.VersionResponse, error) {
	return &proto.VersionResponse{
		GrpcVersion:   proto.Version,
		ServerVersion: version.String(),
	}, nil
}

func (s *WoodpeckerServer) Next(c context.Context, req *proto.NextRequest) (*proto.NextResponse, error) {
	filter := rpc.Filter{
		Labels: req.GetFilter().GetLabels(),
	}

	res := new(proto.NextResponse)
	pipeline, err := s.peer.Next(c, filter)
	if err != nil || pipeline == nil {
		return res, err
	}

	res.Workflow = new(proto.Workflow)
	res.Workflow.Id = pipeline.ID
	res.Workflow.Timeout = pipeline.Timeout
	res.Workflow.Payload, err = json.Marshal(pipeline.Config)

	return res, err
}

func (s *WoodpeckerServer) Init(c context.Context, req *proto.InitRequest) (*proto.Empty, error) {
	state := rpc.WorkflowState{
		Started:  req.GetState().GetStarted(),
		Finished: req.GetState().GetFinished(),
		Error:    req.GetState().GetError(),
	}
	res := new(proto.Empty)
	err := s.peer.Init(c, req.GetId(), state)
	return res, err
}

func (s *WoodpeckerServer) Update(c context.Context, req *proto.UpdateRequest) (*proto.Empty, error) {
	state := rpc.StepState{
		StepUUID: req.GetState().GetStepUuid(),
		Started:  req.GetState().GetStarted(),
		Finished: req.GetState().GetFinished(),
		Exited:   req.GetState().GetExited(),
		Error:    req.GetState().GetError(),
		ExitCode: int(req.GetState().GetExitCode()),
	}
	res := new(proto.Empty)
	err := s.peer.Update(c, req.GetId(), state)
	return res, err
}

func (s *WoodpeckerServer) Done(c context.Context, req *proto.DoneRequest) (*proto.Empty, error) {
	state := rpc.WorkflowState{
		Started:  req.GetState().GetStarted(),
		Finished: req.GetState().GetFinished(),
		Error:    req.GetState().GetError(),
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
	logEntry := &rpc.LogEntry{
		Data:     req.GetLogEntry().GetData(),
		Line:     int(req.GetLogEntry().GetLine()),
		Time:     req.GetLogEntry().GetTime(),
		StepUUID: req.GetLogEntry().GetStepUuid(),
		Type:     int(req.GetLogEntry().GetType()),
	}
	res := new(proto.Empty)
	err := s.peer.Log(c, logEntry)
	return res, err
}

func (s *WoodpeckerServer) RegisterAgent(c context.Context, req *proto.RegisterAgentRequest) (*proto.RegisterAgentResponse, error) {
	res := new(proto.RegisterAgentResponse)
	agentID, err := s.peer.RegisterAgent(c, req.GetPlatform(), req.GetBackend(), req.GetVersion(), req.GetCapacity())
	res.AgentId = agentID
	return res, err
}

func (s *WoodpeckerServer) UnregisterAgent(ctx context.Context, _ *proto.Empty) (*proto.Empty, error) {
	err := s.peer.UnregisterAgent(ctx)
	return new(proto.Empty), err
}

func (s *WoodpeckerServer) ReportHealth(c context.Context, req *proto.ReportHealthRequest) (*proto.Empty, error) {
	res := new(proto.Empty)
	err := s.peer.ReportHealth(c, req.GetStatus())
	return res, err
}
