# Symphonix HealthCI: Autonomous Self-Healing Requirements Specification

**Document Identifier**: `DOC-REQ-HEALTHCI-HEAL-001`
**Standard Compliance**: DCB0129 / DCB0160 (Clinical Risk Management), ISO/IEC 42001 (AI Management System), IEC 62304 (Medical Device Software Lifecycle)
**Author**: Symphonix Health Forward Deployed Engineering (FDE) & Clinical AI Safety
**Status**: APPROVED
**Target Matrix**: `contrib/symphonix/tests/reduced_json_matrices/healthci_self_healing_matrix.14col.json`

---

## 1. Functional Requirements (FR)

### FR-HEAL-001: DCB0129 Autonomous Code Provenance Gate
* **Description**: Any code patch generated autonomously by an AI agent must carry immutable cryptographic provenance, machine DID lineage, unified diff checksums, and explicit DCB0129 Hazard Impact statements. Direct merges to protected branches (`main`, `master`, `production`, `release/*`) without human review are strictly forbidden.
* **Acceptance Criteria**:
  * **FR-HEAL-001-AC01**: Patch records must be attributed to `symphonix-healer[bot]` with machine identity `did:symphonix:agent:healer-v1` and compute a SHA-256 hash of the unified diff.
  * **FR-HEAL-001-AC02**: Merging to protected branches without human reviews must fail closed with `ErrDCB0129ProvenanceViolation`.
  * **FR-HEAL-001-AC03**: Protected branch merges require a minimum of 2 human reviews, at least one of which must be a certified Clinical Safety Officer (CSO). Non-CSO combinations must be rejected.
  * **FR-HEAL-001-AC04**: Merges to non-protected draft branches (`feature/*`, `heal/*`) are permitted without full CSO sign-off for iterative developer evaluation.

### FR-HEAL-002: Tiered Flaky Test Isolation with Clinical Immune Boundary
* **Description**: The CI engine must distinguish between non-clinical integration/UI flakes and clinical safety evaluation benchmarks. Clinical evaluations can never be quarantined under any circumstances.
* **Acceptance Criteria**:
  * **FR-HEAL-002-AC01**: Any attempt to quarantine tests in suites matching `clinical/*`, `evals/*`, `golden_dataset/*`, `sepsis`, `triage`, or `fhir` must immediately fail closed with `ErrClinicalTestCannotBeQuarantined`.
  * **FR-HEAL-002-AC02**: Non-clinical tests (e.g. UI layout timing, end-to-end browser animation flakes) may be quarantined with an auto-generated Jira/issue ticket ID.
  * **FR-HEAL-002-AC03**: Non-clinical quarantines enforce a strict 72-hour Time-to-Live (TTL). Once expired, `IsQuarantined` returns `false` with `ErrQuarantineExpired`, forcing test re-execution.

### FR-HEAL-003: Predictive Pre-emptive Resource Rightsizing
* **Description**: To prevent runtime Exit 137 (OOM) failures before container initialization, the system must utilize historical execution telemetry stored in PostgreSQL 16 to pre-allocate memory headroom.
* **Acceptance Criteria**:
  * **FR-HEAL-003-AC01**: Calculate the 95th percentile (P95) peak memory usage across historical runs of a specific pipeline step.
  * **FR-HEAL-003-AC02**: Apply a 20% safety headroom margin (`P95 * 1.20`). If the rightsized memory exceeds the step's default limit, pre-allocate the higher limit on startup.
  * **FR-HEAL-003-AC03**: In the absence of historical telemetry, safely fall back to the default configured limit without throwing errors.

### FR-HEAL-004: Self-Healing Pipeline Configurations (.woodpecker.yaml)
* **Description**: AST/YAML inspection must identify clinical CI compliance violations and generate interactive Unified Diffs for 1-click remediation.
* **Acceptance Criteria**:
  * **FR-HEAL-004-AC01**: Detect `L-CAID-10` Zero-Mock violations (such as `--mock`, `jest.mock`, `?demo=true`) and propose replacement with `--real-services`.
  * **FR-HEAL-004-AC02**: Detect plaintext secret usage without mandatory Google Tink AEAD keyset configuration (`WOODPECKER_ENCRYPTION_TINK_KEYSET_FILE`) and inject the required encryption keyset environment variable.
  * **FR-HEAL-004-AC03**: Detect unconstrained privileged containers (`privileged: true`) and propose unprivileged execution (`privileged: false`) with `RequiresCSOReview: true`.

### FR-HEAL-005: Multi-Agent Dual-Turn Cross-Check
* **Description**: Autonomous fixes proposed by a developer agent must undergo independent clinical scrutiny by a dedicated Clinical Safety Auditor agent persona prior to human presentation.
* **Acceptance Criteria**:
  * **FR-HEAL-005-AC01**: Turn 1 candidate patches that attempt to delete or relax test assertions (`assert.Equal`, `expect`, `require`) must be rejected with `HazardCritical`.
  * **FR-HEAL-005-AC02**: Turn 1 candidate patches that tamper with clinical decision thresholds (e.g. vital signs, triage cutoff scores, sensitivity bounds) must be rejected with `HazardCritical`.
  * **FR-HEAL-005-AC03**: Benign technical patches (e.g. package import path updates, compiler syntax corrections) that introduce no clinical risks must be approved with `HazardNegligible`.

### FR-HEAL-006: Bounded Counterfactual FHIR Synthetic Evaluations
* **Description**: When a clinical evaluation benchmark misses a sensitivity threshold, the system must generate minimal counterfactual test vectors across decision boundaries to diagnose sensitivity cliffs without distribution drift.
* **Acceptance Criteria**:
  * **FR-HEAL-006-AC01**: Perturbations must be bounded strictly to micro-intervals ($\pm 1.0\%$, $\pm 2.0\%$) around the decision boundary.
  * **FR-HEAL-006-AC02**: Output must adhere to standard FHIR R4 `Observation` JSON schemas, including valid LOINC coding and interpretation flags (`CRITICAL` vs `NORMAL`).

---

## 2. Non-Functional Requirements (NFR)

### NFR-HEAL-001: Air-Gapped Local LLM Inference (Qwen 2.5 Coder 32B)
* **Description**: The self-healing engine must operate 100% offline within hospital perimeters, interfacing with local Ollama or vLLM instances without cloud data egress.
* **Acceptance Criteria**:
  * **NFR-HEAL-001-AC01**: Communicate with OpenAI-compatible `/v1/chat/completions` API targeting `qwen2.5-coder:32b`.
  * **NFR-HEAL-001-AC02**: In the event of network disconnection, timeout, or unreachable endpoint, gracefully fall back to local deterministic regex heuristics within 10 seconds without build crashes or hangs.

### NFR-HEAL-002: Zero Down-Scrolling Single-Page Ergonomics
* **Description**: The clinical dashboard must present all self-healing telemetry, DCB0129 provenance status, and pipeline matrices in a single viewport (`overflow: hidden`) without down-scrolling.
* **Acceptance Criteria**:
  * **NFR-HEAL-002-AC01**: SignalBox layout audit must confirm `has_unnecessary_scroll: false` with page content bottom $\le 991\text{px}$.
  * **NFR-HEAL-002-AC02**: Tabbed segmentation must be used to partition clinical gate, self-healing telemetry, and CDR 1.0.0 views.
