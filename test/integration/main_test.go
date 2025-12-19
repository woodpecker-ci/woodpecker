package integration_test

import (
	"testing"
	"time"

	"go.woodpecker-ci.org/woodpecker/v3/test/integration/utils"
)

// TestEnvStart verifies that all components can start successfully
func TestEnvStart(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// Start server
	server, err := utils.StartServer(t)
	if err != nil {
		t.Fatalf("Could not start server: %s", err)
	}
	defer func() {
		if err := server.Stop(); err != nil {
			t.Logf("Warning: failed to stop server: %v", err)
		}
	}()

	// Start agent
	agent, err := utils.StartAgent(t)
	if err != nil {
		t.Fatalf("Could not start agent: %s", err)
	}
	defer func() {
		if err := agent.Stop(); err != nil {
			t.Logf("Warning: failed to stop agent: %v", err)
		}
	}()

	// Wait for services to stabilize
	time.Sleep(3 * time.Second)

	// Verify server health endpoint
	if err := utils.WaitForHTTP("http://localhost:8000/healthz", 5*time.Second); err != nil {
		t.Fatalf("Server health check failed: %v", err)
	}

	t.Log("✓ All components started successfully")
	t.Log("✓ Server is healthy and responding")
	t.Log("✓ Agent is running")
}
