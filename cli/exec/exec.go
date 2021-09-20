package exec

import (
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/drone/envsubst"
	"github.com/woodpecker-ci/woodpecker/cncd/pipeline/pipeline"
	"github.com/woodpecker-ci/woodpecker/cncd/pipeline/pipeline/backend"
	"github.com/woodpecker-ci/woodpecker/cncd/pipeline/pipeline/backend/docker"
	"github.com/woodpecker-ci/woodpecker/cncd/pipeline/pipeline/frontend"
	"github.com/woodpecker-ci/woodpecker/cncd/pipeline/pipeline/frontend/yaml"
	"github.com/woodpecker-ci/woodpecker/cncd/pipeline/pipeline/frontend/yaml/compiler"
	"github.com/woodpecker-ci/woodpecker/cncd/pipeline/pipeline/frontend/yaml/linter"
	"github.com/woodpecker-ci/woodpecker/cncd/pipeline/pipeline/frontend/yaml/matrix"
	"github.com/woodpecker-ci/woodpecker/cncd/pipeline/pipeline/interrupt"
	"github.com/woodpecker-ci/woodpecker/cncd/pipeline/pipeline/multipart"

	"github.com/urfave/cli"
)

// Command exports the exec command.
var Command = cli.Command{
	Name:      "exec",
	Usage:     "execute a local build",
	ArgsUsage: "[path/to/.drone.yml]",
	Action: func(c *cli.Context) {
		if err := exec(c); err != nil {
			log.Fatalln(err)
		}
	},
	Flags: []cli.Flag{
		cli.BoolTFlag{
			Name:   "local",
			Usage:  "build from local directory",
			EnvVar: "DRONE_LOCAL,WOODPECKER_LOCAL",
		},
		cli.DurationFlag{
			Name:   "timeout",
			Usage:  "build timeout",
			Value:  time.Hour,
			EnvVar: "DRONE_TIMEOUT,WOODPECKER_TIMEOUT",
		},
		cli.StringSliceFlag{
			Name:   "volumes",
			Usage:  "build volumes",
			EnvVar: "DRONE_VOLUMES,WOODPECKER_VOLUMES",
		},
		cli.StringSliceFlag{
			Name:   "network",
			Usage:  "external networks",
			EnvVar: "DRONE_NETWORKS,WOODPECKER_NETWORKS",
		},
		cli.StringFlag{
			Name:   "prefix",
			Value:  "drone",
			Usage:  "prefix containers created by drone",
			EnvVar: "DRONE_DOCKER_PREFIX,WOODPECKER_DOCKER_PREFIX",
			Hidden: true,
		},
		cli.StringSliceFlag{
			Name:  "privileged",
			Usage: "privileged plugins",
			Value: &cli.StringSlice{
				"plugins/docker",
				"plugins/gcr",
				"plugins/ecr",
			},
		},

		//
		// Please note the below flags are mirrored in the pipec and
		// should be kept synchronized. Do not edit directly
		// https://github.com/cncd/pipeline/pipec
		//

		//
		// workspace default
		//
		cli.StringFlag{
			Name:   "workspace-base",
			Value:  "/drone",
			EnvVar: "DRONE_WORKSPACE_BASE,WOODPECKER_WORKSPACE_BASE",
		},
		cli.StringFlag{
			Name:   "workspace-path",
			Value:  "src",
			EnvVar: "DRONE_WORKSPACE_PATH,WOODPECKER_WORKSPACE_PATH",
		},
		//
		// netrc parameters
		//
		cli.StringFlag{
			Name:   "netrc-username",
			EnvVar: "DRONE_NETRC_USERNAME,WOODPECKER_NETRC_USERNAME",
		},
		cli.StringFlag{
			Name:   "netrc-password",
			EnvVar: "DRONE_NETRC_PASSWORD,WOODPECKER_NETRC_PASSWORD",
		},
		cli.StringFlag{
			Name:   "netrc-machine",
			EnvVar: "DRONE_NETRC_MACHINE,WOODPECKER_NETRC_MACHINE",
		},
		//
		// metadata parameters
		//
		cli.StringFlag{
			Name:   "system-arch",
			Value:  "linux/amd64",
			EnvVar: "DRONE_SYSTEM_ARCH,WOODPECKER_SYSTEM_ARCH",
		},
		cli.StringFlag{
			Name:   "system-name",
			Value:  "pipec",
			EnvVar: "DRONE_SYSTEM_NAME,WOODPECKER_SYSTEM_NAME",
		},
		cli.StringFlag{
			Name:   "system-link",
			Value:  "https://github.com/cncd/pipec",
			EnvVar: "DRONE_SYSTEM_LINK,WOODPECKER_SYSTEM_LINK",
		},
		cli.StringFlag{
			Name:   "repo-name",
			EnvVar: "DRONE_REPO_NAME,WOODPECKER_REPO_NAME",
		},
		cli.StringFlag{
			Name:   "repo-link",
			EnvVar: "DRONE_REPO_LINK,WOODPECKER_REPO_LINK",
		},
		cli.StringFlag{
			Name:   "repo-remote-url",
			EnvVar: "DRONE_REPO_REMOTE,WOODPECKER_REPO_REMOTE",
		},
		cli.StringFlag{
			Name:   "repo-private",
			EnvVar: "DRONE_REPO_PRIVATE,WOODPECKER_REPO_PRIVATE",
		},
		cli.IntFlag{
			Name:   "build-number",
			EnvVar: "DRONE_BUILD_NUMBER,WOODPECKER_BUILD_NUMBER",
		},
		cli.IntFlag{
			Name:   "parent-build-number",
			EnvVar: "DRONE_PARENT_BUILD_NUMBER,WOODPECKER_PARENT_BUILD_NUMBER",
		},
		cli.Int64Flag{
			Name:   "build-created",
			EnvVar: "DRONE_BUILD_CREATED,WOODPECKER_BUILD_CREATED",
		},
		cli.Int64Flag{
			Name:   "build-started",
			EnvVar: "DRONE_BUILD_STARTED,WOODPECKER_BUILD_STARTED",
		},
		cli.Int64Flag{
			Name:   "build-finished",
			EnvVar: "DRONE_BUILD_FINISHED,WOODPECKER_BUILD_FINISHED",
		},
		cli.StringFlag{
			Name:   "build-status",
			EnvVar: "DRONE_BUILD_STATUS,WOODPECKER_BUILD_STATUS",
		},
		cli.StringFlag{
			Name:   "build-event",
			EnvVar: "DRONE_BUILD_EVENT,WOODPECKER_BUILD_EVENT",
		},
		cli.StringFlag{
			Name:   "build-link",
			EnvVar: "DRONE_BUILD_LINK,WOODPECKER_BUILD_LINK",
		},
		cli.StringFlag{
			Name:   "build-target",
			EnvVar: "DRONE_BUILD_TARGET,WOODPECKER_BUILD_TARGET",
		},
		cli.StringFlag{
			Name:   "commit-sha",
			EnvVar: "DRONE_COMMIT_SHA,WOODPECKER_COMMIT_SHA",
		},
		cli.StringFlag{
			Name:   "commit-ref",
			EnvVar: "DRONE_COMMIT_REF,WOODPECKER_COMMIT_REF",
		},
		cli.StringFlag{
			Name:   "commit-refspec",
			EnvVar: "DRONE_COMMIT_REFSPEC,WOODPECKER_COMMIT_REFSPEC",
		},
		cli.StringFlag{
			Name:   "commit-branch",
			EnvVar: "DRONE_COMMIT_BRANCH,WOODPECKER_COMMIT_BRANCH",
		},
		cli.StringFlag{
			Name:   "commit-message",
			EnvVar: "DRONE_COMMIT_MESSAGE,WOODPECKER_COMMIT_MESSAGE",
		},
		cli.StringFlag{
			Name:   "commit-author-name",
			EnvVar: "DRONE_COMMIT_AUTHOR_NAME,WOODPECKER_COMMIT_AUTHOR_NAME",
		},
		cli.StringFlag{
			Name:   "commit-author-avatar",
			EnvVar: "DRONE_COMMIT_AUTHOR_AVATAR,WOODPECKER_COMMIT_AUTHOR_AVATAR",
		},
		cli.StringFlag{
			Name:   "commit-author-email",
			EnvVar: "DRONE_COMMIT_AUTHOR_EMAIL,WOODPECKER_COMMIT_AUTHOR_EMAIL",
		},
		cli.IntFlag{
			Name:   "prev-build-number",
			EnvVar: "DRONE_PREV_BUILD_NUMBER,WOODPECKER_PREV_BUILD_NUMBER",
		},
		cli.Int64Flag{
			Name:   "prev-build-created",
			EnvVar: "DRONE_PREV_BUILD_CREATED,WOODPECKER_PREV_BUILD_CREATED",
		},
		cli.Int64Flag{
			Name:   "prev-build-started",
			EnvVar: "DRONE_PREV_BUILD_STARTED,WOODPECKER_PREV_BUILD_STARTED",
		},
		cli.Int64Flag{
			Name:   "prev-build-finished",
			EnvVar: "DRONE_PREV_BUILD_FINISHED,WOODPECKER_PREV_BUILD_FINISHED",
		},
		cli.StringFlag{
			Name:   "prev-build-status",
			EnvVar: "DRONE_PREV_BUILD_STATUS,WOODPECKER_PREV_BUILD_STATUS",
		},
		cli.StringFlag{
			Name:   "prev-build-event",
			EnvVar: "DRONE_PREV_BUILD_EVENT,WOODPECKER_PREV_BUILD_EVENT",
		},
		cli.StringFlag{
			Name:   "prev-build-link",
			EnvVar: "DRONE_PREV_BUILD_LINK,WOODPECKER_PREV_BUILD_LINK",
		},
		cli.StringFlag{
			Name:   "prev-commit-sha",
			EnvVar: "DRONE_PREV_COMMIT_SHA,WOODPECKER_PREV_COMMIT_SHA",
		},
		cli.StringFlag{
			Name:   "prev-commit-ref",
			EnvVar: "DRONE_PREV_COMMIT_REF,WOODPECKER_PREV_COMMIT_REF",
		},
		cli.StringFlag{
			Name:   "prev-commit-refspec",
			EnvVar: "DRONE_PREV_COMMIT_REFSPEC,WOODPECKER_PREV_COMMIT_REFSPEC",
		},
		cli.StringFlag{
			Name:   "prev-commit-branch",
			EnvVar: "DRONE_PREV_COMMIT_BRANCH,WOODPECKER_PREV_COMMIT_BRANCH",
		},
		cli.StringFlag{
			Name:   "prev-commit-message",
			EnvVar: "DRONE_PREV_COMMIT_MESSAGE,WOODPECKER_PREV_COMMIT_MESSAGE",
		},
		cli.StringFlag{
			Name:   "prev-commit-author-name",
			EnvVar: "DRONE_PREV_COMMIT_AUTHOR_NAME,WOODPECKER_PREV_COMMIT_AUTHOR_NAME",
		},
		cli.StringFlag{
			Name:   "prev-commit-author-avatar",
			EnvVar: "DRONE_PREV_COMMIT_AUTHOR_AVATAR,WOODPECKER_PREV_COMMIT_AUTHOR_AVATAR",
		},
		cli.StringFlag{
			Name:   "prev-commit-author-email",
			EnvVar: "DRONE_PREV_COMMIT_AUTHOR_EMAIL,WOODPECKER_PREV_COMMIT_AUTHOR_EMAIL",
		},
		cli.IntFlag{
			Name:   "job-number",
			EnvVar: "DRONE_JOB_NUMBER,WOODPECKER_JOB_NUMBER",
		},
		cli.StringSliceFlag{
			Name:   "env, e",
			EnvVar: "DRONE_ENV,WOODPECKER_ENV",
		},
	},
}

func exec(c *cli.Context) error {
	file := c.Args().First()
	if file == "" {
		file = ".drone.yml"
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
		file = ".drone.yml"
	}

	metadata := metadataFromContext(c)
	environ := metadata.Environ()
	secrets := []compiler.Secret{}
	for k, v := range metadata.EnvironDrone() {
		environ[k] = v
	}
	for key, val := range metadata.Job.Matrix {
		environ[key] = val
		secrets = append(secrets, compiler.Secret{
			Name:  key,
			Value: val,
		})
	}

	if len(axis) != 0 {
		for k, v := range axis {
			environ[k] = v
		}
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
func metadataFromContext(c *cli.Context) frontend.Metadata {
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
			Matrix: availableEnvironment(),
		},
		Sys: frontend.System{
			Name: c.String("system-name"),
			Link: c.String("system-link"),
			Arch: c.String("system-arch"),
		},
	}
}

func availableEnvironment() map[string]string {
	result := make(map[string]string, 0)

	for _, env := range os.Environ() {
		pair := strings.SplitN(env, "=", 2)
		result[pair[0]] = pair[1]
	}

	return result
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
