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
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/containerd/errdefs"
	"github.com/docker/go-connections/tlsconfig"
	"github.com/moby/moby/api/pkg/stdcopy"
	"github.com/moby/moby/api/types/network"
	"github.com/moby/moby/api/types/system"
	"github.com/moby/moby/client"
	"github.com/moby/moby/client/pkg/jsonmessage"
	"github.com/moby/term"
	"github.com/rs/zerolog/log"
	"github.com/urfave/cli/v3"
	"golang.org/x/sync/errgroup"

	backend_types "go.woodpecker-ci.org/woodpecker/v3/pipeline/backend/types"
	"go.woodpecker-ci.org/woodpecker/v3/shared/httputil"
	"go.woodpecker-ci.org/woodpecker/v3/shared/utils"
)

var containerKillTimeout = 5 // seconds

type docker struct {
	client client.APIClient
	info   system.Info
	config config
}

const (
	EngineName          = "docker"
	networkDriverNAT    = "nat"
	networkDriverBridge = "bridge"
	volumeDriver        = "local"
)

// New returns a new Docker Backend.
func New() backend_types.Backend {
	return &docker{
		client: nil,
	}
}

func (e *docker) Name() string {
	return EngineName
}

func (e *docker) IsAvailable(ctx context.Context) bool {
	if c, ok := ctx.Value(backend_types.CliCommand).(*cli.Command); ok {
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

	options := tlsconfig.Options{
		CAFile:             filepath.Join(dockerCertPath, "ca.pem"),
		CertFile:           filepath.Join(dockerCertPath, "cert.pem"),
		KeyFile:            filepath.Join(dockerCertPath, "key.pem"),
		InsecureSkipVerify: !verifyTLS,
	}
	tlsConf, err := tlsconfig.Client(options)
	if err != nil {
		log.Error().Err(err).Msg("could not create http client out of docker backend options")
		return nil
	}

	return &http.Client{
		Transport: httputil.NewUserAgentRoundTripper(
			&http.Transport{TLSClientConfig: tlsConf},
			"backend-docker"),
		CheckRedirect: client.CheckRedirect,
	}
}

func (e *docker) Flags() []cli.Flag {
	return Flags
}

// Load new client for Docker Backend using environment variables.
func (e *docker) Load(ctx context.Context) (*backend_types.BackendInfo, error) {
	c, ok := ctx.Value(backend_types.CliCommand).(*cli.Command)
	if !ok {
		return nil, backend_types.ErrNoCliContextFound
	}

	var dockerClientOpts []client.Opt
	if httpClient := httpClientOfOpts(c.String("backend-docker-cert"), c.Bool("backend-docker-tls-verify")); httpClient != nil {
		dockerClientOpts = append(dockerClientOpts, client.WithHTTPClient(httpClient))
	}
	if dockerHost := c.String("backend-docker-host"); dockerHost != "" {
		dockerClientOpts = append(dockerClientOpts, client.WithHost(dockerHost))
	}
	if dockerAPIVersion := c.String("backend-docker-api-version"); dockerAPIVersion != "" {
		dockerClientOpts = append(dockerClientOpts, client.WithAPIVersion(dockerAPIVersion))
	}

	cl, err := client.New(dockerClientOpts...)
	if err != nil {
		return nil, err
	}
	e.client = cl

	info, err := cl.Info(ctx, client.InfoOptions{})
	if err != nil {
		return nil, err
	}

	e.info = info.Info

	e.config, err = configFromCli(c)
	if err != nil {
		return nil, err
	}

	return &backend_types.BackendInfo{
		Platform: e.info.OSType + "/" + normalizeArchType(e.info.Architecture),
	}, nil
}

func (e *docker) SetupWorkflow(ctx context.Context, conf *backend_types.Config, taskUUID string) error {
	log.Trace().Str("taskUUID", taskUUID).Msg("create workflow environment")

	_, err := e.client.VolumeCreate(ctx, client.VolumeCreateOptions{
		Name:   conf.Volume,
		Driver: volumeDriver,
	})
	if err != nil {
		return err
	}

	networkDriver := networkDriverBridge
	if e.info.OSType == "windows" {
		networkDriver = networkDriverNAT
	}
	_, err = e.client.NetworkCreate(ctx, conf.Network, client.NetworkCreateOptions{
		Driver:     networkDriver,
		EnableIPv6: &e.config.enableIPv6,
	})
	return err
}

func (e *docker) StartStep(ctx context.Context, step *backend_types.Step, taskUUID string) error {
	options, err := parseBackendOptions(step)
	if err != nil {
		log.Error().Err(err).Msg("could not parse backend options")
	}

	log.Trace().Str("taskUUID", taskUUID).Msgf("start step %s", step.Name)

	config := e.toConfig(step, options)
	hostConfig, err := toHostConfig(step, &e.config)
	if err != nil {
		return err
	}
	containerName := toContainerName(step)

	// create pull options with encoded authorization credentials.
	pullOpts := client.ImagePullOptions{}
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
			if err := jsonmessage.DisplayJSONMessagesStream(responseBody, os.Stdout, fd, isTerminal, nil); err != nil {
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

	_, err = e.client.ContainerCreate(ctx, client.ContainerCreateOptions{
		Config:     config,
		HostConfig: hostConfig,
		Name:       containerName,
	})
	if errdefs.IsNotFound(err) {
		// automatically pull and try to re-create the image if the
		// failure is caused because the image does not exist.
		responseBody, pErr := e.client.ImagePull(ctx, config.Image, pullOpts)
		if pErr != nil {
			return pErr
		}
		// TODO(1936): show image pull progress in web-ui
		fd, isTerminal := term.GetFdInfo(os.Stdout)
		if err := jsonmessage.DisplayJSONMessagesStream(responseBody, os.Stdout, fd, isTerminal, nil); err != nil {
			log.Error().Err(err).Msg("DisplayJSONMessagesStream")
		}
		responseBody.Close()

		_, err = e.client.ContainerCreate(ctx, client.ContainerCreateOptions{
			Config:     config,
			HostConfig: hostConfig,
			Name:       containerName,
		})
	}
	if err != nil {
		return err
	}

	if len(step.NetworkMode) == 0 {
		for _, net := range step.Networks {
			_, err = e.client.NetworkConnect(ctx, net.Name, client.NetworkConnectOptions{
				EndpointConfig: &network.EndpointSettings{
					Aliases: net.Aliases,
				},
				Container: containerName,
			})
			if err != nil {
				return err
			}
		}

		// join the container to an existing network
		if e.config.network != "" {
			_, err = e.client.NetworkConnect(ctx, e.config.network, client.NetworkConnectOptions{
				Container: containerName,
			})
			if err != nil {
				return err
			}
		}
	}

	_, err = e.client.ContainerStart(ctx, containerName, client.ContainerStartOptions{})
	return err
}

func (e *docker) WaitStep(ctx context.Context, step *backend_types.Step, taskUUID string) (*backend_types.State, error) {
	log := log.Logger.With().Str("taskUUID", taskUUID).Str("stepUUID", step.UUID).Logger()
	log.Trace().Msgf("wait for step %s", step.Name)

	containerName := toContainerName(step)

	wait := e.client.ContainerWait(ctx, containerName, client.ContainerWaitOptions{})
	select {
	case resp := <-wait.Result:
		log.Trace().Msgf("ContainerWait returned with resp: %v", resp)
		if resp.Error != nil {
			return nil, fmt.Errorf("ContainerWait error: %s", resp.Error.Message)
		}
	case err := <-wait.Error:
		log.Trace().Msgf("ContainerWait returned with err: %v", err)
		return nil, err
	case <-ctx.Done():
		return nil, ctx.Err()
	}

	info, err := e.client.ContainerInspect(ctx, containerName, client.ContainerInspectOptions{})
	if err != nil {
		return nil, err
	}

	exitCode := info.Container.State.ExitCode
	// Windows Docker may return 4294967295 (uint32 max, i.e. int32(-1)) for abnormal exits.
	if exitCode == 4294967295 { //nolint:mnd // because it is int(^uint32(0))
		exitCode = int(int32(exitCode))
	}

	return &backend_types.State{
		Exited:    true,
		ExitCode:  exitCode,
		OOMKilled: info.Container.State.OOMKilled,
	}, nil
}

func (e *docker) TailStep(ctx context.Context, step *backend_types.Step, taskUUID string) (io.ReadCloser, error) {
	log.Trace().Str("taskUUID", taskUUID).Msgf("tail logs of step %s", step.Name)

	logs, err := e.client.ContainerLogs(ctx, toContainerName(step), client.ContainerLogsOptions{
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
		_, _ = stdcopy.StdCopy(wc, wc, logs)
		_ = logs.Close()
		_ = wc.Close()
	}()
	return rc, nil
}

func (e *docker) DestroyStep(ctx context.Context, step *backend_types.Step, taskUUID string) error {
	log.Trace().Str("taskUUID", taskUUID).Msgf("stop step %s", step.Name)

	containerName := toContainerName(step)
	var stopErr error

	// we first signal to the container to stop ...
	if _, err := e.client.ContainerStop(ctx, containerName, client.ContainerStopOptions{
		Timeout: &containerKillTimeout,
	}); err != nil && !isErrContainerNotFoundOrNotRunning(err) {
		// we do not return error yet as we try to kill it first
		stopErr = fmt.Errorf("could not stop container '%s': %w", step.Name, err)
	}

	// ... and if stop does not work just force kill it
	if _, err := e.client.ContainerKill(ctx, containerName, client.ContainerKillOptions{
		Signal: "9",
	}); err != nil && !isErrContainerNotFoundOrNotRunning(err) {
		return errors.Join(stopErr, fmt.Errorf("could not kill container '%s': %w", step.Name, err))
	}

	// now we clean up files left
	if _, err := e.client.ContainerRemove(ctx, containerName, removeOpts); err != nil && !isErrContainerNotFoundOrNotRunning(err) {
		return fmt.Errorf("could not remove container '%s': %w", step.Name, err)
	}

	return nil
}

func (e *docker) DestroyWorkflow(ctx context.Context, conf *backend_types.Config, taskUUID string) error {
	log.Trace().Str("taskUUID", taskUUID).Msgf("delete workflow environment")

	errWG := errgroup.Group{}

	for _, stage := range conf.Stages {
		for _, step := range stage.Steps {
			errWG.Go(func() error {
				return e.DestroyStep(ctx, step, taskUUID)
			})
		}
	}

	if err := errWG.Wait(); err != nil {
		log.Error().Err(err).Msgf("could not destroy all containers")
	}

	if _, err := e.client.VolumeRemove(ctx, conf.Volume, client.VolumeRemoveOptions{
		Force: true,
	}); err != nil {
		log.Error().Err(err).Msgf("could not remove volume '%s'", conf.Volume)
	}
	if _, err := e.client.NetworkRemove(ctx, conf.Network, client.NetworkRemoveOptions{}); err != nil {
		log.Error().Err(err).Msgf("could not remove network '%s'", conf.Network)
	}
	return nil
}

var removeOpts = client.ContainerRemoveOptions{
	RemoveVolumes: true,
	RemoveLinks:   false,
	Force:         false,
}

func isErrContainerNotFoundOrNotRunning(err error) bool {
	// Error response from daemon: Cannot kill container: ...: No such container: ...
	// Error response from daemon: Cannot kill container: ...: Container ... is not running"
	// Error response from podman daemon: can only kill running containers. ... is in state exited
	// Error response from daemon: removal of container ... is already in progress
	// Error: No such container: ...
	return err != nil &&
		(strings.Contains(err.Error(), "No such container") ||
			strings.Contains(err.Error(), "is not running") ||
			strings.Contains(err.Error(), "can only kill running containers") ||
			(strings.Contains(err.Error(), "removal of container") && strings.Contains(err.Error(), "is already in progress")))
}

// normalizeArchType converts the arch type reported by docker info into
// the runtime.GOARCH format
// TODO: find out if we we need to convert other arch types too
func normalizeArchType(s string) string {
	switch s {
	case "x86_64":
		return "amd64"
	case "aarch64":
		return "arm64"
	default:
		return s
	}
}
