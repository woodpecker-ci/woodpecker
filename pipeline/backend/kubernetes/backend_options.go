// Copyright 2024 Woodpecker Authors
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

package kubernetes

import (
	"github.com/go-viper/mapstructure/v2"
	kube_core_v1 "k8s.io/api/core/v1"

	backend_types "go.woodpecker-ci.org/woodpecker/v3/pipeline/backend/types"
)

// BackendOptions defines all the advanced options for the kubernetes backend.
type BackendOptions struct {
	Resources          Resources              `mapstructure:"resources"`
	RuntimeClassName   *string                `mapstructure:"runtimeClassName"`
	ServiceAccountName string                 `mapstructure:"serviceAccountName"`
	Labels             map[string]string      `mapstructure:"labels"`
	Annotations        map[string]string      `mapstructure:"annotations"`
	NodeSelector       map[string]string      `mapstructure:"nodeSelector"`
	Tolerations        []Toleration           `mapstructure:"tolerations"`
	Affinity           *kube_core_v1.Affinity `mapstructure:"affinity"`
	SecurityContext    *SecurityContext       `mapstructure:"securityContext"`
	Secrets            []SecretRef            `mapstructure:"secrets"`
}

// Resources defines two maps for kubernetes resource definitions.
type Resources struct {
	Requests map[string]string `mapstructure:"requests"`
	Limits   map[string]string `mapstructure:"limits"`
}

// Toleration defines Kubernetes toleration.
type Toleration struct {
	Key               string             `mapstructure:"key"`
	Operator          TolerationOperator `mapstructure:"operator"`
	Value             string             `mapstructure:"value"`
	Effect            TaintEffect        `mapstructure:"effect"`
	TolerationSeconds *int64             `mapstructure:"tolerationSeconds"`
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
	Privileged               *bool                                `mapstructure:"privileged"`
	RunAsNonRoot             *bool                                `mapstructure:"runAsNonRoot"`
	RunAsUser                *int64                               `mapstructure:"runAsUser"`
	RunAsGroup               *int64                               `mapstructure:"runAsGroup"`
	FSGroup                  *int64                               `mapstructure:"fsGroup"`
	FsGroupChangePolicy      *kube_core_v1.PodFSGroupChangePolicy `mapstructure:"fsGroupChangePolicy"`
	SeccompProfile           *SecProfile                          `mapstructure:"seccompProfile"`
	ApparmorProfile          *SecProfile                          `mapstructure:"apparmorProfile"`
	AllowPrivilegeEscalation *bool                                `mapstructure:"allowPrivilegeEscalation"`
	Capabilities             *Capabilities                        `mapstructure:"capabilities"`
}

type SecProfile struct {
	Type             SecProfileType `mapstructure:"type"`
	LocalhostProfile string         `mapstructure:"localhostProfile"`
}

type SecProfileType string

type Capabilities struct {
	Drop []string `mapstructure:"drop"`
}

// SecretRef defines Kubernetes secret reference.
type SecretRef struct {
	Name   string       `mapstructure:"name"`
	Key    string       `mapstructure:"key"`
	Target SecretTarget `mapstructure:"target"`
}

// SecretTarget defines secret mount target.
type SecretTarget struct {
	Env  string `mapstructure:"env"`
	File string `mapstructure:"file"`
}

const (
	SecProfileTypeRuntimeDefault SecProfileType = "RuntimeDefault"
	SecProfileTypeLocalhost      SecProfileType = "Localhost"
)

func parseBackendOptions(step *backend_types.Step) (BackendOptions, error) {
	var result BackendOptions
	if step == nil || step.BackendOptions == nil {
		return result, nil
	}
	err := mapstructure.WeakDecode(step.BackendOptions[EngineName], &result)
	return result, err
}
