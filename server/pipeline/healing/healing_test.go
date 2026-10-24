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
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"go.woodpecker-ci.org/woodpecker/v3/server"
	"go.woodpecker-ci.org/woodpecker/v3/server/model"
)

// ----------------------------------------------------------------------------
// Feature 1: Code Provenance & DCB0129 Clinical Merge Gate Tests
// ----------------------------------------------------------------------------

func TestDCB0129ProvenanceTracking_DirectMergeBlocked(t *testing.T) {
	tracker := NewInMemoryProvenanceTracker()
	ctx := context.Background()

	patch := &DCB0129ProvenanceRecord{
		PatchID:            "patch-clinical-001",
		PipelineID:         1042,
		StepName:           "fhir-schema-validation",
		TargetBranch:       "main",
		DiffUnified:        "--- a/fhir.go\n+++ b/fhir.go\n@@ -10,1 +10,1 @@\n- var Valid = false\n+ var Valid = true\n",
		ModelDiagnosticRCA: "Fixed syntax validation failure in FHIR R4 parser",
		HazardImpact: DCB0129HazardImpact{
			HazardLevel:            HazardMinor,
			ClinicalJustification:  "Correction of schema parsing logic for observation status",
			MitigationVerification: "Passed full synthetic FHIR suite",
			TouchesClinicalLogic:   true,
		},
	}

	err := tracker.RecordPatch(ctx, patch)
	require.NoError(t, err)

	// Invariant Check 1: Patch must be authored by symphonix-healer[bot]
	saved, err := tracker.GetPatch(ctx, patch.PatchID)
	require.NoError(t, err)
	assert.Equal(t, "symphonix-healer[bot]", saved.AuthorBot)
	assert.NotEmpty(t, saved.DiffChecksumSHA256)
	assert.True(t, saved.DualHumanReviewRequired)

	// Invariant Check 2: Attempting to merge to protected branch without sign-offs MUST FAIL
	err = tracker.VerifyMergeEligibility(ctx, patch.PatchID, "main")
	require.Error(t, err)
	assert.True(t, errors.Is(err, ErrDCB0129ProvenanceViolation))
	assert.Contains(t, err.Error(), "requires at least 2 human sign-offs")

	// Sign-off 1: Senior DevOps Engineer (non-CSO)
	err = tracker.AddHumanSignOff(ctx, patch.PatchID, HumanSignOff{
		ReviewerID:   "usr-eng-01",
		ReviewerName: "Lead DevOps",
		Role:         "SENIOR_ENGINEER",
		IsCSO:        false,
		Comments:     "Verified unified diff and container build passes",
	})
	require.NoError(t, err)

	// Attempting merge with only 1 sign-off MUST FAIL
	err = tracker.VerifyMergeEligibility(ctx, patch.PatchID, "main")
	require.Error(t, err)
	assert.True(t, errors.Is(err, ErrDCB0129ProvenanceViolation))
	assert.Contains(t, err.Error(), "requires at least 2 human sign-offs")

	// Sign-off 2: Certified Clinical Safety Officer (CSO)
	err = tracker.AddHumanSignOff(ctx, patch.PatchID, HumanSignOff{
		ReviewerID:   "usr-cso-99",
		ReviewerName: "Dr. Evelyn Reed (CSO)",
		Role:         "CLINICAL_SAFETY_OFFICER",
		IsCSO:        true,
		Comments:     "Reviewed DCB0129 hazard impact; approved for staging/main cutover",
	})
	require.NoError(t, err)

	// Now verification MUST PASS!
	err = tracker.VerifyMergeEligibility(ctx, patch.PatchID, "main")
	require.NoError(t, err, "Merge must succeed once dual human sign-off with certified CSO is recorded")

	// Lock as merged
	err = tracker.MarkMerged(ctx, patch.PatchID)
	require.NoError(t, err)
	savedMerged, _ := tracker.GetPatch(ctx, patch.PatchID)
	assert.True(t, savedMerged.Merged)
}

