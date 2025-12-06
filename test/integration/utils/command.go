package utils

import (
	"fmt"
	"log"
	"os/exec"
	"strings"
	"testing"
)

type Command struct {
	cmd *exec.Cmd
	env map[string]string
}

func NewTask(cmdName string, args ...string) *Command {
	cmd := exec.Command(cmdName, args...)
	return &Command{
		cmd: cmd,
	}
}

func (t *Command) WorkDir(workDir string) *Command {
	t.cmd.Dir = workDir
	return t
}

func (t *Command) SetEnv(key, value string) *Command {
	t.env[key] = value
	return t
}

func (t *Command) Run() (string, error) {
	log.Printf("# %s %s", t.cmd.Path, strings.Join(t.cmd.Args, " "))
	env := []string{}
	for key, value := range t.env {
		env = append(env, fmt.Sprintf("%s=%s", key, value))
	}
	t.cmd.Env = env

	output, err := t.cmd.Output()
	return string(output), err
}

func (t *Command) RunOrFail(te *testing.T) {
	output, err := t.Run()
	if err != nil {
		te.Fatalf("Failed to execute command '%s %s': %v\n%s", t.cmd.Path, strings.Join(t.cmd.Args, " "), err, output)
	}
}
