package integration_test

import (
	"testing"

	"go.woodpecker-ci.org/woodpecker/v3/test/integration/utils"
)

func TestEnvStart(t *testing.T) {
	forge, err := utils.StartForge(t)
	if err != nil {
		t.Fatalf("Could not start forge %s", err)
	}
	defer forge.Stop()

	server, err := utils.StartServer(t)
	if err != nil {
		t.Fatalf("Could not start server %s", err)
	}
	defer server.Stop()

	agent, err := utils.StartAgent(t)
	if err != nil {
		t.Fatalf("Could not start agent %s", err)
	}
	defer agent.Stop()
}
