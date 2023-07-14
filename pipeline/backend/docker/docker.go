// Copyright 2022 Woodpecker Authors
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

package docker

import (
	"context"
	"io"
	"os"
	"runtime"
	"strings"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/api/types/volume"
	"github.com/moby/moby/client"
	"github.com/moby/moby/pkg/jsonmessage"
	"github.com/moby/moby/pkg/stdcopy"
	"github.com/moby/term"
	"github.com/rs/zerolog/log"
	"github.com/urfave/cli/v2"

	backend "github.com/woodpecker-ci/woodpecker/pipeline/backend/types"
	"github.com/woodpecker-ci/woodpecker/shared/utils"
)

type docker struct {
	client     client.APIClient
	enableIPv6 bool
	network    string
	volumes    []string
}

const (
	networkDriverNAT    = "nat"
	networkDriverBridge = "bridge"
	volumeDriver        = "local"
)

// New returns a new Docker Engine.
func New() backend.Engine {
	return &docker{
		client: nil,
	}
}

func (e *docker) Name() string {
	return "docker"
}

func (e *docker) IsAvailable(context.Context) bool {
	if os.Getenv("DOCKER_HOST") != "" {
		return true
	}
	_, err := os.Stat("/var/run/docker.sock")
	return err == nil
}

// Load new client for Docker Engine using environment variables.
func (e *docker) Load(ctx context.Context) error {
	cl, err := client.NewClientWithOpts(client.FromEnv)
	if err != nil {
		return err
	}
	e.client = cl

	c, ok := ctx.Value(backend.CliContext).(*cli.Context)
	if !ok {
		return backend.ErrNoCliContextFound
	}
	e.enableIPv6 = c.Bool("backend-docker-ipv6")

	e.network = c.String("backend-docker-network")

	volumes := strings.Split(c.String("backend-docker-volumes"), ",")
	e.volumes = make([]string, 0, len(volumes))
	// Validate provided volume definitions
	for _, v := range volumes {
		if v == "" {
			continue
		}
		parts, err := splitVolumeParts(v)
		if err != nil {
			log.Error().Err(err).Msgf("invalid volume '%s' provided in WOODPECKER_BACKEND_DOCKER_VOLUMES", v)
			continue
		}
		e.volumes = append(e.volumes, strings.Join(parts, ":"))
	}

	return nil
}

func (e *docker) Setup(_ context.Context, conf *backend.Config) error {
	for _, vol := range conf.Volumes {
		_, err := e.client.VolumeCreate(noContext, volume.VolumeCreateBody{
			Name:   vol.Name,
			Driver: volumeDriver,
		})
		if err != nil {
			return err
		}
	}

	networkDriver := networkDriverBridge
	if runtime.GOOS == "windows" {
		networkDriver = networkDriverNAT
	}
	for _, n := range conf.Networks {
		_, err := e.client.NetworkCreate(noContext, n.Name, types.NetworkCreate{
			Driver:     networkDriver,
			EnableIPv6: e.enableIPv6,
		})
		if err != nil {
			return err
		}
	}
	return nil
}

