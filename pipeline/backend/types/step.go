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
	UUID           string            `json:"uuid"`
	Type           StepType          `json:"type,omitempty"`
	Image          string            `json:"image,omitempty"`
	Pull           bool              `json:"pull,omitempty"`
	Detached       bool              `json:"detach,omitempty"`
	Privileged     bool              `json:"privileged,omitempty"`
	WorkingDir     string            `json:"working_dir,omitempty"`
	Environment    map[string]string `json:"environment,omitempty"`
	Entrypoint     []string          `json:"entrypoint,omitempty"`
	Commands       []string          `json:"commands,omitempty"`
	ExtraHosts     []HostAlias       `json:"extra_hosts,omitempty"`
	Volumes        []string          `json:"volumes,omitempty"`
	Tmpfs          []string          `json:"tmpfs,omitempty"`
	Devices        []string          `json:"devices,omitempty"`
	Networks       []Conn            `json:"networks,omitempty"`
	DNS            []string          `json:"dns,omitempty"`
	DNSSearch      []string          `json:"dns_search,omitempty"`
	OnFailure      bool              `json:"on_failure,omitempty"`
	OnSuccess      bool              `json:"on_success,omitempty"`
	Failure        string            `json:"failure,omitempty"`
	AuthConfig     Auth              `json:"auth_config,omitempty"`
	NetworkMode    string            `json:"network_mode,omitempty"`
	Ports          []Port            `json:"ports,omitempty"`
	BackendOptions map[string]any    `json:"backend_options,omitempty"`
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
