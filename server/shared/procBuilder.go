// Copyright 2022 Woodpecker Authors
// Copyright 2018 Drone.IO Inc.
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

package shared

import (
	"fmt"
	"math/rand"
	"net/url"
	"path/filepath"
	"sort"
	"strings"

	"github.com/drone/envsubst"
	"github.com/rs/zerolog/log"

	backend "github.com/woodpecker-ci/woodpecker/pipeline/backend/types"
	"github.com/woodpecker-ci/woodpecker/pipeline/frontend"
	"github.com/woodpecker-ci/woodpecker/pipeline/frontend/yaml"
	"github.com/woodpecker-ci/woodpecker/pipeline/frontend/yaml/compiler"
	"github.com/woodpecker-ci/woodpecker/pipeline/frontend/yaml/linter"
	"github.com/woodpecker-ci/woodpecker/pipeline/frontend/yaml/matrix"
	"github.com/woodpecker-ci/woodpecker/server"
	"github.com/woodpecker-ci/woodpecker/server/forge"
	"github.com/woodpecker-ci/woodpecker/server/model"
)

// TODO(974) move to pipeline/*

// ProcBuilder Takes the hook data and the yaml and returns in internal data model
type ProcBuilder struct {
	Repo  *model.Repo
	Curr  *model.Pipeline
	Last  *model.Pipeline
	Netrc *model.Netrc
	Secs  []*model.Secret
	Regs  []*model.Registry
	Link  string
	Yamls []*forge.FileMeta
	Envs  map[string]string
}

type PipelineItem struct {
	Proc      *model.Proc
	Platform  string
	Labels    map[string]string
	DependsOn []string
	RunsOn    []string
	Config    *backend.Config
}

func (b *ProcBuilder) Build() ([]*PipelineItem, error) {
	var items []*PipelineItem

	sort.Sort(forge.ByName(b.Yamls))

	pidSequence := 1

	for _, y := range b.Yamls {
		// matrix axes
		axes, err := matrix.ParseString(string(y.Data))
		if err != nil {
			return nil, err
		}
		if len(axes) == 0 {
			axes = append(axes, matrix.Axis{})
		}

		for _, axis := range axes {
			proc := &model.Proc{
				PipelineID: b.Curr.ID,
				PID:        pidSequence,
				PGID:       pidSequence,
				State:      model.StatusPending,
				Environ:    axis,
				Name:       SanitizePath(y.Name),
			}

			metadata := metadataFromStruct(b.Repo, b.Curr, b.Last, proc, b.Link)
			environ := b.environmentVariables(metadata, axis)

			// add global environment variables for substituting
			for k, v := range b.Envs {
				if _, exists := environ[k]; exists {
					// don't override existing values
					continue
				}
				environ[k] = v
			}

			// substitute vars
			substituted, err := b.envsubst(string(y.Data), environ)
			if err != nil {
				return nil, err
			}

			// parse yaml pipeline
			parsed, err := yaml.ParseString(substituted)
			if err != nil {
				return nil, &yaml.PipelineParseError{Err: err}
			}

			// lint pipeline
			if err := linter.New(
				linter.WithTrusted(b.Repo.IsTrusted),
			).Lint(parsed); err != nil {
				return nil, &yaml.PipelineParseError{Err: err}
			}

			// checking if filtered.
			if match, err := parsed.When.Match(metadata, true); !match && err == nil {
				log.Debug().Str("pipeline", proc.Name).Msg(
					"Marked as skipped, dose not match metadata",
				)
				proc.State = model.StatusSkipped
			} else if err != nil {
				log.Debug().Str("pipeline", proc.Name).Msg(
					"Pipeline config could not be parsed",
				)
				return nil, err
			}

			// TODO: deprecated branches filter => remove after some time
			if !parsed.Branches.Match(b.Curr.Branch) && (b.Curr.Event != model.EventDeploy && b.Curr.Event != model.EventTag) {
				log.Debug().Str("pipeline", proc.Name).Msg(
					"Marked as skipped, dose not match branch",
				)
				proc.State = model.StatusSkipped
			}

			ir, err := b.toInternalRepresentation(parsed, environ, metadata, proc.ID)
			if err != nil {
				return nil, err
			}

			if len(ir.Stages) == 0 {
				continue
			}

			item := &PipelineItem{
				Proc:      proc,
				Config:    ir,
				Labels:    parsed.Labels,
				DependsOn: parsed.DependsOn,
				RunsOn:    parsed.RunsOn,
				Platform:  parsed.Platform,
			}
			if item.Labels == nil {
				item.Labels = map[string]string{}
			}

			items = append(items, item)
			pidSequence++
		}
	}

	items = filterItemsWithMissingDependencies(items)

	// check if at least one proc can start, if list is not empty
	if len(items) > 0 && !procListContainsItemsToRun(items) {
		return nil, fmt.Errorf("pipeline has no startpoint")
	}

	return items, nil
}

