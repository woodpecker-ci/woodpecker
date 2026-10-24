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
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"
)

var (
	// ErrDCB0129ProvenanceViolation is returned when an AI patch attempts merge without dual human review.
	ErrDCB0129ProvenanceViolation = errors.New("DCB0129 safety invariant violated: AI code cannot be merged without dual human review")
	// ErrInsufficientSignOffs is returned when required sign-offs (e.g. Clinical Safety Officer) are missing.
	ErrInsufficientSignOffs = errors.New("insufficient human sign-offs: minimum 2 reviewers including certified CSO required")
	// ErrPatchNotFound is returned when the specified patch ID is unknown.
	ErrPatchNotFound = errors.New("code provenance patch record not found")
	// ErrPatchAlreadyApproved is returned when attempting to modify a frozen approved patch.
	ErrPatchAlreadyApproved = errors.New("patch is already approved and locked")
)

// HazardLevel represents statutory clinical safety risk levels under DCB0129 / ISO-14971.
type HazardLevel string

const (
	HazardNegligible HazardLevel = "NEGLIGIBLE"
	HazardMinor      HazardLevel = "MINOR"
	HazardModerate   HazardLevel = "MODERATE"
	HazardMajor      HazardLevel = "MAJOR"
	HazardCritical   HazardLevel = "CRITICAL"
)

// HumanSignOff records an accountable human reviewer's approval.
type HumanSignOff struct {
	ReviewerID   string    `json:"reviewer_id"`
	ReviewerName string    `json:"reviewer_name"`
	Role         string    `json:"role"` // e.g. "CLINICAL_SAFETY_OFFICER", "SENIOR_ENGINEER"
	Timestamp    time.Time `json:"timestamp"`
	Comments     string    `json:"comments"`
	IsCSO        bool      `json:"is_cso"`
}

// DCB0129HazardImpact outlines clinical hazard analysis for code modification.
type DCB0129HazardImpact struct {
	HazardLevel            HazardLevel `json:"hazard_level"`
	ClinicalJustification  string      `json:"clinical_justification"`
	MitigationVerification string      `json:"mitigation_verification"`
	TouchesClinicalLogic   bool        `json:"touches_clinical_logic"`
}

// DCB0129ProvenanceRecord tracks full cryptographic lineage of automated code patches.
type DCB0129ProvenanceRecord struct {
	PatchID                string              `json:"patch_id"`
	AuthorBot              string              `json:"author_bot"`   // e.g. "symphonix-healer[bot]"
	MachineDID             string              `json:"machine_did"`  // e.g. "did:symphonix:agent:healer-v1"
	RepoID                 int64               `json:"repo_id"`
	PipelineID             int64               `json:"pipeline_id"`
	StepName               string              `json:"step_name"`
	TargetBranch           string              `json:"target_branch"`
	DiffUnified            string              `json:"diff_unified"`
	DiffChecksumSHA256     string              `json:"diff_checksum_sha256"`
	ModelDiagnosticRCA     string              `json:"model_diagnostic_rca"`
	HazardImpact           DCB0129HazardImpact `json:"hazard_impact"`
	DualHumanReviewRequired bool                `json:"dual_human_review_required"`
	CreatedAt              time.Time           `json:"created_at"`
	SignOffs               []HumanSignOff      `json:"sign_offs"`
	Merged                 bool                `json:"merged"`
}

// ProvenanceTracker manages the lifecycle and clinical governance gate for AI-generated patches.
type ProvenanceTracker interface {
	RecordPatch(ctx context.Context, record *DCB0129ProvenanceRecord) error
	GetPatch(ctx context.Context, patchID string) (*DCB0129ProvenanceRecord, error)
	AddHumanSignOff(ctx context.Context, patchID string, signOff HumanSignOff) error
	VerifyMergeEligibility(ctx context.Context, patchID string, targetBranch string) error
	MarkMerged(ctx context.Context, patchID string) error
}

// InMemoryProvenanceTracker is a thread-safe implementation of ProvenanceTracker.
type InMemoryProvenanceTracker struct {
	mu      sync.RWMutex
	patches map[string]*DCB0129ProvenanceRecord
}