// ----------------------------------------------------------------------------
// Feature 2: Tiered Flaky Test Isolation Tests
// ----------------------------------------------------------------------------

func TestTieredQuarantine_ClinicalSafetyInvariant(t *testing.T) {
	qm := NewTieredQuarantineManager()
	ctx := context.Background()

	// Case A: Clinical evaluation test CANNOT be quarantined
	clinicalTest := QuarantineRecord{
		TestIdentifier: "evals/sepsis/triage_benchmark_test.py::test_zero_miss_sensitivity",
		SuiteName:      "clinical-evals",
		Reason:         "Model sensitivity dropped to 99.85% (threshold is 99.9%)",
		IssueTicketID:  "CLIN-702",
	}

	err := qm.QuarantineTest(ctx, clinicalTest)
	require.Error(t, err)
	assert.True(t, errors.Is(err, ErrClinicalTestCannotBeQuarantined))
	assert.Contains(t, err.Error(), "clinical safety critical")

	// Verify test is NOT in quarantine
	isQuarantined, _ := qm.IsQuarantined(ctx, clinicalTest.TestIdentifier)
	assert.False(t, isQuarantined)

	// Case B: Non-clinical UI / Integration test CAN be quarantined
	nonClinicalTest := QuarantineRecord{
		TestIdentifier: "web/e2e/navbar_toggle.spec.ts::should animate sidebar smoothly",
		SuiteName:      "frontend-e2e",
		Reason:         "Chromium CSS animation timing variance in headless test runner",
		IssueTicketID:  "UI-312",
	}

	err = qm.QuarantineTest(ctx, nonClinicalTest)
	require.NoError(t, err)

	isQuarantined, err = qm.IsQuarantined(ctx, nonClinicalTest.TestIdentifier)
	require.NoError(t, err)
	assert.True(t, isQuarantined)

	// Case C: Quarantine TTL Expiration (72h)
	expiredTest := QuarantineRecord{
		TestIdentifier: "integration/db_lock_test.go::TestHighContention",
		SuiteName:      "integration-tests",
		Reason:         "Lock timeout under virtualized runner load",
		IssueTicketID:  "INFRA-99",
		QuarantinedAt:  time.Now().UTC().Add(-80 * time.Hour), // 80 hours ago
		ExpiresAt:      time.Now().UTC().Add(-8 * time.Hour),  // Expired 8 hours ago
	}

	// Directly insert expired record
	qm.records[expiredTest.TestIdentifier] = expiredTest

	isExpiredActive, err := qm.IsQuarantined(ctx, expiredTest.TestIdentifier)
	assert.False(t, isExpiredActive)
	require.Error(t, err)
	assert.True(t, errors.Is(err, ErrQuarantineExpired))
}

// ----------------------------------------------------------------------------
// Feature 3: Predictive Pre-emptive Resource Sizing Tests
// ----------------------------------------------------------------------------

func TestPredictiveMemorySizing_PreemptsExit137(t *testing.T) {
	predictor := NewTelemetryMemoryPredictor()
	ctx := context.Background()

	stepName := "dicom-segmentation-benchmark"
	repoID := int64(42)
	defaultLimitMB := int64(1024) // 1GB default

	// With no history, should return default limit
	predMB, resized := predictor.PredictStepMemory(ctx, repoID, stepName, defaultLimitMB)
	assert.Equal(t, defaultLimitMB, predMB)
	assert.False(t, resized)

	// Feed historical step telemetry (peak usage around 1500MB - 1800MB)
	samples := []int64{1200, 1400, 1550, 1600, 1750, 1800, 1820, 1850, 1900, 2000}
	for _, m := range samples {
		_ = predictor.RecordStepMetric(ctx, StepMetricRecord{
			RepoID:       repoID,
			StepName:     stepName,
			PeakMemoryMB: m,
			DurationSec:  45.2,
			ExitCode:     0,
		})
	}

	// Calculate prediction: P95 is ~1950MB * 1.20 headroom = ~2340MB
	predMB, resized = predictor.PredictStepMemory(ctx, repoID, stepName, defaultLimitMB)
	assert.True(t, resized, "Should pre-emptively scale memory before container starts")
	assert.Greater(t, predMB, int64(2000))
}

