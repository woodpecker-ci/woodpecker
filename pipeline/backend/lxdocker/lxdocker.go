// Copyright 2023 Woodpecker Authors
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

package lxdocker

import (
	"context"
	"fmt"
	"io"
	"strings"

	"github.com/docker/docker/api/types/volume"
	"github.com/moby/moby/client"

	"github.com/woodpecker-ci/woodpecker/pipeline/backend/docker"
	"github.com/woodpecker-ci/woodpecker/pipeline/backend/lxd"
	"github.com/woodpecker-ci/woodpecker/pipeline/backend/types"
)

type lxdocker struct {
	dockerClient *client.Client
	docker       types.Engine
	lxd          types.Engine
	workflows    map[string]*meldWorkflow
}

type meldWorkflow struct {
	volumes    map[string]string
	dockerConf *types.Config
	lxdConf    *types.Config
}

// New returns a new lxdocker Engine.
func New() types.Engine {
	return &lxdocker{
		docker:    docker.New(),
		lxd:       lxd.New(),
		workflows: make(map[string]*meldWorkflow),
	}
}

func (e *lxdocker) Name() string {
	return "lxdocker"
}

func (e *lxdocker) IsAvailable(ctx context.Context) bool {
	return e.docker.IsAvailable(ctx) && e.lxd.IsAvailable(ctx)
}

func (e *lxdocker) Load(ctx context.Context) error {
	dockerClient, err := client.NewClientWithOpts(client.FromEnv)
	if err != nil {
		return err
	}
	e.dockerClient = dockerClient
	err = e.docker.Load(ctx)
	if err != nil {
		return err
	}
	return e.lxd.Load(ctx)
}

// SetupWorkflow the pipeline environment.
func (e *lxdocker) SetupWorkflow(ctx context.Context, conf *types.Config, taskUUID string) (err error) {
	cleanups := []func(){}
	defer func() {
		if err != nil {
			for _, f := range cleanups {
				f()
			}
		}
	}()

	volumes := map[string]string{}
	for _, v := range conf.Volumes {
		v := v
		vol, err := e.dockerClient.VolumeCreate(ctx, volume.VolumeCreateBody{
			Name:   volID(v.Name, taskUUID),
			Driver: "local",
		})
		if err != nil {
			return err
		}
		cleanups = append(cleanups, func() {
			_ = e.dockerClient.VolumeRemove(ctx, volID(v.Name, taskUUID), true)
		})
		volumes[v.Name] = vol.Mountpoint
	}

	mw := &meldWorkflow{
		volumes: volumes,
		dockerConf: filterConfig(conf, volumes, func(step *types.Step) bool {
			return !strings.HasPrefix(step.Image, "lxd:")
		}),
		lxdConf: filterConfig(conf, volumes, func(step *types.Step) bool {
			return strings.HasPrefix(step.Image, "lxd:")
		}),
	}

	err = e.docker.SetupWorkflow(ctx, mw.dockerConf, taskUUID)
	if err != nil {
		return err
	}
	cleanups = append(cleanups, func() {
		_ = e.docker.DestroyWorkflow(ctx, mw.dockerConf, taskUUID)
	})

	err = e.lxd.SetupWorkflow(ctx, mw.lxdConf, taskUUID)
	if err != nil {
		return err
	}
	e.workflows[taskUUID] = mw
	return nil
}

// StartStep the pipeline step.
func (e *lxdocker) StartStep(ctx context.Context, step *types.Step, taskUUID string) error {
	mw, ok := e.workflows[taskUUID]
	if !ok {
		return fmt.Errorf("workflow %s not found", taskUUID)
	}
	step = mapStep(step, mw.volumes)
	if strings.HasPrefix(step.Image, "lxd:") {
		return e.lxd.StartStep(ctx, step, taskUUID)
	}
	return e.docker.StartStep(ctx, step, taskUUID)
}

// WaitStep for the pipeline step to complete and returns
// the completion results.
func (e *lxdocker) WaitStep(ctx context.Context, step *types.Step, taskUUID string) (*types.State, error) {
	mw, ok := e.workflows[taskUUID]
	if !ok {
		return nil, fmt.Errorf("workflow %s not found", taskUUID)
	}
	step = mapStep(step, mw.volumes)
	if strings.HasPrefix(step.Image, "lxd:") {
		return e.lxd.WaitStep(ctx, step, taskUUID)
	}
	return e.docker.WaitStep(ctx, step, taskUUID)
}

// TailStep the pipeline step logs.
func (e *lxdocker) TailStep(ctx context.Context, step *types.Step, taskUUID string) (io.ReadCloser, error) {
	mw, ok := e.workflows[taskUUID]
	if !ok {
		return nil, fmt.Errorf("workflow %s not found", taskUUID)
	}
	step = mapStep(step, mw.volumes)
	if strings.HasPrefix(step.Image, "lxd:") {
		return e.lxd.TailStep(ctx, step, taskUUID)
	}
	return e.docker.TailStep(ctx, step, taskUUID)
}

// DestroyWorkflow the pipeline environment.
func (e *lxdocker) DestroyWorkflow(ctx context.Context, _ *types.Config, taskUUID string) error {
	mw, ok := e.workflows[taskUUID]
	if !ok {
		return nil
	}

	err := e.lxd.DestroyWorkflow(ctx, mw.lxdConf, taskUUID)
	if err != nil {
		return err
	}

	err = e.docker.DestroyWorkflow(ctx, mw.dockerConf, taskUUID)
	if err != nil {
		return err
	}

	for volName := range mw.volumes {
		err := e.dockerClient.VolumeRemove(ctx, volID(volName, taskUUID), true)
		if err != nil {
			return err
		}
	}

	delete(e.workflows, taskUUID)
	return nil
}

func filterConfig(conf *types.Config, volumes map[string]string, stepFilter func(*types.Step) bool) *types.Config {
	filtered := &types.Config{
		Networks: conf.Networks,
		Secrets:  conf.Secrets,
	}
	for _, src := range conf.Stages {
		stage := &types.Stage{
			Name:  src.Name,
			Alias: src.Alias,
		}
		for _, step := range src.Steps {
			if !stepFilter(step) {
				continue
			}
			stage.Steps = append(stage.Steps, mapStep(step, volumes))
		}
		filtered.Stages = append(filtered.Stages, stage)
	}
	return filtered
}

func mapStep(step *types.Step, volumes map[string]string) *types.Step {
	mapped := *step
	mapped.Volumes = nil

	for _, volume := range step.Volumes {
		parts := strings.Split(volume, ":")
		if len(parts) < 2 {
			continue
		}
		if strings.HasPrefix(parts[0], "/") {
			mapped.Volumes = append(mapped.Volumes, volume)
			continue
		}
		mappedPath, ok := volumes[parts[0]]
		if !ok {
			continue
		}
		mappedVol := strings.Join(append([]string{mappedPath}, parts[1:]...), ":")
		mapped.Volumes = append(mapped.Volumes, mappedVol)
	}

	return &mapped
}

func volID(name, taskUUID string) string {
	return fmt.Sprintf("%s-%s", taskUUID, name)
}
