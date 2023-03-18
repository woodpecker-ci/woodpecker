package ssh

import (
	"context"
	"io"
	"strings"

	"github.com/melbahja/goph"
	"github.com/urfave/cli/v2"

	"github.com/woodpecker-ci/woodpecker/pipeline/backend/common"
	"github.com/woodpecker-ci/woodpecker/pipeline/backend/types"
	"github.com/woodpecker-ci/woodpecker/shared/constant"
)

type ssh struct {
	cmd        *goph.Cmd
	output     io.ReadCloser
	client     *goph.Client
	workingdir string
}

type readCloser struct {
	io.Reader
}

func (c readCloser) Close() error {
	return nil
}

// New returns a new ssh Engine.
func New() types.Engine {
	return &ssh{}
}

func (e *ssh) Name() string {
	return "ssh"
}

func (e *ssh) IsAvailable(ctx context.Context) bool {
	c, ok := ctx.Value(types.CliContext).(*cli.Context)
	return ok && c.String("backend-ssh-address") != "" && c.String("backend-ssh-user") != "" && (c.String("backend-ssh-key") != "" || c.String("backend-ssh-password") != "")
}

func (e *ssh) Load(ctx context.Context) error {
	cmd, err := e.client.Command("/bin/env", "mktemp", "-d", "-p", "/tmp", "woodpecker-ssh-XXXXXXXXXX")
	if err != nil {
		return err
	}

	dir, err := cmd.Output()
	if err != nil {
		return err
	}

	e.workingdir = string(dir)
	c, ok := ctx.Value(types.CliContext).(*cli.Context)
	if !ok {
		return types.ErrNoCliContextFound
	}
	address := c.String("backend-ssh-address")
	user := c.String("backend-ssh-user")
	var auth goph.Auth
	if file := c.String("backend-ssh-key"); file != "" {
		keyAuth, err := goph.Key(file, c.String("backend-ssh-key-password"))
		if err != nil {
			return err
		}
		auth = append(auth, keyAuth...)
	}
	if password := c.String("backend-ssh-password"); password != "" {
		auth = append(auth, goph.Password(password)...)
	}
	client, err := goph.New(user, address, auth)
	if err != nil {
		return err
	}
	e.client = client
	return nil
}

// Setup the pipeline environment.
func (e *ssh) Setup(ctx context.Context, config *types.Config) error {
	return nil
}

// Exec the pipeline step.
func (e *ssh) Exec(ctx context.Context, step *types.Step) error {
	// Get environment variables
	var command []string
	for a, b := range step.Environment {
		if a != "HOME" && a != "SHELL" { // Don't override $HOME and $SHELL
			command = append(command, a+"="+b)
		}
	}

	if step.Image == constant.DefaultCloneImage {
		// Default clone step
		command = append(command, "CI_WORKSPACE="+e.workingdir+"/"+step.Environment["CI_REPO"])
		command = append(command, "plugin-git")
	} else {
		// Use "image name" as run command
		command = append(command, step.Image)
		command = append(command, "-c")

		// TODO: use commands directly
		script := common.GenerateScript(step.Commands)
		// Deleting the initial lines removes netrc support but adds compatibility for more shells like fish
		command = append(command, "cd "+e.workingdir+"/"+step.Environment["CI_REPO"]+" && "+script[strings.Index(script, "\n\n")+2:])
	}

	// Prepare command
	var err error
	e.cmd, err = e.client.CommandContext(ctx, "/bin/env", command...)
	if err != nil {
		return err
	}

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
	sftp, err := e.client.NewSftp()
	if err != nil {
		return err
	}

	return sftp.RemoveDirectory(e.workingdir)
}
