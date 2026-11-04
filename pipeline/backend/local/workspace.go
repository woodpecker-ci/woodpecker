// Copyright 2026 Woodpecker Authors
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

package local

import (
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"

	"go.woodpecker-ci.org/woodpecker/v3/pipeline/backend/types"
)

const overlayDirName = "overlay"

// stepWorkspace returns the directory a step runs in. Steps with the overlay mode get a
// copy of the workspace, readonly can not be enforced as steps run directly in the directory.
func stepWorkspace(step *types.Step, state *workflowState) (string, error) {
	switch step.WorkspaceMode {
	case types.WorkspaceModeReadOnly:
		return "", fmt.Errorf("local backend can not mount the workspace read-only, remove workspace: readonly from step %q", step.Name)
	case types.WorkspaceModeOverlay:
		dir := overlayDir(step, state)
		if err := copyWorkspace(state.workspaceDir, dir); err != nil {
			return "", fmt.Errorf("could not copy the workspace for step %q: %w", step.Name, err)
		}
		return dir, nil
	}
	return state.workspaceDir, nil
}

func overlayDir(step *types.Step, state *workflowState) string {
	return filepath.Join(state.baseDir, overlayDirName, step.UUID)
}

// copyWorkspace copies the directory tree src to dst, keeping file modes and symlinks.
func copyWorkspace(src, dst string) error {
	return filepath.WalkDir(src, func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}
		target := filepath.Join(dst, rel)
		info, err := entry.Info()
		if err != nil {
			return err
		}

		switch {
		case entry.IsDir():
			return os.MkdirAll(target, info.Mode().Perm())
		case info.Mode()&fs.ModeSymlink != 0:
			link, err := os.Readlink(path)
			if err != nil {
				return err
			}
			return os.Symlink(link, target)
		case info.Mode().IsRegular():
			return copyFile(path, target, info.Mode().Perm())
		}
		// sockets, devices and pipes are no workspace content
		return nil
	})
}

func copyFile(src, dst string, perm fs.FileMode) (err error) {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	out, err := os.OpenFile(dst, os.O_CREATE|os.O_EXCL|os.O_WRONLY, perm)
	if err != nil {
		return err
	}
	defer func() {
		err = errors.Join(err, out.Close())
	}()

	if _, err := io.Copy(out, in); err != nil {
		return err
	}
	// the mode of OpenFile is masked by the umask
	return out.Chmod(perm)
}
