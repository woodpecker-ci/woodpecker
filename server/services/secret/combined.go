// Copyright 2024 Woodpecker Authors
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

package secret

import (
	"go.woodpecker-ci.org/woodpecker/v3/server/model"
)

// combined merges secrets from multiple services.
// For SecretListPipeline, external (HTTP) secrets override DB secrets by name.
// CRUD operations are delegated to the primary (first) service only.
type combined struct {
	primary  Service
	services []Service
}

// NewCombined returns a secret service that merges secrets from multiple sources.
// The first service is the primary and handles CRUD operations.
// For SecretListPipeline, secrets from later services override earlier ones by name.
func NewCombined(services ...Service) Service {
	return &combined{
		primary:  services[0],
		services: services,
	}
}

func (c *combined) SecretListPipeline(repo *model.Repo, pipeline *model.Pipeline, netrc *model.Netrc) ([]*model.Secret, error) {
	secretMap := make(map[string]*model.Secret)
	var orderedNames []string

	for _, svc := range c.services {
		secrets, err := svc.SecretListPipeline(repo, pipeline, netrc)
		if err != nil {
			return nil, err
		}
		for _, s := range secrets {
			if _, exists := secretMap[s.Name]; !exists {
				orderedNames = append(orderedNames, s.Name)
			}
			// Later services override earlier ones by name
			secretMap[s.Name] = s
		}
	}

	result := make([]*model.Secret, 0, len(secretMap))
	for _, name := range orderedNames {
		result = append(result, secretMap[name])
	}

	return result, nil
}

// CRUD operations delegate to the primary service.

func (c *combined) SecretFind(repo *model.Repo, name string) (*model.Secret, error) {
	return c.primary.SecretFind(repo, name)
}

func (c *combined) SecretList(repo *model.Repo, p *model.ListOptions) ([]*model.Secret, error) {
	return c.primary.SecretList(repo, p)
}

func (c *combined) SecretCreate(repo *model.Repo, secret *model.Secret) error {
	return c.primary.SecretCreate(repo, secret)
}

func (c *combined) SecretUpdate(repo *model.Repo, secret *model.Secret) error {
	return c.primary.SecretUpdate(repo, secret)
}

func (c *combined) SecretDelete(repo *model.Repo, name string) error {
	return c.primary.SecretDelete(repo, name)
}

func (c *combined) OrgSecretFind(orgID int64, name string) (*model.Secret, error) {
	return c.primary.OrgSecretFind(orgID, name)
}

func (c *combined) OrgSecretList(orgID int64, p *model.ListOptions) ([]*model.Secret, error) {
	return c.primary.OrgSecretList(orgID, p)
}

func (c *combined) OrgSecretCreate(orgID int64, secret *model.Secret) error {
	return c.primary.OrgSecretCreate(orgID, secret)
}

func (c *combined) OrgSecretUpdate(orgID int64, secret *model.Secret) error {
	return c.primary.OrgSecretUpdate(orgID, secret)
}

func (c *combined) OrgSecretDelete(orgID int64, name string) error {
	return c.primary.OrgSecretDelete(orgID, name)
}

func (c *combined) GlobalSecretFind(name string) (*model.Secret, error) {
	return c.primary.GlobalSecretFind(name)
}

func (c *combined) GlobalSecretList(p *model.ListOptions) ([]*model.Secret, error) {
	return c.primary.GlobalSecretList(p)
}

func (c *combined) GlobalSecretCreate(secret *model.Secret) error {
	return c.primary.GlobalSecretCreate(secret)
}

func (c *combined) GlobalSecretUpdate(secret *model.Secret) error {
	return c.primary.GlobalSecretUpdate(secret)
}

func (c *combined) GlobalSecretDelete(name string) error {
	return c.primary.GlobalSecretDelete(name)
}