// ----------------------------------------------------------------------------
// Feature 4: Self-Healing .woodpecker.yaml Tests
// ----------------------------------------------------------------------------

func TestConfigRepairer_DetectsMockAndMissingTinkKeyset(t *testing.T) {
	repairer := NewDefaultConfigRepairer()
	ctx := context.Background()

	brokenYAML := `
when:
  branch: main

steps:
  test:
    image: node:20
    commands:
      - npm test --mock
    environment:
      API_SECRET: supersecret
`

	suggestion, err := repairer.AnalyzeAndRepair(ctx, brokenYAML)
	require.NoError(t, err)

	assert.True(t, suggestion.RequiresCSOReview)
	assert.NotEmpty(t, suggestion.IssuesDetected)

	// Must detect L-CAID-10 zero-mock violation and SEC-TINK-01 missing Tink AEAD keyset
	hasMockIssue := false
	hasTinkIssue := false
	for _, issue := range suggestion.IssuesDetected {
		if issue.RuleID == "L-CAID-10" {
			hasMockIssue = true
		}
		if issue.RuleID == "SEC-TINK-01" {
			hasTinkIssue = true
		}
	}
	assert.True(t, hasMockIssue, "Should detect L-CAID-10 mock test violation")
	assert.True(t, hasTinkIssue, "Should detect SEC-TINK-01 missing Tink keyset")

	// Verify remediated YAML replaces --mock with --real-services
	assert.Contains(t, suggestion.RemediatedYAML, "--real-services")
	assert.NotContains(t, suggestion.RemediatedYAML, "--mock")
	assert.Contains(t, suggestion.RemediatedYAML, "WOODPECKER_ENCRYPTION_TINK_KEYSET_FILE")
}

// ----------------------------------------------------------------------------
// Feature 5: Multi-Agent Dual-Turn Cross-Check Tests
// ----------------------------------------------------------------------------

func TestMultiAgentCrossCheck_RejectsAssertionRelaxation(t *testing.T) {
	auditor := NewClinicalSafetyMultiAgentAuditor()
	ctx := context.Background()

	// Turn 1 generates a candidate patch that relaxes a clinical test assertion
	maliciousDevPatch := CandidatePatch{
		StepName: "sepsis-vital-triage",
		ProposedDiff: `
--- a/triage_test.go
+++ b/triage_test.go
@@ -40,2 +40,2 @@
- assert.Equal(t, HighRisk, result.RiskLevel)
+ // assert.Equal(t, HighRisk, result.RiskLevel)
`,
		DiagnosticRCA:   "Test was failing due to boundary mismatch; temporarily commented out assertion",
		ConfidenceScore: 0.88,
	}

	result, err := auditor.AuditCandidatePatch(ctx, maliciousDevPatch)
	require.NoError(t, err)

	// Turn 2 MUST REJECT
	assert.False(t, result.Approved)
	assert.Equal(t, HazardCritical, result.HazardLevel)
	assert.Contains(t, result.AuditorNotes, "REJECTED")

	// Test a legitimate benign technical patch (fixing import syntax)
	benignPatch := CandidatePatch{
		StepName: "client-registry-sync",
		ProposedDiff: `
--- a/client.go
+++ b/client.go
@@ -5,1 +5,1 @@
- import "github.com/legacy/pkg"
+ import "github.com/modern/pkg"
`,
		DiagnosticRCA:   "Updated deprecated import path",
		ConfidenceScore: 0.95,
	}

	resultBenign, err := auditor.AuditCandidatePatch(ctx, benignPatch)
	require.NoError(t, err)
	assert.True(t, resultBenign.Approved)
	assert.Equal(t, HazardNegligible, resultBenign.HazardLevel)
	assert.Contains(t, resultBenign.AuditorNotes, "PASSED")
}

