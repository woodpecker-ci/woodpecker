package utils

import (
	"log"
	"os/exec"
	"strings"
	"testing"
)

func runOrFail(t *testing.T, workingDir, cmdName string, args ...string) {
	log.Printf("# %s %s", cmdName, strings.Join(args, " "))

	cmd := exec.Command(cmdName, args...)
	cmd.Dir = workingDir
	output, err := cmd.Output()
	if err != nil {
		t.Fatalf("Failed to execute command '%s %s': %v\n%s", cmdName, strings.Join(args, " "), err, output)
	}
}

func StartForge(t *testing.T) (*Service, error) {
	runOrFail(t, "", "mkdir", "-p", "web/dist")
	runOrFail(t, "", "echo", "\"test\"", "web/dist/index.html")

	service := NewService("echo", "start", "forge")
	service.WorkDir("forge/")

	return service, service.Start()
}

func StartServer(_ *testing.T) (*Service, error) {
	service := NewService("go", "run", "./cmd/server/")

	return service, service.Start()
}

func StartAgent(_ *testing.T) (*Service, error) {
	service := NewService("go", "run", "./cmd/agent/")

	return service, service.Start()
}
