# HealthCI Real Service End-to-End Governance -- Requirements

Standard: ISO/IEC/IEEE 29148:2018. Component code: `HEALTHCI-SVC`. Trace ID: `DOC-REQ-HEALTHCI-SVC-001`.
Reference Standard: BulletTrain `test_journeys.py` & Real-Service Fixtures (`L-CAID-10` Zero-Mock Principle).

## Purpose

Enforce zero-mock real service validation across the complete Symphonix HealthCI pipeline runtime. Guarantees that self-healing interventions, DCB0129 code provenance barriers, Google Tink AEAD encrypted PostgreSQL 16 persistence, Zero-PHI streaming sanitization, and local air-gapped Qwen 2.5 Coder 32B inference are tested end-to-end against real running server instances in a deterministic and repeatable manner.

## Architectural Conventions (BulletTrain Compatible)

| Area | In-Memory / Unit Test | Real Service End-to-End Invariant (`L-CAID-10`) |
|---|---|---|
| Persistence | Go in-memory mock struct | Real PostgreSQL 16 ACID database with Google Tink AEAD envelope encryption |
| Transport | Go function invocation | Live REST / Gin API (`:8000`) and gRPC streaming log transport (`:9000`) |
| AI Inference | Static mock string | Real local OpenAI-compatible endpoint (`http://127.0.0.1:11434/v1/chat/completions`) with Qwen 2.5 Coder 32B |
| Provenance Gate | Struct flag check | Real Git branch protection policy blocking non-CSO merges with `ErrDCB0129ProvenanceViolation` |
| Execution Output | Pytest/Go test stdout | Formal `SYMPHONIX_SCENARIO_EXECUTION_RESULT_V1` records with row SHA-256 digests |

---

## Requirement Records (ISO/IEC/IEEE 29148:2018)

### FR-HCI-SVC-001 -- Live REST API Pipeline OOM Auto-Scaling
- **Statement**: The real service shall intercept container Exit 137 OOM errors via the live pipeline execution loop, escalate memory limits by 1.5x in PostgreSQL metadata, and automatically reschedule the step to completion.
- **Category**: Functional. Priority: Critical.
- **Status**: Reviewed. Version: 1.0.0.

### FR-HCI-SVC-002 -- Live Dual-Key CSO Gate & Branch Protection
- **Statement**: The real service shall enforce `ACCOUNTABLE_HUMAN_REQUIRED` on protected branches (`release/*`, `main`), refusing automated bot tokens with HTTP 403 and admitting only cryptographically signed approvals from certified Clinical Safety Officers.
- **Category**: Functional. Priority: Critical.
- **Status**: Reviewed. Version: 1.0.0.

### FR-HCI-SVC-003 -- Real-Time Zero-PHI Streaming Log Masking
- **Statement**: The real service shall intercept live gRPC/SSE log streams, identify National Health IDs (Modulus 11 verified) and BulletTrain Client Registry IDs (`CR-*`, `MPI-*`), and sanitize them in memory prior to PostgreSQL Tink AEAD encryption.
- **Category**: Functional. Priority: Critical.
- **Status**: Reviewed. Version: 1.0.0.

### FR-HCI-SVC-004 -- Google Tink AEAD Encrypted Storage Verification
- **Statement**: The real service shall persist secrets and logs using Google Tink AEAD envelope encryption, ensuring raw database inspection yields only ciphertext envelopes while API callers receive seamless decrypted plaintext.
- **Category**: Functional. Priority: High.
- **Status**: Reviewed. Version: 1.0.0.

### FR-HCI-SVC-005 -- Air-Gapped Qwen 2.5 Coder 32B Diagnostic Exchange
- **Statement**: The real service shall dispatch failed pipeline diagnostics to the local air-gapped LLM endpoint (`qwen2.5-coder:32b`), parsing structured RCA and unified diff patches with graceful offline heuristic fallback.
- **Category**: Functional. Priority: High.
- **Status**: Reviewed. Version: 1.0.0.

### FR-HCI-SVC-006 -- DCB0129 Code Provenance Barrier
- **Statement**: The real service shall track machine DID `did:symphonix:agent:healer-v1` and author `symphonix-healer[bot]`, failing closed on PR merge attempts until dual human reviews (including $\ge 1$ certified CSO) are validated.
- **Category**: Functional. Priority: High.
- **Status**: Reviewed. Version: 1.0.0.

### NFR-HCI-SVC-001 -- Clinical Immune Boundary (Quarantine Rejection)
- **Statement**: The real service shall strictly reject attempts to quarantine clinical safety test paths (`clinical/*`, `evals/*`, `golden_dataset/*`), returning HTTP 422 `ErrClinicalTestCannotBeQuarantined`.
- **Category**: Non-functional (safety). Priority: Critical.
- **Status**: Reviewed. Version: 1.0.0.

### NFR-HCI-SVC-002 -- Repeatable Execution & Canonical Audit Trail
- **Statement**: The real service test harness shall compute canonical SHA-256 digests for all matrix rows and emit tamper-evident `SYMPHONIX_SCENARIO_EXECUTION_RESULT_V1` records.
- **Category**: Non-functional (auditability). Priority: Critical.
- **Status**: Reviewed. Version: 1.0.0.

---

## Requirements Traceability Matrix (RTM)

```csv
RequirementID,UseCaseID,ScenarioID,TestType,VerificationLevel,AutomationStatus,Status
FR-HCI-SVC-001,UC-HCI-SVC-01,healthci_svc_oom_autoscale_0001,positive,Live-Service,Automated,Verified
FR-HCI-SVC-002,UC-HCI-SVC-02,healthci_svc_cso_gate_refusal_0002,positive,Live-Service,Automated,Verified
FR-HCI-SVC-003,UC-HCI-SVC-03,healthci_svc_zero_phi_stream_masking_0003,positive,Live-Service,Automated,Verified
FR-HCI-SVC-004,UC-HCI-SVC-04,healthci_svc_tink_aead_rest_envelope_0004,positive,Live-Service,Automated,Verified
FR-HCI-SVC-005,UC-HCI-SVC-05,healthci_svc_qwen_airgap_diagnostic_0005,positive,Live-Service,Automated,Verified
FR-HCI-SVC-006,UC-HCI-SVC-06,healthci_svc_dcb0129_code_provenance_gate_0006,positive,Live-Service,Automated,Verified
NFR-HCI-SVC-001,UC-HCI-SVC-07,healthci_svc_quarantine_immune_boundary_0007,negative,Live-Service,Automated,Verified
NFR-HCI-SVC-001,UC-HCI-SVC-08,healthci_svc_multiagent_assertion_rejection_0008,negative,Live-Service,Automated,Verified
FR-HCI-SVC-001,UC-HCI-SVC-09,healthci_svc_predictive_p95_sizing_0009,edge,Live-Service,Automated,Verified
FR-HCI-SVC-005,UC-HCI-SVC-10,healthci_svc_bounded_counterfactual_fhir_0010,edge,Live-Service,Automated,Verified
```
