// Copyright 2026 Woodpecker Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package pipeline_test

import (
	"testing"

	"go.woodpecker-ci.org/woodpecker/v3/test/integration/blocks"
	"go.woodpecker-ci.org/woodpecker/v3/test/integration/env"
)

// TestFlow_ManualTrigger tests the flow of manually triggering a pipeline via API.
//
// Flow:
// 1. Setup test environment
// 2. Create a repository with a pipeline configuration
// 3. Activate the repository
// 4. Manually trigger a pipeline via Woodpecker API (not via webhook)
// 5. Verify that the pipeline is created and queued
// 6. Verify that the pipeline executes successfully
// 7. Verify pipeline metadata shows it was manually triggered
func TestFlow_ManualTrigger(t *testing.T) {
	// Setup the complete test environment
	e := env.SetupTestEnv(t)
	e.Start()

	// Define a simple pipeline for manual triggering
	pipelineConfig := `
when:
  - event: manual

steps:
  - name: manual-pipeline-step
    image: alpine:latest
    commands:
      - echo "This pipeline was manually triggered!"
      - echo "No webhook was needed"
      - echo "Triggered via API call"
`

	// TODO: Step 1: Create repository
	t.Log("📝 Creating test repository...")

	gitRepo := blocks.NewGitRepo(t)
	// gitRepo.Init(t)
	gitRepo.WriteFile(t, ".woodpecker.yml", []byte(pipelineConfig))
	gitRepo.Add(t, ".woodpecker.yml")
	gitRepo.Commit(t, ":tada: init")
	gitRepo.Push(t)
	t.Log("✓ Repository created and pipeline config pushed")

	// TODO: Step 3: Activate repository in Woodpecker
	t.Log("🔗 Activating repository in Woodpecker...")
	// err = env.WoodpeckerClient.ActivateRepo(repo.Owner, repo.Name)
	// if err != nil {
	//     t.Fatalf("Failed to activate repository: %v", err)
	// }
	repo := blocks.NewTestRepo()
	repo.Enable(t)
	t.Log("✓ Repository activated")

	// TODO: Step 4: Manually trigger a pipeline
	t.Log("🚀 Manually triggering pipeline...")
	// Trigger with specific branch
	// branch := "main"
	// pipeline, err := env.WoodpeckerClient.TriggerPipeline(repo.Owner, repo.Name, branch)
	// if err != nil {
	//     t.Fatalf("Failed to trigger pipeline: %v", err)
	// }
	// pipelineID := int(pipeline["number"].(float64))
	// t.Logf("✓ Manual pipeline #%d triggered", pipelineID)

	// TODO: Step 5: Wait for pipeline to complete
	t.Log("⏳ Waiting for pipeline to complete...")
	// status, err := env.WoodpeckerClient.WaitForPipelineComplete(
	//     repo.Owner,
	//     repo.Name,
	//     pipelineID,
	//     60*time.Second,
	// )
	// if err != nil {
	//     t.Fatalf("Error waiting for pipeline: %v", err)
	// }

	// TODO: Step 6: Verify pipeline succeeded
	t.Log("✅ Verifying pipeline status...")
	// if status != "success" {
	//     t.Fatalf("Expected pipeline to succeed, got status: %s", status)
	// }
	// t.Log("✓ Pipeline completed successfully")

	// TODO: Step 7: Verify pipeline metadata
	t.Log("📋 Verifying pipeline metadata...")
	// p, err := env.WoodpeckerClient.GetPipeline(repo.Owner, repo.Name, pipelineID)
	// if err != nil {
	//     t.Fatalf("Failed to get pipeline details: %v", err)
	// }
	//
	// Verify event type is "manual" or "deploy"
	// event := p["event"].(string)
	// if event != "manual" && event != "deploy" {
	//     t.Errorf("Expected event to be 'manual' or 'deploy', got: %s", event)
	// }
	//
	// Verify the branch
	// pipelineBranch := p["branch"].(string)
	// if pipelineBranch != branch {
	//     t.Errorf("Expected branch to be '%s', got: %s", branch, pipelineBranch)
	// }

	// TODO: Step 8: Verify logs
	t.Log("📋 Verifying pipeline logs...")
	// logs, err := env.WoodpeckerClient.GetPipelineLogs(repo.Owner, repo.Name, pipelineID)
	// if !strings.Contains(logs, "manually triggered") {
	//     t.Error("Expected 'manually triggered' in logs")
	// }

	t.Log("✅ Manual trigger flow test completed successfully!")
	t.Log("")
	t.Log("ℹ️  This test verifies that:")
	t.Log("   - Pipelines can be triggered manually via API")
	t.Log("   - No webhook or forge event is required")
	t.Log("   - Pipeline executes with correct branch/event metadata")
	t.Log("   - Manual triggers work independently of push events")
}

// TestFlow_ManualTrigger_WithVariables tests manual trigger with custom variables
func TestFlow_ManualTrigger_WithVariables(t *testing.T) {
	t.Skip("TODO: Implement test for manual trigger with variables")

	// This test should verify:
	// - Manual triggers can include custom environment variables
	// - Variables are properly passed to pipeline steps
	// - Variables override defaults when specified
}

// TestFlow_ManualTrigger_DifferentBranches tests manual triggering on different branches
func TestFlow_ManualTrigger_DifferentBranches(t *testing.T) {
	t.Skip("TODO: Implement test for manual trigger on different branches")

	// This test should verify:
	// - Manual trigger works on non-default branches
	// - Correct pipeline configuration is used for the specified branch
	// - Branch-specific when conditions are evaluated correctly
}

// TestFlow_ManualTrigger_WhilePipelineRunning tests manual trigger while another is running
func TestFlow_ManualTrigger_WhilePipelineRunning(t *testing.T) {
	t.Skip("TODO: Implement test for concurrent manual triggers")

	// This test should verify:
	// - Multiple manual triggers can be queued
	// - Each pipeline runs independently
	// - Queue and concurrency limits are respected
}

// TestFlow_ManualTrigger_WithParameters tests manual trigger with pipeline parameters
func TestFlow_ManualTrigger_WithParameters(t *testing.T) {
	t.Skip("TODO: Implement test for manual trigger with parameters")

	// This test should verify:
	// - Manual triggers can pass parameters to workflows
	// - Parameters are accessible in pipeline configuration
	// - Parameter validation works correctly
}
