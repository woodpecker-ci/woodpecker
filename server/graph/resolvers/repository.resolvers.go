package resolvers

// This file will be automatically regenerated based on the schema, any resolver implementations
// will be copied through when generating and any unknown code will be moved to the end.

import (
	"context"
	"fmt"

	"github.com/woodpecker-ci/woodpecker/server/graph/generated"
	"github.com/woodpecker-ci/woodpecker/server/model"
)

func (r *repoResolver) Active(ctx context.Context, obj *model.Repo) (bool, error) {
	panic(fmt.Errorf("not implemented"))
}

func (r *repoResolver) Scm(ctx context.Context, obj *model.Repo) (string, error) {
	panic(fmt.Errorf("not implemented"))
}

func (r *repoResolver) AvatarURL(ctx context.Context, obj *model.Repo) (string, error) {
	panic(fmt.Errorf("not implemented"))
}

func (r *repoResolver) LinkURL(ctx context.Context, obj *model.Repo) (string, error) {
	panic(fmt.Errorf("not implemented"))
}

func (r *repoResolver) CloneURL(ctx context.Context, obj *model.Repo) (string, error) {
	panic(fmt.Errorf("not implemented"))
}

func (r *repoResolver) DefaultBranch(ctx context.Context, obj *model.Repo) (string, error) {
	panic(fmt.Errorf("not implemented"))
}

func (r *repoResolver) Private(ctx context.Context, obj *model.Repo) (bool, error) {
	panic(fmt.Errorf("not implemented"))
}

func (r *repoResolver) Trusted(ctx context.Context, obj *model.Repo) (bool, error) {
	panic(fmt.Errorf("not implemented"))
}

func (r *repoResolver) AllowPr(ctx context.Context, obj *model.Repo) (bool, error) {
	panic(fmt.Errorf("not implemented"))
}

func (r *repoResolver) ConfigFile(ctx context.Context, obj *model.Repo) (string, error) {
	panic(fmt.Errorf("not implemented"))
}

func (r *repoResolver) Visibility(ctx context.Context, obj *model.Repo) (string, error) {
	panic(fmt.Errorf("not implemented"))
}

func (r *repoResolver) LastBuild(ctx context.Context, obj *model.Repo) (int, error) {
	panic(fmt.Errorf("not implemented"))
}

func (r *repoResolver) Gated(ctx context.Context, obj *model.Repo) (bool, error) {
	panic(fmt.Errorf("not implemented"))
}

func (r *repoResolver) CancelPreviousPipelineEvents(ctx context.Context, obj *model.Repo) ([]string, error) {
	panic(fmt.Errorf("not implemented"))
}

// Repo returns generated.RepoResolver implementation.
func (r *Resolver) Repo() generated.RepoResolver { return &repoResolver{r} }

type repoResolver struct{ *Resolver }
