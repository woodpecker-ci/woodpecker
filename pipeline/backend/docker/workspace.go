// Copyright 2026 Woodpecker Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
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
	"strings"

	"github.com/containerd/errdefs"
	"github.com/moby/moby/api/types/container"
	"github.com/moby/moby/client"

	backend_types "go.woodpecker-ci.org/woodpecker/v3/pipeline/backend/types"
)

// workspaceBindIndex returns the index of the workspace volume bind, a named
// volume mounted at the workspace base, or -1 if there is none.
func workspaceBindIndex(binds []string, workspaceBase string) int {
	for i, bind := range binds {
		parts, err := splitVolumeParts(bind)
		if err != nil || len(parts) < minVolumeComponents {
			continue
		}
		if parts[1] == workspaceBase && !strings.HasPrefix(parts[0], "/") {
			return i
		}
	}
	return -1
}

// overlayVolumeNames returns the names of the overlay volume of a step and of
// the volumes holding its upper and work directory.
func overlayVolumeNames(step *backend_types.Step) (overlay, upper, work string) {
	name := toContainerName(step)
	return name + "_overlay", name + "_upper", name + "_work"
}

// overlayDriverOpts returns the local volume driver options that mount an overlayfs.
func overlayDriverOpts(lower, upper, work string) map[string]string {
	return map[string]string{
		"type":   "overlay",
		"device": "overlay",
		"o":      fmt.Sprintf("lowerdir=%s,upperdir=%s,workdir=%s", lower, upper, work),
	}
}

// setupWorkspaceOverlay replaces the workspace volume of an overlay step with an
// overlay volume, so the step can change the workspace without keeping the changes.
func (e *docker) setupWorkspaceOverlay(ctx context.Context, step *backend_types.Step, hostConfig *container.HostConfig) error {
	if e.info.OSType == "windows" {
		return errors.New("workspace overlay is not supported on Windows")
	}

	index := workspaceBindIndex(hostConfig.Binds, step.WorkspaceBase)
	if index < 0 {
		return fmt.Errorf("step %q has no workspace volume to overlay", step.Name)
	}
	parts, err := splitVolumeParts(hostConfig.Binds[index])
	if err != nil {
		return err
	}

	workspace, err := e.client.VolumeInspect(ctx, parts[0], client.VolumeInspectOptions{})
	if err != nil {
		return fmt.Errorf("could not inspect workspace volume: %w", err)
	}
	if workspace.Volume.Driver != volumeDriver {
		return fmt.Errorf("workspace overlay needs the %s volume driver, got %s", volumeDriver, workspace.Volume.Driver)
	}

	overlayName, upperName, workName := overlayVolumeNames(step)
	upper, err := e.client.VolumeCreate(ctx, client.VolumeCreateOptions{Name: upperName, Driver: volumeDriver})
	if err != nil {
		return err
	}
	work, err := e.client.VolumeCreate(ctx, client.VolumeCreateOptions{Name: workName, Driver: volumeDriver})
	if err != nil {
		return err
	}
	_, err = e.client.VolumeCreate(ctx, client.VolumeCreateOptions{
		Name:       overlayName,
		Driver:     volumeDriver,
		DriverOpts: overlayDriverOpts(workspace.Volume.Mountpoint, upper.Volume.Mountpoint, work.Volume.Mountpoint),
	})
	if err != nil {
		return err
	}

	hostConfig.Binds[index] = overlayName + ":" + parts[1]
	return nil
}

// removeWorkspaceOverlay removes the overlay volumes of a step, the overlay first
// as it uses the others.
func (e *docker) removeWorkspaceOverlay(ctx context.Context, step *backend_types.Step) error {
	overlayName, upperName, workName := overlayVolumeNames(step)
	var errs []error
	for _, name := range []string{overlayName, upperName, workName} {
		if _, err := e.client.VolumeRemove(ctx, name, client.VolumeRemoveOptions{Force: true}); err != nil && !errdefs.IsNotFound(err) {
			errs = append(errs, fmt.Errorf("could not remove volume '%s': %w", name, err))
		}
	}
	return errors.Join(errs...)
}
