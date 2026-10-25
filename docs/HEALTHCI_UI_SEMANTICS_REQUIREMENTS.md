# HealthCI UI Semantics & Headed Browser Governance -- Requirements

Standard: ISO/IEC/IEEE 29148:2018. Component code: `HEALTHCI-UI`. Trace ID: `DOC-REQ-HEALTHCI-UI-001`.
Reference Standard: BulletTrain `UI_SEMANTICS_REQUIREMENTS.md` (Workstream 8).

## Purpose

Treat accessible, semantically labelled UI as control-plane infrastructure so agent-driven browser automation and human oversight are deterministic and verifiable: stable agent semantic hooks, assertion regions exposing machine-readable state, single-page zero-scroll constraints, and multi-persona headed browser execution via SignalBox MCP under real operator conditions.

## Architectural Conventions (BulletTrain Compatible)

| Need | BulletTrain Convention | HealthCI Implementation Decision |
|---|---|---|
| Element-locating selector chain | `data-agent-id` > `data-testid` > `aria-label` > `role` | EXTEND: `data-agent-id` top priority; `data-testid` fallback for test tooling. |
| Machine-readable hook convention | `data-agent-id` / `data-agent-region` / `data-agent-status` | BUILD: Every interactive element and region exposes `data-agent-id`, `data-agent-region`, and `data-agent-status`. |
| Zero-scroll single-page layout | `has_unnecessary_scroll: false`, `overflow: hidden`, max bottom $\le$ viewport height | ENFORCE: Single-page desktop layout with zero vertical or horizontal overflow. |
| Headed browser execution | Playwright `headless: false`, named agent personas | MANDATORY: SignalBox MCP drives headed Chromium window (`headless: false`) with persona attribution (`fde-superpersona`, `clinician`, `cso`). |
| Dual-Audience Explainability | Clinician technical view (Merkle root, guidelines) vs. Patient plain-language view | VALIDATE: Instant tab switching between Clinician (technical) and Patient (non-technical, no hashes) CDR 1.0.0 views. |

---

## Requirement Records (ISO/IEC/IEEE 29148:2018)

### FR-HCI-UI-001 -- Stable Agent Semantic Hooks
- **Statement**: The system shall publish stable semantic hooks (`data-agent-id`, `data-agent-region`, `data-agent-status`) across all interactive and status components in the Clinical Control Plane.
- **Category**: Functional. Priority: High.
- **Acceptance Criteria**:
  - `FR-HCI-UI-001-AC01`: Every interactive element (buttons, selectors, toggles) exposes a stable, deterministic `data-agent-id`.
  - `FR-HCI-UI-001-AC02`: Every functional region exposes a `data-agent-region` identifier.
  - `FR-HCI-UI-001-AC03`: Machine-readable state is published dynamically through `data-agent-status`.
- **Status**: Reviewed. Version: 1.0.0.

### FR-HCI-UI-002 -- Assertion Regions for Agent-Visible State
- **Statement**: The system shall provide designated assertion regions exposing machine-readable state to autonomous agents and human auditors.
- **Category**: Functional. Priority: High.
- **Acceptance Criteria**:
  - `FR-HCI-UI-002-AC01`: Operating mode transitions (`live`, `simulation`, `replay`, `audit`) update `data-agent-status` on the banner region.
  - `FR-HCI-UI-002-AC02`: SignalBox panel and tab state reflect active/inactive visibility status without polling.
- **Status**: Reviewed. Version: 1.0.0.

### FR-HCI-UI-003 -- Dual-Audience CDR 1.0.0 Presentation
- **Statement**: The system shall provide distinct, dual-audience clinical decision rationale views conforming to CDR 1.0.0 specifications.
- **Category**: Functional. Priority: High.
- **Acceptance Criteria**:
  - `FR-HCI-UI-003-AC01`: Clinician view displays cryptographic WORM SHA-256 Merkle root and clinical guideline citations.
  - `FR-HCI-UI-003-AC02`: Patient view presents plain-language reassurance, ruled-out dangers, and strictly suppresses technical hashes and cryptographic identifiers.
- **Status**: Reviewed. Version: 1.0.0.

### FR-HCI-UI-004 -- Self-Healing Telemetry & Air-Gap Model Assurance
- **Statement**: The system shall surface real-time telemetry for all 6 self-healing primitives and display local air-gapped LLM inference status (Qwen 2.5 Coder 32B).
- **Category**: Functional. Priority: Medium.
- **Acceptance Criteria**:
  - `FR-HCI-UI-004-AC01`: Displays counts of L1 OOM scaling, L2 AI retries, and L3 quarantines.
  - `FR-HCI-UI-004-AC02`: Explicitly highlights local air-gapped inference engine badge and DCB0129 dual-key provenance status.
