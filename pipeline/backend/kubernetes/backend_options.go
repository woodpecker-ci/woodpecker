package kubernetes

import (
	"github.com/mitchellh/mapstructure"

	backend "go.woodpecker-ci.org/woodpecker/v2/pipeline/backend/types"
)

// BackendOptions defines all the advanced options for the kubernetes backend.
type BackendOptions struct {
	Resources          Resources         `mapstructure:"resources"`
	RuntimeClassName   *string           `mapstructure:"runtimeClassName"`
	ServiceAccountName string            `mapstructure:"serviceAccountName"`
	Labels             map[string]string `mapstructure:"labels"`
	Annotations        map[string]string `mapstructure:"annotations"`
	NodeSelector       map[string]string `mapstructure:"nodeSelector"`
	Tolerations        []Toleration      `mapstructure:"tolerations"`
	SecurityContext    *SecurityContext  `mapstructure:"securityContext"`
	Secrets            []SecretRef       `mapstructure:"secrets"`
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
	Privileged      *bool       `mapstructure:"privileged"`
	RunAsNonRoot    *bool       `mapstructure:"runAsNonRoot"`
	RunAsUser       *int64      `mapstructure:"runAsUser"`
	RunAsGroup      *int64      `mapstructure:"runAsGroup"`
	FSGroup         *int64      `mapstructure:"fsGroup"`
	SeccompProfile  *SecProfile `mapstructure:"seccompProfile"`
	ApparmorProfile *SecProfile `mapstructure:"apparmorProfile"`
}

type SecProfile struct {
	Type             SecProfileType `mapstructure:"type"`
	LocalhostProfile string         `mapstructure:"localhostProfile"`
}

type SecProfileType string

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

func parseBackendOptions(step *backend.Step) (BackendOptions, error) {
	var result BackendOptions
	if step.BackendOptions == nil {
		return result, nil
	}
	err := mapstructure.Decode(step.BackendOptions[EngineName], &result)
	return result, err
}
