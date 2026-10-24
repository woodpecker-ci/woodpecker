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
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"
)

var (
	// ErrClinicalTestCannotBeQuarantined enforces the strict safety invariant that clinical evals cannot be silenced.
	ErrClinicalTestCannotBeQuarantined = errors.New("DCB0129 safety violation: clinical safety evaluations and golden datasets can NEVER be quarantined")
	// ErrQuarantineExpired indicates a quarantine record has passed its statutory 72-hour TTL.
	ErrQuarantineExpired = errors.New("test quarantine has expired (TTL exceeded), test must be executed")
	// ErrTestNotFound indicates the requested test quarantine entry does not exist.
	ErrTestNotFound = errors.New("quarantine record not found for test")
)

// DefaultQuarantineTTL defines maximum lifespan of a quarantine before mandatory re-execution.
const DefaultQuarantineTTL = 72 * time.Hour

// QuarantineRecord stores details of an isolated non-clinical flaky test.
type QuarantineRecord struct {
	TestIdentifier string    `json:"test_identifier"` // e.g. "web/e2e/navbar_toggle.spec.ts::should toggle"
	SuiteName      string    `json:"suite_name"`      // e.g. "frontend-e2e"
	Reason         string    `json:"reason"`          // e.g. "Transient UI rendering race condition"
	IssueTicketID  string    `json:"issue_ticket_id"` // e.g. "ISSUE-HEAL-104"
	QuarantinedAt  time.Time `json:"quarantined_at"`
	ExpiresAt      time.Time `json:"expires_at"`
	QuarantinedBy  string    `json:"quarantined_by"`  // e.g. "symphonix-healer[bot]"
}

// QuarantineManager controls test quarantine admissions and enforcement.
type QuarantineManager interface {
	QuarantineTest(ctx context.Context, record QuarantineRecord) error
	IsQuarantined(ctx context.Context, testIdentifier string) (bool, error)
	LiftQuarantine(ctx context.Context, testIdentifier string) error
	ListQuarantined(ctx context.Context) ([]QuarantineRecord, error)
}

// TieredQuarantineManager enforces strict clinical vs non-clinical test boundaries.
type TieredQuarantineManager struct {
	mu      sync.RWMutex
	records map[string]QuarantineRecord
}

// NewTieredQuarantineManager instantiates a new TieredQuarantineManager.
func NewTieredQuarantineManager() *TieredQuarantineManager {
	return &TieredQuarantineManager{
		records: make(map[string]QuarantineRecord),
	}
}

// isClinicalTest inspects test identifiers for clinical evaluation patterns.
func isClinicalTest(identifier string, suite string) bool {
	combined := strings.ToLower(fmt.Sprintf("%s::%s", suite, identifier))
	clinicalKeywords := []string{
		"clinical",
		"eval",
		"golden",
		"sepsis",
		"triage",
		"fhir",
		"patient",
		"dcb0129",
		"dcb0160",
		"iso42001",
		"dosage",
		"vital",
		"diagnostic",
	}

	for _, kw := range clinicalKeywords {
		if strings.Contains(combined, kw) {
			return true
		}
	}
	return false
}

// QuarantineTest places a flaky test into quarantine if and only if it is NOT clinical.
func (m *TieredQuarantineManager) QuarantineTest(ctx context.Context, record QuarantineRecord) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Invariant: Clinical safety tests CANNOT be quarantined
	if isClinicalTest(record.TestIdentifier, record.SuiteName) {
		return fmt.Errorf("%w: test %q in suite %q is classified as clinical safety critical",
			ErrClinicalTestCannotBeQuarantined, record.TestIdentifier, record.SuiteName)
	}

	now := time.Now().UTC()
	if record.QuarantinedAt.IsZero() {
		record.QuarantinedAt = now
	}
	if record.ExpiresAt.IsZero() {
		record.ExpiresAt = record.QuarantinedAt.Add(DefaultQuarantineTTL)
	}
	if record.QuarantinedBy == "" {
		record.QuarantinedBy = "symphonix-healer[bot]"
	}

	m.records[record.TestIdentifier] = record
	return nil
}

// IsQuarantined checks if a test is actively quarantined and within its 72-hour TTL.
func (m *TieredQuarantineManager) IsQuarantined(ctx context.Context, testIdentifier string) (bool, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	record, exists := m.records[testIdentifier]
	if !exists {
		return false, nil
	}

	// Check if TTL has expired
	if time.Now().UTC().After(record.ExpiresAt) {
		delete(m.records, testIdentifier)
		return false, fmt.Errorf("%w for test %q (expired at %s)", ErrQuarantineExpired, testIdentifier, record.ExpiresAt.Format(time.RFC3339))
	}

	return true, nil
}

// LiftQuarantine removes a test from quarantine once resolved.
func (m *TieredQuarantineManager) LiftQuarantine(ctx context.Context, testIdentifier string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.records[testIdentifier]; !exists {
		return ErrTestNotFound
	}

	delete(m.records, testIdentifier)
	return nil
}

// ListQuarantined returns all currently active quarantine records.
func (m *TieredQuarantineManager) ListQuarantined(ctx context.Context) ([]QuarantineRecord, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	now := time.Now().UTC()
	var active []QuarantineRecord
	for _, rec := range m.records {
		if now.Before(rec.ExpiresAt) {
			active = append(active, rec)
		}
	}
	return active, nil
}
