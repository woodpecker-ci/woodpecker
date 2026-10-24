// Copyright 2026 Woodpecker Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package healing

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/rs/zerolog/log"

	"go.woodpecker-ci.org/woodpecker/v3/server"
	"go.woodpecker-ci.org/woodpecker/v3/server/model"
	"go.woodpecker-ci.org/woodpecker/v3/server/pipeline"
	"go.woodpecker-ci.org/woodpecker/v3/server/store"
)

// Classification represents the diagnosed category of a pipeline failure.
type Classification string

const (
	ClassTransient   Classification = "TRANSIENT"
	ClassCodeDefect  Classification = "CODE_DEFECT"
	ClassConfigDrift Classification = "CONFIG_DRIFT"
	ClassUnknown     Classification = "UNKNOWN"
)

// Diagnosis stores the result of an automated root-cause analysis.
type Diagnosis struct {
	Classification    Classification `json:"classification"`
	RootCause         string         `json:"root_cause"`
	Confidence        float64        `json:"confidence"`
	RecommendedAction string         `json:"recommended_action"`
}

// DiagnoseAndHeal is the main entry point called asynchronously upon pipeline failure.
func DiagnoseAndHeal(ctx context.Context, _store store.Store, repo *model.Repo, failedPipeline *model.Pipeline) {
	if !server.Config.SelfHealing.Enabled {
		return
	}

	logger := log.With().
		Str("repo", repo.FullName).
		Int64("pipeline", failedPipeline.Number).
		Int64("rerun_count", failedPipeline.RerunCount).
		Logger()

	if failedPipeline.RerunCount >= int64(server.Config.SelfHealing.MaxRetries) {
		logger.Warn().Msg("Self-healing: max retry limit reached, skipping auto-remediation")
		return
	}

	// 1. Locate the failed workflow and step
	failedStep, logExcerpt := findFailedStepAndLogs(_store, failedPipeline)
	if failedStep == nil {
		logger.Debug().Msg("Self-healing: could not identify specific failed step")
		return
	}

	logger.Info().
		Str("failed_step", failedStep.Name).
		Int("exit_code", failedStep.ExitCode).
		Msg("Self-healing: initiating automated root cause analysis")

	// 2. Perform diagnosis (fast pattern match + optional AI deep path)
	diagnosis := analyzeFailure(ctx, failedStep, logExcerpt)
	logger.Info().
		Str("classification", string(diagnosis.Classification)).
		Str("root_cause", diagnosis.RootCause).
		Float64("confidence", diagnosis.Confidence).
		Msg("Self-healing: diagnostic classification complete")

	// 3. Dispatch remediation action
	if diagnosis.Classification == ClassTransient && server.Config.SelfHealing.AutoRestartFlaky {
		repoUser, err := _store.GetUser(repo.UserID)
		if err != nil {
			logger.Error().Err(err).Msg("Self-healing: failed to retrieve repo owner for restart")
			return
		}

		logger.Warn().
			Str("root_cause", diagnosis.RootCause).
			Msg("Self-healing: transient failure detected, autonomously restarting pipeline")

		envs := map[string]string{
			"CI_SELF_HEALED":           "true",
			"CI_SELF_HEALED_REASON":    diagnosis.RootCause,
			"CI_SELF_HEALED_TIMESTAMP": fmt.Sprint(time.Now().Unix()),
		}

		restarted, err := pipeline.Restart(ctx, _store, failedPipeline, repoUser, repo, envs)
		if err != nil {
			logger.Error().Err(err).Msg("Self-healing: failed to trigger automated pipeline restart")
			return
		}

		logger.Info().
			Int64("new_pipeline_number", restarted.Number).
			Msg("Self-healing: pipeline successfully restarted with healing context")
	}
}

// findFailedStepAndLogs scans pipeline steps to find the step that failed and extracts its tail log.
func findFailedStepAndLogs(_store store.Store, p *model.Pipeline) (*model.Step, string) {
	steps, err := _store.StepList(p.ID)
	if err != nil || len(steps) == 0 {
		return nil, ""
	}

	var targetStep *model.Step
	for _, step := range steps {
		if step.State == model.StatusFailure {
			targetStep = step
			break
		}
	}

	if targetStep == nil {
		return nil, ""
	}

	// Fetch log entries
	logEntries, err := server.Config.Services.LogStore.LogFind(targetStep)
	if err != nil || len(logEntries) == 0 {
		return targetStep, ""
	}

	// Capture last 50 lines
	startIdx := 0
	if len(logEntries) > 50 {
		startIdx = len(logEntries) - 50
	}

	var sb strings.Builder
	for _, entry := range logEntries[startIdx:] {
		sb.Write(entry.Data)
		sb.WriteString("\n")
	}

	return targetStep, sb.String()
}

// analyzeFailure combines deterministic heuristics with the optional AI diagnostic endpoint.
func analyzeFailure(ctx context.Context, step *model.Step, logs string) Diagnosis {
	// Fast Path: Deterministic pattern matching on logs
	if diag, found := matchTransientPatterns(logs); found {
		return diag
	}

	// Deep Path: AI Endpoint if configured
	endpoint := server.Config.SelfHealing.Endpoint
	if endpoint != "" {
		if diag, err := queryAIDiagnostic(ctx, endpoint, step, logs); err == nil {
			return diag
		} else {
			log.Warn().Err(err).Msg("Self-healing: AI diagnostic query failed, falling back to heuristic")
		}
	}

	// Exit code 137 typically corresponds to external kill / OOM
	if step.ExitCode == 137 {
		return Diagnosis{
			Classification:    ClassTransient,
			RootCause:         "Container killed by host/runtime (Exit 137 / potential OOM)",
			Confidence:        0.85,
			RecommendedAction: "RESTART",
		}
	}

	return Diagnosis{
		Classification:    ClassCodeDefect,
		RootCause:         fmt.Sprintf("Step %s exited with non-zero status %d", step.Name, step.ExitCode),
		Confidence:        0.70,
		RecommendedAction: "MANUAL_INSPECTION",
	}
}

