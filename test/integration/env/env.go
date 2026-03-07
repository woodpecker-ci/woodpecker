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

package env

import (
	"context"
	"testing"
	"time"

	"go.woodpecker-ci.org/woodpecker/v3/test/integration/utils"
)

// TestEnv represents the complete integration test environment
// with all necessary components (Forge, Woodpecker Server, Woodpecker Agent)
type TestEnv struct {
	t      *testing.T
	ctx    context.Context
	cancel context.CancelFunc

	projectRoot string

	// Components
	Forge  *TestForge
	Server *TestServer
	Agent  *utils.Service

	// API Clients
	GiteaClient      *GiteaClient
	WoodpeckerClient *utils.WoodpeckerClient
}

func SetupTestEnv(t *testing.T) *TestEnv {
	t.Helper()

	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)

	env := &TestEnv{
		t:      t,
		ctx:    ctx,
		cancel: cancel,
	}

	return env
}

func (e *TestEnv) Start() {
	t := e.t
	t.Helper()

	t.Log("🚀 Setting up integration test environment...")

	// Step 1: Start Forge (Gitea)
	e.Forge = NewTestForge()

	err := e.Forge.Start(t)
	if err != nil {
		e.Stop()
		t.Fatalf("Failed to start forge: %v", err)
	}

	giteaClient := "test-client"
	giteaClientSecret := "test-secret"

	e.GiteaClient = NewGiteaClient(e.Forge.URL, e.Forge.AdminToken)

	// Step 2: Start Woodpecker Server
	t.Log("  🔧 Starting Woodpecker Server...")
	e.Server = &TestServer{
		URL: "http://localhost:8000",
	}
	err = e.Server.Start(t, e.Forge.URL, giteaClient, giteaClientSecret)
	if err != nil {
		e.Stop()
		t.Fatalf("Failed to start Woodpecker Server: %v", err)
	}

	// woodpeckerURL := "http://localhost:8000"
	woodpeckerGRPC_URL := "http://localhost:9000"
	woodpeckerAgentToken := "woodpecker-agent-token"

	// Step 3: Start Woodpecker Agent with dummy backend
	err = e.StartAgent(woodpeckerGRPC_URL, woodpeckerAgentToken)
	if err != nil {
		e.Stop()
		t.Fatalf("Failed to start Woodpecker Agent: %v", err)
	}

	t.Log("✅ Integration test environment setup complete!")
}

func (e *TestEnv) Stop() {
	t := e.t

	t.Helper()
	t.Log("🧹 Cleaning up test environment...")

	if e.cancel != nil {
		e.cancel()
	}

	if e.Agent != nil {
		e.Agent.Stop()
	}

	if e.Server != nil {
		e.Server.Stop()
	}

	if e.Forge != nil {
		e.Forge.Stop()
	}

	t.Log("✓ Cleanup complete")
}
