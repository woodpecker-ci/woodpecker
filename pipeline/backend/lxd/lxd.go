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

package lxd

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	lxd "github.com/canonical/lxd/client"
	"github.com/canonical/lxd/shared/api"
	"github.com/rs/zerolog/log"

	"github.com/woodpecker-ci/woodpecker/pipeline/backend/common"
	"github.com/woodpecker-ci/woodpecker/pipeline/backend/types"
	"github.com/woodpecker-ci/woodpecker/shared/constant"
)

const (
	defaultSimpleStreamsRepo = "https://images.linuxcontainers.org"
	lxdSnapUnixSocket        = "/var/snap/lxd/common/lxd/unix.socket"
)

type backend struct {
	inst lxd.InstanceServer

	active map[string]*backendStep
}

type backendStep struct {
	exec   lxd.Operation
	stdout *io.PipeReader
	done   chan bool
}

// New returns a new lxd Engine.
func New() types.Engine {
	return &backend{
		active: make(map[string]*backendStep),
	}
}

func (e *backend) Name() string {
	return "lxd"
}

func (e *backend) IsAvailable(ctx context.Context) bool {
	socketPath := lxdSnapUnixSocket
	_, err := os.Stat(socketPath)
	if errors.Is(err, os.ErrNotExist) {
		socketPath = ""
	}
	_, err = lxd.ConnectLXDUnixWithContext(ctx, socketPath, nil)
	return err == nil
}

func (e *backend) Load(ctx context.Context) error {
	socketPath := lxdSnapUnixSocket
	_, err := os.Stat(socketPath)
	if errors.Is(err, os.ErrNotExist) {
		socketPath = ""
	} else if err != nil {
		return err
	}
	inst, err := lxd.ConnectLXDUnixWithContext(ctx, socketPath, nil)
	if err != nil {
		return err
	}
	e.inst = inst
	return nil
}

// SetupWorkflow the pipeline environment.
func (e *backend) SetupWorkflow(_ context.Context, config *types.Config, taskUUID string) error {
	log.Trace().Str("taskUUID", taskUUID).Msg("create workflow environment")
	return nil
}

// StartStep the pipeline step.
func (e *backend) StartStep(ctx context.Context, step *types.Step, taskUUID string) error {
	log.Trace().Str("taskUUID", taskUUID).Msgf("start step %s", step.Name)

	if step.Image == constant.DefaultCloneImage {
		return fmt.Errorf("clone not implemented")
	}

	if !strings.HasPrefix(step.Image, "lxd:") {
		return fmt.Errorf("lxd step image name must have lxd: prefix")
	}
	alias := strings.TrimPrefix(step.Image, "lxd:")

	var image *api.Image

	aliasInfo, _, err := e.inst.GetImageAlias(alias)
	if err != nil && !api.StatusErrorCheck(err, http.StatusNotFound) {
		return fmt.Errorf("failed to find image %q: %w", alias, err)
	} else if err == nil {
		image, _, err = e.inst.GetImage(aliasInfo.Target)
		if err != nil {
			return fmt.Errorf("failed to find image %q: %w", alias, err)
		}
	}

	if image == nil {
		var err error
		imageServer, err := lxd.ConnectSimpleStreams(defaultSimpleStreamsRepo, nil)
		if err != nil {
			return fmt.Errorf("failed to connect to lxd image server: %w", err)
		}
		aliasInfo, _, err := imageServer.GetImageAlias(alias)
		if err != nil {
			return fmt.Errorf("failed to find image %q: %w", alias, err)
		}
		srcImage, _, err := imageServer.GetImage(aliasInfo.Target)
		if err != nil {
			return fmt.Errorf("failed to get image %q: %w", alias, err)
		}
		log.Trace().Str("taskUUID", taskUUID).Msgf("downloading image %q", srcImage.Fingerprint)
		copyOp, err := e.inst.CopyImage(imageServer, *srcImage, &lxd.ImageCopyArgs{
			CopyAliases: true,
		})
		if err != nil {
			return fmt.Errorf("failed to download image: %w", err)
		}
		err = copyOp.Wait()
		if err != nil {
			return fmt.Errorf("failed to download image: %w", err)
		}
		image, _, err = e.inst.GetImage(aliasInfo.Target)
		if err != nil {
			return fmt.Errorf("failed to get image %q: %w", alias, err)
		}
	}

	devices := map[string]map[string]string{}
	for i, volume := range step.Volumes {
		parts := strings.Split(volume, ":")
		if len(parts) != 2 {
			continue
		}
		if !strings.HasPrefix(parts[0], "/") {
			continue
		}
		if !strings.HasPrefix(parts[1], "/") {
			continue
		}
		name := fmt.Sprintf("step-vol-%d", i)
		devices[name] = map[string]string{
			"type":   "disk",
			"source": parts[0],
			"path":   parts[1],
		}
	}

	log.Trace().Str("taskUUID", taskUUID).Msgf("devices %#v", devices)

	createOp, err := e.inst.CreateInstanceFromImage(e.inst, *image, api.InstancesPost{
		Name: instanceName(step),
		InstancePut: api.InstancePut{
			Ephemeral: true,
			Devices:   devices,
			Config: map[string]string{
				"security.nesting":    "true",
				"security.privileged": "true",
			},
		},
	})
	if err != nil {
		return err
	}
	err = createOp.Wait()
	if err != nil {
		return err
	}

	startOp, err := e.inst.UpdateInstanceState(instanceName(step), api.InstanceStatePut{
		Action: "start",
	}, "")
	if err != nil {
		return err
	}
	err = startOp.WaitContext(ctx)
	if err != nil {
		return err
	}

	env, entry, cmd := common.GenerateContainerConf(step.Commands)
	for k, v := range env {
		step.Environment[k] = v
	}
	command := append(entry, cmd...)

	reader, writer := io.Pipe()
	done := make(chan bool)
	execOp, err := e.inst.ExecInstance(instanceName(step), api.InstanceExecPost{
		Command:     command,
		Environment: step.Environment,
		Cwd:         step.WorkingDir,
		WaitForWS:   true,
	}, &lxd.InstanceExecArgs{
		Stdout:   &writerNopCloser{writer},
		Stderr:   &writerNopCloser{writer},
		DataDone: done,
	})
	if err != nil {
		return err
	}
	go func() {
		<-done
		writer.Close()
	}()
	e.active[step.UUID] = &backendStep{
		exec:   execOp,
		stdout: reader,
		done:   done,
	}

	return nil
}