// ----------------------------------------------------------------------------
// Feature 6: Bounded Counterfactual Synthetic Case Generation Tests
// ----------------------------------------------------------------------------

func TestBoundedCounterfactualGenerator_GeneratesFHIRObservations(t *testing.T) {
	generator := NewBoundedCounterfactualGenerator()
	ctx := context.Background()

	evalCase := ClinicalEvaluationCase{
		CaseID:           "case-sepsis-border-01",
		DiagnosisContext: "SIRS Temperature Boundary",
		ExpectedOutcome:  "FLAG_HIGH_RISK",
		ActualOutcome:    "FLAG_NORMAL",
		Parameters: []ClinicalParameter{
			{
				Code:      "8310-5",
				Name:      "Body Temperature",
				Value:     38.0,
				Unit:      "Cel",
				Threshold: 38.3,
			},
		},
	}

	vectors, err := generator.GenerateCounterfactualVectors(ctx, evalCase)
	require.NoError(t, err)
	assert.NotEmpty(t, vectors)
	assert.Len(t, vectors, 4) // ±1%, ±2%

	for _, v := range vectors {
		assert.Contains(t, v.FHIRResourceJSON, `"resourceType": "Observation"`)
		assert.Contains(t, v.FHIRResourceJSON, `"http://loinc.org"`)
		assert.Contains(t, v.FHIRResourceJSON, `"8310-5"`)
	}
}

// ----------------------------------------------------------------------------
// Local Air-Gapped LLM Diagnostic Endpoint Tests (Qwen 2.5 Coder Chat Mock)
// ----------------------------------------------------------------------------

func TestQueryAIDiagnostic_OpenAICompatLocalServer(t *testing.T) {
	// Mock local Ollama / vLLM / OpenAI server
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "application/json", r.Header.Get("Content-Type"))

		var req map[string]any
		_ = json.NewDecoder(r.Body).Decode(&req)
		assert.Equal(t, "qwen2.5-coder:32b", req["model"])

		resp := map[string]any{
			"choices": []any{
				map[string]any{
					"message": map[string]any{
						"role": "assistant",
						"content": `{"classification": "TRANSIENT", "root_cause": "Network DNS transient during client registry sync", "confidence": 0.96, "recommended_action": "RESTART"}`,
					},
				},
			},
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer mockServer.Close()

	server.Config.SelfHealing.Model = "qwen2.5-coder:32b"
	ctx := context.Background()
	step := &model.Step{
		Name:     "sync-client-registry",
		ExitCode: 1,
	}

	diag, err := queryAIDiagnostic(ctx, mockServer.URL+"/v1/chat/completions", step, "dial tcp: lookup registry.internal: i/o timeout")
	require.NoError(t, err)
	assert.Equal(t, ClassTransient, diag.Classification)
	assert.Contains(t, diag.RootCause, "Network DNS transient")
	assert.Equal(t, "RESTART", diag.RecommendedAction)
	assert.InDelta(t, 0.96, diag.Confidence, 0.01)
}

func TestQueryAIDiagnostic_OfflineInternetFailureGracefulFallback(t *testing.T) {
	// Unreachable port simulating network/internet drop
	server.Config.SelfHealing.Model = "qwen2.5-coder:32b"
	ctx := context.Background()
	step := &model.Step{
		Name:     "fhir-validate",
		ExitCode: 137,
	}

	diag := analyzeFailure(ctx, step, "Container killed by OOM")
	// Must gracefully fall back to local heuristic without panic or hang!
	assert.Equal(t, ClassTransient, diag.Classification)
	assert.Contains(t, diag.RootCause, "Exit 137 / potential OOM")
	assert.Equal(t, "RESTART", diag.RecommendedAction)
}
