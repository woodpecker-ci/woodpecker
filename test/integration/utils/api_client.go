package utils

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// WoodpeckerClient provides methods to interact with Woodpecker API
type WoodpeckerClient struct {
	baseURL string
	token   string
	client  *http.Client
}

// NewWoodpeckerClient creates a new API client
func NewWoodpeckerClient(baseURL, token string) *WoodpeckerClient {
	return &WoodpeckerClient{
		baseURL: baseURL,
		token:   token,
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// doRequest performs an HTTP request with authentication
func (c *WoodpeckerClient) doRequest(method, path string, body interface{}) (*http.Response, error) {
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
		req.Header.Set("Authorization", "Bearer "+c.token)
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	return c.client.Do(req)
}

// GetRepos fetches the list of repositories
func (c *WoodpeckerClient) GetRepos() ([]map[string]interface{}, error) {
	resp, err := c.doRequest("GET", "/api/repos", nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("unexpected status code: %d, body: %s", resp.StatusCode, string(body))
	}

	var repos []map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&repos); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return repos, nil
}

// ActivateRepo activates a repository
func (c *WoodpeckerClient) ActivateRepo(owner, name string) error {
	path := fmt.Sprintf("/api/repos/%s/%s", owner, name)
	resp, err := c.doRequest("POST", path, nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("failed to activate repo: status %d, body: %s", resp.StatusCode, string(body))
	}

	return nil
}

// GetPipeline fetches a specific pipeline
func (c *WoodpeckerClient) GetPipeline(owner, name string, pipelineID int) (map[string]interface{}, error) {
	path := fmt.Sprintf("/api/repos/%s/%s/pipelines/%d", owner, name, pipelineID)
	resp, err := c.doRequest("GET", path, nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("unexpected status code: %d, body: %s", resp.StatusCode, string(body))
	}

	var pipeline map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&pipeline); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return pipeline, nil
}

// TriggerPipeline manually triggers a pipeline
func (c *WoodpeckerClient) TriggerPipeline(owner, name, branch string) (map[string]interface{}, error) {
	path := fmt.Sprintf("/api/repos/%s/%s/pipelines", owner, name)
	body := map[string]string{
		"branch": branch,
	}

	resp, err := c.doRequest("POST", path, body)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("failed to trigger pipeline: status %d, body: %s", resp.StatusCode, string(respBody))
	}

	var pipeline map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&pipeline); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return pipeline, nil
}

// WaitForPipelineComplete waits for a pipeline to complete (success or failure)
func (c *WoodpeckerClient) WaitForPipelineComplete(owner, name string, pipelineID int, timeout time.Duration) (string, error) {
	deadline := time.Now().Add(timeout)

	for time.Now().Before(deadline) {
		pipeline, err := c.GetPipeline(owner, name, pipelineID)
		if err != nil {
			return "", err
		}

		status, ok := pipeline["status"].(string)
		if !ok {
			return "", fmt.Errorf("pipeline status not found")
		}

		// Check if pipeline is in a terminal state
		switch status {
		case "success":
			return "success", nil
		case "failure", "error", "killed":
			return status, nil
		}

		time.Sleep(1 * time.Second)
	}

	return "", fmt.Errorf("timeout waiting for pipeline to complete")
}
