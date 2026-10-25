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
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"go.woodpecker-ci.org/woodpecker/v3/server"
	"go.woodpecker-ci.org/woodpecker/v3/server/model"
)

// Canonical14ColTestCase represents one scenario in the standard 14-column matrix.
type Canonical14ColTestCase struct {
	UseCaseID         string                 `json:"use_case_id"`
	Component         string                 `json:"component"`
	Scenario          string                 `json:"scenario"`
	TestType          string                 `json:"test_type"`
	Priority          string                 `json:"priority"`
	ExpectedOutcomes  []string               `json:"expected_outcomes"`
	Preconditions     string                 `json:"preconditions"`
	TestData          map[string]any         `json:"test_data"`
	ValidationRules   []string               `json:"validation_rules"`
	Dependencies      []string               `json:"dependencies"`
	Tags              []string               `json:"tags"`
	EstimatedDuration int                    `json:"estimated_duration"`
	AutomationStatus  string                 `json:"automation_status"`
	Notes             string                 `json:"notes"`
}

// Canonical14ColMatrix represents the root JSON schema for 14-column test matrices.
type Canonical14ColMatrix struct {
	Metadata struct {
		Format         string         `json:"format"`
		TotalScenarios int            `json:"total_scenarios"`
		Component      string         `json:"component"`
		DataSource     string         `json:"data_source"`
		RequirementsDoc string        `json:"requirements_doc"`
		Executor       string         `json:"executor"`
		Distribution   map[string]int `json:"distribution"`
		RequirementIDs []string       `json:"requirement_ids"`
	} `json:"metadata"`
	TestCases []Canonical14ColTestCase `json:"test_cases"`
}

// loadCanonicalMatrix locates and deserializes the 14-column matrix file.
func loadCanonicalMatrix(t *testing.T) Canonical14ColMatrix {
	t.Helper()
	// Candidate search paths
	paths := []string{
		"../../../contrib/symphonix/tests/reduced_json_matrices/healthci_self_healing_matrix.14col.json",
		"contrib/symphonix/tests/reduced_json_matrices/healthci_self_healing_matrix.14col.json",
		"/workspace/contrib/symphonix/tests/reduced_json_matrices/healthci_self_healing_matrix.14col.json",
	}

	var data []byte
	var err error
	var foundPath string
	for _, p := range paths {
		if abs, absErr := filepath.Abs(p); absErr == nil {
			if b, readErr := os.ReadFile(abs); readErr == nil {
				data = b
				foundPath = abs
				break
			}
		}
	}

	require.NotEmpty(t, data, "Must locate healthci_self_healing_matrix.14col.json across candidate paths (checked %v, last err: %v)", paths, err)
	t.Logf("Loaded 14-column matrix from: %s", foundPath)

	var matrix Canonical14ColMatrix
	err = json.Unmarshal(data, &matrix)
	require.NoError(t, err, "Matrix JSON must unmarshal cleanly into 14-column schema")
	return matrix
}

// TestCanonical14ColumnMatrixSchema validates structural adherence to the 14-column standard.
func TestCanonical14ColumnMatrixSchema(t *testing.T) {
	matrix := loadCanonicalMatrix(t)

	// Validate Metadata
	assert.Equal(t, "Standard BulletTrain 14-column format", matrix.Metadata.Format)
	assert.Equal(t, 17, matrix.Metadata.TotalScenarios)
	assert.Equal(t, "healthci_self_healing", matrix.Metadata.Component)
	assert.NotEmpty(t, matrix.Metadata.RequirementsDoc)
	assert.Len(t, matrix.Metadata.RequirementIDs, 8)
	assert.Len(t, matrix.TestCases, matrix.Metadata.TotalScenarios)

	// Validate distribution
	posCount, negCount, edgeCount := 0, 0, 0
	for _, tc := range matrix.TestCases {
		// Verify all 14 columns are populated
		assert.NotEmpty(t, tc.UseCaseID, "Column 1: use_case_id")
		assert.NotEmpty(t, tc.Component, "Column 2: component")
		assert.NotEmpty(t, tc.Scenario, "Column 3: scenario")
		assert.NotEmpty(t, tc.TestType, "Column 4: test_type")
		assert.NotEmpty(t, tc.Priority, "Column 5: priority")
		assert.NotEmpty(t, tc.ExpectedOutcomes, "Column 6: expected_outcomes")
		assert.NotEmpty(t, tc.Preconditions, "Column 7: preconditions")
		assert.NotEmpty(t, tc.TestData, "Column 8: test_data")
		assert.NotEmpty(t, tc.ValidationRules, "Column 9: validation_rules")
		assert.NotEmpty(t, tc.Dependencies, "Column 10: dependencies")
		assert.NotEmpty(t, tc.Tags, "Column 11: tags")
		assert.Greater(t, tc.EstimatedDuration, 0, "Column 12: estimated_duration")
		assert.Equal(t, "automated", tc.AutomationStatus, "Column 13: automation_status")
		assert.NotEmpty(t, tc.Notes, "Column 14: notes")

		switch tc.TestType {
		case "positive":
			posCount++
		case "negative":
			negCount++
		case "edge":
			edgeCount++
		default:
			t.Fatalf("Unknown test_type %q in %s", tc.TestType, tc.UseCaseID)
		}
	}

	assert.Equal(t, matrix.Metadata.Distribution["positive"], posCount)
	assert.Equal(t, matrix.Metadata.Distribution["negative"], negCount)
	assert.Equal(t, matrix.Metadata.Distribution["edge"], edgeCount)
}

