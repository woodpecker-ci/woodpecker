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
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/api/types/system"
	"github.com/docker/docker/api/types/volume"
	tls_config "github.com/docker/go-connections/tlsconfig"
	"github.com/moby/moby/client"
	json_message "github.com/moby/moby/pkg/jsonmessage"
	std_copy "github.com/moby/moby/pkg/stdcopy"
	"github.com/moby/term"
	"github.com/rs/zerolog/log"
	"github.com/urfave/cli/v3"

	backend "go.woodpecker-ci.org/woodpecker/v2/pipeline/backend/types"
	"go.woodpecker-ci.org/woodpecker/v2/shared/utils"
)

type docker struct {
	client client.APIClient
	info   system.Info
	config config
}

const (
	networkDriverNAT    = "nat"
	networkDriverBridge = "bridge"
	volumeDriver        = "local"
)

// New returns a new Docker Backend.
func New() backend.Backend {
	return &docker{
		client: nil,
	}
}

func (e *docker) Name() string {
	return "docker"
}

func (e *docker) IsAvailable(ctx context.Context) bool {
	if c, ok := ctx.Value(backend.CliCommand).(*cli.Command); ok {
		if c.IsSet("backend-docker-host") {
			return true
		}
	}
	_, err := os.Stat("/var/run/docker.sock")
	return err == nil
}

func httpClientOfOpts(dockerCertPath string, verifyTLS bool) *http.Client {
	if dockerCertPath == "" {
		return nil
	}

	options := tls_config.Options{
		CAFile:             filepath.Join(dockerCertPath, "ca.pem"),
		CertFile:           filepath.Join(dockerCertPath, "cert.pem"),
		KeyFile:            filepath.Join(dockerCertPath, "key.pem"),
		InsecureSkipVerify: !verifyTLS,
	}
	tlsConf, err := tls_config.Client(options)
	if err != nil {
		log.Error().Err(err).Msg("could not create http client out of docker backend options")
		return nil
	}

	return &http.Client{
		Transport:     &http.Transport{TLSClientConfig: tlsConf},
		CheckRedirect: client.CheckRedirect,
	}
}

func (e *docker) Flags() []cli.Flag {
	return Flags
}

// Load new client for Docker Backend using environment variables.
func (e *docker) Load(ctx context.Context) (*backend.BackendInfo, error) {
	c, ok := ctx.Value(backend.CliCommand).(*cli.Command)
	if !ok {
		return nil, backend.ErrNoCliContextFound
	}

	var dockerClientOpts []client.Opt
	if httpClient := httpClientOfOpts(c.String("backend-docker-cert"), c.Bool("backend-docker-tls-verify")); httpClient != nil {
		dockerClientOpts = append(dockerClientOpts, client.WithHTTPClient(httpClient))
	}
	if dockerHost := c.String("backend-docker-host"); dockerHost != "" {
		dockerClientOpts = append(dockerClientOpts, client.WithHost(dockerHost))
	}
	if dockerAPIVersion := c.String("backend-docker-api-version"); dockerAPIVersion != "" {
		dockerClientOpts = append(dockerClientOpts, client.WithVersion(dockerAPIVersion))
	} else {
		dockerClientOpts = append(dockerClientOpts, client.WithAPIVersionNegotiation())
	}

	cl, err := client.NewClientWithOpts(dockerClientOpts...)
	if err != nil {
		return nil, err
	}
	e.client = cl

	e.info, err = cl.Info(ctx)
	if err != nil {
		return nil, err
	}

	e.config, err = configFromCli(c)
	if err != nil {
		return nil, err
	}

	return &backend.BackendInfo{
		Platform: e.info.OSType + "/" + normalizeArchType(e.info.Architecture),
	}, nil
}

func (e *docker) SetupWorkflow(ctx context.Context, conf *backend.Config, taskUUID string) error {
	log.Trace().Str("taskUUID", taskUUID).Msg("create workflow environment")

	for _, vol := range conf.Volumes {
		_, err := e.client.VolumeCreate(ctx, volume.CreateOptions{
			Name:   vol.Name,
			Driver: volumeDriver,
		})
		if err != nil {
			return err
		}
	}

	networkDriver := networkDriverBridge
	if e.info.OSType == "windows" {
		networkDriver = networkDriverNAT
	}
	for _, n := range conf.Networks {
		_, err := e.client.NetworkCreate(ctx, n.Name, network.CreateOptions{
			Driver:     networkDriver,
			EnableIPv6: &e.config.enableIPv6,
		})
		if err != nil {
			return err
		}
	}
	return nil
}

