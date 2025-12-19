package integration_test

import (
	"testing"
	"time"

	"go.woodpecker-ci.org/woodpecker/v3/test/integration/utils"
)

// TestCompleteWorkflow demonstrates a full end-to-end workflow
// NOTE: This test requires Gitea to be running and properly configured
// Uncomment and use once forge integration is complete
func TestCompleteWorkflow(t *testing.T) {
	t.Skip("Skipping until forge integration is complete")

	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// Step 1: Start all services
	t.Log("Starting Gitea forge...")
	forge, err := utils.StartForge(t)
	if err != nil {
		t.Fatalf("Could not start forge: %s", err)
	}
	defer forge.Stop()

	t.Log("Starting Woodpecker server...")
	server, err := utils.StartServer(t)
	if err != nil {
		t.Fatalf("Could not start server: %s", err)
	}
	defer server.Stop()

	t.Log("Starting Woodpecker agent...")
	agent, err := utils.StartAgent(t)
	if err != nil {
		t.Fatalf("Could not start agent: %s", err)
	}
	defer agent.Stop()

	time.Sleep(5 * time.Second)
	t.Log("✓ All services started")

	// Step 2: Create and configure test repository
	config := utils.TestRepoConfig{
		Name:           "test-pipeline-repo",
		PipelineConfig: utils.MultiStepPipelineConfig(),
	}
	repoPath := utils.CreateTestRepo(t, config)
	t.Logf("✓ Created test repository at: %s", repoPath)

	// Step 3: Push to Gitea
	// TODO: Implement pushing to Gitea
	// This would involve:
	// - Creating a repository in Gitea via API
	// - Adding Gitea as a remote
	// - Pushing the repository
	t.Log("TODO: Push repository to Gitea")

	// Step 4: Activate repository in Woodpecker
	client := utils.NewWoodpeckerClient("http://localhost:8000", "your-token-here")

	owner := "woodpecker"
	repoName := "test-pipeline-repo"

	err = client.ActivateRepo(owner, repoName)
	if err != nil {
		t.Fatalf("Failed to activate repository: %v", err)
	}
	t.Log("✓ Repository activated in Woodpecker")

	// Step 5: Trigger a pipeline
	pipeline, err := client.TriggerPipeline(owner, repoName, "main")
	if err != nil {
		t.Fatalf("Failed to trigger pipeline: %v", err)
	}

	pipelineID := int(pipeline["id"].(float64))
	t.Logf("✓ Pipeline #%d triggered", pipelineID)

	// Step 6: Wait for pipeline to complete
	status, err := client.WaitForPipelineComplete(owner, repoName, pipelineID, 5*time.Minute)
	if err != nil {
		t.Fatalf("Pipeline did not complete: %v", err)
	}

	// Step 7: Verify success
	if status != "success" {
		t.Fatalf("Pipeline failed with status: %s", status)
	}

	t.Logf("✓ Pipeline completed successfully!")
	t.Log("✓ Complete workflow test passed!")
}

// TestPipelineFailure tests that pipeline failures are handled correctly
func TestPipelineFailure(t *testing.T) {
	t.Skip("Skipping until forge integration is complete")

	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// Start services (abbreviated)
	server, _ := utils.StartServer(t)
	defer server.Stop()

	agent, _ := utils.StartAgent(t)
	defer agent.Stop()

	time.Sleep(3 * time.Second)

	// Create a repository with a failing pipeline
	config := utils.TestRepoConfig{
		Name:           "failing-pipeline",
		PipelineConfig: utils.FailingPipelineConfig(),
	}
	utils.CreateTestRepo(t, config)

	// TODO: Complete the test once forge integration is ready
	// 1. Push to Gitea
	// 2. Activate repo
	// 3. Trigger pipeline
	// 4. Verify it fails with expected status
	// 5. Check error logs

	t.Log("TODO: Complete pipeline failure test")
}

// TestMultipleConcurrentPipelines tests handling of multiple pipelines
func TestMultipleConcurrentPipelines(t *testing.T) {
	t.Skip("Skipping until basic workflow is stable")

	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// This test would verify:
	// 1. Multiple repositories can be activated
	// 2. Pipelines from different repos can run concurrently
	// 3. Agent properly handles workflow queue
	// 4. Results are correctly associated with respective repos

	t.Log("TODO: Implement concurrent pipeline test")
}
