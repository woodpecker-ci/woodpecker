package utils

import (
	"fmt"
	"os/exec"
)

type Service struct {
	cmd *exec.Cmd
	env map[string]string
}

func NewService(cmdName string, args ...string) *Service {
	cmd := exec.Command(cmdName, args...)
	return &Service{
		cmd: cmd,
	}
}

func (s *Service) Start() error {
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

func (s *Service) SetEnv(key, value string) *Service {
	s.env[key] = value
	return s
}

func (s *Service) WorkDir(workDir string) *Service {
	s.cmd.Dir = workDir
	return s
}
