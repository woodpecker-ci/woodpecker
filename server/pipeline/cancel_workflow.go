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

package pipeline

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/rs/zerolog/log"

	"go.woodpecker-ci.org/woodpecker/v3/server"
	"go.woodpecker-ci.org/woodpecker/v3/server/forge"
	"go.woodpecker-ci.org/woodpecker/v3/server/model"
	"go.woodpecker-ci.org/woodpecker/v3/server/queue"
	"go.woodpecker-ci.org/woodpecker/v3/server/store"
	"go.woodpecker-ci.org/woodpecker/v3/server/store/types"
)

// CancelWorkflow stops only the selected workflow; running agents own cleanup.
func CancelWorkflow(ctx context.Context, f forge.Forge, s store.Store, repo *model.Repo, pl *model.Pipeline, id int64) (err error) {
	pipelineID := pl.ID
	unlock := LockLifecycle(pipelineID)
	defer unlock()
	defer func() {
		if err != nil {
			log.Error().Err(err).Int64("repo_id", repo.ID).Int64("pipeline_id", pipelineID).Int64("workflow_id", id).Msg("cancel workflow")
		}
	}()
	w, err := s.WorkflowLoad(id)
	if errors.Is(err, types.ErrRecordNotExist) {
		return &ErrNotFound{Msg: "Workflow not found"}
	}
	if err != nil {
		return err
	}
	if w.PipelineID != pl.ID {
		return &ErrNotFound{Msg: "Workflow not found"}
	}
	pl, err = s.GetPipeline(pl.ID)
	if err != nil {
		return err
	}
	if (pl.Status != model.StatusPending && pl.Status != model.StatusRunning) || !w.Running() {
		return &ErrBadRequest{Msg: "Only pending or running workflows in active pipelines can be canceled"}
	}
	if err = server.Config.Services.Scheduler.CancelWorkflows(ctx, []string{strconv.FormatInt(id, 10)}); err != nil {
		if !errors.Is(err, queue.ErrNotFound) {
			return fmt.Errorf("cancel scheduled workflow: %w", err)
		}
		fresh, loadErr := s.WorkflowLoad(id)
		if loadErr != nil {
			return loadErr
		}
		if fresh.State == model.StatusRunning || fresh.State == model.StatusBlocked || fresh.State == model.StatusCreated {
			return fmt.Errorf("active workflow missing from scheduler: %w", err)
		}
		if fresh.State != model.StatusPending {
			return nil
		}
		// The pending workflow can be visible before startup inserts its task.
		// Mark it terminal now so a later queue insertion cannot initialize it.
		w = fresh
	}
	if w.State == model.StatusRunning {
		return nil
	}
	canceled, err := s.WorkflowCancelPending(id, time.Now().Unix())
	if err != nil {
		return fmt.Errorf("finalize pending workflow: %w", err)
	}
	if !canceled {
		// Initialization or completion can win the conditional update. A
		// canceled scheduler task still carries the stop signal for its agent.
		fresh, err := s.WorkflowLoad(id)
		if err != nil {
			return err
		}
		if fresh.State == model.StatusPending {
			return fmt.Errorf("workflow %d remained pending after cancellation", id)
		}
		return nil
	}
	pl, err = s.GetPipeline(pl.ID)
	if err != nil {
		return err
	}
	pl.Workflows, err = s.WorkflowGetTree(pl)
	if err != nil {
		return err
	}
	if pl.CancelInfo == nil && !model.IsThereRunningStage(pl.Workflows) {
		pl, err = UpdateStatusToDone(s, *pl, PipelineStatus(pl.Workflows), time.Now().Unix())
		if err != nil {
			return err
		}
	}
	for _, fresh := range pl.Workflows {
		if fresh.ID == id {
			w = fresh
			break
		}
	}
	// Use the repository owner credentials, as the RPC completion path does.
	user, err := s.GetUser(repo.UserID)
	if err != nil {
		return err
	}
	forge.Refresh(ctx, f, s, user)
	if err := f.Status(ctx, user, repo, pl, w); err != nil {
		log.Error().Err(err).Int64("workflow_id", id).Msg("update canceled workflow forge status")
	}
	return server.Config.Services.Scheduler.PublishPipelineEvent(ctx, repo, pl)
}