// matchTransientPatterns detects common infrastructure, network, and lock transients.
func matchTransientPatterns(logContent string) (Diagnosis, bool) {
	lower := strings.ToLower(logContent)

	transientKeywords := []struct {
		substr string
		cause  string
	}{
		{"connection refused", "Network connection refused"},
		{"i/o timeout", "I/O timeout during network transfer"},
		{"tls handshake timeout", "TLS handshake timeout"},
		{"temporary failure in name resolution", "DNS resolution transient"},
		{"network is unreachable", "Network temporarily unreachable"},
		{"context deadline exceeded", "Context deadline exceeded"},
		{"server misbehaving", "Upstream server misbehaving"},
		{"502 bad gateway", "Upstream HTTP 502 Bad Gateway"},
		{"503 service unavailable", "Upstream HTTP 503 Service Unavailable"},
		{"504 gateway timeout", "Upstream HTTP 504 Gateway Timeout"},
		{"database is locked", "SQLite/Database temporary lock contention"},
		{"deadlock detected", "Database deadlock detected"},
		{"too many open files", "System file descriptor exhaustion"},
	}

	for _, item := range transientKeywords {
		if strings.Contains(lower, item.substr) {
			return Diagnosis{
				Classification:    ClassTransient,
				RootCause:         item.cause,
				Confidence:        0.95,
				RecommendedAction: "RESTART",
			}, true
		}
	}

	return Diagnosis{}, false
}

// queryAIDiagnostic sends a structured request to a local LLM (vLLM/Ollama) or CAID service.
func queryAIDiagnostic(ctx context.Context, endpoint string, step *model.Step, logExcerpt string) (Diagnosis, error) {
	modelName := server.Config.SelfHealing.Model
	if modelName == "" {
		modelName = "qwen2.5-coder:32b"
	}

	var reqBody []byte
	var err error

	// If endpoint points to an OpenAI / Ollama compatible completions API
	if strings.Contains(endpoint, "chat/completions") || strings.Contains(endpoint, "api/chat") {
		prompt := fmt.Sprintf("Analyze this CI step failure:\nStep Name: %s\nExit Code: %d\nLog Excerpt:\n%s\n\nClassify root cause into TRANSIENT, CODE_DEFECT, or CONFIG_DRIFT.",
			step.Name, step.ExitCode, logExcerpt)

		payload := map[string]any{
			"model": modelName,
			"messages": []map[string]string{
				{
					"role":    "system",
					"content": "You are Symphonix HealthCI Autonomous Self-Healing Diagnostic Agent. Respond ONLY with valid JSON having the exact keys: classification (TRANSIENT, CODE_DEFECT, or CONFIG_DRIFT), root_cause (concise explanation), confidence (float between 0 and 1.0), recommended_action (RESTART, MANUAL_INSPECTION, or AUTO_FIX).",
				},
				{
					"role":    "user",
					"content": prompt,
				},
			},
			"temperature": 0.1,
		}
		reqBody, err = json.Marshal(payload)
	} else {
		// Native direct diagnostic endpoint payload
		payload := map[string]any{
			"model":       modelName,
			"step_name":   step.Name,
			"exit_code":   step.ExitCode,
			"log_excerpt": logExcerpt,
		}
		reqBody, err = json.Marshal(payload)
	}

	if err != nil {
		return Diagnosis{}, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(reqBody))
	if err != nil {
		return Diagnosis{}, err
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return Diagnosis{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return Diagnosis{}, fmt.Errorf("ai diagnostic endpoint returned status %d", resp.StatusCode)
	}

	var rawResp map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&rawResp); err != nil {
		return Diagnosis{}, err
	}

	// Case A: OpenAI / Ollama chat completion format
	if choices, ok := rawResp["choices"].([]any); ok && len(choices) > 0 {
		if firstChoice, ok := choices[0].(map[string]any); ok {
			if msg, ok := firstChoice["message"].(map[string]any); ok {
				if content, ok := msg["content"].(string); ok {
					var diag Diagnosis
					// Strip potential markdown code blocks ```json ... ```
					cleanedContent := strings.TrimSpace(content)
					cleanedContent = strings.TrimPrefix(cleanedContent, "```json")
					cleanedContent = strings.TrimPrefix(cleanedContent, "```")
					cleanedContent = strings.TrimSuffix(cleanedContent, "```")
					cleanedContent = strings.TrimSpace(cleanedContent)

					if err := json.Unmarshal([]byte(cleanedContent), &diag); err == nil && diag.Classification != "" {
						return diag, nil
					}
				}
			}
		}
	}

	// Case B: Direct diagnosis object
	rawBytes, _ := json.Marshal(rawResp)
	var diag Diagnosis
	if err := json.Unmarshal(rawBytes, &diag); err == nil && diag.Classification != "" {
		return diag, nil
	}

	return Diagnosis{
		Classification:    ClassUnknown,
		RootCause:         "AI diagnostic returned unparseable response structure",
		Confidence:        0.50,
		RecommendedAction: "MANUAL_INSPECTION",
	}, nil
}