func procListContainsItemsToRun(items []*PipelineItem) bool {
	for i := range items {
		if items[i].Proc.State == model.StatusPending {
			return true
		}
	}
	return false
}

func filterItemsWithMissingDependencies(items []*PipelineItem) []*PipelineItem {
	itemsToRemove := make([]*PipelineItem, 0)

	for _, item := range items {
		for _, dep := range item.DependsOn {
			if !containsItemWithName(dep, items) {
				itemsToRemove = append(itemsToRemove, item)
			}
		}
	}

	if len(itemsToRemove) > 0 {
		filtered := make([]*PipelineItem, 0)
		for _, item := range items {
			if !containsItemWithName(item.Proc.Name, itemsToRemove) {
				filtered = append(filtered, item)
			}
		}
		// Recursive to handle transitive deps
		return filterItemsWithMissingDependencies(filtered)
	}

	return items
}

func containsItemWithName(name string, items []*PipelineItem) bool {
	for _, item := range items {
		if name == item.Proc.Name {
			return true
		}
	}
	return false
}

func (b *ProcBuilder) envsubst(y string, environ map[string]string) (string, error) {
	return envsubst.Eval(y, func(name string) string {
		env := environ[name]
		if strings.Contains(env, "\n") {
			env = fmt.Sprintf("%q", env)
		}
		return env
	})
}

func (b *ProcBuilder) environmentVariables(metadata frontend.Metadata, axis matrix.Axis) map[string]string {
	environ := metadata.Environ()
	for k, v := range axis {
		environ[k] = v
	}
	return environ
}

func (b *ProcBuilder) toInternalRepresentation(parsed *yaml.Config, environ map[string]string, metadata frontend.Metadata, procID int64) (*backend.Config, error) {
	var secrets []compiler.Secret
	for _, sec := range b.Secs {
		if !sec.Match(b.Curr.Event) {
			continue
		}
		secrets = append(secrets, compiler.Secret{
			Name:       sec.Name,
			Value:      sec.Value,
			Match:      sec.Images,
			PluginOnly: sec.PluginsOnly,
		})
	}

	var registries []compiler.Registry
	for _, reg := range b.Regs {
		registries = append(registries, compiler.Registry{
			Hostname: reg.Address,
			Username: reg.Username,
			Password: reg.Password,
			Email:    reg.Email,
		})
	}

	return compiler.New(
		compiler.WithEnviron(environ),
		compiler.WithEnviron(b.Envs),
		compiler.WithEscalated(server.Config.Pipeline.Privileged...),
		compiler.WithResourceLimit(server.Config.Pipeline.Limits.MemSwapLimit, server.Config.Pipeline.Limits.MemLimit, server.Config.Pipeline.Limits.ShmSize, server.Config.Pipeline.Limits.CPUQuota, server.Config.Pipeline.Limits.CPUShares, server.Config.Pipeline.Limits.CPUSet),
		compiler.WithVolumes(server.Config.Pipeline.Volumes...),
		compiler.WithNetworks(server.Config.Pipeline.Networks...),
		compiler.WithLocal(false),
		compiler.WithOption(
			compiler.WithNetrc(
				b.Netrc.Login,
				b.Netrc.Password,
				b.Netrc.Machine,
			),
			b.Repo.IsSCMPrivate || server.Config.Pipeline.AuthenticatePublicRepos,
		),
		compiler.WithDefaultCloneImage(server.Config.Pipeline.DefaultCloneImage),
		compiler.WithRegistry(registries...),
		compiler.WithSecret(secrets...),
		compiler.WithPrefix(
			fmt.Sprintf(
				"wp_%d_%d",
				procID,
				rand.Int(),
			),
		),
		compiler.WithProxy(),
		compiler.WithWorkspaceFromURL("/woodpecker", b.Repo.Link),
		compiler.WithMetadata(metadata),
	).Compile(parsed)
}

