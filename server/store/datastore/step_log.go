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
	"time"

	"xorm.io/builder"
	"xorm.io/xorm"

	"go.woodpecker-ci.org/woodpecker/v3/server/model"
)

// unfinalizedCond selects the steps whose log is complete but not finalized yet,
// the counterpart of allowAppendingLogs: a running step may still get lines, even in a stopped pipeline.
// Steps that never started are included, so they leave the selection too.
func unfinalizedCond(finishedBefore int64) builder.Cond {
	return builder.And(
		builder.Eq{"steps.log_state": model.LogStateOpen},
		builder.Neq{"steps.state": model.StatusRunning},
		builder.Lt{"steps.finished": finishedBefore},
		builder.NotIn("pipelines.status", model.StatusPending, model.StatusRunning, model.StatusBlocked, model.StatusCreated),
		builder.Lt{"pipelines.finished": finishedBefore},
	)
}

func (s storage) unfinalizedSession(ctx context.Context, cond builder.Cond) *xorm.Session {
	return s.engine.Context(ctx).Table("steps").Join("INNER", "pipelines", "pipelines.id = steps.pipeline_id").Where(cond)
}

func (s storage) StepListLogUnfinalized(ctx context.Context, finishedBefore, afterPipelineID, untilPipelineID int64, batch int) ([]*model.Step, error) {
	cond := builder.And(
		unfinalizedCond(finishedBefore),
		builder.Gt{"steps.pipeline_id": afterPipelineID},
		builder.Lte{"steps.pipeline_id": untilPipelineID},
	)

	var pipelineIDs []int64
	if err := s.unfinalizedSession(ctx, cond).Distinct("steps.pipeline_id").
		OrderBy("steps.pipeline_id ASC").Limit(batch).Find(&pipelineIDs); err != nil {
		return nil, err
	}

	steps := make([]*model.Step, 0)
	if len(pipelineIDs) == 0 {
		return steps, nil
	}

	err := s.unfinalizedSession(ctx, builder.And(cond, builder.In("steps.pipeline_id", pipelineIDs))).
		Cols("steps.*").OrderBy("steps.pipeline_id ASC, steps.id ASC").Find(&steps)
	return steps, err
}

func (s storage) StepLogUnfinalizedStats(ctx context.Context, finishedBefore int64) (count, maxPipelineID int64, err error) {
	_, err = s.unfinalizedSession(ctx, unfinalizedCond(finishedBefore)).
		Select("COUNT(*), COALESCE(MAX(steps.pipeline_id), 0)").Get(&count, &maxPipelineID)
	return count, maxPipelineID, err
}

func (s storage) StepListLogClaimed(ctx context.Context) ([]*model.Step, error) {
	steps := make([]*model.Step, 0)
	return steps, s.engine.Context(ctx).In("log_state", model.LogStateFinalizing, model.LogStateDeleting).Find(&steps)
}

func (s storage) StepLogStateSwap(ctx context.Context, step *model.Step, to model.LogState) (bool, error) {
	claimed := time.Now().UnixNano()
	var swapped int64
	err := retryWriteCtx(ctx, func() (err error) {
		swapped, err = s.engine.Context(ctx).Table("steps").
			Where(builder.Eq{"id": step.ID, "log_state": step.LogState, "log_claimed": step.LogClaimed}).
			Update(map[string]any{"log_state": to, "log_claimed": claimed})
		return wrapUpdate(swapped, err)
	})
	if err != nil || swapped == 0 {
		return false, err
	}
	step.LogState, step.LogClaimed = to, claimed
	return true, nil
}
