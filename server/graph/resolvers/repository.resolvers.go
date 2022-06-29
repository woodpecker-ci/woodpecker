package resolvers

// This file will be automatically regenerated based on the schema, any resolver implementations
// will be copied through when generating and any unknown code will be moved to the end.

import (
	"context"

	"github.com/woodpecker-ci/woodpecker/server/graph/generated"
	"github.com/woodpecker-ci/woodpecker/server/model"
)

func (r *repoResolver) SCMKind(ctx context.Context, obj *model.Repo) (string, error) {
	return string(obj.SCMKind), nil
}

func (r *repoResolver) Visibility(ctx context.Context, obj *model.Repo) (string, error) {
	return string(obj.Visibility), nil
}

func (r *repoResolver) CancelPreviousPipelineEvents(ctx context.Context, obj *model.Repo) ([]string, error) {
	events := make([]string, len(obj.CancelPreviousPipelineEvents))
	for _, event := range obj.CancelPreviousPipelineEvents {
		events = append(events, string(event))
	}
	return events, nil
}

func (r *repoResolver) LastBuild(ctx context.Context, obj *model.Repo) (int, error) {
	// TODO: panic(fmt.Errorf("not implemented"))
	return 1, nil
}

// Repo returns generated.RepoResolver implementation.
func (r *Resolver) Repo() generated.RepoResolver { return &repoResolver{r} }

type repoResolver struct{ *Resolver }
