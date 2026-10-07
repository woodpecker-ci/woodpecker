// Copyright 2023 Woodpecker Authors
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

package repo

import (
	"context"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/urfave/cli/v3"

	"go.woodpecker-ci.org/woodpecker/v3/cli/internal"
	"go.woodpecker-ci.org/woodpecker/v3/woodpecker-go/woodpecker"
)

var repoAddCmd = &cli.Command{
	Name:      "add",
	Usage:     "add a repository",
	ArgsUsage: "<forge-remote-id|repo-full-name>",
	Action:    repoAdd,
}

func repoAdd(ctx context.Context, c *cli.Command) error {
	client, err := internal.NewClient(ctx, c)
	if err != nil {
		return err
	}

	return repoAddWithClient(c.Args().First(), client, c.Writer)
}

func repoAddWithClient(repoArg string, client woodpecker.Client, out io.Writer) error {
	forgeRemoteID, err := repoAddForgeRemoteID(repoArg, client)
	if err != nil {
		return err
	}

	opt := woodpecker.RepoPostOptions{
		ForgeRemoteID: forgeRemoteID,
	}

	repo, err := client.RepoPost(opt)
	if err != nil {
		return err
	}
	if repo == nil {
		return fmt.Errorf("server returned no repository")
	}

	if out == nil {
		out = os.Stdout
	}
	_, err = fmt.Fprintf(out, "Successfully activated repository %s\n", repo.FullName)
	return err
}

func repoAddForgeRemoteID(repoArg string, client woodpecker.Client) (string, error) {
	repoArg = strings.TrimSpace(repoArg)
	if repoArg == "" {
		return "", fmt.Errorf("repository or forge remote id required")
	}

	if !strings.Contains(repoArg, "/") {
		return repoArg, nil
	}

	repos, err := client.RepoList(woodpecker.RepoListOptions{
		All:  true,
		Name: repoNameFromFullName(repoArg),
	})
	if err != nil {
		return "", fmt.Errorf("lookup repository %q: %w", repoArg, err)
	}

	for _, repo := range repos {
		if repo == nil || !strings.EqualFold(repo.FullName, repoArg) {
			continue
		}
		if !validForgeRemoteID(repo.ForgeRemoteID) {
			return "", fmt.Errorf("repository %q has no forge remote id", repoArg)
		}
		return repo.ForgeRemoteID, nil
	}

	return "", fmt.Errorf("repository %q not found", repoArg)
}

func repoNameFromFullName(repoFullName string) string {
	lastSlash := strings.LastIndex(repoFullName, "/")
	if lastSlash == -1 || lastSlash == len(repoFullName)-1 {
		return repoFullName
	}
	return repoFullName[lastSlash+1:]
}

func validForgeRemoteID(forgeRemoteID string) bool {
	return forgeRemoteID != "" && forgeRemoteID != "0"
}