// NewInMemoryProvenanceTracker creates a new in-memory provenance tracker.
func NewInMemoryProvenanceTracker() *InMemoryProvenanceTracker {
	return &InMemoryProvenanceTracker{
		patches: make(map[string]*DCB0129ProvenanceRecord),
	}
}

// ComputeDiffChecksum generates a deterministic SHA-256 hash of the git diff.
func ComputeDiffChecksum(diff string) string {
	h := sha256.Sum256([]byte(diff))
	return hex.EncodeToString(h[:])
}

// RecordPatch stores a new patch with cryptographic hash and mandatory human review invariant.
func (t *InMemoryProvenanceTracker) RecordPatch(ctx context.Context, record *DCB0129ProvenanceRecord) error {
	t.mu.Lock()
	defer t.mu.Unlock()

	if record.PatchID == "" {
		record.PatchID = fmt.Sprintf("patch-%d-%d", record.PipelineID, time.Now().UnixNano())
	}
	record.AuthorBot = "symphonix-healer[bot]"
	record.MachineDID = "did:symphonix:agent:healer-v1"
	record.DiffChecksumSHA256 = ComputeDiffChecksum(record.DiffUnified)
	record.DualHumanReviewRequired = true // Hard clinical invariant
	record.CreatedAt = time.Now().UTC()
	record.Merged = false

	t.patches[record.PatchID] = record
	return nil
}

// GetPatch retrieves a provenance record by patch ID.
func (t *InMemoryProvenanceTracker) GetPatch(ctx context.Context, patchID string) (*DCB0129ProvenanceRecord, error) {
	t.mu.RLock()
	defer t.mu.RUnlock()

	record, exists := t.patches[patchID]
	if !exists {
		return nil, ErrPatchNotFound
	}
	return record, nil
}

// AddHumanSignOff appends a verified human sign-off to the patch record.
func (t *InMemoryProvenanceTracker) AddHumanSignOff(ctx context.Context, patchID string, signOff HumanSignOff) error {
	t.mu.Lock()
	defer t.mu.Unlock()

	record, exists := t.patches[patchID]
	if !exists {
		return ErrPatchNotFound
	}

	if record.Merged {
		return ErrPatchAlreadyApproved
	}

	if signOff.Timestamp.IsZero() {
		signOff.Timestamp = time.Now().UTC()
	}

	record.SignOffs = append(record.SignOffs, signOff)
	return nil
}

// VerifyMergeEligibility enforces DCB0129 clinical safety rules before code merge.
func (t *InMemoryProvenanceTracker) VerifyMergeEligibility(ctx context.Context, patchID string, targetBranch string) error {
	t.mu.RLock()
	defer t.mu.RUnlock()

	record, exists := t.patches[patchID]
	if !exists {
		return ErrPatchNotFound
	}

	// Rule 1: Protected branch checks (main, master, release/*, production)
	isProtected := targetBranch == "main" ||
		targetBranch == "master" ||
		targetBranch == "production" ||
		strings.HasPrefix(targetBranch, "release/")

	if isProtected {
		if len(record.SignOffs) < 2 {
			return fmt.Errorf("%w: target branch %q requires at least 2 human sign-offs (got %d)",
				ErrDCB0129ProvenanceViolation, targetBranch, len(record.SignOffs))
		}

		// Rule 2: At least one sign-off must be a certified Clinical Safety Officer (CSO)
		hasCSO := false
		for _, s := range record.SignOffs {
			if s.IsCSO || strings.Contains(strings.ToUpper(s.Role), "CSO") || strings.Contains(strings.ToUpper(s.Role), "CLINICAL_SAFETY_OFFICER") {
				hasCSO = true
				break
			}
		}

		if !hasCSO {
			return fmt.Errorf("%w: missing mandatory Clinical Safety Officer (CSO) sign-off for protected branch %q",
				ErrDCB0129ProvenanceViolation, targetBranch)
		}
	}

	return nil
}

// MarkMerged locks the record as merged into target branch.
func (t *InMemoryProvenanceTracker) MarkMerged(ctx context.Context, patchID string) error {
	t.mu.Lock()
	defer t.mu.Unlock()

	record, exists := t.patches[patchID]
	if !exists {
		return ErrPatchNotFound
	}

	record.Merged = true
	return nil
}
