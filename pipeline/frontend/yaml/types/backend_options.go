// Copyright 2023 Woodpecker Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package types

// BackendOptions are advanced options for specific backends
type BackendOptions struct {
	Kubernetes KubernetesBackendOptions `yaml:"kubernetes,omitempty"`
}

type KubernetesBackendOptions struct {
	Resources          Resources         `yaml:"resources,omitempty"`
	ServiceAccountName string            `yaml:"serviceAccountName,omitempty"`
	NodeSelector       map[string]string `yaml:"nodeSelector,omitempty"`
	Tolerations        []Toleration      `yaml:"tolerations,omitempty"`
	SecurityContext    *SecurityContext  `yaml:"securityContext,omitempty"`
}

type Resources struct {
	Requests map[string]string `yaml:"requests,omitempty"`
	Limits   map[string]string `yaml:"limits,omitempty"`
}

type Toleration struct {
	Key               string             `yaml:"key,omitempty"`
	Operator          TolerationOperator `yaml:"operator,omitempty"`
	Value             string             `yaml:"value,omitempty"`
	Effect            TaintEffect        `yaml:"effect,omitempty"`
	TolerationSeconds *int64             `yaml:"tolerationSeconds,omitempty"`
}

type TaintEffect string

const (
	TaintEffectNoSchedule       TaintEffect = "NoSchedule"
	TaintEffectPreferNoSchedule TaintEffect = "PreferNoSchedule"
	TaintEffectNoExecute        TaintEffect = "NoExecute"
)

type TolerationOperator string

const (
	TolerationOpExists TolerationOperator = "Exists"
	TolerationOpEqual  TolerationOperator = "Equal"
)

type SecurityContext struct {
	Privileged      *bool       `yaml:"privileged,omitempty"`
	RunAsNonRoot    *bool       `yaml:"runAsNonRoot,omitempty"`
	RunAsUser       *int64      `yaml:"runAsUser,omitempty"`
	RunAsGroup      *int64      `yaml:"runAsGroup,omitempty"`
	FSGroup         *int64      `yaml:"fsGroup,omitempty"`
	SeccompProfile  *SecProfile `yaml:"seccompProfile,omitempty"`
	ApparmorProfile *SecProfile `yaml:"apparmorProfile,omitempty"`
}

type SecProfile struct {
	Type             string `yaml:"type,omitempty"`
	LocalhostProfile string `yaml:"localhostProfile,omitempty"`
}
