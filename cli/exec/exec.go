package exec

import (
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"path"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/drone/envsubst"
	"github.com/urfave/cli/v2"

	"github.com/woodpecker-ci/woodpecker/cli/common"
	"github.com/woodpecker-ci/woodpecker/pipeline"
	"github.com/woodpecker-ci/woodpecker/pipeline/backend"
	"github.com/woodpecker-ci/woodpecker/pipeline/backend/docker"
	"github.com/woodpecker-ci/woodpecker/pipeline/frontend"
	"github.com/woodpecker-ci/woodpecker/pipeline/frontend/yaml"
	"github.com/woodpecker-ci/woodpecker/pipeline/frontend/yaml/compiler"
	"github.com/woodpecker-ci/woodpecker/pipeline/frontend/yaml/linter"
	"github.com/woodpecker-ci/woodpecker/pipeline/frontend/yaml/matrix"
	"github.com/woodpecker-ci/woodpecker/pipeline/interrupt"
	"github.com/woodpecker-ci/woodpecker/pipeline/multipart"
)

// Command exports the exec command.
var Command = &cli.Command{
	Name:      "exec",
	Usage:     "execute a local build",
	ArgsUsage: "[path/to/.woodpecker.yml]",
	Action:    exec,
	Flags:     append(common.GlobalFlags, flags...),
}

func exec(c *cli.Context) error {
	file := c.Args().First()
	if file == "" {
		file = ".woodpecker.yml"
	}

	dat, err := ioutil.ReadFile(file)
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
		err := execWithAxis(c, axis)
		if err != nil {
			return err
		}
	}
	return nil
}

func execWithAxis(c *cli.Context, axis matrix.Axis) error {
	file := c.Args().First()
	if file == "" {
		file = ".woodpecker.yml"
	}

	metadata := metadataFromContext(c, axis)
	environ := metadata.Environ()
	var secrets []compiler.Secret
	for key, val := range metadata.Job.Matrix {
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
		dir, _ := filepath.Abs(filepath.Dir(file))

		if runtime.GOOS == "windows" {
			dir = convertPathForWindows(dir)
		}
		volumes = append(volumes, c.String("prefix")+"_default:"+workspaceBase)
		volumes = append(volumes, dir+":"+path.Join(workspaceBase, workspacePath))
	}

	// lint the yaml file
	if lerr := linter.New(linter.WithTrusted(true)).Lint(conf); lerr != nil {
		return lerr
	}

	// compiles the yaml file
	compiled := compiler.New(
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
	engine, err := docker.NewEnv()
	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), c.Duration("timeout"))
	defer cancel()
	ctx = interrupt.WithContext(ctx)

	return pipeline.New(compiled,
		pipeline.WithContext(ctx),
		pipeline.WithTracer(pipeline.DefaultTracer),
		pipeline.WithLogger(defaultLogger),
		pipeline.WithEngine(engine),
	).Run()
}

// return the metadata from the cli context.
func metadataFromContext(c *cli.Context, axis matrix.Axis) frontend.Metadata {
	return frontend.Metadata{
		Repo: frontend.Repo{
			Name:    c.String("repo-name"),
			Link:    c.String("repo-link"),
			Remote:  c.String("repo-remote-url"),
			Private: c.Bool("repo-private"),
		},
		Curr: frontend.Build{
			Number:   c.Int("build-number"),
			Parent:   c.Int("parent-build-number"),
			Created:  c.Int64("build-created"),
			Started:  c.Int64("build-started"),
			Finished: c.Int64("build-finished"),
			Status:   c.String("build-status"),
			Event:    c.String("build-event"),
			Link:     c.String("build-link"),
			Target:   c.String("build-target"),
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
		Prev: frontend.Build{
			Number:   c.Int("prev-build-number"),
			Created:  c.Int64("prev-build-created"),
			Started:  c.Int64("prev-build-started"),
			Finished: c.Int64("prev-build-finished"),
			Status:   c.String("prev-build-status"),
			Event:    c.String("prev-build-event"),
			Link:     c.String("prev-build-link"),
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
		Job: frontend.Job{
			Number: c.Int("job-number"),
			Matrix: axis,
		},
		Sys: frontend.System{
			Name: c.String("system-name"),
			Link: c.String("system-link"),
			Arch: c.String("system-arch"),
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

var defaultLogger = pipeline.LogFunc(func(proc *backend.Step, rc multipart.Reader) error {
	part, err := rc.NextPart()
	if err != nil {
		return err
	}

	logstream := NewLineWriter(proc.Alias)
	io.Copy(logstream, part)

	return nil
})
