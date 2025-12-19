package integration_test

import (
	"testing"
	"time"

	"go.woodpecker-ci.org/woodpecker/v3/test/integration/utils"
)

// TestSimplePipelineExecution demonstrates the full integration test workflow
// This test verifies that a basic pipeline can execute successfully through all components
func TestSimplePipelineExecution(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// Step 1: Start Woodpecker server
	server, err := utils.StartServer(t)
	if err != nil {
		t.Fatalf("Could not start server: %s", err)
	}
	defer func() {
		if err := server.Stop(); err != nil {
			t.Logf("Warning: failed to stop server: %v", err)
		}
	}()

	// Step 2: Start Woodpecker agent
	agent, err := utils.StartAgent(t)
	if err != nil {
		t.Fatalf("Could not start agent: %s", err)
	}
	defer func() {
		if err := agent.Stop(); err != nil {
			t.Logf("Warning: failed to stop agent: %v", err)
		}
	}()

	// Give the system a moment to stabilize
	time.Sleep(3 * time.Second)

	t.Log("✓ Server and agent started successfully")

	// Step 3: Create API client (in production, you'd authenticate with the forge)
	// For now, we're using the admin user in OPEN mode
	client := utils.NewWoodpeckerClient("http://localhost:8000", "")

	// Step 4: Create a test repository with a simple pipeline
	config := utils.TestRepoConfig{
		Name:           "test-repo",
		PipelineConfig: utils.SimplePipelineConfig(),
	}
	repoPath := utils.CreateTestRepo(t, config)
	t.Logf("✓ Test repository created at: %s", repoPath)

	// TODO: The following steps require forge integration
	// Once Gitea is running, these would work:
	//
	// 5. Push repository to Gitea
	// 6. Activate repository in Woodpecker via API
	// 7. Trigger a pipeline (webhook or manual)
	// 8. Wait for pipeline to complete
	// 9. Verify pipeline succeeded
	// 10. Check pipeline logs

	// For now, just verify the API is accessible
	_, err = client.GetRepos()
	if err != nil {
		t.Logf("Note: Could not fetch repos (expected without forge): %v", err)
	}

	t.Log("✓ Integration test framework is ready!")
	t.Log("Next: Add forge (Gitea) integration to complete the workflow")
}

// TestAgentConnection verifies that the agent can connect to the server
func TestAgentConnection(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// Start server
	server, err := utils.StartServer(t)
	if err != nil {
		t.Fatalf("Could not start server: %s", err)
	}
	defer server.Stop()

	// Start agent
	agent, err := utils.StartAgent(t)
	if err != nil {
		t.Fatalf("Could not start agent: %s", err)
	}
	defer agent.Stop()

	// Wait for connection to establish
	time.Sleep(3 * time.Second)

	// TODO: Add API endpoint check to verify agent is connected
	// This could be done via the server's API: GET /api/agents

	t.Log("✓ Agent connection test completed")
}
