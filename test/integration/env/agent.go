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
	"fmt"

	"go.woodpecker-ci.org/woodpecker/v3/test/integration/utils"
)

func (e *TestEnv) StartAgent(serverURL, agentToken string) error {
	t := e.t

	if e.Agent != nil {
		return fmt.Errorf("agent already started")
	}

	t.Log("  🤖 Starting Woodpecker Agent with mock backend...")

	service := utils.NewService("go", "run", "./cmd/agent/").
		WorkDir(e.projectRoot).
		// Agent configuration
		SetEnv("WOODPECKER_SERVER", serverURL).
		SetEnv("WOODPECKER_AGENT_SECRET", agentToken).
		// SetEnv("WOODPECKER_MAX_WORKFLOWS", "1").
		// SetEnv("WOODPECKER_HEALTHCHECK", "false").
		SetEnv("WOODPECKER_BACKEND", "dummy").
		// Log level
		SetEnv("WOODPECKER_LOG_LEVEL", "debug")

	if err := service.Start(); err != nil {
		return fmt.Errorf("failed to start agent: %w", err)
	}

	t.Cleanup(e.StopAgent)

	e.Agent = service

	// TODO: wait for agent to be ready
	// if err := utils.WaitForHTTP("http://localhost:3000", 30*time.Second); err != nil {
	// 	return fmt.Errorf("forge did not become ready: %w", err)
	// }

	t.Logf("  ✓ Woodpecker Agent started successfully")
	return nil
}

func (e *TestEnv) StopAgent() {
	t := e.t
	if e.Agent != nil {
		if err := e.Agent.Stop(); err != nil {
			t.Errorf("Warning: Failed to stop agent: %v", err)
		} else {
			t.Logf("Woodpecker agent stopped successfully")
		}
	}
}