func (e *docker) Exec(ctx context.Context, step *backend.Step) error {
	config := toConfig(step)
	hostConfig := toHostConfig(step)
	containerName := toContainerName(step)

	// create pull options with encoded authorization credentials.
	pullopts := types.ImagePullOptions{}
	if step.AuthConfig.Username != "" && step.AuthConfig.Password != "" {
		pullopts.RegistryAuth, _ = encodeAuthToBase64(step.AuthConfig)
	}

	// automatically pull the latest version of the image if requested
	// by the process configuration.
	if step.Pull {
		responseBody, perr := e.client.ImagePull(ctx, config.Image, pullopts)
		if perr == nil {
			// TODO(1936): show image pull progress in web-ui
			fd, isTerminal := term.GetFdInfo(os.Stdout)
			if err := jsonmessage.DisplayJSONMessagesStream(responseBody, os.Stdout, fd, isTerminal, nil); err != nil {
				log.Error().Err(err).Msg("DisplayJSONMessagesStream")
			}
			responseBody.Close()
		}
		// Fix "Show warning when fail to auth to docker registry"
		// (https://web.archive.org/web/20201023145804/https://github.com/drone/drone/issues/1917)
		if perr != nil && step.AuthConfig.Password != "" {
			return perr
		}
	}

	// add default volumes to the host configuration
	hostConfig.Binds = utils.DedupStrings(append(hostConfig.Binds, e.volumes...))

	_, err := e.client.ContainerCreate(ctx, config, hostConfig, nil, nil, containerName)
	if client.IsErrNotFound(err) {
		// automatically pull and try to re-create the image if the
		// failure is caused because the image does not exist.
		responseBody, perr := e.client.ImagePull(ctx, config.Image, pullopts)
		if perr != nil {
			return perr
		}
		// TODO(1936): show image pull progress in web-ui
		fd, isTerminal := term.GetFdInfo(os.Stdout)
		if err := jsonmessage.DisplayJSONMessagesStream(responseBody, os.Stdout, fd, isTerminal, nil); err != nil {
			log.Error().Err(err).Msg("DisplayJSONMessagesStream")
		}
		responseBody.Close()

		_, err = e.client.ContainerCreate(ctx, config, hostConfig, nil, nil, containerName)
	}
	if err != nil {
		return err
	}

	if len(step.NetworkMode) == 0 {
		for _, net := range step.Networks {
			err = e.client.NetworkConnect(ctx, net.Name, containerName, &network.EndpointSettings{
				Aliases: net.Aliases,
			})
			if err != nil {
				return err
			}
		}

		// join the container to an existing network
		if e.network != "" {
			err = e.client.NetworkConnect(ctx, e.network, containerName, &network.EndpointSettings{})
			if err != nil {
				return err
			}
		}
	}

	return e.client.ContainerStart(ctx, containerName, startOpts)
}

func (e *docker) Wait(ctx context.Context, step *backend.Step) (*backend.State, error) {
	containerName := toContainerName(step)

	wait, errc := e.client.ContainerWait(ctx, containerName, "")
	select {
	case <-wait:
	case <-errc:
	}

	info, err := e.client.ContainerInspect(ctx, containerName)
	if err != nil {
		return nil, err
	}
	// if info.State.Running {
	// TODO
	// }

	return &backend.State{
		Exited:    true,
		ExitCode:  info.State.ExitCode,
		OOMKilled: info.State.OOMKilled,
	}, nil
}

func (e *docker) Tail(ctx context.Context, step *backend.Step) (io.ReadCloser, error) {
	logs, err := e.client.ContainerLogs(ctx, toContainerName(step), logsOpts)
	if err != nil {
		return nil, err
	}
	rc, wc := io.Pipe()

	// de multiplex 'logs' who contains two streams, previously multiplexed together using StdWriter
	go func() {
		_, _ = stdcopy.StdCopy(wc, wc, logs)
		_ = logs.Close()
		_ = wc.Close()
	}()
	return rc, nil
}

func (e *docker) Destroy(_ context.Context, conf *backend.Config) error {
	for _, stage := range conf.Stages {
		for _, step := range stage.Steps {
			containerName := toContainerName(step)
			if err := e.client.ContainerKill(noContext, containerName, "9"); err != nil && !isErrContainerNotFoundOrNotRunning(err) {
				log.Error().Err(err).Msgf("could not kill container '%s'", stage.Name)
			}
			if err := e.client.ContainerRemove(noContext, containerName, removeOpts); err != nil && !isErrContainerNotFoundOrNotRunning(err) {
				log.Error().Err(err).Msgf("could not remove container '%s'", stage.Name)
			}
		}
	}
	for _, v := range conf.Volumes {
		if err := e.client.VolumeRemove(noContext, v.Name, true); err != nil {
			log.Error().Err(err).Msgf("could not remove volume '%s'", v.Name)
		}
	}
	for _, n := range conf.Networks {
		if err := e.client.NetworkRemove(noContext, n.Name); err != nil {
			log.Error().Err(err).Msgf("could not remove network '%s'", n.Name)
		}
	}
	return nil
}

var (
	noContext = context.Background()

	startOpts = types.ContainerStartOptions{}

	removeOpts = types.ContainerRemoveOptions{
		RemoveVolumes: true,
		RemoveLinks:   false,
		Force:         false,
	}

	logsOpts = types.ContainerLogsOptions{
		Follow:     true,
		ShowStdout: true,
		ShowStderr: true,
		Details:    false,
		Timestamps: false,
	}
)

func isErrContainerNotFoundOrNotRunning(err error) bool {
	// Error response from daemon: Cannot kill container: ...: No such container: ...
	// Error response from daemon: Cannot kill container: ...: Container ... is not running"
	// Error: No such container: ...
	return err != nil && (strings.Contains(err.Error(), "No such container") || strings.Contains(err.Error(), "is not running"))
}
