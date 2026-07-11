package env

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// GiteaClient provides methods to interact with Gitea API
type GiteaClient struct {
	baseURL string
	token   string
	client  *http.Client
}

// NewGiteaClient creates a new Gitea API client
func NewGiteaClient(baseURL, token string) *GiteaClient {
	return &GiteaClient{
		baseURL: baseURL,
		token:   token,
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// doRequest performs an HTTP request to Gitea API
func (c *GiteaClient) doRequest(method, path string, body any) (*http.Response, error) {
	var bodyReader io.Reader
	if body != nil {
		jsonData, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal request body: %w", err)
		}
		bodyReader = bytes.NewBuffer(jsonData)
	}

	req, err := http.NewRequest(method, c.baseURL+path, bodyReader)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	if c.token != "" {
		req.Header.Set("Authorization", "token "+c.token)
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	return c.client.Do(req)
}

// CreateRepository creates a new repository in Gitea
func (c *GiteaClient) CreateRepository(name, description string) (map[string]any, error) {
	body := map[string]any{
		"name":        name,
		"description": description,
		"private":     false,
		"auto_init":   true, // Initialize with README
	}

	resp, err := c.doRequest("POST", "/api/v1/user/repos", body)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		respBody, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("failed to create repository: status %d, body: %s", resp.StatusCode, string(respBody))
	}

	var repo map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&repo); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return repo, nil
}

// CreateFile creates or updates a file in a repository
func (c *GiteaClient) CreateFile(owner, repo, filepath, content, message string) error {
	body := map[string]any{
		"content": content, // Should be base64 encoded
		"message": message,
	}

	path := fmt.Sprintf("/api/v1/repos/%s/%s/contents/%s", owner, repo, filepath)
	resp, err := c.doRequest("POST", path, body)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		respBody, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("failed to create file: status %d, body: %s", resp.StatusCode, string(respBody))
	}

	return nil
}

// TriggerWebhook simulates a push webhook from Gitea to Woodpecker
func (c *GiteaClient) TriggerWebhook(webhookURL string, payload map[string]any) error {
	jsonData, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal webhook payload: %w", err)
	}

	req, err := http.NewRequest("POST", webhookURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create webhook request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Gitea-Event", "push")

	resp, err := c.client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send webhook: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		respBody, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("webhook returned error: status %d, body: %s", resp.StatusCode, string(respBody))
	}

	return nil
}

// CreateWebhook creates a webhook in a Gitea repository
func (c *GiteaClient) CreateWebhook(owner, repo, webhookURL string) error {
	body := map[string]any{
		"type":   "gitea",
		"active": true,
		"config": map[string]string{
			"url":          webhookURL,
			"content_type": "json",
		},
		"events": []string{"push", "pull_request", "create", "delete"},
	}

	path := fmt.Sprintf("/api/v1/repos/%s/%s/hooks", owner, repo)
	resp, err := c.doRequest("POST", path, body)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		respBody, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("failed to create webhook: status %d, body: %s", resp.StatusCode, string(respBody))
	}

	return nil
}

// GetRepository gets repository information
func (c *GiteaClient) GetRepository(owner, repo string) (map[string]any, error) {
	path := fmt.Sprintf("/api/v1/repos/%s/%s", owner, repo)
	resp, err := c.doRequest("GET", path, nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("failed to get repository: status %d, body: %s", resp.StatusCode, string(respBody))
	}

	var repository map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&repository); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return repository, nil
}
