// Copyright 2025 Woodpecker Authors
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

package registry

import (
	"context"

	"github.com/rs/zerolog/log"

	"go.woodpecker-ci.org/woodpecker/v3/server/model"
)

type withExtension struct {
	base      Service
	extension *httpExtension
}

// NewWithExtension returns a registry service that combines a base service with an HTTP extension.
// The extension is called during RegistryListPipeline to fetch additional registry credentials.
func NewWithExtension(base Service, extension *httpExtension) Service {
	return &withExtension{base, extension}
}

func (w *withExtension) RegistryListPipeline(ctx context.Context, repo *model.Repo, pipeline *model.Pipeline) ([]*model.Registry, error) {
	// Get registries from base service
	baseRegistries, err := w.base.RegistryListPipeline(ctx, repo, pipeline)
	if err != nil {
		return nil, err
	}

	// Get registries from HTTP extension
	extensionRegistries, err := w.extension.RegistryListPipeline(ctx, repo, pipeline)
	if err != nil {
		// Log the error but don't fail - use base registries only
		log.Warn().Err(err).Msg("failed to fetch registries from extension")
		return baseRegistries, nil
	}

	if len(extensionRegistries) == 0 {
		return baseRegistries, nil
	}

	// Merge registries, with extension registries taking priority (no duplicates by address)
	exists := make(map[string]struct{}, len(extensionRegistries))
	for _, reg := range extensionRegistries {
		exists[reg.Address] = struct{}{}
	}

	merged := make([]*model.Registry, 0, len(baseRegistries)+len(extensionRegistries))
	merged = append(merged, extensionRegistries...)

	for _, reg := range baseRegistries {
		if _, ok := exists[reg.Address]; ok {
			continue
		}
		exists[reg.Address] = struct{}{}
		merged = append(merged, reg)
	}

	return merged, nil
}

// All other methods delegate to the base service.

func (w *withExtension) RegistryFind(repo *model.Repo, addr string) (*model.Registry, error) {
	return w.base.RegistryFind(repo, addr)
}

func (w *withExtension) RegistryList(repo *model.Repo, p *model.ListOptions) ([]*model.Registry, error) {
	return w.base.RegistryList(repo, p)
}

func (w *withExtension) RegistryCreate(repo *model.Repo, registry *model.Registry) error {
	return w.base.RegistryCreate(repo, registry)
}

func (w *withExtension) RegistryUpdate(repo *model.Repo, registry *model.Registry) error {
	return w.base.RegistryUpdate(repo, registry)
}

func (w *withExtension) RegistryDelete(repo *model.Repo, addr string) error {
	return w.base.RegistryDelete(repo, addr)
}

func (w *withExtension) OrgRegistryFind(owner int64, addr string) (*model.Registry, error) {
	return w.base.OrgRegistryFind(owner, addr)
}

func (w *withExtension) OrgRegistryList(owner int64, p *model.ListOptions) ([]*model.Registry, error) {
	return w.base.OrgRegistryList(owner, p)
}

func (w *withExtension) OrgRegistryCreate(owner int64, registry *model.Registry) error {
	return w.base.OrgRegistryCreate(owner, registry)
}

func (w *withExtension) OrgRegistryUpdate(owner int64, registry *model.Registry) error {
	return w.base.OrgRegistryUpdate(owner, registry)
}

func (w *withExtension) OrgRegistryDelete(owner int64, addr string) error {
	return w.base.OrgRegistryDelete(owner, addr)
}

func (w *withExtension) GlobalRegistryFind(addr string) (*model.Registry, error) {
	return w.base.GlobalRegistryFind(addr)
}

func (w *withExtension) GlobalRegistryList(p *model.ListOptions) ([]*model.Registry, error) {
	return w.base.GlobalRegistryList(p)
}

func (w *withExtension) GlobalRegistryCreate(registry *model.Registry) error {
	return w.base.GlobalRegistryCreate(registry)
}

func (w *withExtension) GlobalRegistryUpdate(registry *model.Registry) error {
	return w.base.GlobalRegistryUpdate(registry)
}

func (w *withExtension) GlobalRegistryDelete(addr string) error {
	return w.base.GlobalRegistryDelete(addr)
}
