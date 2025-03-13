package utils

import (
	"fmt"
	"os/exec"
)

type Service struct {
	cmdName    string
	args       []string
	env        map[string]string
	workingDir string
	cmd        *exec.Cmd
}

func (s *Service) Start() error {
	s.cmd = exec.Command(s.cmdName, s.args...)
	s.cmd.Dir = s.workingDir
	env := []string{}
	for key, value := range s.env {
		env = append(env, fmt.Sprintf("%s=%s", key, value))
	}
	s.cmd.Env = env
	return s.cmd.Start()
}

func (s *Service) Stop() error {
	if s.cmd == nil || s.cmd.Process == nil {
		return fmt.Errorf("process not found / running")
	}

	return s.cmd.Process.Kill()
}

func (s *Service) SetEnv(key, value string) {
	s.env[key] = value
}

func (s *Service) WorkDir(workDir string) {
	s.workingDir = workDir
}

func NewService(cmdName string, args ...string) *Service {
	return &Service{
		cmdName: cmdName,
		args:    args,
	}
}