- **Status**: Reviewed. Version: 1.0.0.

### NFR-HCI-UI-001 -- Zero-Scroll Single-Page Fit
- **Statement**: The application UI shall fit entirely within the desktop viewport without vertical down-scrolling or unnecessary horizontal scrollbars.
- **Category**: Non-functional (usability/layout). Priority: Critical.
- **Acceptance Criteria**:
  - `NFR-HCI-UI-001-AC01`: SignalBox layout audit verifies `has_unnecessary_scroll: false`.
  - `NFR-HCI-UI-001-AC02`: `scrollHeight == clientHeight` and max bottom coordinate $\le$ viewport height.
- **Status**: Verified. Version: 1.0.0.

### NFR-HCI-UI-002 -- Headed Browser Verification with Named Superpersonas
- **Statement**: Frontend verification shall execute in a visible, headed Chromium browser (`headless: false`) driven by SignalBox MCP with explicit persona attribution.
- **Category**: Non-functional (verification). Priority: Critical.
- **Acceptance Criteria**:
  - `NFR-HCI-UI-002-AC01`: Runs against local preview served on literal IPv4 `127.0.0.1`.
  - `NFR-HCI-UI-002-AC02`: Binds personas: `fde-superpersona`, `clinician`, and `cso`.
  - `NFR-HCI-UI-002-AC03`: Emits persistent execution records adhering to `SYMPHONIX_SCENARIO_EXECUTION_RESULT_V1`.
- **Status**: Verified. Version: 1.0.0.

### NFR-HCI-UI-003 -- WCAG 2.2 AA Contrast & Accessibility
- **Statement**: All UI controls, text elements, and assertion badges shall satisfy WCAG 2.2 AA contrast ratios ($\ge 4.5:1$ normal text, $\ge 3:1$ large/bold).
- **Category**: Non-functional (accessibility). Priority: High.
- **Acceptance Criteria**:
  - `NFR-HCI-UI-003-AC01`: Automated layout audit reports zero critical contrast violations.
- **Status**: Reviewed. Version: 1.0.0.

---

## Requirements Traceability Matrix (RTM)

```csv
RequirementID,UseCaseID,ScenarioID,TestType,VerificationLevel,AutomationStatus,Status
FR-HCI-UI-001,UC-HCI-UI-01,healthci_ui_fde_view_levels_0001,positive,E2E-Headed,Automated,Verified
FR-HCI-UI-002,UC-HCI-UI-02,healthci_ui_mode_switch_reactive_styling_0006,positive,E2E-Headed,Automated,Verified
FR-HCI-UI-003,UC-HCI-UI-03,healthci_ui_clinician_gate_review_0002,positive,E2E-Headed,Automated,Verified
FR-HCI-UI-003,UC-HCI-UI-04,healthci_ui_cso_patient_cdr_0003,positive,E2E-Headed,Automated,Verified
FR-HCI-UI-004,UC-HCI-UI-05,healthci_ui_fde_self_healing_telemetry_0004,positive,E2E-Headed,Automated,Verified
FR-HCI-UI-001,UC-HCI-UI-06,healthci_ui_cso_code_provenance_gate_0005,positive,E2E-Headed,Automated,Verified
FR-HCI-UI-001,UC-HCI-UI-07,healthci_ui_missing_agent_hook_0007,negative,E2E-Headed,Automated,Verified
NFR-HCI-UI-001,UC-HCI-UI-08,healthci_ui_zero_scroll_overflow_rejection_0008,negative,E2E-Headed,Automated,Verified
FR-HCI-UI-002,UC-HCI-UI-09,healthci_ui_unauthorized_override_rejection_0009,negative,E2E-Headed,Automated,Verified
FR-HCI-UI-002,UC-HCI-UI-10,healthci_ui_rapid_tab_switching_resilience_0010,edge,E2E-Headed,Automated,Verified
NFR-HCI-UI-003,UC-HCI-UI-11,healthci_ui_high_contrast_theme_edge_0011,edge,E2E-Headed,Automated,Verified
NFR-HCI-UI-002,UC-HCI-UI-12,healthci_ui_headed_browser_visual_attestation_0012,positive,E2E-Headed,Automated,Verified
```
