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

package exec

import (
	"context"
	"fmt"
	"io"
	"path"
	"path/filepath"
	"strings"

	"github.com/urfave/cli/v2"

	"go.woodpecker-ci.org/woodpecker/v2/cli/common"
	"go.woodpecker-ci.org/woodpecker/v2/pipeline"
	"go.woodpecker-ci.org/woodpecker/v2/pipeline/backend"
	"go.woodpecker-ci.org/woodpecker/v2/pipeline/backend/docker"
	"go.woodpecker-ci.org/woodpecker/v2/pipeline/backend/dummy"
	"go.woodpecker-ci.org/woodpecker/v2/pipeline/backend/kubernetes"
	"go.woodpecker-ci.org/woodpecker/v2/pipeline/backend/local"
	backend_types "go.woodpecker-ci.org/woodpecker/v2/pipeline/backend/types"
	"go.woodpecker-ci.org/woodpecker/v2/pipeline/frontend/metadata"
	"go.woodpecker-ci.org/woodpecker/v2/pipeline/frontend/yaml/compiler"
	"go.woodpecker-ci.org/woodpecker/v2/pipeline/frontend/yaml/stepbuilder"
	pipelineLog "go.woodpecker-ci.org/woodpecker/v2/pipeline/log"
	"go.woodpecker-ci.org/woodpecker/v2/server/model"
	"go.woodpecker-ci.org/woodpecker/v2/shared/utils"
)

// Command exports the exec command.
var Command = &cli.Command{
	Name:      "exec",
	Usage:     "execute a local pipeline",
	ArgsUsage: "[path/to/.woodpecker.yaml]",
	Action:    run,
	Flags:     utils.MergeSlices(flags, docker.Flags, kubernetes.Flags, local.Flags),
}

type Item struct {
	Workflow  *model.Workflow
	Labels    map[string]string
	DependsOn []string
	RunsOn    []string
	Config    *backend_types.Config
}

func run(c *cli.Context) error {
	dir := c.Args().First()
	if dir == "" {
		dir = "."
	}

	yamls, err := common.GetConfigs(c, path.Join(dir, ".woodpecker"))
	if err != nil {
		return err
	}

	envs := make(map[string]string)
	for _, env := range c.StringSlice("env") {
		before, after, _ := strings.Cut(env, "=")
		envs[before] = after
	}

	// configure volumes for local execution
	volumes := c.StringSlice("volumes")
	// if c.Bool("local") {
	// 	var (
	// 		workspaceBase = conf.Workspace.Base
	// 		workspacePath = conf.Workspace.Path
	// 	)
	// 	if workspaceBase == "" {
	// 		workspaceBase = c.String("workspace-base")
	// 	}
	// 	if workspacePath == "" {
	// 		workspacePath = c.String("workspace-path")
	// 	}

	// 	volumes = append(volumes, c.String("prefix")+"_default:"+workspaceBase)
	// 	volumes = append(volumes, repoPath+":"+path.Join(workspaceBase, workspacePath))
	// }

	getWorkflowMetadata := func(workflow *model.Workflow) metadata.Metadata {
		return metadata.Metadata{} // TODO: metadata
	}

	repoIsTrusted := false
	host := "localhost"

	b := stepbuilder.NewStepBuilder(yamls, getWorkflowMetadata, repoIsTrusted, host, envs,
		compiler.WithEscalated(
			c.StringSlice("privileged")...,
		),
		compiler.WithVolumes(volumes...),
		compiler.WithWorkspace(
			c.String("workspace-base"),
			c.String("workspace-path"),
		),
		compiler.WithNetworks(
			c.StringSlice("network")...,
		),
		compiler.WithPrefix(
			c.String("prefix"),
		),
		compiler.WithProxy(compiler.ProxyOptions{
			NoProxy:    c.String("backend-no-proxy"),
			HTTPProxy:  c.String("backend-http-proxy"),
			HTTPSProxy: c.String("backend-https-proxy"),
		}),
		compiler.WithLocal(
			c.Bool("local"),
		),
		compiler.WithNetrc(
			c.String("netrc-username"),
			c.String("netrc-password"),
			c.String("netrc-machine"),
		),
		// compiler.WithMetadata(metadata),
		// compiler.WithSecret(secrets...), // TODO: secrets
		// compiler.WithEnviron(pipelineEnv), // TODO: pipelineEnv
	)
	items, err := b.Build()
	if err != nil {
		return err
	}

	for _, item := range items {
		// TODO: check dependencies
		err := runWorkflow(c, item.Config)
		if err != nil {
			return err
		}
	}

	return nil
}

func runWorkflow(c *cli.Context, compiled *backend_types.Config) error {
	backendCtx := context.WithValue(c.Context, backend_types.CliContext, c)
	backends := []backend_types.Backend{
		kubernetes.New(),
		docker.New(),
		local.New(),
		dummy.New(),
	}
	backendEngine, err := backend.FindBackend(backendCtx, backends, c.String("backend-engine"))
	if err != nil {
		return err
	}

	if _, err = backendEngine.Load(backendCtx); err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), c.Duration("timeout"))
	defer cancel()
	ctx = utils.WithContextSigtermCallback(ctx, func() {
		fmt.Println("ctrl+c received, terminating process")
	})

	return pipeline.New(compiled,
		pipeline.WithContext(ctx),
		pipeline.WithTracer(pipeline.DefaultTracer),
		pipeline.WithLogger(defaultLogger),
		pipeline.WithBackend(backendEngine),
		pipeline.WithDescription(map[string]string{
			"CLI": "exec",
		}),
	).Run(c.Context)
}

// convertPathForWindows converts a path to use slash separators
// for Windows. If the path is a Windows volume name like C:, it
// converts it to an absolute root path starting with slash (e.g.
// C: -> /c). Otherwise it just converts backslash separators to
// slashes.
func convertPathForWindows(path string) string {
	base := filepath.VolumeName(path)

	// Check if path is volume name like C:
	//nolint:mnd
	if len(base) == 2 {
		path = strings.TrimPrefix(path, base)
		base = strings.ToLower(strings.TrimSuffix(base, ":"))
		return "/" + base + filepath.ToSlash(path)
	}

	return filepath.ToSlash(path)
}

const maxLogLineLength = 1024 * 1024 // 1mb
var defaultLogger = pipeline.Logger(func(step *backend_types.Step, rc io.ReadCloser) error {
	logWriter := NewLineWriter(step.Name, step.UUID)
	return pipelineLog.CopyLineByLine(logWriter, rc, maxLogLineLength)
})
