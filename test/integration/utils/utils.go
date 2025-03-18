package utils

import (
	"testing"
)

func StartForge(_ *testing.T) (*Service, error) {
	service := NewService("echo", "start", "forge")

	return service, service.Start()
}

func StartServer(t *testing.T) (*Service, error) {
	NewTask("mkdir", "-p", "web/dist").RunOrFail(t)
	NewTask("echo", "\"test\"", ">", "web/dist/index.html").RunOrFail(t)

	service := NewService("go", "run", "./cmd/server/")
	return service, service.Start()
}

func StartAgent(_ *testing.T) (*Service, error) {
	service := NewService("go", "run", "./cmd/agent/")
	return service, service.Start()
}
