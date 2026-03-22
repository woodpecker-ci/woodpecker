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

//go:build test

package env

import (
	"fmt"
	"net/url"
	"path/filepath"
	"testing"
	"time"

	"go.woodpecker-ci.org/woodpecker/v3/test/integration/utils"
)

type TestForge struct {
	URL           string
	AdminUser     string
	AdminPassword string
	AdminEmail    string
	AdminToken    string

	service *utils.Service
}

func NewTestForge() *TestForge {
	return &TestForge{
		URL:           "http://localhost:8000",
		AdminUser:     "woodpecker",
		AdminPassword: "woodpecker123",
		AdminEmail:    "woodpecker@localhost",
	}
}

func (f *TestForge) Start(t *testing.T) error {
	if f.service != nil {
		return fmt.Errorf("forge already started")
	}

	projectRoot := "."

	t.Log("  📦 Starting forge (Gitea) ...")

	composeFile := filepath.Join(projectRoot, "data", "gitea", "docker-compose.yml")
	f.service = utils.NewService("docker", "compose", "-f", composeFile, "up", "-d")
	if err := f.service.Start(); err != nil {
		return fmt.Errorf("failed to start forge: %w", err)
	}

	// Wait for Gitea to be ready
	if err := utils.WaitForHTTP("http://localhost:3000", 30*time.Second); err != nil {
		return fmt.Errorf("forge did not become ready: %w", err)
	}

	t.Log("  ✓ Forge started successfully")

	return nil
}

func (f *TestForge) SetupAdmin(t *testing.T) error {
	utils.NewCommand("docker", "compose", "exec", "gitea",
		"gitea", "admin", "user", "create",
		"--username", f.AdminUser,
		"--password", f.AdminPassword,
		"--email", f.AdminEmail,
		"--admin",
	).RunOrFail(t)

	adminToken, err := utils.NewCommand("docker", "compose", "exec", "-T", "gitea",
		"gitea", "admin", "user", "generate-access-token",
		"-u", f.AdminUser,
		"--scopes", "write:repository,write:user",
		"--raw",
	).Run()
	if err != nil {
		return fmt.Errorf("failed to generate admin token: %w", err)
	}

	f.AdminToken = adminToken

	return nil
}

func (f *TestForge) SetupOAuthApp(t *testing.T, clientName, clientSecret string) error {
	appID, err := utils.NewCommand("docker", "compose", "exec", "-T", "gitea",
		"gitea", "admin", "oauth2", "add",
		"--name", clientName,
		"--redirect-uris", "http://localhost:8000/callback",
		"--client-secret", clientSecret,
	).Run()
	if err != nil {
		return fmt.Errorf("failed to create OAuth app: %w", err)
	}

	t.Logf("  ✓ OAuth app created with ID: %s", appID)
	return nil
}

func (f *TestForge) GetRepositoryCloneURL(repo string) (string, error) {
	u, err := url.Parse(f.URL)
	if err != nil {
		return "", fmt.Errorf("invalid forge URL: %w", err)
	}
	u.User = url.UserPassword(f.AdminUser, f.AdminPassword)

	return fmt.Sprintf("%s/%s.git", u.String(), repo), nil
}

func (f *TestForge) Stop() error {
	if f.service == nil {
		return nil
	}

	if err := f.service.Stop(); err != nil {
		return fmt.Errorf("failed to stop forge: %w", err)
	}

	f.service = nil
	return nil
}
