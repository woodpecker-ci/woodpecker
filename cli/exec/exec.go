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
	"os"
	"path"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/drone/envsubst"
	"github.com/rs/zerolog/log"
	"github.com/urfave/cli/v2"

	"go.woodpecker-ci.org/woodpecker/v2/cli/common"
	"go.woodpecker-ci.org/woodpecker/v2/cli/lint"
	"go.woodpecker-ci.org/woodpecker/v2/pipeline"
	"go.woodpecker-ci.org/woodpecker/v2/pipeline/backend"
	"go.woodpecker-ci.org/woodpecker/v2/pipeline/backend/docker"
	"go.woodpecker-ci.org/woodpecker/v2/pipeline/backend/kubernetes"
	"go.woodpecker-ci.org/woodpecker/v2/pipeline/backend/local"
	backend_types "go.woodpecker-ci.org/woodpecker/v2/pipeline/backend/types"
	"go.woodpecker-ci.org/woodpecker/v2/pipeline/frontend/yaml"
	"go.woodpecker-ci.org/woodpecker/v2/pipeline/frontend/yaml/compiler"
	"go.woodpecker-ci.org/woodpecker/v2/pipeline/frontend/yaml/linter"
	"go.woodpecker-ci.org/woodpecker/v2/pipeline/frontend/yaml/matrix"
	pipelineLog "go.woodpecker-ci.org/woodpecker/v2/pipeline/log"
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

var backends = []backend_types.Backend{
	kubernetes.New(),
	docker.New(),
	local.New(),
}

func run(c *cli.Context) error {
	return common.RunPipelineFunc(c, execFile, execDir)
}

func execDir(c *cli.Context, dir string) error {
	// TODO: respect pipeline dependency
	repoPath := c.String("repo-path")
	if repoPath != "" {
		repoPath, _ = filepath.Abs(repoPath)
	} else {
		repoPath, _ = filepath.Abs(filepath.Dir(dir))
	}
	if runtime.GOOS == "windows" {
		repoPath = convertPathForWindows(repoPath)
	}
	return filepath.Walk(dir, func(path string, info os.FileInfo, e error) error {
		if e != nil {
			return e
		}

		// check if it is a regular file (not dir)
		if info.Mode().IsRegular() && (strings.HasSuffix(info.Name(), ".yaml") || strings.HasSuffix(info.Name(), ".yml")) {
			fmt.Println("#", info.Name())
			_ = runExec(c, path, repoPath) // TODO: should we drop errors or store them and report back?
			fmt.Println("")
			return nil
		}

		return nil
	})
}

func execFile(c *cli.Context, file string) error {
	repoPath := c.String("repo-path")
	if repoPath != "" {
		repoPath, _ = filepath.Abs(repoPath)
	} else {
		repoPath, _ = filepath.Abs(filepath.Dir(file))
	}
	if runtime.GOOS == "windows" {
		repoPath = convertPathForWindows(repoPath)
	}
	return runExec(c, file, repoPath)
}

func runExec(c *cli.Context, file, repoPath string) error {
	dat, err := os.ReadFile(file)
	if err != nil {
		return err
	}

	axes, err := matrix.ParseString(string(dat))
	if err != nil {
		return fmt.Errorf("parse matrix fail")
	}

	if len(axes) == 0 {
		axes = append(axes, matrix.Axis{})
	}
	for _, axis := range axes {
		err := execWithAxis(c, file, repoPath, axis)
		if err != nil {
			return err
		}
	}
	return nil
}

func execWithAxis(c *cli.Context, file, repoPath string, axis matrix.Axis) error {
	metadata := metadataFromContext(c, axis)
	environ := metadata.Environ()
	var secrets []compiler.Secret
	for key, val := range metadata.Workflow.Matrix {
		environ[key] = val
		secrets = append(secrets, compiler.Secret{
			Name:  key,
			Value: val,
		})
	}

	pipelineEnv := make(map[string]string)
	for _, env := range c.StringSlice("env") {
		before, after, _ := strings.Cut(env, "=")
		pipelineEnv[before] = after
		if oldVar, exists := environ[before]; exists {
			// override existing values, but print a warning
			log.Warn().Msgf("environment variable '%s' had value '%s', but got overwritten", before, oldVar)
		}
		environ[before] = after
	}

	tmpl, err := envsubst.ParseFile(file)
	if err != nil {
		return err
	}
	confStr, err := tmpl.Execute(func(name string) string {
		return environ[name]
	})
	if err != nil {
		return err
	}

	conf, err := yaml.ParseString(confStr)
	if err != nil {
		return err
	}

	// configure volumes for local execution
	volumes := c.StringSlice("volumes")
	if c.Bool("local") {
		var (
			workspaceBase = conf.Workspace.Base
			workspacePath = conf.Workspace.Path
		)
		if workspaceBase == "" {
			workspaceBase = c.String("workspace-base")
		}
		if workspacePath == "" {
			workspacePath = c.String("workspace-path")
		}

		volumes = append(volumes, c.String("prefix")+"_default:"+workspaceBase)
		volumes = append(volumes, repoPath+":"+path.Join(workspaceBase, workspacePath))
	}

	// lint the yaml file
	err = linter.New(linter.WithTrusted(true)).Lint([]*linter.WorkflowConfig{{
		File:      path.Base(file),
		RawConfig: confStr,
		Workflow:  conf,
	}})
	if err != nil {
		str, err := lint.FormatLintError(file, err)
		fmt.Print(str)
		if err != nil {
			return err
		}
	}

	// compiles the yaml file
	compiled, err := compiler.New(
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
		compiler.WithMetadata(metadata),
		compiler.WithSecret(secrets...),
		compiler.WithEnviron(pipelineEnv),
	).Compile(conf)
	if err != nil {
		return err
	}

	backendCtx := context.WithValue(c.Context, backend_types.CliContext, c)
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
