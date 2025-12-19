package utils

import (
	"os"
	"path/filepath"
	"testing"
)

// TestRepoConfig holds configuration for creating a test repository
type TestRepoConfig struct {
	Name           string
	PipelineConfig string
}

// CreateTestRepo creates a test repository with a Woodpecker pipeline configuration
func CreateTestRepo(t *testing.T, config TestRepoConfig) string {
	// Create temporary directory for the repo
	repoDir := t.TempDir()

	// Create .woodpecker.yml with the provided config
	woodpeckerYml := filepath.Join(repoDir, ".woodpecker.yml")
	if err := os.WriteFile(woodpeckerYml, []byte(config.PipelineConfig), 0644); err != nil {
		t.Fatalf("Failed to create .woodpecker.yml: %v", err)
	}

	// Initialize git repository
	NewTask("git", "init").WorkDir(repoDir).RunOrFail(t)
	NewTask("git", "config", "user.name", "Test User").WorkDir(repoDir).RunOrFail(t)
	NewTask("git", "config", "user.email", "test@example.com").WorkDir(repoDir).RunOrFail(t)
	NewTask("git", "add", ".").WorkDir(repoDir).RunOrFail(t)
	NewTask("git", "commit", "-m", "Initial commit").WorkDir(repoDir).RunOrFail(t)

	t.Logf("Created test repository at: %s", repoDir)
	return repoDir
}

// SimplePipelineConfig returns a basic pipeline configuration for testing
func SimplePipelineConfig() string {
	return `
when:
  - event: push
    branch: main

steps:
  - name: greeting
    image: alpine:latest
    commands:
      - echo "Hello from Woodpecker!"
      - echo "Pipeline is working correctly"
`
}

// MultiStepPipelineConfig returns a pipeline with multiple steps
func MultiStepPipelineConfig() string {
	return `
when:
  - event: push
    branch: main

steps:
  - name: step1
    image: alpine:latest
    commands:
      - echo "Step 1: Starting"
      - sleep 1

  - name: step2
    image: alpine:latest
    commands:
      - echo "Step 2: Running"
      - sleep 1

  - name: step3
    image: alpine:latest
    commands:
      - echo "Step 3: Completed"
`
}

// FailingPipelineConfig returns a pipeline that will fail
func FailingPipelineConfig() string {
	return `
when:
  - event: push
    branch: main

steps:
  - name: will-fail
    image: alpine:latest
    commands:
      - echo "This step will fail"
      - exit 1
`
}
