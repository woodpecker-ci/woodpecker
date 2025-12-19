package integration_test

import (
	"testing"

	"go.woodpecker-ci.org/woodpecker/v3/test/integration/utils"
)

// TestBasicPipelineExecution tests that a simple pipeline can be executed end-to-end
// This is the foundational integration test that verifies:
// 1. Server starts and is accessible
// 2. Agent connects to server
// 3. A simple "hello world" pipeline can be triggered and executed
func TestBasicPipelineExecution(t *testing.T) {
	t.Parallel()

	// Start the woodpecker server
	server, err := utils.StartServer(t)
	if err != nil {
		t.Fatalf("Could not start server: %s", err)
	}
	defer server.Stop()

	// Start the woodpecker agent
	agent, err := utils.StartAgent(t)
	if err != nil {
		t.Fatalf("Could not start agent: %s", err)
	}
	defer agent.Stop()

	// TODO: check server api if agent is connected

	// TODO: Once forge integration is working:
	// 1. Create a test repository with a simple .woodpecker.yml
	// 2. Register the repository with Woodpecker
	// 3. Trigger a pipeline (e.g., via webhook or manual trigger)
	// 4. Wait for pipeline to complete
	// 5. Verify pipeline succeeded

	t.Log("✓ Server and agent started successfully")
	t.Log("Next steps: Add repository creation, pipeline trigger, and result verification")
}