// WaitStep for the pipeline step to complete and returns
// the completion results.
func (e *backend) WaitStep(ctx context.Context, step *types.Step, taskUUID string) (*types.State, error) {
	log.Trace().Str("taskUUID", taskUUID).Msgf("wait for step %s", step.Name)

	current, ok := e.active[step.UUID]
	if !ok {
		return nil, fmt.Errorf("step %s not found", step.UUID)
	}

	err := current.exec.WaitContext(ctx)
	state := &types.State{
		Exited:   true,
		ExitCode: 0,
		Error:    err,
	}

	res := current.exec.Get()
	if res.Metadata != nil {
		if statusCode, ok := res.Metadata["return"].(float64); ok {
			state.ExitCode = int(statusCode)
		}
	}

	return state, nil
}

// TailStep the pipeline step logs.
func (e *backend) TailStep(_ context.Context, step *types.Step, taskUUID string) (io.ReadCloser, error) {
	log.Trace().Str("taskUUID", taskUUID).Msgf("tail logs of step %s", step.Name)

	current, ok := e.active[step.UUID]
	if !ok {
		return nil, fmt.Errorf("step %s not found", step.UUID)
	}

	return current.stdout, nil
}

// DestroyWorkflow the pipeline environment.
func (e *backend) DestroyWorkflow(ctx context.Context, config *types.Config, taskUUID string) error {
	log.Trace().Str("taskUUID", taskUUID).Msgf("delete workflow environment")

	for _, stage := range config.Stages {
		for _, step := range stage.Steps {
			_, ok := e.active[step.UUID]
			if !ok {
				continue
			}

			log.Trace().Str("taskUUID", taskUUID).Msgf("stopping %s", instanceName(step))
			stop, err := e.inst.UpdateInstanceState(instanceName(step), api.InstanceStatePut{
				Action: "stop",
				Force:  true,
			}, "")
			if err != nil {
				return err
			}
			err = stop.WaitContext(ctx)
			if err != nil {
				return err
			}

			delete(e.active, step.UUID)
		}
	}

	return nil
}

func instanceName(step *types.Step) string {
	return "wp-" + step.UUID
}

type writerNopCloser struct {
	io.Writer
}

func (*writerNopCloser) Close() error {
	return nil
}
