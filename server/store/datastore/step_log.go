// Copyright 2026 Woodpecker Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package datastore

import (
	"context"

	"xorm.io/builder"
	"xorm.io/xorm"

	"go.woodpecker-ci.org/woodpecker/v3/server/model"
)

// stepLogCond selects the steps of filter, ignoring Limit. A complete log is the counterpart of
// allowAppendingLogs: a running step may still get lines, even in a stopped pipeline.
// Steps that never started are included, so they leave the selection too.
func stepLogCond(filter *model.StepLogFilter) builder.Cond {
	cond := builder.NewCond()
	if len(filter.States) > 0 {
		cond = cond.And(builder.In("steps.log_state", filter.States))
	}
	if filter.CompleteBefore != 0 {
		cond = cond.And(
			builder.Neq{"steps.state": model.StatusRunning},
			builder.Lt{"steps.finished": filter.CompleteBefore},
			builder.NotIn("pipelines.status", model.StatusPending, model.StatusRunning, model.StatusBlocked, model.StatusCreated),
			builder.Lt{"pipelines.finished": filter.CompleteBefore},
		)
	}
	return cond
}

func (s storage) stepLogSession(ctx context.Context, cond builder.Cond) *xorm.Session {
	return s.engine.Context(ctx).Table("steps").Join("INNER", "pipelines", "pipelines.id = steps.pipeline_id").Where(cond)
}

// StepListByLogState orders by step ID, so the log_state index (log_state, id) serves the order and
// the limit: steps are created with their pipeline, oldest pipeline first.
func (s storage) StepListByLogState(ctx context.Context, filter *model.StepLogFilter) ([]*model.Step, error) {
	sess := s.stepLogSession(ctx, stepLogCond(filter)).Cols("steps.*").OrderBy("steps.id ASC")
	if filter.Limit > 0 {
		sess = sess.Limit(filter.Limit)
	}
	steps := make([]*model.Step, 0)
	return steps, sess.Find(&steps)
}

func (s storage) StepCountByLogState(ctx context.Context, filter *model.StepLogFilter) (int64, error) {
	return s.stepLogSession(ctx, stepLogCond(filter)).Count()
}

func (s storage) StepLogStateSwap(ctx context.Context, step *model.Step, to model.LogState) (bool, error) {
	var swapped int64
	err := retryWriteCtx(ctx, func() (err error) {
		swapped, err = s.engine.Context(ctx).Table("steps").
			Where(builder.Eq{"id": step.ID, "log_state": step.LogState}).
			Update(map[string]any{"log_state": to})
		return wrapUpdate(swapped, err)
	})
	if err != nil || swapped == 0 {
		return false, err
	}
	step.LogState = to
	return true, nil
}