// TestCanonical14ColumnExecution runs all 17 scenarios against the live implementation.
func TestCanonical14ColumnExecution(t *testing.T) {
	matrix := loadCanonicalMatrix(t)
	ctx := context.Background()

	for _, tc := range matrix.TestCases {
		t.Run(tc.UseCaseID+"_"+tc.Scenario, func(t *testing.T) {
			switch tc.UseCaseID {

			// ----------------------------------------------------------------
			// FR-HEAL-001: DCB0129 Code Provenance Gate Scenarios
			// ----------------------------------------------------------------
			case "healthci_heal_prov_0001":
				tracker := NewInMemoryProvenanceTracker()
				patch := &DCB0129ProvenanceRecord{
					PatchID:      "patch-001",
					PipelineID:   1001,
					StepName:     "fhir-schema-validation",
					TargetBranch: "main",
					DiffUnified:  "--- a/fhir.go\n+++ b/fhir.go\n@@ -1,1 +1,1 @@\n- old\n+ new\n",
				}
				err := tracker.RecordPatch(ctx, patch)
				require.NoError(t, err)
				assert.Equal(t, "symphonix-healer[bot]", patch.AuthorBot)
				assert.Equal(t, "did:symphonix:agent:healer-v1", patch.MachineDID)

				// Add dual human reviews (Senior Dev + Certified CSO)
				_ = tracker.AddHumanSignOff(ctx, patch.PatchID, HumanSignOff{ReviewerID: "dev-01", Role: "SENIOR_DEV"})
				_ = tracker.AddHumanSignOff(ctx, patch.PatchID, HumanSignOff{ReviewerID: "cso-01", Role: "CLINICAL_SAFETY_OFFICER", IsCSO: true})

				err = tracker.VerifyMergeEligibility(ctx, patch.PatchID, "main")
				assert.NoError(t, err, "Dual sign-off with CSO must succeed on main")

			case "healthci_heal_prov_0002":
				tracker := NewInMemoryProvenanceTracker()
				patch := &DCB0129ProvenanceRecord{
					PatchID:      "patch-002",
					TargetBranch: "main",
					DiffUnified:  "--- a/test\n+++ b/test\n",
				}
				_ = tracker.RecordPatch(ctx, patch)
				err := tracker.VerifyMergeEligibility(ctx, patch.PatchID, "main")
				require.Error(t, err)
				assert.True(t, errors.Is(err, ErrDCB0129ProvenanceViolation))

			case "healthci_heal_prov_0003":
				tracker := NewInMemoryProvenanceTracker()
				patch := &DCB0129ProvenanceRecord{
					PatchID:      "patch-003",
					TargetBranch: "main",
					DiffUnified:  "--- a/test\n+++ b/test\n",
				}
				_ = tracker.RecordPatch(ctx, patch)
				_ = tracker.AddHumanSignOff(ctx, patch.PatchID, HumanSignOff{ReviewerID: "dev-01", Role: "DEV"})
				_ = tracker.AddHumanSignOff(ctx, patch.PatchID, HumanSignOff{ReviewerID: "dev-02", Role: "QA_ENGINEER"})
				err := tracker.VerifyMergeEligibility(ctx, patch.PatchID, "main")
				require.Error(t, err)
				assert.True(t, errors.Is(err, ErrDCB0129ProvenanceViolation))
				assert.Contains(t, err.Error(), "missing mandatory Clinical Safety Officer")

			case "healthci_heal_prov_0004":
				tracker := NewInMemoryProvenanceTracker()
				patch := &DCB0129ProvenanceRecord{
					PatchID:      "patch-004",
					TargetBranch: "feature/sandbox-eval",
					DiffUnified:  "--- a/test\n+++ b/test\n",
				}
				_ = tracker.RecordPatch(ctx, patch)
				err := tracker.VerifyMergeEligibility(ctx, patch.PatchID, "feature/sandbox-eval")
				assert.NoError(t, err, "Feature sandbox branches permit iterative developer experimentation")

			// ----------------------------------------------------------------
			// FR-HEAL-002: Tiered Flaky Test Isolation Scenarios
			// ----------------------------------------------------------------
			case "healthci_heal_quar_0005":
				qm := NewTieredQuarantineManager()
				err := qm.QuarantineTest(ctx, QuarantineRecord{
					TestIdentifier: "evals/sepsis/triage_benchmark_test.py::test_zero_miss",
					SuiteName:      "clinical-evals",
					Reason:         "Sensitivity boundary variance",
				})
				require.Error(t, err)
				assert.True(t, errors.Is(err, ErrClinicalTestCannotBeQuarantined))
				isQuarantined, _ := qm.IsQuarantined(ctx, "evals/sepsis/triage_benchmark_test.py::test_zero_miss")
				assert.False(t, isQuarantined)

			case "healthci_heal_quar_0006":
				qm := NewTieredQuarantineManager()
				rec := QuarantineRecord{
					TestIdentifier: "web/e2e/navbar_toggle.spec.ts::should animate",
					SuiteName:      "frontend-e2e",
					Reason:         "UI rendering timing variance",
					IssueTicketID:  "UI-312",
				}
				err := qm.QuarantineTest(ctx, rec)
				require.NoError(t, err)
				isQuarantined, _ := qm.IsQuarantined(ctx, rec.TestIdentifier)
				assert.True(t, isQuarantined)

			case "healthci_heal_quar_0007":
				qm := NewTieredQuarantineManager()
				expiredRec := QuarantineRecord{
					TestIdentifier: "integration/db_lock_test.go::TestHighContention",
					SuiteName:      "integration",
					Reason:         "Lock timeout under load",
					QuarantinedAt:  time.Now().UTC().Add(-80 * time.Hour),
					ExpiresAt:      time.Now().UTC().Add(-8 * time.Hour),
				}
				qm.records[expiredRec.TestIdentifier] = expiredRec
				isQuarantined, err := qm.IsQuarantined(ctx, expiredRec.TestIdentifier)
				assert.False(t, isQuarantined)
				require.Error(t, err)
				assert.True(t, errors.Is(err, ErrQuarantineExpired))

			// ----------------------------------------------------------------
			// FR-HEAL-003: Predictive Memory Rightsizing Scenarios
			// ----------------------------------------------------------------
			case "healthci_heal_pred_0008":
				predictor := NewTelemetryMemoryPredictor()
				repoID := int64(100)
				step := "dicom_segmentation_benchmark"
				samples := []int64{1200, 1400, 1550, 1600, 1750, 1800, 1820, 1850, 1900, 2000}
				for _, m := range samples {
					_ = predictor.RecordStepMetric(ctx, StepMetricRecord{RepoID: repoID, StepName: step, PeakMemoryMB: m})
				}
				recMB, resized := predictor.PredictStepMemory(ctx, repoID, step, 1024)
				assert.True(t, resized)
				assert.Greater(t, recMB, int64(2000))

			case "healthci_heal_pred_0009":
				predictor := NewTelemetryMemoryPredictor()
				recMB, resized := predictor.PredictStepMemory(ctx, 999, "brand_new_step", 1024)
				assert.False(t, resized)
				assert.Equal(t, int64(1024), recMB)

			// ----------------------------------------------------------------
			// FR-HEAL-004: Self-Healing YAML Config Scenarios
			// ----------------------------------------------------------------
			case "healthci_heal_yaml_0010":
				repairer := NewDefaultConfigRepairer()
				broken := "steps:\n  test:\n    commands:\n      - npm test --mock\n    environment:\n      SECRET: 123\n"
				sug, err := repairer.AnalyzeAndRepair(ctx, broken)
				require.NoError(t, err)
				assert.Contains(t, sug.RemediatedYAML, "--real-services")
				assert.NotContains(t, sug.RemediatedYAML, "--mock")
				assert.Contains(t, sug.RemediatedYAML, "WOODPECKER_ENCRYPTION_TINK_KEYSET_FILE")
				assert.NotEmpty(t, sug.UnifiedDiff)

			case "healthci_heal_yaml_0011":
				repairer := NewDefaultConfigRepairer()
				privYAML := "steps:\n  build:\n    privileged: true\n"
				sug, err := repairer.AnalyzeAndRepair(ctx, privYAML)
				require.NoError(t, err)
				assert.True(t, sug.RequiresCSOReview)
				assert.Contains(t, sug.RemediatedYAML, "privileged: false")

			// ----------------------------------------------------------------
			// FR-HEAL-005: Multi-Agent Dual-Turn Cross-Check Scenarios
			// ----------------------------------------------------------------
			case "healthci_heal_multi_0012":
				auditor := NewClinicalSafetyMultiAgentAuditor()
				patch := CandidatePatch{
					StepName:     "client-sync",
					ProposedDiff: "--- a/client.go\n+++ b/client.go\n@@ -1,1 +1,1 @@\n- import \"old/pkg\"\n+ import \"new/pkg\"\n",
				}
				res, err := auditor.AuditCandidatePatch(ctx, patch)
				require.NoError(t, err)
				assert.True(t, res.Approved)
				assert.Equal(t, HazardNegligible, res.HazardLevel)

			case "healthci_heal_multi_0013":
				auditor := NewClinicalSafetyMultiAgentAuditor()
				patch := CandidatePatch{
					StepName:     "triage-test",
					ProposedDiff: "--- a/test.go\n+++ b/test.go\n@@ -1,1 +1,1 @@\n- assert.Equal(t, true, result)\n+ // assert.Equal(t, true, result)\n",
				}
				res, err := auditor.AuditCandidatePatch(ctx, patch)
				require.NoError(t, err)
				assert.False(t, res.Approved)
				assert.Equal(t, HazardCritical, res.HazardLevel)

			case "healthci_heal_multi_0014":
				auditor := NewClinicalSafetyMultiAgentAuditor()
				patch := CandidatePatch{
					StepName:     "vital-check",
					ProposedDiff: "--- a/bounds.go\n+++ b/bounds.go\n@@ -1,1 +1,1 @@\n- const TemperatureThreshold = 38.3\n+ const TemperatureThreshold = 38.0\n",
				}
				res, err := auditor.AuditCandidatePatch(ctx, patch)
				require.NoError(t, err)
				assert.False(t, res.Approved)
				assert.Equal(t, HazardCritical, res.HazardLevel)

			// ----------------------------------------------------------------
			// FR-HEAL-006: Bounded Counterfactual Synthetic Case Scenarios
			// ----------------------------------------------------------------
			case "healthci_heal_syn_0015":
				gen := NewBoundedCounterfactualGenerator()
				evalCase := ClinicalEvaluationCase{
					CaseID: "case-01",
					Parameters: []ClinicalParameter{
						{Code: "8310-5", Name: "Body Temperature", Value: 38.0, Threshold: 38.3, Unit: "Cel"},
					},
				}
				vectors, err := gen.GenerateCounterfactualVectors(ctx, evalCase)
				require.NoError(t, err)
				assert.Len(t, vectors, 4)
				for _, v := range vectors {
					assert.Contains(t, v.FHIRResourceJSON, `"Observation"`)
					assert.Contains(t, v.FHIRResourceJSON, `"8310-5"`)
				}

			// ----------------------------------------------------------------
			// NFR-HEAL-001: Air-Gapped Qwen 2.5 Coder 32B & Offline Resilience
			// ----------------------------------------------------------------
			case "healthci_heal_llm_0016":
				mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					w.Header().Set("Content-Type", "application/json")
					resp := map[string]any{
						"choices": []any{
							map[string]any{
								"message": map[string]any{
									"role":    "assistant",
									"content": `{"classification": "TRANSIENT", "root_cause": "Network DNS transient", "confidence": 0.95, "recommended_action": "RESTART"}`,
								},
							},
						},
					}
					_ = json.NewEncoder(w).Encode(resp)
				}))
				defer mockServer.Close()

				server.Config.SelfHealing.Model = "qwen2.5-coder:32b"
				diag, err := queryAIDiagnostic(ctx, mockServer.URL+"/v1/chat/completions", &model.Step{Name: "step"}, "i/o timeout")
				require.NoError(t, err)
				assert.Equal(t, ClassTransient, diag.Classification)
				assert.Equal(t, "RESTART", diag.RecommendedAction)
				assert.GreaterOrEqual(t, diag.Confidence, 0.90)

			case "healthci_heal_llm_0017":
				server.Config.SelfHealing.Model = "qwen2.5-coder:32b"
				step := &model.Step{Name: "step-137", ExitCode: 137}
				diag := analyzeFailure(ctx, step, "exit 137 OOM")
				assert.Equal(t, ClassTransient, diag.Classification)
				assert.True(t, strings.Contains(diag.RootCause, "Exit 137") || strings.Contains(diag.RootCause, "OOM"))
				assert.Equal(t, "RESTART", diag.RecommendedAction)

			default:
				t.Fatalf("Unimplemented test execution handler for canonical use case: %s", tc.UseCaseID)
			}
		})
	}
}
