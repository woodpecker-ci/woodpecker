package kubernetes

import (
	"testing"

	"github.com/stretchr/testify/assert"

	backend "go.woodpecker-ci.org/woodpecker/v2/pipeline/backend/types"
)

func Test_parseBackendOptions(t *testing.T) {
	got, err := parseBackendOptions(&backend.Step{BackendOptions: nil})
	assert.NoError(t, err)
	assert.Equal(t, BackendOptions{}, got)
	got, err = parseBackendOptions(&backend.Step{BackendOptions: map[string]any{}})
	assert.NoError(t, err)
	assert.Equal(t, BackendOptions{}, got)
	got, err = parseBackendOptions(&backend.Step{
		BackendOptions: map[string]any{
			"kubernetes": map[string]any{
				"nodeSelector":       map[string]string{"storage": "ssd"},
				"serviceAccountName": "wp-svc-acc",
				"labels":             map[string]string{"app": "test"},
				"annotations":        map[string]string{"apps.kubernetes.io/pod-index": "0"},
				"tolerations": []map[string]any{
					{"key": "net-port", "value": "100Mbit", "effect": TaintEffectNoSchedule},
				},
				"resources": map[string]any{
					"requests": map[string]string{"memory": "128Mi", "cpu": "1000m"},
					"limits":   map[string]string{"memory": "256Mi", "cpu": "2"},
				},
				"securityContext": map[string]any{
					"privileged":   newBool(true),
					"runAsNonRoot": newBool(true),
					"runAsUser":    newInt64(101),
					"runAsGroup":   newInt64(101),
					"fsGroup":      newInt64(101),
					"seccompProfile": map[string]any{
						"type":             "Localhost",
						"localhostProfile": "profiles/audit.json",
					},
					"apparmorProfile": map[string]any{
						"type":             "Localhost",
						"localhostProfile": "k8s-apparmor-example-deny-write",
					},
				},
				"secrets": []map[string]any{
					{
						"name": "aws",
						"key":  "access-key",
						"target": map[string]any{
							"env": "AWS_SECRET_ACCESS_KEY",
						},
					},
					{
						"name": "reg-cred",
						"key":  ".dockerconfigjson",
						"target": map[string]any{
							"file": "~/.docker/config.json",
						},
					},
				},
			},
		},
	})
	assert.NoError(t, err)
	assert.Equal(t, BackendOptions{
		NodeSelector:       map[string]string{"storage": "ssd"},
		ServiceAccountName: "wp-svc-acc",
		Labels:             map[string]string{"app": "test"},
		Annotations:        map[string]string{"apps.kubernetes.io/pod-index": "0"},
		Tolerations:        []Toleration{{Key: "net-port", Value: "100Mbit", Effect: TaintEffectNoSchedule}},
		Resources: Resources{
			Requests: map[string]string{"memory": "128Mi", "cpu": "1000m"},
			Limits:   map[string]string{"memory": "256Mi", "cpu": "2"},
		},
		SecurityContext: &SecurityContext{
			Privileged:   newBool(true),
			RunAsNonRoot: newBool(true),
			RunAsUser:    newInt64(101),
			RunAsGroup:   newInt64(101),
			FSGroup:      newInt64(101),
			SeccompProfile: &SecProfile{
				Type:             "Localhost",
				LocalhostProfile: "profiles/audit.json",
			},
			ApparmorProfile: &SecProfile{
				Type:             "Localhost",
				LocalhostProfile: "k8s-apparmor-example-deny-write",
			},
		},
		Secrets: []SecretRef{
			{
				Name:   "aws",
				Key:    "access-key",
				Target: SecretTarget{Env: "AWS_SECRET_ACCESS_KEY"},
			},
			{
				Name:   "reg-cred",
				Key:    ".dockerconfigjson",
				Target: SecretTarget{File: "~/.docker/config.json"},
			},
		},
	}, got)
}