func (e *docker) StartStep(ctx context.Context, step *backend.Step, taskUUID string) error {
	log.Trace().Str("taskUUID", taskUUID).Msgf("start step %s", step.Name)

	config := e.toConfig(step)
	hostConfig := toHostConfig(step, &e.config)
	containerName := toContainerName(step)

	// create pull options with encoded authorization credentials.
	pullOpts := image.PullOptions{}
	if step.AuthConfig.Username != "" && step.AuthConfig.Password != "" {
		pullOpts.RegistryAuth, _ = encodeAuthToBase64(step.AuthConfig)
	}

	// automatically pull the latest version of the image if requested
	// by the process configuration.
	if step.Pull {
		responseBody, pErr := e.client.ImagePull(ctx, config.Image, pullOpts)
		if pErr == nil {
			// TODO(1936): show image pull progress in web-ui
			fd, isTerminal := term.GetFdInfo(os.Stdout)
			if err := json_message.DisplayJSONMessagesStream(responseBody, os.Stdout, fd, isTerminal, nil); err != nil {
				log.Error().Err(err).Msg("DisplayJSONMessagesStream")
			}
			responseBody.Close()
		}
		// Fix "Show warning when fail to auth to docker registry"
		// (https://web.archive.org/web/20201023145804/https://github.com/drone/drone/issues/1917)
		if pErr != nil && step.AuthConfig.Password != "" {
			return pErr
		}
	}

	// add default volumes to the host configuration
	hostConfig.Binds = utils.DeduplicateStrings(append(hostConfig.Binds, e.config.volumes...))

	_, err := e.client.ContainerCreate(ctx, config, hostConfig, nil, nil, containerName)
	if client.IsErrNotFound(err) {
		// automatically pull and try to re-create the image if the
		// failure is caused because the image does not exist.
		responseBody, pErr := e.client.ImagePull(ctx, config.Image, pullOpts)
		if pErr != nil {
			return pErr
		}
		// TODO(1936): show image pull progress in web-ui
		fd, isTerminal := term.GetFdInfo(os.Stdout)
		if err := json_message.DisplayJSONMessagesStream(responseBody, os.Stdout, fd, isTerminal, nil); err != nil {
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
		if e.config.network != "" {
			err = e.client.NetworkConnect(ctx, e.config.network, containerName, &network.EndpointSettings{})
			if err != nil {
				return err
			}
		}
	}

	return e.client.ContainerStart(ctx, containerName, container.StartOptions{})
}

func (e *docker) WaitStep(ctx context.Context, step *backend.Step, taskUUID string) (*backend.State, error) {
	log.Trace().Str("taskUUID", taskUUID).Msgf("wait for step %s", step.Name)

	containerName := toContainerName(step)

	wait, errC := e.client.ContainerWait(ctx, containerName, "")
	select {
	case <-wait:
	case <-errC:
	}

	info, err := e.client.ContainerInspect(ctx, containerName)
	if err != nil {
		return nil, err
	}

	return &backend.State{
		Exited:    true,
		ExitCode:  info.State.ExitCode,
		OOMKilled: info.State.OOMKilled,
	}, nil
}

func (e *docker) TailStep(ctx context.Context, step *backend.Step, taskUUID string) (io.ReadCloser, error) {
	log.Trace().Str("taskUUID", taskUUID).Msgf("tail logs of step %s", step.Name)

	logs, err := e.client.ContainerLogs(ctx, toContainerName(step), container.LogsOptions{
		Follow:     true,
		ShowStdout: true,
		ShowStderr: true,
		Details:    false,
		Timestamps: false,
	})
	if err != nil {
		return nil, err
	}
	rc, wc := io.Pipe()

	// de multiplex 'logs' who contains two streams, previously multiplexed together using StdWriter
	go func() {
		_, _ = std_copy.StdCopy(wc, wc, logs)
		_ = logs.Close()
		_ = wc.Close()
	}()
	return rc, nil
}

func (e *docker) DestroyStep(ctx context.Context, step *backend.Step, taskUUID string) error {
	log.Trace().Str("taskUUID", taskUUID).Msgf("stop step %s", step.Name)

	containerName := toContainerName(step)

	if err := e.client.ContainerKill(ctx, containerName, "9"); err != nil && !isErrContainerNotFoundOrNotRunning(err) {
		return err
	}

	if err := e.client.ContainerRemove(ctx, containerName, removeOpts); err != nil && !isErrContainerNotFoundOrNotRunning(err) {
		return err
	}

	return nil
}

func (e *docker) DestroyWorkflow(ctx context.Context, conf *backend.Config, taskUUID string) error {
	log.Trace().Str("taskUUID", taskUUID).Msgf("delete workflow environment")

	for _, stage := range conf.Stages {
		for _, step := range stage.Steps {
			containerName := toContainerName(step)
			if err := e.client.ContainerKill(ctx, containerName, "9"); err != nil && !isErrContainerNotFoundOrNotRunning(err) {
				log.Error().Err(err).Msgf("could not kill container '%s'", step.Name)
			}
			if err := e.client.ContainerRemove(ctx, containerName, removeOpts); err != nil && !isErrContainerNotFoundOrNotRunning(err) {
				log.Error().Err(err).Msgf("could not remove container '%s'", step.Name)
			}
		}
	}
	for _, v := range conf.Volumes {
		if err := e.client.VolumeRemove(ctx, v.Name, true); err != nil {
			log.Error().Err(err).Msgf("could not remove volume '%s'", v.Name)
		}
	}
	for _, n := range conf.Networks {
		if err := e.client.NetworkRemove(ctx, n.Name); err != nil {
			log.Error().Err(err).Msgf("could not remove network '%s'", n.Name)
		}
	}
	return nil
}

var removeOpts = container.RemoveOptions{
	RemoveVolumes: true,
	RemoveLinks:   false,
	Force:         false,
}

func isErrContainerNotFoundOrNotRunning(err error) bool {
	// Error response from daemon: Cannot kill container: ...: No such container: ...
	// Error response from daemon: Cannot kill container: ...: Container ... is not running"
	// Error response from podman daemon: can only kill running containers. ... is in state exited
	// Error: No such container: ...
	return err != nil && (strings.Contains(err.Error(), "No such container") || strings.Contains(err.Error(), "is not running") || strings.Contains(err.Error(), "can only kill running containers"))
}

// normalizeArchType converts the arch type reported by docker info into
// the runtime.GOARCH format
// TODO: find out if we we need to convert other arch types too
func normalizeArchType(s string) string {
	switch s {
	case "x86_64":
		return "amd64"
	default:
		return s
	}
}
