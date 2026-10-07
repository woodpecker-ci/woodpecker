package env

import (
	"fmt"
	"net/http"
	"path/filepath"
	"testing"
	"time"

	"go.woodpecker-ci.org/woodpecker/v3/test/integration/utils"
)

type TestServer struct {
	URL     string
	service *utils.Service
}

func (s *TestServer) Start(t *testing.T, giteaURL, giteaClient, giteaClientSecret string) error {
	if s.service != nil {
		return fmt.Errorf("server already started")
	}

	projectRoot := "."

	t.Log("  🔧 Starting Woodpecker Server...")

	// Prepare web dist directory
	utils.NewCommand("mkdir", "-p", filepath.Join(projectRoot, "web/dist")).RunOrFail(t)
	utils.NewCommand("sh", "-c", fmt.Sprintf("echo test > %s", filepath.Join(projectRoot, "web/dist/index.html"))).RunOrFail(t)

	s.service = utils.NewService("go", "run", "./cmd/server/").
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
		SetEnv("WOODPECKER_GITEA_URL", giteaURL).
		SetEnv("WOODPECKER_GITEA_CLIENT", giteaClient).
		SetEnv("WOODPECKER_GITEA_SECRET", giteaClientSecret).
		// Log level
		SetEnv("WOODPECKER_LOG_LEVEL", "debug")

	if err := s.service.Start(); err != nil {
		return fmt.Errorf("failed to start server: %w", err)
	}

	// Wait for server to be ready
	if err := utils.WaitForHTTP("http://localhost:8000/healthz", 30*time.Second); err != nil {
		return fmt.Errorf("server did not become ready: %w", err)
	}

	t.Logf("  ✓ Woodpecker server started successfully")
	return nil
}

// Simulate user login
func (s *TestServer) Login(code, state string) (string, error) {
	client := http.Client{Timeout: 10 * time.Second}
	resp, err := client.Get(fmt.Sprintf("http://localhost:8000/authorize?code=%s&state=%s", code, state))
	if err != nil {
		return "", fmt.Errorf("failed to perform login request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("login request failed with status: %s", resp.Status)
	}

	cookies := resp.Cookies()
	for _, cookie := range cookies {
		if cookie.Name == "user_sess" {
			return cookie.Value, nil
		}
	}

	return "", fmt.Errorf("user_sess cookie not found in login response")
}

func (s *TestServer) Stop() error {
	if s.service == nil {
		return nil
	}

	if err := s.service.Stop(); err != nil {
		return fmt.Errorf("failed to stop server: %w", err)
	}

	return nil
}
