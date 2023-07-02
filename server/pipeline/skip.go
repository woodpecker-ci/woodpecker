package pipeline

import (
	"context"
	"fmt"

	"github.com/rs/zerolog/log"

	"github.com/woodpecker-ci/woodpecker/server"
	"github.com/woodpecker-ci/woodpecker/server/model"
	"github.com/woodpecker-ci/woodpecker/server/store"
)

func SkipWorkflow(ctx context.Context, store store.Store, pipeline *model.Pipeline, workflowPid int, repo *model.Repo, user *model.User) (*model.Pipeline, error) {
	workflowToSkip, err := store.WorkflowFind(pipeline, workflowPid)
	if err != nil {
		log.Error().Err(err).Msg("can not get workflow list from store")
		return nil, fmt.Errorf("cannot find the workflow %d in pipeline", workflowPid)
	}

	if err = server.Config.Services.Queue.EvictCurrent(ctx, fmt.Sprint(workflowToSkip.ID), model.StatusSkipped); err != nil {
		log.Error().Err(err).Msgf("queue: evict: %v", workflowToSkip.ID)
		return nil, fmt.Errorf("cannot evict %d in pipeline", workflowPid)
	}

	if _, err = UpdateWorkflowToStatusSkipped(store, *workflowToSkip); err != nil {
		log.Error().Msgf("error: done: cannot update step_id %d state: %s", workflowToSkip.ID, err)
		return nil, fmt.Errorf("cannot skip %d in pipeline", workflowPid)
	}

	steps, err := store.StepListFromWorkflowFind(workflowToSkip)
	if err != nil {
		log.Error().Err(err).Msg("can not get step list from store")
	}

	// Skip the children of the skipped step
	for _, step := range steps {
		if _, err = UpdateStepToStatusSkipped(store, *step, 0); err != nil {
			log.Error().Msgf("error: done: cannot update step_id %d state: %s", step.ID, err)
			return nil, fmt.Errorf("cannot skip %d in pipeline", step.PID)
		}
	}

	if pipeline.Workflows, err = store.WorkflowGetTree(pipeline); err != nil {
		log.Error().Err(err).Msg("can not get workflow list from store")
	}

	if err := publishToTopic(ctx, pipeline, repo); err != nil {
		log.Error().Err(err).Msg("publishToTopic")
	}

	updatePipelineStatus(ctx, pipeline, repo, user)

	return pipeline, nil
}
