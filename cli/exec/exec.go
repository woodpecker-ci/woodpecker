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

	"github.com/rs/zerolog/log"
	"github.com/urfave/cli/v3"

	"go.woodpecker-ci.org/woodpecker/v2/cli/common"
	"go.woodpecker-ci.org/woodpecker/v2/pipeline"
	"go.woodpecker-ci.org/woodpecker/v2/pipeline/backend"
	"go.woodpecker-ci.org/woodpecker/v2/pipeline/backend/docker"
	"go.woodpecker-ci.org/woodpecker/v2/pipeline/backend/kubernetes"
	"go.woodpecker-ci.org/woodpecker/v2/pipeline/backend/local"
	backend_types "go.woodpecker-ci.org/woodpecker/v2/pipeline/backend/types"
	pipeline_errors "go.woodpecker-ci.org/woodpecker/v2/pipeline/errors"
	"go.woodpecker-ci.org/woodpecker/v2/pipeline/frontend/metadata"
	"go.woodpecker-ci.org/woodpecker/v2/pipeline/frontend/yaml/compiler"
	"go.woodpecker-ci.org/woodpecker/v2/pipeline/frontend/yaml/stepbuilder"
	pipeline_log "go.woodpecker-ci.org/woodpecker/v2/pipeline/log"
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

func run(ctx context.Context, c *cli.Command) error {
	repoPath := c.Args().First()
	if repoPath == "" {
		repoPath = "."
	}

	yamls, err := common.GetConfigs(ctx, path.Join(repoPath, ".woodpecker"))
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
	workspaceBase := c.String("workspace-base")
	workspacePath := c.String("workspace-path")
	if c.Bool("local") {
		volumes = append(volumes, c.String("prefix")+"_default:"+workspaceBase)
		volumes = append(volumes, repoPath+":"+path.Join(workspaceBase, workspacePath))
	}

	getWorkflowMetadata := func(workflow *model.Workflow) metadata.Metadata {
		return metadataFromCommand(c, workflow)
	}

	repoIsTrusted := false
	host := "localhost"
	privilegedPlugins := c.StringSlice("plugins-privileged")
	secrets := []compiler.Secret{}

	b := stepbuilder.NewStepBuilder(yamls, getWorkflowMetadata, repoIsTrusted, host, envs,
		compiler.WithEscalated(
			privilegedPlugins...,
		),
		compiler.WithVolumes(volumes...),
		compiler.WithWorkspace(
			workspaceBase,
			workspacePath,
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
		compiler.WithSecret(secrets...),
		// compiler.WithEnviron(pipelineEnv), // TODO: pipelineEnv
	)
	b.PrivilegedPlugins = privilegedPlugins

	items, err := b.Build()
	if pipeline_errors.HasBlockingErrors(err) {
		return err
	} else if err != nil {
		log.Error().Err(err).Msg("error building pipeline")
	}

	done := make(map[string]bool)
	for len(done) < len(items) {
		for _, item := range items {
			// skip already executed workflows
			if done[item.Workflow.Name] {
				continue
			}

			if waitingForDependencies(item, done) {
				continue
			}

			fmt.Printf("# %s\n", item.Workflow.Name)

			err := runWorkflow(ctx, c, item.Config, item.Workflow.Name)
			if err != nil {
				return err
			}

			done[item.Workflow.Name] = true
		}
	}

	return nil
}

var backends = []backend_types.Backend{
	kubernetes.New(),
	docker.New(),
	local.New(),
}

func runWorkflow(ctx context.Context, c *cli.Command, compiled *backend_types.Config, workflowName string) error {
	backendCtx := context.WithValue(ctx, backend_types.CliCommand, c)
	backendEngine, err := backend.FindBackend(backendCtx, backends, c.String("backend-engine"))
	if err != nil {
		return err
	}

	if _, err = backendEngine.Load(backendCtx); err != nil {
		return err
	}

	pipelineCtx, cancel := context.WithTimeout(context.Background(), c.Duration("timeout"))
	defer cancel()
	pipelineCtx = utils.WithContextSigtermCallback(pipelineCtx, func() {
		fmt.Printf("ctrl+c received, terminating current pipeline '%s'\n", workflowName)
	})

	return pipeline.New(compiled,
		pipeline.WithContext(pipelineCtx), //nolint:contextcheck
		pipeline.WithTracer(pipeline.DefaultTracer),
		pipeline.WithLogger(defaultLogger),
		pipeline.WithBackend(backendEngine),
		pipeline.WithDescription(map[string]string{
			"CLI": "exec",
		}),
	).Run(ctx)
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

var defaultLogger = pipeline.Logger(func(step *backend_types.Step, rc io.ReadCloser) error {
	logWriter := NewLineWriter(step.Name, step.UUID)
	return pipeline_log.CopyLineByLine(logWriter, rc, pipeline.MaxLogLineLength)
})

func waitingForDependencies(item *stepbuilder.Item, done map[string]bool) bool {
	for _, step := range item.DependsOn {
		if v, ok := done[step]; !ok || !v {
			return true
		}
	}

	return false
}