func SetPipelineStepsOnPipeline(pipeline *model.Pipeline, pipelineItems []*PipelineItem) *model.Pipeline {
	var pidSequence int
	for _, item := range pipelineItems {
		pipeline.Procs = append(pipeline.Procs, item.Proc)
		if pidSequence < item.Proc.PID {
			pidSequence = item.Proc.PID
		}
	}

	for _, item := range pipelineItems {
		for _, stage := range item.Config.Stages {
			var gid int
			for _, step := range stage.Steps {
				pidSequence++
				if gid == 0 {
					gid = pidSequence
				}
				proc := &model.Proc{
					PipelineID: pipeline.ID,
					Name:       step.Alias,
					PID:        pidSequence,
					PPID:       item.Proc.PID,
					PGID:       gid,
					State:      model.StatusPending,
				}
				if item.Proc.State == model.StatusSkipped {
					proc.State = model.StatusSkipped
				}
				pipeline.Procs = append(pipeline.Procs, proc)
			}
		}
	}

	return pipeline
}

// return the metadata from the cli context.
func metadataFromStruct(repo *model.Repo, pipeline, last *model.Pipeline, proc *model.Proc, link string) frontend.Metadata {
	host := link
	uri, err := url.Parse(link)
	if err == nil {
		host = uri.Host
	}
	return frontend.Metadata{
		Repo: frontend.Repo{
			Name:     repo.FullName,
			Link:     repo.Link,
			CloneURL: repo.Clone,
			Private:  repo.IsSCMPrivate,
			Branch:   repo.Branch,
		},
		Curr: metadataPipelineFromModelPipeline(pipeline, true),
		Prev: metadataPipelineFromModelPipeline(last, false),
		Job: frontend.Job{
			Number: proc.PID,
			Matrix: proc.Environ,
		},
		Sys: frontend.System{
			Name:     "woodpecker",
			Link:     link,
			Host:     host,
			Platform: "", // will be set by pipeline platform option or by agent
		},
	}
}

func metadataPipelineFromModelPipeline(pipeline *model.Pipeline, includeParent bool) frontend.Pipeline {
	cron := ""
	if pipeline.Event == model.EventCron {
		cron = pipeline.Sender
	}

	parent := int64(0)
	if includeParent {
		parent = pipeline.Parent
	}

	return frontend.Pipeline{
		Number:   pipeline.Number,
		Parent:   parent,
		Created:  pipeline.Created,
		Started:  pipeline.Started,
		Finished: pipeline.Finished,
		Status:   string(pipeline.Status),
		Event:    string(pipeline.Event),
		Link:     pipeline.Link,
		Target:   pipeline.Deploy,
		Commit: frontend.Commit{
			Sha:     pipeline.Commit,
			Ref:     pipeline.Ref,
			Refspec: pipeline.Refspec,
			Branch:  pipeline.Branch,
			Message: pipeline.Message,
			Author: frontend.Author{
				Name:   pipeline.Author,
				Email:  pipeline.Email,
				Avatar: pipeline.Avatar,
			},
			ChangedFiles: pipeline.ChangedFiles,
		},
		Cron: cron,
	}
}

func SanitizePath(path string) string {
	path = filepath.Base(path)
	path = strings.TrimSuffix(path, ".yml")
	path = strings.TrimPrefix(path, ".")
	return path
}
