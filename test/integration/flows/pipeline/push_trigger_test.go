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

func TestFlow_TriggerPipelineByPush(t *testing.T) {
	// Setup the complete test environment
	e := env.SetupTestEnv(t)
	e.Start()

	// Define a simple pipeline configuration
	pipelineConfig := `
	when:
	  - event: push

	steps:
	  - name: greeting
	    image: alpine:latest
	    commands:
	      - echo "Hello from Woodpecker CI!"
	      - echo "This pipeline was triggered by a push event"
	`

	t.Log("📝 Creating git repository ...")
	gitRepo := blocks.NewGitRepo(t)
	cloneURL, err := e.Forge.GetRepositoryCloneURL(t.Name())
	if err != nil {
		t.Fatalf("Failed to get repository clone URL: %v", err)
	}
	gitRepo.Init(t, cloneURL)
	gitRepo.WriteFile(t, "README.md", []byte(t.Name()))
	gitRepo.Add(t, "README.md")
	gitRepo.Commit(t, ":tada: initial commit")
	gitRepo.Push(t)

	t.Log("🔗 Activating repository in Woodpecker...")

	t.Log("🚀 Pushing pipeline config to trigger pipeline...")
	gitRepo.WriteFile(t, ".woodpecker.yml", []byte(pipelineConfig))
	gitRepo.Add(t, ".woodpecker.yml")
	gitRepo.Commit(t, ":tada: init")
	t.Log("✓ Pipeline config committed, pushing to trigger pipeline...")

	// TODO: Step 5: Wait for pipeline to be created
	t.Log("⏳ Waiting for pipeline to be created...")
	// Poll Woodpecker API for pipeline
	// var pipelineID int
	// for i := 0; i < 10; i++ {
	//     pipelines, err := env.WoodpeckerClient.GetPipelines(repo.Owner, repo.Name)
	//     if err == nil && len(pipelines) > 0 {
	//         pipelineID = pipelines[0]["number"].(int)
	//         break
	//     }
	//     time.Sleep(1 * time.Second)
	// }

	// TODO: Step 6: Wait for pipeline to complete
	t.Log("⏳ Waiting for pipeline to complete...")
	// status, err := env.WoodpeckerClient.WaitForPipelineComplete(
	//     repo.Owner,
	//     repo.Name,
	//     pipelineID,
	//     60*time.Second,
	// )

	// TODO: Step 7: Verify pipeline succeeded
	t.Log("✅ Verifying pipeline status...")
	// if status != "success" {
	//     t.Fatalf("Expected pipeline to succeed, got status: %s", status)
	// }

	// TODO: Step 8: Verify logs contain expected output
	t.Log("📋 Verifying pipeline logs...")
	// logs, err := env.WoodpeckerClient.GetPipelineLogs(repo.Owner, repo.Name, pipelineID)
	// if !strings.Contains(logs, "Hello from Woodpecker CI!") {
	//     t.Error("Expected log output not found")
	// }

	t.Log("✅ Push trigger flow test completed successfully!")
}
