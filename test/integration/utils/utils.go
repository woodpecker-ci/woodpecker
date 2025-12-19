package utils

import (
	"fmt"
	"path/filepath"
	"testing"
	"time"
)

func StartForge(t *testing.T) (*Service, error) {
	// Start Gitea using docker-compose
	// Use the existing docker-compose file
	projectRoot := getProjectRoot()
	composeFile := filepath.Join(projectRoot, "data", "gitea", "docker-compose.yml")

	service := NewService("docker-compose", "-f", composeFile, "up", "-d")

	if err := service.Start(); err != nil {
		return nil, fmt.Errorf("failed to start forge: %w", err)
	}

	// Wait for Gitea to be ready
	if err := WaitForHTTP("http://localhost:3000", 30*time.Second); err != nil {
		return nil, fmt.Errorf("forge did not become ready: %w", err)
	}

	t.Logf("Forge (Gitea) started successfully")
	return service, nil
}

func StartServer(t *testing.T) (*Service, error) {
	projectRoot := getProjectRoot()

	// Prepare web dist directory
	NewTask("mkdir", "-p", filepath.Join(projectRoot, "web/dist")).RunOrFail(t)
	NewTask("sh", "-c", fmt.Sprintf("echo test > %s", filepath.Join(projectRoot, "web/dist/index.html"))).RunOrFail(t)

	service := NewService("go", "run", "./cmd/server/").
		WorkDir(projectRoot).
		// Server configuration
		SetEnv("WOODPECKER_OPEN", "true").
		SetEnv("WOODPECKER_ADMIN", "woodpecker").
		SetEnv("WOODPECKER_HOST", "http://localhost:8000").
		SetEnv("WOODPECKER_SERVER_ADDR", ":8000").
		SetEnv("WOODPECKER_GRPC_ADDR", ":9000").
		SetEnv("WOODPECKER_WEBHOOK_HOST", "http://localhost:8000").
		SetEnv("WOODPECKER_AGENT_SECRET", "test-secret-123").
		// Gitea forge configuration
		SetEnv("WOODPECKER_GITEA", "true").
		SetEnv("WOODPECKER_GITEA_URL", "http://localhost:3000").
		SetEnv("WOODPECKER_GITEA_CLIENT", "test-client").
		SetEnv("WOODPECKER_GITEA_SECRET", "test-secret").
		// Log level
		SetEnv("WOODPECKER_LOG_LEVEL", "debug")

	if err := service.Start(); err != nil {
		return nil, fmt.Errorf("failed to start server: %w", err)
	}

	// Wait for server to be ready
	if err := WaitForHTTP("http://localhost:8000/healthz", 30*time.Second); err != nil {
		return nil, fmt.Errorf("server did not become ready: %w", err)
	}

	t.Logf("Woodpecker server started successfully")
	return service, nil
}

func StartAgent(t *testing.T) (*Service, error) {
	projectRoot := getProjectRoot()

	service := NewService("go", "run", "./cmd/agent/").
		WorkDir(projectRoot).
		// Agent configuration
		SetEnv("WOODPECKER_SERVER", "localhost:9000").
		SetEnv("WOODPECKER_AGENT_SECRET", "test-secret-123").
		SetEnv("WOODPECKER_MAX_WORKFLOWS", "1").
		SetEnv("WOODPECKER_HEALTHCHECK", "false").
		SetEnv("WOODPECKER_BACKEND", "docker").
		// Log level
		SetEnv("WOODPECKER_LOG_LEVEL", "debug")

	if err := service.Start(); err != nil {
		return nil, fmt.Errorf("failed to start agent: %w", err)
	}

	// Give agent time to connect
	time.Sleep(2 * time.Second)

	t.Logf("Woodpecker agent started successfully")
	return service, nil
}

func getProjectRoot() string {
	// This assumes tests run from test/integration/
	return "../.."
}
