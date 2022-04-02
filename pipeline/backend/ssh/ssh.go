package ssh

import (
	"context"
	"encoding/base64"
	"io"
	"os"
	"strings"

	"github.com/melbahja/goph"

	"github.com/woodpecker-ci/woodpecker/pipeline/backend/types"
	"github.com/woodpecker-ci/woodpecker/server"
)

type ssh struct {
	cmd    *goph.Cmd
	output io.ReadCloser
	client *goph.Client
}

type readCloser struct {
	io.Reader
}

func (c readCloser) Close() error {
	return nil
}

// make sure local implements Engine
var _ types.Engine = &ssh{}

// New returns a new ssh Engine.
func New() types.Engine {
	return &ssh{}
}

func (e *ssh) Name() string {
	return "ssh"
}

func (e *ssh) IsAvailable() bool {
	return true
}

func (e *ssh) Load() error {
	var auth goph.Auth
	if file, has := os.LookupEnv("WOODPECKER_SSH_KEY"); has {
		var err error
		auth, err = goph.Key(file, os.Getenv("WOODPECKER_SSH_KEY_PASSWORD"))
		if err != nil {
			return err
		}
	} else {
		auth = goph.Password(os.Getenv("WOODPECKER_SSH_PASSWORD"))
	}
	client, err := goph.New(os.Getenv("WOODPECKER_SSH_USER"), os.Getenv("WOODPECKER_SSH_HOST"), auth)
	if err != nil {
		return err
	}
	e.client = client
	return nil
}

// Setup the pipeline environment.
func (e *ssh) Setup(ctx context.Context, proc *types.Config) error {
	return nil
}

// Exec the pipeline step.
func (e *ssh) Exec(ctx context.Context, proc *types.Step) error {
	// Get environment variables
	Command := []string{}
	for a, b := range proc.Environment {
		if a != "HOME" && a != "SHELL" { // Don't override $HOME and $SHELL
			Command = append(Command, a+"="+b)
		}
	}

	// Get default clone image
	defaultCloneImage := "docker.io/woodpeckerci/plugin-git:latest"
	if len(server.Config.Pipeline.DefaultCloneImage) > 0 {
		defaultCloneImage = server.Config.Pipeline.DefaultCloneImage
	}

	if proc.Image == defaultCloneImage {
		// Default clone step
		Command = append(Command, "CI_WORKSPACE=/tmp/woodpecker/"+proc.Environment["CI_REPO"])
		Command = append(Command, "plugin-git")
	} else {
		// Use "image name" as run command
		Command = append(Command, proc.Image[18:len(proc.Image)-7])
		Command = append(Command, "-c")

		// Decode script and delete initial lines
		// Deleting the initial lines removes netrc support but adds compatibility for more shells like fish
		Script, _ := base64.RawStdEncoding.DecodeString(proc.Environment["CI_SCRIPT"])
		Command = append(Command, string(Script)[strings.Index(string(Script), "\n\n")+2:])
	}

	// Prepare command
	var err error
	e.cmd, err = e.client.CommandContext(ctx, "/bin/env", Command...)
	if err != nil {
		return err
	}

	// Prepare working directory
	/*if proc.Image == defaultCloneImage {
		e.cmd.Dir = "/tmp/woodpecker/" + proc.Environment["CI_REPO_OWNER"]
	} else {
		e.cmd.Dir = "/tmp/woodpecker/" + proc.Environment["CI_REPO"]
	}*/

	// Get output and redirect Stderr to Stdout
	std, _ := e.cmd.StdoutPipe()
	e.output = readCloser{std}
	e.cmd.Stderr = e.cmd.Stdout

	return e.cmd.Start()
}

// Wait for the pipeline step to complete and returns
// the completion results.
func (e *ssh) Wait(context.Context, *types.Step) (*types.State, error) {
	return &types.State{
		Exited: true,
	}, e.cmd.Wait()
}

// Tail the pipeline step logs.
func (e *ssh) Tail(context.Context, *types.Step) (io.ReadCloser, error) {
	return e.output, nil
}

// Destroy the pipeline environment.
func (e *ssh) Destroy(context.Context, *types.Config) error {
	e.client.Close()
	// os.RemoveAll(e.cmd.Dir)
	return nil
}
