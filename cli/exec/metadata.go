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

package exec

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"runtime"
	"strings"

	"github.com/urfave/cli/v3"

	"go.woodpecker-ci.org/woodpecker/v2/pipeline/frontend/metadata"
	"go.woodpecker-ci.org/woodpecker/v2/pipeline/frontend/yaml/matrix"
	"go.woodpecker-ci.org/woodpecker/v2/version"
)

// return the metadata from the cli context.
func metadataFromContext(_ context.Context, c *cli.Command, axis matrix.Axis, w *metadata.Workflow) (*metadata.Metadata, error) {
	m := &metadata.Metadata{}

	if c.IsSet("metadata-file") {
		metadataFile, err := os.Open(c.String("metadata-file"))
		if err != nil {
			return nil, err
		}
		defer metadataFile.Close()

		if err := json.NewDecoder(metadataFile).Decode(m); err != nil {
			return nil, err
		}
	}

	platform := c.String("system-platform")
	if platform == "" {
		platform = runtime.GOOS + "/" + runtime.GOARCH
	}

	metadataFileAndOverrideOrDefault(c, "repo-name", func(fullRepoName string) {
		if idx := strings.LastIndex(fullRepoName, "/"); idx != -1 {
			m.Repo.Owner = fullRepoName[:idx]
			m.Repo.Name = fullRepoName[idx+1:]
		}
	}, c.String)

	var err error
	metadataFileAndOverrideOrDefault(c, "pipeline-changed-files", func(changedFilesRaw string) {
		var changedFiles []string
		if len(changedFilesRaw) != 0 && changedFilesRaw[0] == '[' {
			if jsonErr := json.Unmarshal([]byte(changedFilesRaw), &changedFiles); jsonErr != nil {
				err = fmt.Errorf("pipeline-changed-files detected json but could not parse it: %w", jsonErr)
			}
		} else {
			for _, file := range strings.Split(changedFilesRaw, ",") {
				changedFiles = append(changedFiles, strings.TrimSpace(file))
			}
		}
		m.Curr.Commit.ChangedFiles = changedFiles
	}, c.String)
	if err != nil {
		return nil, err
	}

	// Repo
	metadataFileAndOverrideOrDefault(c, "repo-remote-id", func(s string) { m.Repo.RemoteID = s }, c.String)
	metadataFileAndOverrideOrDefault(c, "repo-url", func(s string) { m.Repo.ForgeURL = s }, c.String)
	metadataFileAndOverrideOrDefault(c, "repo-scm", func(s string) { m.Repo.SCM = s }, c.String)
	metadataFileAndOverrideOrDefault(c, "repo-default-branch", func(s string) { m.Repo.Branch = s }, c.String)
	metadataFileAndOverrideOrDefault(c, "repo-clone-url", func(s string) { m.Repo.CloneURL = s }, c.String)
	metadataFileAndOverrideOrDefault(c, "repo-clone-ssh-url", func(s string) { m.Repo.CloneSSHURL = s }, c.String)
	metadataFileAndOverrideOrDefault(c, "repo-private", func(b bool) { m.Repo.Private = b }, c.Bool)
	metadataFileAndOverrideOrDefault(c, "repo-trusted", func(b bool) { m.Repo.Trusted = b }, c.Bool)

	// Current Pipeline
	metadataFileAndOverrideOrDefault(c, "pipeline-number", func(i int64) { m.Curr.Number = i }, c.Int)
	metadataFileAndOverrideOrDefault(c, "pipeline-parent", func(i int64) { m.Curr.Parent = i }, c.Int)
	metadataFileAndOverrideOrDefault(c, "pipeline-created", func(i int64) { m.Curr.Created = i }, c.Int)
	metadataFileAndOverrideOrDefault(c, "pipeline-started", func(i int64) { m.Curr.Started = i }, c.Int)
	metadataFileAndOverrideOrDefault(c, "pipeline-finished", func(i int64) { m.Curr.Finished = i }, c.Int)
	metadataFileAndOverrideOrDefault(c, "pipeline-status", func(s string) { m.Curr.Status = s }, c.String)
	metadataFileAndOverrideOrDefault(c, "pipeline-event", func(s string) { m.Curr.Event = s }, c.String)
	metadataFileAndOverrideOrDefault(c, "pipeline-url", func(s string) { m.Curr.ForgeURL = s }, c.String)
	metadataFileAndOverrideOrDefault(c, "pipeline-deploy-to", func(s string) { m.Curr.DeployTo = s }, c.String)
	metadataFileAndOverrideOrDefault(c, "pipeline-deploy-task", func(s string) { m.Curr.DeployTask = s }, c.String)

	// Current Pipeline Commit
	metadataFileAndOverrideOrDefault(c, "commit-sha", func(s string) { m.Curr.Commit.Sha = s }, c.String)
	metadataFileAndOverrideOrDefault(c, "commit-ref", func(s string) { m.Curr.Commit.Ref = s }, c.String)
	metadataFileAndOverrideOrDefault(c, "commit-refspec", func(s string) { m.Curr.Commit.Refspec = s }, c.String)
	metadataFileAndOverrideOrDefault(c, "commit-branch", func(s string) { m.Curr.Commit.Branch = s }, c.String)
	metadataFileAndOverrideOrDefault(c, "commit-message", func(s string) { m.Curr.Commit.Message = s }, c.String)
	metadataFileAndOverrideOrDefault(c, "commit-author-name", func(s string) { m.Curr.Commit.Author.Name = s }, c.String)
	metadataFileAndOverrideOrDefault(c, "commit-author-email", func(s string) { m.Curr.Commit.Author.Email = s }, c.String)
	metadataFileAndOverrideOrDefault(c, "commit-author-avatar", func(s string) { m.Curr.Commit.Author.Avatar = s }, c.String)

	metadataFileAndOverrideOrDefault(c, "commit-pull-labels", func(sl []string) { m.Curr.Commit.PullRequestLabels = sl }, c.StringSlice)
	metadataFileAndOverrideOrDefault(c, "commit-release-is-pre", func(b bool) { m.Curr.Commit.IsPrerelease = b }, c.Bool)

	// Previous Pipeline
	metadataFileAndOverrideOrDefault(c, "prev-pipeline-number", func(i int64) { m.Prev.Number = i }, c.Int)
	metadataFileAndOverrideOrDefault(c, "prev-pipeline-created", func(i int64) { m.Prev.Created = i }, c.Int)
	metadataFileAndOverrideOrDefault(c, "prev-pipeline-started", func(i int64) { m.Prev.Started = i }, c.Int)
	metadataFileAndOverrideOrDefault(c, "prev-pipeline-finished", func(i int64) { m.Prev.Finished = i }, c.Int)
	metadataFileAndOverrideOrDefault(c, "prev-pipeline-status", func(s string) { m.Prev.Status = s }, c.String)
	metadataFileAndOverrideOrDefault(c, "prev-pipeline-event", func(s string) { m.Prev.Event = s }, c.String)
	metadataFileAndOverrideOrDefault(c, "prev-pipeline-url", func(s string) { m.Prev.ForgeURL = s }, c.String)

	// Previous Pipeline Commit
	metadataFileAndOverrideOrDefault(c, "prev-commit-sha", func(s string) { m.Prev.Commit.Sha = s }, c.String)
	metadataFileAndOverrideOrDefault(c, "prev-commit-ref", func(s string) { m.Prev.Commit.Ref = s }, c.String)
	metadataFileAndOverrideOrDefault(c, "prev-commit-refspec", func(s string) { m.Prev.Commit.Refspec = s }, c.String)
	metadataFileAndOverrideOrDefault(c, "prev-commit-branch", func(s string) { m.Prev.Commit.Branch = s }, c.String)
	metadataFileAndOverrideOrDefault(c, "prev-commit-message", func(s string) { m.Prev.Commit.Message = s }, c.String)
	metadataFileAndOverrideOrDefault(c, "prev-commit-author-name", func(s string) { m.Prev.Commit.Author.Name = s }, c.String)
	metadataFileAndOverrideOrDefault(c, "prev-commit-author-email", func(s string) { m.Prev.Commit.Author.Email = s }, c.String)
	metadataFileAndOverrideOrDefault(c, "prev-commit-author-avatar", func(s string) { m.Prev.Commit.Author.Avatar = s }, c.String)

	// Workflow
	metadataFileAndOverrideOrDefault(c, "workflow-name", func(s string) { m.Workflow.Name = s }, c.String)
	metadataFileAndOverrideOrDefault(c, "workflow-number", func(i int64) { m.Workflow.Number = int(i) }, c.Int)
	m.Workflow.Matrix = axis

	// System
	metadataFileAndOverrideOrDefault(c, "system-name", func(s string) { m.Sys.Name = s }, c.String)
	metadataFileAndOverrideOrDefault(c, "system-url", func(s string) { m.Sys.URL = s }, c.String)
	metadataFileAndOverrideOrDefault(c, "system-host", func(s string) { m.Sys.Host = s }, c.String)
	m.Sys.Platform = platform
	m.Sys.Version = version.Version

	// Forge
	metadataFileAndOverrideOrDefault(c, "forge-type", func(s string) { m.Forge.Type = s }, c.String)
	metadataFileAndOverrideOrDefault(c, "forge-url", func(s string) { m.Forge.URL = s }, c.String)

	if w != nil {
		m.Workflow = *w
	}

	return m, nil
}

// metadataFileAndOverrideOrDefault will either use the flag default or if metadata file is set only overload if explicit set.
func metadataFileAndOverrideOrDefault[T any](c *cli.Command, flag string, setter func(T), getter func(string) T) {
	if !c.IsSet("metadata-file") || c.IsSet(flag) {
		setter(getter(flag))
	}
}
