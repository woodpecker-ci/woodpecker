package docker

import (
	"context"
	"io"
	"os"
	"strings"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/api/types/volume"
	"github.com/moby/moby/client"
	"github.com/moby/moby/pkg/jsonmessage"
	"github.com/moby/moby/pkg/stdcopy"
	"github.com/moby/term"
	"github.com/rs/zerolog/log"

	backend "github.com/woodpecker-ci/woodpecker/pipeline/backend/types"
)

type docker struct {
	client     client.APIClient
	enableIPv6 bool
	network    string
}

// make sure docker implements Engine
var _ backend.Engine = &docker{}

// New returns a new Docker Engine.
func New() backend.Engine {
	return &docker{
		client: nil,
	}
}

func (e *docker) Name() string {
	return "docker"
}

func (e *docker) IsAvailable() bool {
	if os.Getenv("DOCKER_HOST") != "" {
		return true
	}
	_, err := os.Stat("/var/run/docker.sock")
	return err == nil
}

// Load new client for Docker Engine using environment variables.
func (e *docker) Load() error {
	cli, err := client.NewClientWithOpts(client.FromEnv)
	if err != nil {
		return err
	}
	e.client = cli

	return nil
}

func (e *docker) Setup(_ context.Context, conf *backend.Config) error {
	for _, vol := range conf.Volumes {
		_, err := e.client.VolumeCreate(noContext, volume.VolumeCreateBody{
			Name:       vol.Name,
			Driver:     vol.Driver,
			DriverOpts: vol.DriverOpts,
			// Labels:     defaultLabels,
		})
		if err != nil {
			return err
		}
	}
	for _, n := range conf.Networks {
		_, err := e.client.NetworkCreate(noContext, n.Name, types.NetworkCreate{
			Driver:     n.Driver,
			Options:    n.DriverOpts,
			EnableIPv6: e.enableIPv6,
			// Labels:  defaultLabels,
		})
		if err != nil {
			return err
		}
	}
	return nil
}

func (e *docker) Exec(ctx context.Context, proc *backend.Step) error {
	config := toConfig(proc)
	hostConfig := toHostConfig(proc)

	// create pull options with encoded authorization credentials.
	pullopts := types.ImagePullOptions{}
	if proc.AuthConfig.Username != "" && proc.AuthConfig.Password != "" {
		pullopts.RegistryAuth, _ = encodeAuthToBase64(proc.AuthConfig)
	}

	// automatically pull the latest version of the image if requested
	// by the process configuration.
	if proc.Pull {
		responseBody, perr := e.client.ImagePull(ctx, config.Image, pullopts)
		if perr == nil {
			defer responseBody.Close()

			fd, isTerminal := term.GetFdInfo(os.Stdout)
			if err := jsonmessage.DisplayJSONMessagesStream(responseBody, os.Stdout, fd, isTerminal, nil); err != nil {
				log.Error().Err(err).Msg("DisplayJSONMessagesStream")
			}
		}
		// fix for drone/drone#1917
		if perr != nil && proc.AuthConfig.Password != "" {
			return perr
		}
	}

	_, err := e.client.ContainerCreate(ctx, config, hostConfig, nil, nil, proc.Name)
	if client.IsErrNotFound(err) {
		// automatically pull and try to re-create the image if the
		// failure is caused because the image does not exist.
		responseBody, perr := e.client.ImagePull(ctx, config.Image, pullopts)
		if perr != nil {
			return perr
		}
		defer responseBody.Close()
		fd, isTerminal := term.GetFdInfo(os.Stdout)
		if err := jsonmessage.DisplayJSONMessagesStream(responseBody, os.Stdout, fd, isTerminal, nil); err != nil {
			log.Error().Err(err).Msg("DisplayJSONMessagesStream")
		}

		_, err = e.client.ContainerCreate(ctx, config, hostConfig, nil, nil, proc.Name)
	}
	if err != nil {
		return err
	}

	if len(proc.NetworkMode) == 0 {
		for _, net := range proc.Networks {
			err = e.client.NetworkConnect(ctx, net.Name, proc.Name, &network.EndpointSettings{
				Aliases: net.Aliases,
			})
			if err != nil {
				return err
			}
		}

		// join the container to an existing network
		if e.network != "" {
			err = e.client.NetworkConnect(ctx, e.network, proc.Name, &network.EndpointSettings{})
			if err != nil {
				return err
			}
		}
	}

	return e.client.ContainerStart(ctx, proc.Name, startOpts)
}

func (e *docker) Wait(ctx context.Context, proc *backend.Step) (*backend.State, error) {
	wait, errc := e.client.ContainerWait(ctx, proc.Name, "")
	select {
	case <-wait:
	case <-errc:
	}

	info, err := e.client.ContainerInspect(ctx, proc.Name)
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

func (e *docker) Tail(ctx context.Context, proc *backend.Step) (io.ReadCloser, error) {
	logs, err := e.client.ContainerLogs(ctx, proc.Name, logsOpts)
	if err != nil {
		return nil, err
	}
	rc, wc := io.Pipe()

	// de multiplex 'logs' who contains two streams, previously multiplexed together using StdWriter
	go func() {
		_, _ = stdcopy.StdCopy(wc, wc, logs)
		_ = logs.Close()
		_ = wc.Close()
		_ = rc.Close()
	}()
	return rc, nil
}

func (e *docker) Destroy(_ context.Context, conf *backend.Config) error {
	for _, stage := range conf.Stages {
		for _, step := range stage.Steps {
			if err := e.client.ContainerKill(noContext, step.Name, "9"); err != nil && !isErrContainerNotFoundOrNotRunning(err) {
				log.Error().Err(err).Msgf("could not kill container '%s'", stage.Name)
			}
			if err := e.client.ContainerRemove(noContext, step.Name, removeOpts); err != nil && !isErrContainerNotFoundOrNotRunning(err) {
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
