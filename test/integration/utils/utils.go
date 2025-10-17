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

	service := NewService("go", "run", "./cmd/server/").
		// some default settings
		SetEnv("WOODPECKER_OPEN", "true").
		SetEnv("WOODPECKER_ADMIN", "woodpecker").
		SetEnv("WOODPECKER_HOST", "http://host.docker.internal:8000").
		SetEnv("WOODPECKER_EXPERT_WEBHOOK_HOST", "http://host.docker.internal:8000").
		SetEnv("WOODPECKER_AGENT_SECRET", "1234").
		SetEnv("WOODPECKER_GITEA", "true").
		SetEnv("WOODPECKER_GITEA_URL", "true").
		SetEnv("WOODPECKER_GITEA_CLIENT", "123").
		SetEnv("WOODPECKER_GITEA_SECRET", "123").
		SetEnv("WOODPECKER_AGENT_SECRET", "1234")

	return service, service.Start()
}

func StartAgent(_ *testing.T) (*Service, error) {
	service := NewService("go", "run", "./cmd/agent/").
		// some default settings
		SetEnv("WOODPECKER_SERVER", "localhost:9000").
		SetEnv("WOODPECKER_AGENT_SECRET", "1234").
		SetEnv("WOODPECKER_MAX_WORKFLOWS", "1").
		SetEnv("WOODPECKER_HEALTHCHECK", "false")

	return service, service.Start()
}
