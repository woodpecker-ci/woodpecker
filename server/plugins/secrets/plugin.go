package secrets

import "go.woodpecker-ci.org/woodpecker/v2/server/model"

// SecretService defines a service for managing secrets.
type SecretService interface {
	SecretListPipeline(*model.Repo, *model.Pipeline, *model.ListOptions) ([]*model.Secret, error)
	// Repository secrets
	SecretFind(*model.Repo, string) (*model.Secret, error)
	SecretList(*model.Repo, *model.ListOptions) ([]*model.Secret, error)
	SecretCreate(*model.Repo, *model.Secret) error
	SecretUpdate(*model.Repo, *model.Secret) error
	SecretDelete(*model.Repo, string) error
	// Organization secrets
	OrgSecretFind(int64, string) (*model.Secret, error)
	OrgSecretList(int64, *model.ListOptions) ([]*model.Secret, error)
	OrgSecretCreate(int64, *model.Secret) error
	OrgSecretUpdate(int64, *model.Secret) error
	OrgSecretDelete(int64, string) error
	// Global secrets
	GlobalSecretFind(string) (*model.Secret, error)
	GlobalSecretList(*model.ListOptions) ([]*model.Secret, error)
	GlobalSecretCreate(*model.Secret) error
	GlobalSecretUpdate(*model.Secret) error
	GlobalSecretDelete(string) error
}
