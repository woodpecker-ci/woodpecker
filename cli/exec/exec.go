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
	"github.com/urfave/cli/v2"

	"github.com/woodpecker-ci/woodpecker/cli/common"
	"github.com/woodpecker-ci/woodpecker/pipeline"
	"github.com/woodpecker-ci/woodpecker/pipeline/backend"
	"github.com/woodpecker-ci/woodpecker/pipeline/backend/types"
	backendTypes "github.com/woodpecker-ci/woodpecker/pipeline/backend/types"
	"github.com/woodpecker-ci/woodpecker/pipeline/frontend"
	"github.com/woodpecker-ci/woodpecker/pipeline/frontend/yaml"
	"github.com/woodpecker-ci/woodpecker/pipeline/frontend/yaml/compiler"
	"github.com/woodpecker-ci/woodpecker/pipeline/frontend/yaml/linter"
	"github.com/woodpecker-ci/woodpecker/pipeline/frontend/yaml/matrix"
	"github.com/woodpecker-ci/woodpecker/pipeline/multipart"
	"github.com/woodpecker-ci/woodpecker/shared/utils"
)

// Command exports the exec command.
var Command = &cli.Command{
	Name:      "exec",
	Usage:     "execute a local pipeline",
	ArgsUsage: "[path/to/.woodpecker.yaml]",
	Action:    run,
	Flags:     append(common.GlobalFlags, flags...),
}

func run(c *cli.Context) error {
	return common.RunPipelineFunc(c, execFile, execDir)
}

func execDir(c *cli.Context, dir string) error {
	// TODO: respect pipeline dependency
	repoPath, _ := filepath.Abs(filepath.Dir(dir))
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
	repoPath, _ := filepath.Abs(filepath.Dir(file))
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
		return fmt.Errorf("Parse matrix fail")
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

	droneEnv := make(map[string]string)
	for _, env := range c.StringSlice("env") {
		envs := strings.SplitN(env, "=", 2)
		droneEnv[envs[0]] = envs[1]
		if _, exists := environ[envs[0]]; exists {
			// don't override existing values
			continue
		}
		environ[envs[0]] = envs[1]
	}

	tmpl, err := envsubst.ParseFile(file)
	if err != nil {
		return err
	}
	confstr, err := tmpl.Execute(func(name string) string {
		return environ[name]
	})
	if err != nil {
		return err
	}

	conf, err := yaml.ParseString(confstr)
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
	if lerr := linter.New(linter.WithTrusted(true)).Lint(conf); lerr != nil {
		return lerr
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
		compiler.WithProxy(),
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
		compiler.WithEnviron(droneEnv),
	).Compile(conf)
	if err != nil {
		return err
	}

	backendCtx := context.WithValue(c.Context, types.CliContext, c)
	backend.Init(backendCtx)

	engine, err := backend.FindEngine(backendCtx, c.String("backend-engine"))
	if err != nil {
		return err
	}

	if err = engine.Load(backendCtx); err != nil {
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
		pipeline.WithEngine(engine),
		pipeline.WithDescription(map[string]string{
			"CLI": "exec",
		}),
	).Run(c.Context)
}

// return the metadata from the cli context.
func metadataFromContext(c *cli.Context, axis matrix.Axis) frontend.Metadata {
	platform := c.String("system-platform")
	if platform == "" {
		platform = runtime.GOOS + "/" + runtime.GOARCH
	}

	return frontend.Metadata{
		Repo: frontend.Repo{
			Name:     c.String("repo-name"),
			Link:     c.String("repo-link"),
			CloneURL: c.String("repo-clone-url"),
			Private:  c.Bool("repo-private"),
		},
		Curr: frontend.Pipeline{
			Number:   c.Int64("pipeline-number"),
			Parent:   c.Int64("pipeline-parent"),
			Created:  c.Int64("pipeline-created"),
			Started:  c.Int64("pipeline-started"),
			Finished: c.Int64("pipeline-finished"),
			Status:   c.String("pipeline-status"),
			Event:    c.String("pipeline-event"),
			Link:     c.String("pipeline-link"),
			Target:   c.String("pipeline-target"),
			Commit: frontend.Commit{
				Sha:     c.String("commit-sha"),
				Ref:     c.String("commit-ref"),
				Refspec: c.String("commit-refspec"),
				Branch:  c.String("commit-branch"),
				Message: c.String("commit-message"),
				Author: frontend.Author{
					Name:   c.String("commit-author-name"),
					Email:  c.String("commit-author-email"),
					Avatar: c.String("commit-author-avatar"),
				},
			},
		},
		Prev: frontend.Pipeline{
			Number:   c.Int64("prev-pipeline-number"),
			Created:  c.Int64("prev-pipeline-created"),
			Started:  c.Int64("prev-pipeline-started"),
			Finished: c.Int64("prev-pipeline-finished"),
			Status:   c.String("prev-pipeline-status"),
			Event:    c.String("prev-pipeline-event"),
			Link:     c.String("prev-pipeline-link"),
			Commit: frontend.Commit{
				Sha:     c.String("prev-commit-sha"),
				Ref:     c.String("prev-commit-ref"),
				Refspec: c.String("prev-commit-refspec"),
				Branch:  c.String("prev-commit-branch"),
				Message: c.String("prev-commit-message"),
				Author: frontend.Author{
					Name:   c.String("prev-commit-author-name"),
					Email:  c.String("prev-commit-author-email"),
					Avatar: c.String("prev-commit-author-avatar"),
				},
			},
		},
		Workflow: frontend.Workflow{
			Name:   c.String("workflow-name"),
			Number: c.Int("workflow-number"),
			Matrix: axis,
		},
		Step: frontend.Step{
			Name:   c.String("step-name"),
			Number: c.Int("step-number"),
		},
		Sys: frontend.System{
			Name:     c.String("system-name"),
			Link:     c.String("system-link"),
			Platform: platform,
		},
	}
}

func convertPathForWindows(path string) string {
	base := filepath.VolumeName(path)
	if len(base) == 2 {
		path = strings.TrimPrefix(path, base)
		base = strings.ToLower(strings.TrimSuffix(base, ":"))
		return "/" + base + filepath.ToSlash(path)
	}

	return filepath.ToSlash(path)
}

var defaultLogger = pipeline.LogFunc(func(step *backendTypes.Step, rc multipart.Reader) error {
	part, err := rc.NextPart()
	if err != nil {
		return err
	}

	logStream := NewLineWriter(step.Alias, step.UUID)
	_, err = io.Copy(logStream, part)
	return err
})
