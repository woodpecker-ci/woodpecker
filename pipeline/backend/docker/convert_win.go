// Copyright 2024 Woodpecker Authors
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
	"path/filepath"
	"regexp"
	"strings"

	"go.woodpecker-ci.org/woodpecker/v3/pipeline/backend/types"
)

const (
	osTypeWindows              = "windows"
	defaultWindowsDriverLetter = "C:"
)

var mustNotAddWindowsLetterPattern = regexp.MustCompile(`^(?:` +
	// Drive letter followed by colon and optional backslash (C: or C:\)
	`[a-zA-Z]:(?:\\|$)|` +

	// Device path starting with \\ or // followed by .\ or ./ (\\.\  or //./  or \\./ or //.\ )
	`(?:\\\\|//)\.(?:\\|/).*|` +

	// UNC path starting with \\ or // followed by non-dot (\server or //server)
	`(?:\\\\|//)[^.]|` +

	// Relative path starting with .\ or ./ (.\path or ./path)
	`\.(?:\\|/)` +
	`)`)

func (e *docker) windowsPathPatch(step *types.Step) {
	// only patch if target is windows
	if strings.ToLower(e.info.OSType) != osTypeWindows {
		return
	}

	// patch volumes to have an letter if not already set
	for i, vol := range step.Volumes {
		volume := windowsVolumePatch(vol)
		if volume != "" {
			step.Volumes[i] = volume
		}
	}

	// patch workspace
	if !mustNotAddWindowsLetterPattern.MatchString(step.WorkspaceBase) {
		step.WorkspaceBase = filepath.Join(defaultWindowsDriverLetter, step.WorkspaceBase)
	}
	if !mustNotAddWindowsLetterPattern.MatchString(step.WorkingDir) {
		step.WorkingDir = filepath.Join(defaultWindowsDriverLetter, step.WorkingDir)
	}
	if ciWorkspace, ok := step.Environment["CI_WORKSPACE"]; ok {
		if !mustNotAddWindowsLetterPattern.MatchString(ciWorkspace) {
			step.Environment["CI_WORKSPACE"] = filepath.Join(defaultWindowsDriverLetter, ciWorkspace)
		}
	}
	if step.WorkspaceVolume != "" {
		volume := windowsVolumePatch(step.WorkspaceVolume)
		if volume != "" {
			step.WorkspaceVolume = volume
		}
	}
}

func windowsVolumePatch(vol string) string {
		volParts, err := splitVolumeParts(vol)
		if err != nil || len(volParts) < 2 {
			// ignore non valid volumes for now
			return ""
		}

		// fix source destination
		if strings.HasPrefix(volParts[0], "/") {
			volParts[0] = filepath.Join(defaultWindowsDriverLetter, volParts[0])
		}

		// fix mount destination
		if !mustNotAddWindowsLetterPattern.MatchString(volParts[1]) {
			volParts[1] = filepath.Join(defaultWindowsDriverLetter, volParts[1])
		}
		return strings.Join(volParts, ":")
}
