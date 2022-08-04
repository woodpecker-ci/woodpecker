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

	backend "github.com/woodpecker-ci/woodpecker/pipeline/backend/types"
	"github.com/woodpecker-ci/woodpecker/pipeline/frontend"
	"github.com/woodpecker-ci/woodpecker/pipeline/frontend/yaml"
	"github.com/woodpecker-ci/woodpecker/pipeline/frontend/yaml/compiler"
	"github.com/woodpecker-ci/woodpecker/pipeline/frontend/yaml/linter"
	"github.com/woodpecker-ci/woodpecker/pipeline/frontend/yaml/matrix"
	"github.com/woodpecker-ci/woodpecker/server"
	"github.com/woodpecker-ci/woodpecker/server/model"
	"github.com/woodpecker-ci/woodpecker/server/remote"
)

// TODO(974) move to pipeline/*

// ProcBuilder Takes the hook data and the yaml and returns in internal data model
type ProcBuilder struct {
	Repo  *model.Repo
	Curr  *model.Build
	Last  *model.Build
	Netrc *model.Netrc
	Secs  []*model.Secret
	Regs  []*model.Registry
	Link  string
	Yamls []*remote.FileMeta
	Envs  map[string]string
}

type BuildItem struct {
	Proc      *model.Proc
	Platform  string
	Labels    map[string]string
	DependsOn []string
	RunsOn    []string
	Config    *backend.Config
}

func (b *ProcBuilder) Build() ([]*BuildItem, error) {
	var items []*BuildItem

	sort.Sort(remote.ByName(b.Yamls))

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
				BuildID: b.Curr.ID,
				PID:     pidSequence,
				PGID:    pidSequence,
				State:   model.StatusPending,
				Environ: axis,
				Name:    SanitizePath(y.Name),
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
				return nil, err
			}

			// lint pipeline
			if err := linter.New(
				linter.WithTrusted(b.Repo.IsTrusted),
			).Lint(parsed); err != nil {
				return nil, err
			}

			if !parsed.Branches.Match(b.Curr.Branch) && (b.Curr.Event != model.EventDeploy && b.Curr.Event != model.EventTag) {
				proc.State = model.StatusSkipped
			}

			ir := b.toInternalRepresentation(parsed, environ, metadata, proc.ID)

			if len(ir.Stages) == 0 {
				continue
			}

			item := &BuildItem{
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
	procListContainsItemsToRun(items)
	if len(items) > 0 && !procListContainsItemsToRun(items) {
		return nil, fmt.Errorf("build has no startpoint")
	}

	return items, nil
}

func procListContainsItemsToRun(items []*BuildItem) bool {
	for i := range items {
		if items[i].Proc.State == model.StatusPending {
			return true
		}
	}
	return false
}

func filterItemsWithMissingDependencies(items []*BuildItem) []*BuildItem {
	itemsToRemove := make([]*BuildItem, 0)

	for _, item := range items {
		for _, dep := range item.DependsOn {
			if !containsItemWithName(dep, items) {
				itemsToRemove = append(itemsToRemove, item)
			}
		}
	}

	if len(itemsToRemove) > 0 {
		filtered := make([]*BuildItem, 0)
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

func containsItemWithName(name string, items []*BuildItem) bool {
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

func (b *ProcBuilder) toInternalRepresentation(parsed *yaml.Config, environ map[string]string, metadata frontend.Metadata, procID int64) *backend.Config {
	var secrets []compiler.Secret
	for _, sec := range b.Secs {
		if !sec.Match(b.Curr.Event) {
			continue
		}
		secrets = append(secrets, compiler.Secret{
			Name:  sec.Name,
			Value: sec.Value,
			Match: sec.Images,
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
				"%d_%d",
				procID,
				rand.Int(),
			),
		),
		compiler.WithProxy(),
		compiler.WithWorkspaceFromURL("/woodpecker", b.Repo.Link),
		compiler.WithMetadata(metadata),
	).Compile(parsed)
}

func SetBuildStepsOnBuild(build *model.Build, buildItems []*BuildItem) *model.Build {
	var pidSequence int
	for _, item := range buildItems {
		build.Procs = append(build.Procs, item.Proc)
		if pidSequence < item.Proc.PID {
			pidSequence = item.Proc.PID
		}
	}

	for _, item := range buildItems {
		for _, stage := range item.Config.Stages {
			var gid int
			for _, step := range stage.Steps {
				pidSequence++
				if gid == 0 {
					gid = pidSequence
				}
				proc := &model.Proc{
					BuildID: build.ID,
					Name:    step.Alias,
					PID:     pidSequence,
					PPID:    item.Proc.PID,
					PGID:    gid,
					State:   model.StatusPending,
				}
				if item.Proc.State == model.StatusSkipped {
					proc.State = model.StatusSkipped
				}
				build.Procs = append(build.Procs, proc)
			}
		}
	}

	return build
}

// return the metadata from the cli context.
func metadataFromStruct(repo *model.Repo, build, last *model.Build, proc *model.Proc, link string) frontend.Metadata {
	host := link
	uri, err := url.Parse(link)
	if err == nil {
		host = uri.Host
	}
	return frontend.Metadata{
		Repo: frontend.Repo{
			Name:    repo.FullName,
			Link:    repo.Link,
			Remote:  repo.Clone,
			Private: repo.IsSCMPrivate,
			Branch:  repo.Branch,
		},
		Curr: frontend.Build{
			Number:   build.Number,
			Parent:   build.Parent,
			Created:  build.Created,
			Started:  build.Started,
			Finished: build.Finished,
			Status:   string(build.Status),
			Event:    string(build.Event),
			Link:     build.Link,
			Target:   build.Deploy,
			Commit: frontend.Commit{
				Sha:     build.Commit,
				Ref:     build.Ref,
				Refspec: build.Refspec,
				Branch:  build.Branch,
				Message: build.Message,
				Author: frontend.Author{
					Name:   build.Author,
					Email:  build.Email,
					Avatar: build.Avatar,
				},
				ChangedFiles: build.ChangedFiles,
			},
		},
		Prev: frontend.Build{
			Number:   last.Number,
			Created:  last.Created,
			Started:  last.Started,
			Finished: last.Finished,
			Status:   string(last.Status),
			Event:    string(last.Event),
			Link:     last.Link,
			Target:   last.Deploy,
			Commit: frontend.Commit{
				Sha:     last.Commit,
				Ref:     last.Ref,
				Refspec: last.Refspec,
				Branch:  last.Branch,
				Message: last.Message,
				Author: frontend.Author{
					Name:   last.Author,
					Email:  last.Email,
					Avatar: last.Avatar,
				},
				ChangedFiles: last.ChangedFiles,
			},
		},
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

func SanitizePath(path string) string {
	path = filepath.Base(path)
	path = strings.TrimSuffix(path, ".yml")
	path = strings.TrimPrefix(path, ".")
	return path
}
