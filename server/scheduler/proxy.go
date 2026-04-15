// Copyright 2026 Woodpecker Authors
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

package scheduler

import (
	"context"

	"go.woodpecker-ci.org/woodpecker/v3/server/model"
	"go.woodpecker-ci.org/woodpecker/v3/server/pubsub"
	"go.woodpecker-ci.org/woodpecker/v3/server/queue"
)

type proxy struct {
	q  queue.Queue
	ps pubsub.PubSub
}

//
// Queue.
//

func (p *proxy) Done(c context.Context, id string, exitStatus model.StatusValue) error {
	return p.q.Done(c, id, exitStatus)
}

func (p *proxy) Error(c context.Context, id string, err error) error {
	return p.q.Error(c, id, err)
}

func (p *proxy) ErrorAtOnce(c context.Context, ids []string, err error) error {
	return p.q.ErrorAtOnce(c, ids, err)
}

func (p *proxy) Extend(c context.Context, agentID int64, workflowID string) error {
	return p.q.Extend(c, agentID, workflowID)
}

func (p *proxy) Info(c context.Context) queue.InfoT {
	return p.q.Info(c)
}

func (p *proxy) KickAgentWorkers(agentID int64) {
	p.q.KickAgentWorkers(agentID)
}

func (p *proxy) Pause() {
	p.q.Pause()
}

func (p *proxy) Poll(c context.Context, agentID int64, f queue.FilterFn) (*model.Task, error) {
	return p.q.Poll(c, agentID, f)
}

func (p *proxy) PushAtOnce(c context.Context, tasks []*model.Task) error {
	return p.q.PushAtOnce(c, tasks)
}

func (p *proxy) Resume() {
	p.q.Resume()
}

func (p *proxy) Wait(c context.Context, id string) error {
	return p.q.Wait(c, id)
}

//
// PubSub.
//

func (p *proxy) Subscribe(c context.Context, t pubsub.Topics, r pubsub.Receiver) error {
	return p.ps.Subscribe(c, t, r)
}

func (p *proxy) Publish(c context.Context, t pubsub.Topics, m pubsub.Message) error {
	return p.ps.Publish(c, t, m)
}
