// Copyright 2023 Woodpecker Authors
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

package types

// Step defines a container process.
type Step struct {
	Name           string            `json:"name"`
	OrgID          int64             `json:"org_id,omitempty"`
	UUID           string            `json:"uuid"`
	Type           StepType          `json:"type,omitempty"`
	Image          string            `json:"image,omitempty"`
	Pull           bool              `json:"pull,omitempty"`
	Detached       bool              `json:"detach,omitempty"`
	Privileged     bool              `json:"privileged,omitempty"`
	WorkingDir     string            `json:"working_dir,omitempty"`
	WorkspaceBase  string            `json:"workspace_base,omitempty"`
	Environment    map[string]string `json:"environment,omitempty"`
	SecretMapping  map[string]string `json:"secret_mapping,omitempty"`
	Entrypoint     []string          `json:"entrypoint,omitempty"`
	Commands       []string          `json:"commands,omitempty"`
	ExtraHosts     []HostAlias       `json:"extra_hosts,omitempty"`
	Volumes        []string          `json:"volumes,omitempty"`
	Tmpfs          []string          `json:"tmpfs,omitempty"`
	Devices        []string          `json:"devices,omitempty"`
	Networks       []Conn            `json:"networks,omitempty"`
	DNS            []string          `json:"dns,omitempty"`
	DNSSearch      []string          `json:"dns_search,omitempty"`
	DependsOn      []string          `json:"depends_on,omitempty"`
	OnFailure      bool              `json:"on_failure,omitempty"`
	OnSuccess      bool              `json:"on_success,omitempty"`
	Failure        string            `json:"failure,omitempty"`
	AuthConfig     Auth              `json:"auth_config"`
	NetworkMode    string            `json:"network_mode,omitempty"`
	Ports          []Port            `json:"ports,omitempty"`
	BackendOptions map[string]any    `json:"backend_options,omitempty"`
	WorkflowLabels map[string]string `json:"workflow_labels,omitempty"`
	// ExitCodeSignals maps exit codes to the signal they send to the runtime.
	ExitCodeSignals map[int]StepSignal `json:"exit_code_signals,omitempty"`
	// MaxRestarts is the number of restarts after the first run for the restart
	// signal. 0 means StepSignalRestartDefaultMax.
	MaxRestarts int `json:"max_restarts,omitempty"`
}

// RestartMaxAttempts returns how often the step runs at most (first run included)
// when it keeps sending the restart signal.
func (s *Step) RestartMaxAttempts() int {
	if s.MaxRestarts > 0 {
		return s.MaxRestarts + 1
	}
	return StepSignalRestartDefaultMax + 1
}

// StepType identifies the type of step.
type StepType string

const (
	StepTypeClone    StepType = "clone"
	StepTypeService  StepType = "service"
	StepTypePlugin   StepType = "plugin"
	StepTypeCommands StepType = "commands"
	StepTypeCache    StepType = "cache"
)

// StepSignal is an action a step can ask the runtime to perform,
// e.g. by exiting with an exit code mapped to it via `signals:`.
type StepSignal string

const (
	// StepSignalSkip marks the step as successful and skips all
	// steps of the workflow that have not started yet.
	StepSignalSkip StepSignal = "skip"
	// StepSignalSuccess marks the step as successful and continues
	// the workflow as usual (e.g. a linter that exits 1 on warnings).
	StepSignalSuccess StepSignal = "success"
	// StepSignalRestart runs the step again, up to Step.RestartMaxAttempts()
	// attempts in total. If the last attempt still exits with a restart code,
	// the step fails with that exit code.
	StepSignalRestart StepSignal = "restart"
	// StepSignalCancel marks the step as successful and cancels its workflow:
	// running steps are stopped, the remaining steps do not run and the
	// workflow is reported as killed. Other workflows are not affected.
	StepSignalCancel StepSignal = "cancel"
)

const (
	// StepSignalRestartDefaultMax is the number of restarts after the first run
	// when `max_restarts` is not set (3 attempts in total).
	StepSignalRestartDefaultMax = 2
	// StepSignalRestartLimit is the highest allowed `max_restarts`.
	StepSignalRestartLimit = 10
)

// ValidStepSignals lists all signals the runtime knows how to handle.
var ValidStepSignals = map[StepSignal]struct{}{
	StepSignalSkip:    {},
	StepSignalSuccess: {},
	StepSignalRestart: {},
	StepSignalCancel:  {},
}

// IsValid reports whether the signal is known to the runtime.
func (s StepSignal) IsValid() bool {
	_, ok := ValidStepSignals[s]
	return ok
}

const (
	// MinSignalExitCode is the lowest exit code that can be mapped to a signal (0 means success).
	MinSignalExitCode = 1
	// MaxSignalExitCode is the highest exit code a process can return.
	MaxSignalExitCode = 255
)
