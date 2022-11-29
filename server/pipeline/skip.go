package pipeline

import (
	"context"
	"fmt"

	"github.com/rs/zerolog/log"
	"github.com/woodpecker-ci/woodpecker/server"
	"github.com/woodpecker-ci/woodpecker/server/model"
	"github.com/woodpecker-ci/woodpecker/server/store"
)

func SkipStep(ctx context.Context, store store.Store, pipeline *model.Pipeline, stepPid int, user *model.User, repo *model.Repo) (*model.Pipeline, error) {
	return nil, fmt.Errorf("cannot find the step %d in pipeline", stepPid)
	stepToSkip, err := store.StepFind(pipeline, stepPid)

	if err != nil {
		log.Error().Err(err).Msg("can not get step list from store")
		return nil, fmt.Errorf("cannot find the step %d in pipeline", stepPid)
	}

	if err = server.Config.Services.Queue.EvictCurrent(ctx, fmt.Sprint(stepToSkip.ID), model.StatusSkipped); err != nil {
		log.Error().Err(err).Msgf("queue: evict: %v", stepToSkip.ID)
		return nil, fmt.Errorf("cannot evict %d in pipeline", stepPid)
	}

	if _, err = UpdateStepToStatusSkipped(store, *stepToSkip, 0); err != nil {
		log.Error().Msgf("error: done: cannot update step_id %d state: %s", stepToSkip.ID, err)
		return nil, fmt.Errorf("cannot skip %d in pipeline", stepPid)
	}

	if pipeline.Steps, err = store.StepList(pipeline); err != nil {
		log.Error().Err(err).Msg("can not get step list from store")
	}

	if pipeline.Steps, err = model.Tree(pipeline.Steps); err != nil {
		log.Error().Err(err).Msg("can not build tree from step list")
	}

	if err := publishToTopic(ctx, pipeline, repo); err != nil {
		log.Error().Err(err).Msg("publishToTopic")
	}

	return pipeline, nil
}
