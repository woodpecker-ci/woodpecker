package utils

import (
	"fmt"
	"os"
	"os/exec"
	"time"
)

type Service struct {
	cmd *exec.Cmd
	env map[string]string
}

func NewService(cmdName string, args ...string) *Service {
	cmd := exec.Command(cmdName, args...)
	return &Service{
		cmd: cmd,
		env: make(map[string]string),
	}
}

func (s *Service) Start() error {
	// Inherit parent environment
	env := os.Environ()
	for key, value := range s.env {
		env = append(env, fmt.Sprintf("%s=%s", key, value))
	}
	s.cmd.Env = env
	// Capture output for debugging
	s.cmd.Stdout = os.Stdout
	s.cmd.Stderr = os.Stderr
	return s.cmd.Start()
}

func (s *Service) Stop() error {
	if s.cmd == nil || s.cmd.Process == nil {
		return fmt.Errorf("process not found / running")
	}

	// Try graceful shutdown first
	if err := s.cmd.Process.Signal(os.Interrupt); err != nil {
		// If interrupt fails, force kill
		return s.cmd.Process.Kill()
	}

	// Wait for graceful shutdown with timeout
	done := make(chan error, 1)
	go func() {
		done <- s.cmd.Wait()
	}()

	select {
	case <-time.After(5 * time.Second):
		// Force kill after timeout
		return s.cmd.Process.Kill()
	case err := <-done:
		return err
	}
}

func (s *Service) SetEnv(key, value string) *Service {
	s.env[key] = value
	return s
}

func (s *Service) WorkDir(workDir string) *Service {
	s.cmd.Dir = workDir
	return s
}
