package graph

// This file will be automatically regenerated based on the schema, any resolver implementations
// will be copied through when generating and any unknown code will be moved to the end.

import (
	"context"
	"fmt"

	generated1 "github.com/woodpecker-ci/woodpecker/server/graph/generated"
	graphModel "github.com/woodpecker-ci/woodpecker/server/graph/model"
	"github.com/woodpecker-ci/woodpecker/server/model"
)

func (r *mutationResolver) CreateTodo(ctx context.Context, input graphModel.NewTodo) (*model.Repo, error) {
	panic(fmt.Errorf("not implemented"))
}

func (r *queryResolver) Repository(ctx context.Context) ([]*model.Repo, error) {
	panic(fmt.Errorf("not implemented"))
}

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

func (r *secretResolver) ID(ctx context.Context, obj *model.Secret) (string, error) {
	panic(fmt.Errorf("not implemented"))
}

func (r *secretResolver) Event(ctx context.Context, obj *model.Secret) ([]string, error) {
	panic(fmt.Errorf("not implemented"))
}

func (r *secretResolver) Image(ctx context.Context, obj *model.Secret) ([]string, error) {
	panic(fmt.Errorf("not implemented"))
}

func (r *userResolver) Name(ctx context.Context, obj *model.User) (string, error) {
	panic(fmt.Errorf("not implemented"))
}

// Mutation returns generated1.MutationResolver implementation.
func (r *Resolver) Mutation() generated1.MutationResolver { return &mutationResolver{r} }

// Query returns generated1.QueryResolver implementation.
func (r *Resolver) Query() generated1.QueryResolver { return &queryResolver{r} }

// Repo returns generated1.RepoResolver implementation.
func (r *Resolver) Repo() generated1.RepoResolver { return &repoResolver{r} }

// Secret returns generated1.SecretResolver implementation.
func (r *Resolver) Secret() generated1.SecretResolver { return &secretResolver{r} }

// User returns generated1.UserResolver implementation.
func (r *Resolver) User() generated1.UserResolver { return &userResolver{r} }

type mutationResolver struct{ *Resolver }
type queryResolver struct{ *Resolver }
type repoResolver struct{ *Resolver }
type secretResolver struct{ *Resolver }
type userResolver struct{ *Resolver }
